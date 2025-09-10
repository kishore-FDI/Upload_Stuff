package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins, tighten for production
		return true
	},
}

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	connections map[string][]*websocket.Conn
	mutex       sync.RWMutex
}

var connManager = &ConnectionManager{
	connections: make(map[string][]*websocket.Conn),
}

// GetConnectionManager returns the global connection manager
func GetConnectionManager() *ConnectionManager {
	return connManager
}

// AddConnection adds a WebSocket connection for an upload ID
func (cm *ConnectionManager) AddConnection(uploadID string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.connections[uploadID] = append(cm.connections[uploadID], conn)
}

// RemoveConnection removes a WebSocket connection
func (cm *ConnectionManager) RemoveConnection(uploadID string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	connections := cm.connections[uploadID]
	for i, c := range connections {
		if c == conn {
			cm.connections[uploadID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	if len(cm.connections[uploadID]) == 0 {
		delete(cm.connections, uploadID)
	}
}

// BroadcastProgress sends progress updates to all connections for an upload ID
func (cm *ConnectionManager) BroadcastProgress(uploadID string, message ProgressMessage) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := cm.connections[uploadID]
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling progress message: %v", err)
		return
	}

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("Error sending progress message: %v", err)
			// Remove the connection if it's no longer valid
			go cm.RemoveConnection(uploadID, conn)
		}
	}
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	uploadID := c.Param("id")
	if uploadID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "upload ID required"}`))
		return
	}

	// Add connection to manager
	connManager.AddConnection(uploadID, conn)
	defer connManager.RemoveConnection(uploadID, conn)

	// Send initial connection confirmation
	conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "connected", "upload_id": "`+uploadID+`"}`))

	// Keep connection alive and handle incoming messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Echo back with upload ID context for debugging
		response := map[string]interface{}{
			"type":      "echo",
			"upload_id": uploadID,
			"message":   string(msg),
		}

		responseBytes, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, responseBytes)
	}
}

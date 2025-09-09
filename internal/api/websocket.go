package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins, tighten for production
		return true
	},
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	id := c.Param("id")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Echo back with id context
		response := []byte("id=" + id + " msg=" + string(msg))
		err = conn.WriteMessage(websocket.TextMessage, response)
		if err != nil {
			break
		}
	}
}


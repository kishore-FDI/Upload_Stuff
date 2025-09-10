# Upload Progress Tracking & Storage API

This document describes the real-time upload progress tracking and storage API features implemented in the Media Pipeline.

## Features Implemented

### 1. Real-time Progress Tracking via WebSocket

The system now provides real-time upload progress updates through WebSocket connections, allowing clients to track upload progress in real-time.

#### WebSocket Endpoint

```
WS: /api/v1/ws/{upload_id}
```

#### Progress Message Format

```json
{
  "type": "progress|complete|created|error",
  "upload_id": "string",
  "progress": 0.0-100.0,
  "bytes_sent": 0,
  "total_size": 0,
  "status": "uploading|completed|failed|created",
  "message": "optional message"
}
```

#### Message Types

- **`created`**: Upload session created
- **`progress`**: Progress update during upload
- **`complete`**: Upload completed successfully
- **`error`**: Upload failed with error message

### 2. Enhanced Status Endpoint

Query upload progress and status via HTTP API.

#### Endpoint

```
GET /api/v1/uploads/meta/{token}/status
```

#### Response Format

```json
{
  "token": "upload_id",
  "status": "created|uploading|completed|failed",
  "progress": 75.5,
  "offset": 1024000,
  "size": 2048000,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:01:00Z",
  "completed_at": "2024-01-01T00:02:00Z"
}
```

### 3. Storage API

Complete file management with download and delete operations.

#### Download File

```
GET /api/v1/storage/{file_id}
```

- Downloads file with proper filename
- Supports range requests for resumable downloads
- Returns appropriate Content-Disposition header

#### Delete File

```
DELETE /api/v1/storage/{file_id}
```

- Deletes file from storage
- Removes associated metadata from Redis
- Cleans up TUS info files

## Implementation Details

### WebSocket Connection Manager

The `ConnectionManager` class manages WebSocket connections per upload ID:

```go
type ConnectionManager struct {
    connections map[string][]*websocket.Conn
    mutex       sync.RWMutex
}
```

- Thread-safe connection management
- Multiple clients can track the same upload
- Automatic cleanup of disconnected clients

### Progress Storage in Redis

Upload progress is stored in Redis with the key pattern `upload:{upload_id}`:

```redis
HSET upload:abc123
  status "uploading"
  progress 75.5
  offset 1024000
  size 2048000
  created_at "2024-01-01T00:00:00Z"
  updated_at "2024-01-01T00:01:00Z"
```

### TUS Integration

The TUS handler now broadcasts progress events:

1. **Upload Created**: When TUS creates a new upload
2. **Progress Updates**: During chunk uploads
3. **Upload Complete**: When upload finishes

## Usage Examples

### Python Client with WebSocket

```python
import asyncio
import websockets
import json

async def track_upload_progress(upload_id):
    uri = f"ws://localhost:8080/api/v1/ws/{upload_id}"

    async with websockets.connect(uri) as websocket:
        async for message in websocket:
            data = json.loads(message)
            print(f"Progress: {data.get('progress', 0):.1f}%")
```

### JavaScript Client

```javascript
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/${uploadId}`);

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === "progress") {
    updateProgressBar(data.progress);
  }
};
```

### HTTP Status Check

```bash
curl -H "X-API-KEY: your-key" \
     http://localhost:8080/api/v1/uploads/meta/abc123/status
```

## Testing

### Test Files Provided

1. **`test/progress_test.py`**: Python client with WebSocket progress tracking
2. **`test/progress_client.html`**: Browser-based client with real-time UI

### Running Tests

1. Start the server:

   ```bash
   go run main.go
   ```

2. Run Python test:

   ```bash
   python test/progress_test.py
   ```

3. Open HTML client:
   ```bash
   open test/progress_client.html
   ```

## API Authentication

All endpoints require authentication via headers:

- **X-API-KEY**: Business API key
- **X-Username**: Username for the upload

## Error Handling

The system handles various error conditions:

- Invalid upload IDs
- Missing authentication
- File not found
- WebSocket connection errors
- Upload failures

## Performance Considerations

- WebSocket connections are lightweight
- Redis storage is efficient for progress tracking
- Progress updates are throttled to prevent spam
- Automatic cleanup of completed uploads after 24 hours

## Security

- CORS enabled for cross-origin requests
- Rate limiting on all endpoints
- Authentication required for all operations
- File access validation before download/delete

## Future Enhancements

- Webhook support for server-to-server notifications
- Upload resumption with progress persistence
- Batch upload progress tracking
- Upload analytics and reporting
- Integration with CDN for faster downloads

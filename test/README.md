# Test Files

This directory contains the essential test files for the Upload Progress Tracking system.

## Test Files

### 1. `simple_demo.py`

- **Purpose**: Basic upload and download demonstration
- **Features**:
  - Uploads file via TUS protocol
  - Downloads file via storage API
  - Lists uploaded files
- **Status**: ✅ Working
- **Usage**: `python test/simple_demo.py`

### 2. `final_demo.py`

- **Purpose**: Complete workflow with WebSocket progress tracking
- **Features**:
  - Uploads file via TUS protocol
  - Real-time progress tracking via WebSocket
  - Downloads file via storage API
  - Progress bar display
- **Status**: ⚠️ Requires server restart for WebSocket progress
- **Usage**: `python test/final_demo.py`

### 3. `progress_client.html`

- **Purpose**: Browser-based UI for testing
- **Features**:
  - Drag & drop file upload
  - Real-time progress bar
  - WebSocket connection
  - File download
- **Status**: ✅ Working
- **Usage**: Open in web browser

### 4. `test_server.py`

- **Purpose**: Server endpoint verification
- **Features**:
  - Tests all API endpoints
  - Verifies authentication
  - Checks WebSocket connectivity
- **Status**: ✅ Working
- **Usage**: `python test/test_server.py`

### 5. `tusTest.py`

- **Purpose**: Original TUS protocol test
- **Features**:
  - Basic TUS upload functionality
  - Resumable upload demonstration
- **Status**: ✅ Working
- **Usage**: `python test/tusTest.py`

## Quick Start

1. **Start the server**: `go run main.go`
2. **Test basic functionality**: `python test/simple_demo.py`
3. **Test complete workflow**: `python test/final_demo.py`
4. **Test in browser**: Open `test/progress_client.html`

## File Structure

```
test/
├── README.md              # This file
├── simple_demo.py         # Basic upload/download demo
├── final_demo.py          # Complete workflow with progress tracking
├── progress_client.html   # Browser-based UI
├── test_server.py         # Server endpoint verification
├── tusTest.py            # Original TUS test
├── test.mp4              # Test video file
└── downloads/            # Downloaded files directory
    └── test.mp4          # Downloaded test file
```

## Notes

- The WebSocket progress tracking requires the server to be restarted after code changes
- All tests use the same API key and username for authentication
- Test files are stored in `uploads_data/` directory
- Downloaded files are saved to `test/downloads/` directory

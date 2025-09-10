#!/usr/bin/env python3
"""
Final demo: Complete workflow with WebSocket progress tracking
"""

import asyncio
import json
import os
import requests
import threading
import time
import websockets
from tusclient import client

# Configuration
BASE_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/api/v1/ws/"
TUS_ENDPOINT = "http://localhost:8080/api/v1/uploads/"
FILE_PATH = "test.mp4"
DOWNLOAD_DIR = "test/downloads"

HEADERS = {
    "X-API-KEY": "5cf74fb0ea2ec521d85c36e0e4468749d6aaee5b8f5ff5c35c0f9a3260a96a60",
    "X-Username": "user123444",
}

class ProgressTracker:
    def __init__(self):
        self.progress = 0.0
        self.bytes_sent = 0
        self.total_size = 0
        self.status = "disconnected"
        self.connected = False
        self.messages_received = 0
        
    def update_progress(self, progress, bytes_sent, total_size):
        """Update progress and display progress bar"""
        self.progress = progress
        self.bytes_sent = bytes_sent
        self.total_size = total_size
        
        # Create progress bar
        bar_width = 50
        filled = int(progress / 100 * bar_width)
        bar = "█" * filled + "░" * (bar_width - filled)
        
        # Format bytes
        def format_bytes(bytes_val):
            if bytes_val == 0:
                return "0 B"
            for unit in ['B', 'KB', 'MB', 'GB']:
                if bytes_val < 1024.0:
                    return f"{bytes_val:.1f} {unit}"
                bytes_val /= 1024.0
            return f"{bytes_val:.1f} TB"
        
        print(f"\r📊 Progress: [{bar}] {progress:.1f}% ({format_bytes(bytes_sent)}/{format_bytes(total_size)})", end="", flush=True)

async def track_progress(upload_id, tracker):
    """Track upload progress via WebSocket"""
    uri = f"{WS_URL}{upload_id}"
    print(f"🔌 Connecting to WebSocket: {uri}")
    
    try:
        async with websockets.connect(uri) as websocket:
            tracker.connected = True
            print("✅ WebSocket connected!")
            
            # Set a timeout for receiving messages
            try:
                async with asyncio.timeout(15):  # 15 second timeout
                    async for message in websocket:
                        tracker.messages_received += 1
                        try:
                            data = json.loads(message)
                            msg_type = data.get("type", "unknown")
                            
                            if msg_type == "connected":
                                print(f"🎯 Connected to upload: {data.get('upload_id')}")
                            elif msg_type == "created":
                                print(f"📁 Upload created: {data.get('upload_id')}")
                                print(f"   Size: {data.get('total_size', 0):,} bytes")
                            elif msg_type == "progress":
                                progress = data.get("progress", 0.0)
                                bytes_sent = data.get("bytes_sent", 0)
                                total_size = data.get("total_size", 0)
                                tracker.update_progress(progress, bytes_sent, total_size)
                            elif msg_type == "complete":
                                print(f"\n✅ Upload completed: {data.get('upload_id')}")
                                print(f"   Final size: {data.get('total_size', 0):,} bytes")
                                tracker.status = "completed"
                                break
                            elif msg_type == "error":
                                print(f"\n❌ Upload error: {data.get('message', 'Unknown error')}")
                                tracker.status = "error"
                                break
                            else:
                                print(f"\n📨 Message: {data}")
                                
                        except json.JSONDecodeError:
                            print(f"📨 Raw message: {message}")
            except asyncio.TimeoutError:
                print(f"\n⏰ WebSocket timeout - no progress messages received")
                tracker.status = "timeout"
                    
    except Exception as e:
        print(f"❌ WebSocket error: {e}")
        tracker.status = "error"

def upload_file():
    """Upload file using TUS protocol"""
    print(f"🚀 Starting TUS upload of {FILE_PATH}")
    
    if not os.path.exists(FILE_PATH):
        print(f"❌ File not found: {FILE_PATH}")
        return None
    
    try:
        tus_client = client.TusClient(TUS_ENDPOINT, headers=HEADERS)
        
        uploader = tus_client.uploader(
            FILE_PATH,
            chunk_size=1024 * 1024,  # 1 MB chunks
            metadata={"filename": os.path.basename(FILE_PATH)},
        )
        
        print(f"📏 File size: {uploader.get_file_size():,} bytes")
        
        # Start upload
        uploader.upload()
        
        # Get upload ID from URL
        if hasattr(uploader, 'url') and uploader.url:
            upload_id = uploader.url.split("/")[-1]
            print(f"📋 Upload ID: {upload_id}")
            return upload_id
        else:
            print("❌ Could not get upload ID from TUS uploader")
            return None
            
    except Exception as e:
        print(f"❌ Upload error: {e}")
        return None

def download_file(filename):
    """Download file using storage API"""
    print(f"\n💾 Downloading file: {filename}")
    
    # Create download directory
    os.makedirs(DOWNLOAD_DIR, exist_ok=True)
    
    try:
        response = requests.get(f"{BASE_URL}/api/v1/storage/{filename}", headers=HEADERS, stream=True)
        
        if response.status_code == 200:
            file_path = os.path.join(DOWNLOAD_DIR, filename)
            
            # Download file
            with open(file_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            
            file_size = os.path.getsize(file_path)
            print(f"✅ File downloaded successfully!")
            print(f"   Saved to: {file_path}")
            print(f"   Size: {file_size:,} bytes")
            return file_path
        else:
            print(f"❌ Download failed: {response.status_code}")
            print(f"   Response: {response.text}")
            return None
            
    except Exception as e:
        print(f"❌ Download error: {e}")
        return None

async def main():
    """Main workflow function"""
    print("🚀 Final Demo: Complete Upload Workflow with Progress Tracking")
    print("=" * 60)
    
    # Create progress tracker
    tracker = ProgressTracker()
    
    # Step 1: Upload file
    upload_id = upload_file()
    if not upload_id:
        print("❌ Failed to upload file. Exiting.")
        return
    
    # Step 2: Start WebSocket tracking
    print(f"\n🔌 Starting progress tracking...")
    
    # Run WebSocket tracking
    await track_progress(upload_id, tracker)
    
    # Step 3: Download the file
    filename = os.path.basename(FILE_PATH)
    downloaded_file = download_file(filename)
    
    # Step 4: Summary
    print(f"\n📋 Final Demo Summary:")
    print(f"   Upload ID: {upload_id}")
    print(f"   File: {FILE_PATH}")
    print(f"   WebSocket Status: {tracker.status}")
    print(f"   Messages Received: {tracker.messages_received}")
    print(f"   Final Progress: {tracker.progress:.1f}%")
    if downloaded_file:
        print(f"   Downloaded to: {downloaded_file}")
    
    print(f"\n✨ Final demo completed!")
    print(f"\n📝 Note: If WebSocket progress tracking didn't work, the server may need to be restarted")
    print(f"   to pick up the latest code changes for progress broadcasting.")

if __name__ == "__main__":
    asyncio.run(main())

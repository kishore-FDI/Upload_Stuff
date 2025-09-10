#!/usr/bin/env python3
"""
Simple demo: Upload file and download it
"""

import os
import requests
from tusclient import client

# Configuration
BASE_URL = "http://localhost:8080"
TUS_ENDPOINT = "http://localhost:8080/api/v1/uploads/"
FILE_PATH = "test/test.mp4"
DOWNLOAD_DIR = "test/downloads"

HEADERS = {
    "X-API-KEY": "5cf74fb0ea2ec521d85c36e0e4468749d6aaee5b8f5ff5c35c0f9a3260a96a60",
    "X-Username": "user123444",
}

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

def list_uploaded_files():
    """List files in uploads_data directory"""
    print(f"\n📁 Checking uploaded files...")
    uploads_dir = "uploads_data"
    
    if os.path.exists(uploads_dir):
        files = os.listdir(uploads_dir)
        # Filter out .info files
        actual_files = [f for f in files if not f.endswith('.info')]
        print(f"   Found {len(actual_files)} files:")
        for file in actual_files:
            file_path = os.path.join(uploads_dir, file)
            file_size = os.path.getsize(file_path)
            print(f"   - {file} ({file_size:,} bytes)")
        return actual_files
    else:
        print(f"   Uploads directory not found: {uploads_dir}")
        return []

def main():
    """Main workflow function"""
    print("🚀 Simple Upload and Download Demo")
    print("=" * 40)
    
    # Step 1: Upload file
    upload_id = upload_file()
    if not upload_id:
        print("❌ Failed to upload file. Exiting.")
        return
    
    # Step 2: List uploaded files
    uploaded_files = list_uploaded_files()
    
    # Step 3: Download the file
    filename = os.path.basename(FILE_PATH)
    if filename in uploaded_files:
        downloaded_file = download_file(filename)
    else:
        print(f"❌ File {filename} not found in uploads directory")
        downloaded_file = None
    
    # Step 4: Summary
    print(f"\n📋 Demo Summary:")
    print(f"   Upload ID: {upload_id}")
    print(f"   File: {FILE_PATH}")
    if downloaded_file:
        print(f"   Downloaded to: {downloaded_file}")
    
    print(f"\n✨ Demo completed!")

if __name__ == "__main__":
    main()

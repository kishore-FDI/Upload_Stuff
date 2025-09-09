import time
from tusclient import client

TUS_ENDPOINT = "http://localhost:8080/api/v1/uploads/"
FILE_PATH = "test.mp4"

HEADERS = {
    "X-API-KEY": "5cf74fb0ea2ec521d85c36e0e4468749d6aaee5b8f5ff5c35c0f9a3260a96a60",
    "X-Username": "user123444",
}

def main():
    tus_client = client.TusClient(TUS_ENDPOINT, headers=HEADERS)

    uploader = tus_client.uploader(
        FILE_PATH,
        chunk_size=5242880,  # 5 MB
        metadata={"filename": "test.mp4"},
    )

    print(f"Starting upload of {FILE_PATH} ({uploader.get_file_size()} bytes)")
    print(f"Upload URL: {uploader.url}")
    
    # Upload first few chunks
    chunks_uploaded = 0
    max_chunks_before_interrupt = 3  # Upload 3 chunks then interrupt
    
    while uploader.offset < uploader.get_file_size() and chunks_uploaded < max_chunks_before_interrupt:
        uploader.upload_chunk()
        chunks_uploaded += 1
        print(f"Uploaded {uploader.offset} / {uploader.get_file_size()} bytes (chunk {chunks_uploaded})")
    
    if uploader.offset < uploader.get_file_size():
        print(f"\n🛑 INTERRUPTING UPLOAD at {uploader.offset} bytes")
        print("Waiting 5 seconds before resuming...")
        
        # Countdown 5 seconds
        for i in range(5, 0, -1):
            print(f"Resuming in {i} seconds...", end="\r")
            time.sleep(1)
        print("\n")
        
        print("🔄 RESUMING UPLOAD...")
        
        # Resume upload from where we left off
        while uploader.offset < uploader.get_file_size():
            uploader.upload_chunk()
            print(f"Resumed: {uploader.offset} / {uploader.get_file_size()} bytes")

    print(f"\n✅ Upload completed! Upload URL: {uploader.url}")

if __name__ == "__main__":
    main()


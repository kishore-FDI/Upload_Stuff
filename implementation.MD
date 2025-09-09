# AI-Powered Content Moderation Pipeline

This project simulates a **content moderation and tiered storage pipeline**. Files are uploaded with resumable transfer, moderated by AI, and stored in simulated storage tiers based on access frequency. Real-time progress and events are available via WebSocket or webhook.

---

## Architecture Blocks

### 1. **Client**

- Uploads and downloads files via HTTP/S.
- Supports resumable uploads using the **TUS protocol** (`POST`, `PATCH`, `HEAD`).
- Receives real-time progress updates through WebSocket or webhook.
- Authenticates with **JWT** or **signed URLs**.

### 2. **Nginx Gateway**

- Acts as the **entry point**.
- Terminates SSL/TLS.
- Handles routing to Go service and AI service.
- Provides **rate limiting** and **load balancing**.

### 3. **Go Service (Core Pipeline)**

- Implements **TUS protocol** for resumable uploads.
- Handles **range requests** for downloads.
- Tracks upload progress and state in Redis.
- Pushes progress updates to WebSocket channel.
- Applies **tier migration logic** based on access frequency.
- Validates authentication.

### 4. **Redis**

- Central **metadata store** for files.
- Tracks:
  - File paths and IDs
  - Upload offsets
  - Access counters
  - Moderation status
- Provides **queues/streams** for:
  - Moderation tasks
  - Tier migration jobs

### 5. **WebSocket / Webhook**

- WebSocket: real-time push of upload progress and completion.
- Webhook: asynchronous notifications (e.g., “moderation complete”).
- Decouples **state change reporting** from the upload channel.

### 6. **AI Moderation Service (Python, FastAPI)**

- Runs in a separate container.
- Handles moderation tasks:
  - Text: toxic content detection via HuggingFace Transformers.
  - Images/videos: preprocessing with OpenCV/Pillow, classification with PyTorch/TensorFlow.
- Returns moderation decision (approved/flagged) back to Go service.
- Isolated from core pipeline for scalability.

### 7. **Tiered Storage Simulation**

- Local folders simulate cloud storage:
  - **CDN (`./storage/cdn`)**: hot files, lowest latency (~0ms).
  - **S3 (`./storage/s3`)**: warm files, added latency (~50–100ms).
  - **R2 (`./storage/r2`)**: cold files, higher latency (~200–300ms).
- Files migrate between tiers based on **access frequency counters** in Redis.
- Optional **Redis cache** simulates CDN-style speedups.

### 8. **Moderation & Migration Workflow**

1. File uploaded into **CDN tier**.
2. Moderation task pushed into Redis queue → AI service processes it.
3. Result stored in Redis and returned to Go service.
4. Access frequency logged on every download.
5. Periodic migration jobs move files to S3 or R2 based on counters.

### 9. **Observability**

- Logs structured for:
  - Upload start/completion
  - Moderation results
  - Tier migration
- Metrics collected for throughput, latency, and queue depth.
- Extendable with **Prometheus + Grafana** dashboards.

### 10. **Deployment**

- **Docker Compose** orchestrates services:
  - Go service
  - Python AI service
  - Redis
  - Nginx
- Local volumes mounted:
  - `./storage/cdn`
  - `./storage/s3`
  - `./storage/r2`

---

## Workflow Summary

1. Client sends resumable upload via TUS (HTTP/S).
2. Go service streams chunks to storage, tracks progress in Redis.
3. Progress updates pushed to WebSocket or queried via TUS `HEAD`.
4. Go service submits moderation job → AI service returns result.
5. Files served from CDN tier initially, migrated to S3 or R2 as they cool.
6. Redis maintains metadata, counters, and queues(If possible more).

---

<!-- Here’s your updated markdown incorporating **tiered latency simulation and caching**: -->

---

# AI-Powered Content Moderation Pipeline

## Project Description

Build a content moderation pipeline that:

- Moderates content using AI.
- Segregates storage based on access frequency:

  - Frequently accessed → CDN (`./storage/cdn`)
  - Moderately accessed → S3 (`./storage/s3`) with simulated latency
  - Rarely accessed → Cloudflare R2 (`./storage/r2`) with higher simulated latency

- Provides real-time upload progress via socket or webhook.
- Supports resumable uploads after interruptions.
- Simulates cloud infrastructure locally, including caching and tiered performance.

---

## Tech Stack

### Core Pipeline (Go)

- **Go (Golang)** – High-performance, concurrency with goroutines; handles uploads/downloads, storage tiering, and metadata operations efficiently.
- **Redis** – In-memory metadata store, access counters, task queues; supports persistence via RDB/AOF.
- **Redis Queue / Streams** – Asynchronous task processing for moderation, storage tier migration, and notifications.
- **Local folders** (`./storage/cdn`, `./storage/s3`, `./storage/r2`) – Simulate tiered storage; S3 and R2 accesses include artificial latency to mimic cloud performance.
- **Optional caching layer** – Redis or in-memory cache for hot assets to demonstrate CDN-like speed.
- **Nginx** – Reverse proxy for routing, SSL termination, load balancing, and rate limiting.

### AI Moderation Service (Python)

- **Python 3.11+** – AI/ML ecosystem for moderation tasks.
- **FastAPI** – Lightweight API server exposing moderation endpoints.
- **Docker** – Containerizes the AI service for isolation and serverless simulation.
- **HuggingFace Transformers** – Text moderation (toxic content detection).
- **OpenCV / Pillow** – Image/video preprocessing.
- **PyTorch / TensorFlow** – Efficient model inference inside container.

### Client SDK / Communication

- **HTTP/HTTPS** – Upload/download with range requests for resumable transfers.
- **WebSocket (python-socketio)** – Optional real-time upload progress.
- **API keys / JWT / Signed URLs** – Authentication and secure access.

### Supporting Tools

- **Tqdm** – CLI progress bars during testing.
- **Watchdog** – Monitor file access for tier migration simulation.
- **Git** – Version control.

---

## Workflow Highlights

- Files uploaded in chunks; progress tracked via Redis and optionally reported over WebSocket.
- AI moderation performed in Python container via FastAPI; results returned to Go pipeline.
- Storage tiering applied based on access frequency; hot assets cached for fast retrieval.
- S3 and R2 folder accesses include artificial latency to mimic cloud storage.
- Tier migration and cache updates handled asynchronously via Redis queue.

---

## Why This Stack Works

- **Performance & concurrency** – Go handles large-scale uploads/downloads efficiently.
- **AI specialization** – Python isolates AI moderation, leveraging mature libraries.
- **Scalable architecture** – Nginx decouples services, enabling independent AI scaling.
- **Persistence & queueing** – Redis ensures task reliability, counters, and real-time notifications.
- **Local simulation ready** – Entire pipeline runs in containers/folders locally, mimicking cloud deployment, including tiered latency and caching.

---

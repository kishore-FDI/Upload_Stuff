from graphviz import Digraph

dot = Digraph("AI_Powered_Content_Moderation_Pipeline", format="png")
dot.attr(rankdir="TB", size="10")

# ===============================
# Client
# ===============================
dot.node("Client", "Client \n(TUS Upload/Download,\nResume Support,\nAuth w/ JWT or Signed URL)", 
         shape="box", style="filled", fillcolor="lightblue")

# ===============================
# Entry Point (Nginx)
# ===============================
dot.node("Nginx", "Nginx Gateway\n(Routing, SSL/TLS,\nRate Limit,\nLoad Balancing)", 
         shape="box", style="filled", fillcolor="lightyellow")

# ===============================
# Go Service (Core + TUS)
# ===============================
dot.node("GoServer", "Go Service\n- TUS Protocol Handling\n- Chunked Upload + Resume\n- Range Download Support\n- Progress Tracking\n- Tier Migration Logic\n- Auth Validation", 
         shape="box", style="filled", fillcolor="lightgrey")

# ===============================
# WebSocket / Webhook
# ===============================
dot.node("WebSocket", "WebSocket / Webhook\n(Upload Progress,\nCompletion Events)", 
         shape="oval", style="filled", fillcolor="lightgreen")

# ===============================
# Redis
# ===============================
dot.node("Redis", "Redis\n- Metadata (File Paths)\n- Access Counters\n- Upload State\n- Queues (Moderation, Migration)", 
         shape="cylinder", style="filled", fillcolor="orange")

# ===============================
# AI Service (Python)
# ===============================
dot.node("PythonAI", "Python AI Service (FastAPI)\n- Text Moderation (Transformers)\n- Image/Video Moderation (OpenCV + PyTorch)\n- Pre/Post Processing", 
         shape="box", style="filled", fillcolor="pink")

# ===============================
# Storage cluster (with latency simulation)
# ===============================
with dot.subgraph(name="cluster_storage") as c:
    c.attr(label="Tiered Storage (Local Simulation)", style="dashed")
    c.node("CDN", "CDN Tier\n- Hot Files\n- Direct Disk Access\n- Optional Redis Cache\n(Latency: ~0ms)", 
           shape="folder", style="filled", fillcolor="lightgreen")
    c.node("S3", "S3 Tier\n- Warm Files\n- Simulated Object Store\n(Latency: +50–100ms)", 
           shape="folder", style="filled", fillcolor="lightgrey")
    c.node("R2", "R2 Tier\n- Cold Files\n- Simulated Rare Access\n(Latency: +200–300ms)", 
           shape="folder", style="filled", fillcolor="lightpink")

# ===============================
# Edges
# ===============================
# Client flow
dot.edge("Client", "Nginx", "TUS Upload/Download Requests")
dot.edge("Nginx", "GoServer", "Proxy Traffic")

# Upload flow
dot.edge("GoServer", "Redis", "Track Upload State,\nUpdate Counters")
dot.edge("GoServer", "WebSocket", "Emit Progress Updates")

# Moderation flow
dot.edge("GoServer", "Redis", "Push Moderation Task", style="dashed")
dot.edge("GoServer", "PythonAI", "Send File/Content for Moderation\n(via Nginx Proxy)")
dot.edge("PythonAI", "GoServer", "Moderation Result\n(Approved/Flagged)")

# Storage flow
dot.edge("GoServer", "CDN", "Store/Serve Hot Files")
dot.edge("GoServer", "S3", "Store/Serve Warm Files")
dot.edge("GoServer", "R2", "Store/Serve Cold Files")

# Access frequency feedback
dot.edge("CDN", "Redis", "Update Access Frequency", style="dotted")
dot.edge("S3", "Redis", "Update Access Frequency", style="dotted")
dot.edge("R2", "Redis", "Update Access Frequency", style="dotted")

# Tier migration
dot.edge("Redis", "GoServer", "Trigger Tier Migration\nvia Queue", style="dashed")

dot.render("ai_pipeline_detailed", format="svg", view=False)

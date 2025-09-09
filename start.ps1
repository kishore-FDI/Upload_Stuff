# start-dev.ps1
# Start Docker Redis and Go server for Media Pipeline

# Launch Docker Desktop (if not already running)
if (-not (Get-Process -Name "Docker Desktop" -ErrorAction SilentlyContinue)) {
    Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    Write-Host "Starting Docker Desktop..."
    Start-Sleep -Seconds 10  # give it time to come up
}

# Launch Postman (optional, adjust path if installed differently)
if (-not (Get-Process -Name "Postman" -ErrorAction SilentlyContinue)) {
    Start-Process "C:\Users\kisho\AppData\Local\Postman\Postman.exe"
    Write-Host "Starting Postman..."
}

# Check Docker
docker version | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker not running" -ForegroundColor Red
    exit 1
}

# Start or reuse Redis
$redisContainer = docker ps -q -f name=mediapipeline-redis
if (-not $redisContainer) {
    $stoppedContainer = docker ps -aq -f name=mediapipeline-redis
    if ($stoppedContainer) {
        docker start $stoppedContainer | Out-Null
    } else {
        docker run -d --name mediapipeline-redis -p 6379:6379 redis:7-alpine redis-server --appendonly yes | Out-Null
    }
}

# Wait for Redis
for ($i = 0; $i -lt 30; $i++) {
    docker exec mediapipeline-redis redis-cli ping | Out-Null
    if ($LASTEXITCODE -eq 0) { break }
    Start-Sleep -Seconds 1
}
if ($LASTEXITCODE -ne 0) {
    Write-Host "Redis failed to start" -ForegroundColor Red
    exit 1
}

# Check Go
go version | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Go not installed" -ForegroundColor Red
    exit 1
}

# Install deps
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to install Go dependencies" -ForegroundColor Red
    exit 1
}

# Env vars
$env:REDIS_HOST = "localhost"
$env:REDIS_PORT = "6379"
$env:ENVIRONMENT = "development"
$env:PORT = "8080"

# Start server
go run main.go
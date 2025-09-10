package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mediapipeline/internal/config"
	"mediapipeline/internal/db"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"github.com/tus/tusd/pkg/memorylocker"
)

// read tusd .info file for metadata
func readTusInfo(id string) (*tusd.FileInfo, error) {
	infoPath := filepath.Join("./uploads_data", id+".info")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}
	var info tusd.FileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// initialize tusd handler
func initTusHandler(_ *config.Config) (*tusd.UnroutedHandler, error) {
	uploadDir := "./uploads_data"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	store := filestore.New(uploadDir)
	locker := memorylocker.New()
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	locker.UseIn(composer)

	config := tusd.Config{
		StoreComposer:           composer,
		BasePath:                "/api/v1/uploads/",
		DisableDownload:         false,
		DisableTermination:      false,
		NotifyCreatedUploads:    true,
		NotifyCompleteUploads:   true,
		NotifyUploadProgress:    true,
		RespectForwardedHeaders: true,
	}

	// pre upload create callback
	config.PreUploadCreateCallback = func(hook tusd.HookEvent) error {
		h := hook.HTTPRequest
		apiKey := h.Header.Get("X-API-KEY")
		username := h.Header.Get("X-Username")
		if apiKey == "" || username == "" {
			return tusd.NewHTTPError(fmt.Errorf("missing auth headers"), http.StatusBadRequest)
		}
		business, err := db.GetBusinessByAPIKey(apiKey)
		if err != nil || business == nil {
			return tusd.NewHTTPError(fmt.Errorf("invalid api key"), http.StatusUnauthorized)
		}
		if hook.Upload.MetaData == nil {
			hook.Upload.MetaData = make(map[string]string)
		}
		hook.Upload.MetaData["business_id"] = fmt.Sprintf("%d", business.ID)
		hook.Upload.MetaData["username"] = username
		return nil
	}

	// pre finish response callback
	config.PreFinishResponseCallback = func(hook tusd.HookEvent) error {
		id := hook.Upload.ID
		meta := hook.Upload.MetaData
		uploadKey := "upload:" + id
		fields := map[string]interface{}{
			"business_id": meta["business_id"],
			"username":    meta["username"],
			"status":      "uploaded",
			"size":        hook.Upload.Size,
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		}
		if fn, ok := meta["filename"]; ok && fn != "" {
			fields["filename"] = fn
			src := filepath.Join("./uploads_data", id)
			dst := filepath.Join("./uploads_data", fn)
			if _, err := os.Stat(src); err == nil {
				if err := os.Rename(src, dst); err != nil {
					in, _ := os.Open(src)
					out, _ := os.Create(dst)
					io.Copy(out, in)
					in.Close()
					out.Close()
				}
			}
		}
		_ = db.RDB.HSet(db.Ctx, uploadKey, fields)
		_ = db.RDB.Expire(db.Ctx, uploadKey, 24*time.Hour)
		return nil
	}

	// initialize tusd handler
	h, err := tusd.NewUnroutedHandler(config)
	if err != nil {
		return nil, err
	}

	// go routine to handle upload progress and broadcast updates
	go func() {
		for {
			select {
			case info := <-h.CreatedUploads:
				log.Printf("Upload %s created (size: %d)", info.Upload.ID, info.Upload.Size)

				// Store initial upload info in Redis
				uploadKey := "upload:" + info.Upload.ID
				fields := map[string]interface{}{
					"status":     "created",
					"size":       info.Upload.Size,
					"offset":     0,
					"progress":   0.0,
					"created_at": time.Now().UTC().Format(time.RFC3339),
				}
				_ = db.RDB.HSet(db.Ctx, uploadKey, fields)
				_ = db.RDB.Expire(db.Ctx, uploadKey, 24*time.Hour)

				// Broadcast upload created event
				GetConnectionManager().BroadcastProgress(info.Upload.ID, ProgressMessage{
					Type:      "created",
					UploadID:  info.Upload.ID,
					Progress:  0.0,
					BytesSent: 0,
					TotalSize: info.Upload.Size,
					Status:    "created",
					Message:   "Upload created",
				})

			case info := <-h.UploadProgress:
				if info.Upload.Size > 0 {
					percent := float64(info.Upload.Offset) / float64(info.Upload.Size) * 100

					// Update progress in Redis
					uploadKey := "upload:" + info.Upload.ID
					fields := map[string]interface{}{
						"status":     "uploading",
						"offset":     info.Upload.Offset,
						"progress":   percent,
						"updated_at": time.Now().UTC().Format(time.RFC3339),
					}
					_ = db.RDB.HSet(db.Ctx, uploadKey, fields)

					// Print progress bar to console
					barWidth := 50
					filled := int(percent / 100 * float64(barWidth))
					bar := fmt.Sprintf("\r[%s%s] %.2f%%",
						strings.Repeat("=", filled),
						strings.Repeat(" ", barWidth-filled),
						percent,
					)
					fmt.Print(bar)

					// Broadcast progress update
					GetConnectionManager().BroadcastProgress(info.Upload.ID, ProgressMessage{
						Type:      "progress",
						UploadID:  info.Upload.ID,
						Progress:  percent,
						BytesSent: info.Upload.Offset,
						TotalSize: info.Upload.Size,
						Status:    "uploading",
					})
				}

			case info := <-h.CompleteUploads:
				fmt.Printf("\r[==================================================] 100.00%%\n")
				log.Printf("Upload %s completed (%d bytes)", info.Upload.ID, info.Upload.Size)

				// Update final status in Redis
				uploadKey := "upload:" + info.Upload.ID
				fields := map[string]interface{}{
					"status":       "completed",
					"offset":       info.Upload.Size,
					"progress":     100.0,
					"completed_at": time.Now().UTC().Format(time.RFC3339),
				}
				_ = db.RDB.HSet(db.Ctx, uploadKey, fields)

				// Broadcast completion event
				GetConnectionManager().BroadcastProgress(info.Upload.ID, ProgressMessage{
					Type:      "complete",
					UploadID:  info.Upload.ID,
					Progress:  100.0,
					BytesSent: info.Upload.Size,
					TotalSize: info.Upload.Size,
					Status:    "completed",
					Message:   "Upload completed successfully",
				})
			}
		}
	}()

	return h, nil
}

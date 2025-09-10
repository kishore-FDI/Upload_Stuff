package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func downloadHandler(c *gin.Context) {
	id := c.Param("id")
	filePath := "./uploads_data/" + id

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot access file"})
		}
		return
	}

	info, err := readTusInfo(id)
	filename := id
	if err == nil {
		if fn, ok := info.MetaData["filename"]; ok && fn != "" {
			filename = fn
			realPath := filepath.Join("./uploads_data", filename)
			if _, err := os.Stat(realPath); err == nil {
				filePath = realPath
			}
		}
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.File(filePath)
}

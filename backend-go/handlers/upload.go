package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wholesale/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxUploadBytes   = 5 * 1024 * 1024   // 5 MB
	minFreeDiskBytes = 500 * 1024 * 1024 // 500 MB
	chunkSize        = 64 * 1024         // 64 KB read buffer
)

var allowedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

// POST /api/upload/image
func UploadImage(c *gin.Context) {
	// 文件大小限制（在 multipart 读取前设置）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+1024)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "请选择要上传的图片文件"})
		return
	}
	defer file.Close()

	if header.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "图片大小不能超过 5MB"})
		return
	}

	// 扩展名白名单
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "仅支持 .jpg/.jpeg/.png/.webp/.gif 格式"})
		return
	}

	uploadDir := "./uploads"
	thumbDir := filepath.Join(uploadDir, "thumbnails")

	// 确保目录存在
	if err := os.MkdirAll(thumbDir, 0750); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "目录创建失败"})
		return
	}

	// 磁盘空间检查
	if !hasSufficientDisk(uploadDir, minFreeDiskBytes) {
		c.JSON(http.StatusInsufficientStorage, gin.H{"detail": "服务器磁盘空间不足，请联系管理员"})
		return
	}

	// 生成唯一文件名
	filename := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, filename)

	// 使用分块读写，避免大文件内存占用
	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "文件保存失败"})
		return
	}

	buf := make([]byte, chunkSize)
	var written int64
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				out.Close()
				os.Remove(destPath)
				c.JSON(http.StatusInternalServerError, gin.H{"detail": "文件写入失败"})
				return
			}
			written += int64(n)
			if written > maxUploadBytes {
				out.Close()
				os.Remove(destPath)
				c.JSON(http.StatusBadRequest, gin.H{"detail": "图片大小不能超过 5MB"})
				return
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			out.Close()
			os.Remove(destPath)
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "读取文件失败"})
			return
		}
	}
	out.Close()

	// 优化原图（最大 1200×1200）
	if err := services.OptimizeImage(destPath); err != nil {
		// 优化失败不影响上传
		fmt.Printf("image optimize warning: %v\n", err)
	}

	// 生成缩略图
	thumbURL := ""
	thumbPath := services.CreateThumbnail(destPath, uploadDir)
	if thumbPath != "" {
		thumbURL = "/uploads/thumbnails/" + filepath.Base(thumbPath)
	}

	imgURL := "/uploads/" + filename
	c.JSON(http.StatusOK, gin.H{
		"url":           imgURL,
		"thumbnail_url": thumbURL,
	})
}

// hasSufficientDisk 检查指定目录所在磁盘的可用空间
func hasSufficientDisk(dir string, minBytes int64) bool {
	free, err := getDiskFree(dir)
	if err != nil {
		return true // 无法检查时放行
	}
	return free >= minBytes
}

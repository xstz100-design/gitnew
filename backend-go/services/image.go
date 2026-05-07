package services

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// CreateThumbnail 为上传图片生成缩略图（200×200 JPEG）
// 返回缩略图路径；失败时返回空字符串
func CreateThumbnail(imagePath string, uploadDir string) string {
	img, err := imaging.Open(imagePath, imaging.AutoOrientation(true))
	if err != nil {
		return ""
	}

	// 合成白色背景（处理透明 PNG）
	bounds := img.Bounds()
	bg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			bg.Set(x, y, color.White)
		}
	}
	// 将原图绘制到白色背景上
	dst := imaging.OverlayCenter(bg, img, 1.0)

	// 等比缩放到 200×200
	thumb := imaging.Fit(dst, 200, 200, imaging.Lanczos)

	// 生成缩略图路径
	base := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))
	thumbDir := filepath.Join(uploadDir, "thumbnails")
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return ""
	}
	thumbPath := filepath.Join(thumbDir, base+"_thumb.jpg")

	f, err := os.Create(thumbPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	if err := jpeg.Encode(f, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return ""
	}
	return thumbPath
}

// OptimizeImage 将上传图片等比压缩到最大 1200×1200，原地覆盖保存
func OptimizeImage(imagePath string) error {
	img, err := imaging.Open(imagePath, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w > 1200 || h > 1200 {
		img = imaging.Fit(img, 1200, 1200, imaging.Lanczos)
	}

	return imaging.Save(img, imagePath, imaging.JPEGQuality(90))
}

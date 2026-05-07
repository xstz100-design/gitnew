package handlers

import (
	"net/http"
	"strconv"
	"wholesale/database"
	"wholesale/models"

	"github.com/gin-gonic/gin"
)

// ─────────────────── 请求结构 ───────────────────

type AnnouncementCreateRequest struct {
	ContentZh string                  `json:"content_zh"`
	ContentEn string                  `json:"content_en"`
	Type      models.AnnouncementType `json:"type"` // notice | contact | about
	IsActive  bool                    `json:"is_active"`
	SortOrder int                     `json:"sort_order"`
}

type AnnouncementUpdateRequest struct {
	ContentZh *string                  `json:"content_zh"`
	ContentEn *string                  `json:"content_en"`
	Type      *models.AnnouncementType `json:"type"`
	IsActive  *bool                    `json:"is_active"`
	SortOrder *int                     `json:"sort_order"`
}

// GET /api/announcements/public — 不需要认证，仅 is_active=true
func ListPublicAnnouncements(c *gin.Context) {
	var items []models.Announcement
	q := database.DB.Where("is_active = ?", true)
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	q.Order("sort_order ASC, id DESC").Find(&items)
	c.JSON(http.StatusOK, items)
}

// GET /api/announcements — 管理员，返回全部
func ListAnnouncements(c *gin.Context) {
	var items []models.Announcement
	q := database.DB
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	q.Order("sort_order ASC, id DESC").Find(&items)
	c.JSON(http.StatusOK, items)
}

// POST /api/announcements
func CreateAnnouncement(c *gin.Context) {
	var req AnnouncementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	item := models.Announcement{
		ContentZh: req.ContentZh,
		ContentEn: req.ContentEn,
		Type:      req.Type,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
		CreatedAt: models.NowCambodia(),
		UpdatedAt: models.NowCambodia(),
	}
	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "创建公告失败"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PATCH /api/announcements/:id
func UpdateAnnouncement(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var item models.Announcement
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "公告不存在"})
		return
	}

	var req AnnouncementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": models.NowCambodia()}
	if req.ContentZh != nil {
		updates["content_zh"] = *req.ContentZh
	}
	if req.ContentEn != nil {
		updates["content_en"] = *req.ContentEn
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	database.DB.Model(&item).Updates(updates)
	database.DB.First(&item, id)
	c.JSON(http.StatusOK, item)
}

// DELETE /api/announcements/:id
func DeleteAnnouncement(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var item models.Announcement
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "公告不存在"})
		return
	}
	database.DB.Delete(&item)
	c.JSON(http.StatusOK, gin.H{"message": "公告已删除"})
}

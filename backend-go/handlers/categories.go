package handlers

import (
	"net/http"
	"strconv"
	"wholesale/database"
	"wholesale/models"

	"github.com/gin-gonic/gin"
)

// ─────────────────── 请求结构 ───────────────────

type CategoryCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type CategoryUpdateRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

// GET /api/categories — 公开，仅返回 is_active=true
func ListCategories(c *gin.Context) {
	var cats []models.Category
	database.DB.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&cats)
	c.JSON(http.StatusOK, cats)
}

// GET /api/categories/all — 管理员，返回全部
func ListAllCategories(c *gin.Context) {
	var cats []models.Category
	database.DB.Order("sort_order ASC, id ASC").Find(&cats)
	c.JSON(http.StatusOK, cats)
}

// POST /api/categories
func CreateCategory(c *gin.Context) {
	var req CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	cat := models.Category{
		Name:      req.Name,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
		CreatedAt: models.NowCambodia(),
	}
	if err := database.DB.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "创建分类失败"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// PATCH /api/categories/:id
func UpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cat models.Category
	if err := database.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "分类不存在"})
		return
	}

	var req CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) > 0 {
		database.DB.Model(&cat).Updates(updates)
		database.DB.First(&cat, id)
	}
	c.JSON(http.StatusOK, cat)
}

// DELETE /api/categories/:id
func DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cat models.Category
	if err := database.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "分类不存在"})
		return
	}
	database.DB.Delete(&cat)
	c.JSON(http.StatusOK, gin.H{"message": "分类已删除"})
}

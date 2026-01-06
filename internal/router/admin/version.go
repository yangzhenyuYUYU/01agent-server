package admin

import (
	"encoding/json"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetVersionList 获取版本列表（CRUD 列表接口）
func (h *AdminHandler) GetVersionList(c *gin.Context) {
	// 使用 CRUD handler 的 List 方法
	versionCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.Version{},
		SearchFields:   []string{"version"},
		DefaultOrderBy: "date",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	versionCRUD.List(c)
}

// GetVersionItems 获取版本列表（用于前端渲染版本更新记录页）
func (h *AdminHandler) GetVersionItems(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"min=1"`
		PageSize int `form:"page_size" binding:"min=1"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 获取总数
	var total int64
	if err := repository.DB.Model(&models.Version{}).Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询，按日期和ID降序排序
	offset := (req.Page - 1) * req.PageSize
	var versions []models.Version
	if err := repository.DB.Order("date DESC, id DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&versions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(versions))
	for _, item := range versions {
		// 解析 highlights（可能是 JSON 字符串或已经是数组）
		var highlights interface{}
		if err := json.Unmarshal([]byte(item.Highlights), &highlights); err != nil {
			// 如果解析失败，尝试作为字符串处理
			highlights = item.Highlights
		}

		versionType := ""
		if item.Type != nil {
			versionType = string(*item.Type)
		}

		result = append(result, gin.H{
			"id":         item.ID,
			"version":    item.Version,
			"date":       item.Date.Format("2006-01-02"),
			"title":      item.Title,
			"type":       versionType,
			"highlights": highlights,
			"created_at": item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      result,
	})
}

// CreateVersion 创建版本
func (h *AdminHandler) CreateVersion(c *gin.Context) {
	var req struct {
		Version    string   `json:"version" binding:"required"`
		Date       string   `json:"date" binding:"required"` // YYYY-MM-DD
		Title      string   `json:"title" binding:"required"`
		Type       *string  `json:"type"` // major / minor / patch
		Highlights []string `json:"highlights" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，应为 YYYY-MM-DD"))
		return
	}

	// 转换 highlights 为 JSON 字符串
	highlightsJSON, err := json.Marshal(req.Highlights)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "高亮信息格式错误"))
		return
	}

	// 转换版本类型
	var versionType *models.VersionType
	if req.Type != nil && *req.Type != "" {
		vt := models.VersionType(*req.Type)
		versionType = &vt
	}

	version := models.Version{
		Version:    req.Version,
		Date:       date,
		Title:      req.Title,
		Highlights: string(highlightsJSON),
		Type:       versionType,
	}

	if err := repository.DB.Create(&version).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	middleware.Success(c, "success", gin.H{})
}

// UpdateVersion 更新版本
func (h *AdminHandler) UpdateVersion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "缺少ID参数"))
		return
	}

	var req struct {
		Version    *string  `json:"version"`
		Date       *string  `json:"date"` // YYYY-MM-DD
		Title      *string  `json:"title"`
		Type       *string  `json:"type"` // major / minor / patch
		Highlights []string `json:"highlights"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 查询现有记录
	var version models.Version
	if err := repository.DB.Where("id = ?", id).First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "记录不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 更新字段
	updateData := make(map[string]interface{})

	if req.Version != nil {
		updateData["version"] = *req.Version
	}

	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，应为 YYYY-MM-DD"))
			return
		}
		updateData["date"] = date
	}

	if req.Title != nil {
		updateData["title"] = *req.Title
	}

	if req.Highlights != nil {
		// 转换 highlights 为 JSON 字符串
		highlightsJSON, err := json.Marshal(req.Highlights)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "高亮信息格式错误"))
			return
		}
		updateData["highlights"] = string(highlightsJSON)
	}

	if req.Type != nil {
		if *req.Type == "" {
			updateData["type"] = nil
		} else {
			vt := models.VersionType(*req.Type)
			updateData["type"] = vt
		}
	}

	// 执行更新
	if err := repository.DB.Model(&version).Updates(updateData).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 查询更新后的记录
	if err := repository.DB.Where("id = ?", id).First(&version).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 解析 highlights 用于返回
	var highlights interface{}
	if err := json.Unmarshal([]byte(version.Highlights), &highlights); err != nil {
		highlights = version.Highlights
	}

	versionType := ""
	if version.Type != nil {
		versionType = string(*version.Type)
	}

	middleware.Success(c, "更新成功", gin.H{
		"id":         version.ID,
		"version":    version.Version,
		"date":       version.Date.Format("2006-01-02"),
		"title":      version.Title,
		"type":       versionType,
		"highlights": highlights,
		"created_at": version.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at": version.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}


package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserCustomHandler user custom handler
type UserCustomHandler struct {
	db *gorm.DB
}

// NewUserCustomHandler create user custom handler
func NewUserCustomHandler() *UserCustomHandler {
	return &UserCustomHandler{
		db: repository.DB,
	}
}

// ========================= Request/Response Models =========================

// SaveCustomSizeParams save custom size params
type SaveCustomSizeParams struct {
	ID        *int   `json:"id"`
	Name      *string `json:"name"`
	Data      map[string]interface{} `json:"data"`
	Status    *int16 `json:"status"`
	IsDefault *bool  `json:"is_default"`
	SortOrder *int   `json:"sort_order"`
}

// SaveCustomThemeParams save custom theme params
type SaveCustomThemeParams struct {
	ID        *int   `json:"id"`
	Name      *string `json:"name"`
	Data      map[string]interface{} `json:"data"`
	Status    *int16 `json:"status"`
	IsDefault *bool  `json:"is_default"`
	SortOrder *int   `json:"sort_order"`
}

// ========================= Custom Size Handlers =========================

// GetCustomSizeList get custom size list
func (h *UserCustomHandler) GetCustomSizeList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusStr := c.Query("status")
	name := c.Query("name")

	query := h.db.Model(&models.UserCustomSize{}).Where("user_id = ?", userID)

	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	var sizes []models.UserCustomSize
	if err := query.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&sizes).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(sizes))
	for _, size := range sizes {
		var data interface{}
		json.Unmarshal([]byte(size.Data), &data)

		items = append(items, map[string]interface{}{
			"id":         size.ID,
			"user_id":   size.UserID,
			"name":      size.Name,
			"data":      data,
			"status":    size.Status,
			"is_default": size.IsDefault,
			"sort_order": size.SortOrder,
			"created_at": size.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": size.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items":      items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// SaveCustomSize save custom size
func (h *UserCustomHandler) SaveCustomSize(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req SaveCustomSizeParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	var size models.UserCustomSize
	var isNew bool

	if req.ID != nil && *req.ID > 0 {
		if err := h.db.Where("id = ? AND user_id = ?", *req.ID, userID).First(&size).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "尺寸配置不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
			return
		}
		isNew = false
	} else {
		size.UserID = userID
		size.Name = ""
		if req.Name != nil {
			size.Name = *req.Name
		}
		size.Status = models.CustomConfigStatusActive
		if req.Status != nil {
			size.Status = models.CustomConfigStatus(*req.Status)
		}
		size.IsDefault = false
		if req.IsDefault != nil {
			size.IsDefault = *req.IsDefault
		}
		size.SortOrder = 0
		if req.SortOrder != nil {
			size.SortOrder = *req.SortOrder
		}
		dataBytes, _ := json.Marshal(req.Data)
		size.Data = string(dataBytes)
		isNew = true
	}

	if req.Name != nil {
		if len(*req.Name) > 100 {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "尺寸名称不能超过100个字符"))
			return
		}
		size.Name = *req.Name
	}

	if req.Data != nil {
		dataBytes, _ := json.Marshal(req.Data)
		size.Data = string(dataBytes)
	}

	if req.Status != nil {
		size.Status = models.CustomConfigStatus(*req.Status)
	}

	if req.IsDefault != nil {
		size.IsDefault = *req.IsDefault
	}

	if req.SortOrder != nil {
		size.SortOrder = *req.SortOrder
	}

	if isNew {
		if err := h.db.Create(&size).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
			return
		}
	} else {
		if err := h.db.Save(&size).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
	}

	var data interface{}
	json.Unmarshal([]byte(size.Data), &data)

	middleware.Success(c, "success", gin.H{
		"id":         size.ID,
		"user_id":   size.UserID,
		"name":      size.Name,
		"data":      data,
		"status":    size.Status,
		"is_default": size.IsDefault,
		"sort_order": size.SortOrder,
		"created_at": size.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": size.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// SetDefaultSize set default size
func (h *UserCustomHandler) SetDefaultSize(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	sizeIDStr := c.Param("id")

	sizeID, err := strconv.Atoi(sizeIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var size models.UserCustomSize
	if err := h.db.Where("id = ? AND user_id = ?", sizeID, userID).First(&size).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到尺寸配置"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 先取消其他默认
	h.db.Model(&models.UserCustomSize{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false)

	// 设置当前为默认
	size.IsDefault = true
	if err := h.db.Save(&size).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// GetCustomSizeDetail get custom size detail
func (h *UserCustomHandler) GetCustomSizeDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	sizeIDStr := c.Param("id")

	sizeID, err := strconv.Atoi(sizeIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var size models.UserCustomSize
	if err := h.db.Where("id = ? AND user_id = ?", sizeID, userID).First(&size).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到对应尺寸配置"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var data interface{}
	json.Unmarshal([]byte(size.Data), &data)

	middleware.Success(c, "success", gin.H{
		"id":         size.ID,
		"user_id":   size.UserID,
		"name":      size.Name,
		"data":      data,
		"status":    size.Status,
		"is_default": size.IsDefault,
		"sort_order": size.SortOrder,
		"created_at": size.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": size.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteCustomSize delete custom size
func (h *UserCustomHandler) DeleteCustomSize(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	sizeIDStr := c.Param("id")

	sizeID, err := strconv.Atoi(sizeIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var size models.UserCustomSize
	if err := h.db.Where("id = ? AND user_id = ?", sizeID, userID).First(&size).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到尺寸配置，无法删除"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&size).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// ========================= Custom Theme Handlers =========================

// GetCustomThemeList get custom theme list
func (h *UserCustomHandler) GetCustomThemeList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusStr := c.Query("status")
	name := c.Query("name")

	query := h.db.Model(&models.UserCustomTheme{}).Where("user_id = ?", userID)

	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	var themes []models.UserCustomTheme
	if err := query.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&themes).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(themes))
	for _, theme := range themes {
		var data interface{}
		json.Unmarshal([]byte(theme.Data), &data)

		items = append(items, map[string]interface{}{
			"id":         theme.ID,
			"user_id":   theme.UserID,
			"name":      theme.Name,
			"data":      data,
			"status":    theme.Status,
			"is_default": theme.IsDefault,
			"sort_order": theme.SortOrder,
			"created_at": theme.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": theme.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items":      items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// SaveCustomTheme save custom theme
func (h *UserCustomHandler) SaveCustomTheme(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req SaveCustomThemeParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	var theme models.UserCustomTheme
	var isNew bool

	if req.ID != nil && *req.ID > 0 {
		if err := h.db.Where("id = ? AND user_id = ?", *req.ID, userID).First(&theme).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "主题配置不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
			return
		}
		isNew = false
	} else {
		theme.UserID = userID
		theme.Name = ""
		if req.Name != nil {
			theme.Name = *req.Name
		}
		theme.Status = models.CustomConfigStatusActive
		if req.Status != nil {
			theme.Status = models.CustomConfigStatus(*req.Status)
		}
		theme.IsDefault = false
		if req.IsDefault != nil {
			theme.IsDefault = *req.IsDefault
		}
		theme.SortOrder = 0
		if req.SortOrder != nil {
			theme.SortOrder = *req.SortOrder
		}
		dataBytes, _ := json.Marshal(req.Data)
		theme.Data = string(dataBytes)
		isNew = true
	}

	if req.Name != nil {
		if len(*req.Name) > 100 {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "主题名称不能超过100个字符"))
			return
		}
		theme.Name = *req.Name
	}

	if req.Data != nil {
		dataBytes, _ := json.Marshal(req.Data)
		theme.Data = string(dataBytes)
	}

	if req.Status != nil {
		theme.Status = models.CustomConfigStatus(*req.Status)
	}

	if req.IsDefault != nil {
		theme.IsDefault = *req.IsDefault
	}

	if req.SortOrder != nil {
		theme.SortOrder = *req.SortOrder
	}

	if isNew {
		if err := h.db.Create(&theme).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
			return
		}
	} else {
		if err := h.db.Save(&theme).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
	}

	var data interface{}
	json.Unmarshal([]byte(theme.Data), &data)

	middleware.Success(c, "success", gin.H{
		"id":         theme.ID,
		"user_id":   theme.UserID,
		"name":      theme.Name,
		"data":      data,
		"status":    theme.Status,
		"is_default": theme.IsDefault,
		"sort_order": theme.SortOrder,
		"created_at": theme.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": theme.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// SetDefaultTheme set default theme
func (h *UserCustomHandler) SetDefaultTheme(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	themeIDStr := c.Param("id")

	themeID, err := strconv.Atoi(themeIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var theme models.UserCustomTheme
	if err := h.db.Where("id = ? AND user_id = ?", themeID, userID).First(&theme).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到主题配置"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 先取消其他默认
	h.db.Model(&models.UserCustomTheme{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false)

	// 设置当前为默认
	theme.IsDefault = true
	if err := h.db.Save(&theme).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// GetCustomThemeDetail get custom theme detail
func (h *UserCustomHandler) GetCustomThemeDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	themeIDStr := c.Param("id")

	themeID, err := strconv.Atoi(themeIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var theme models.UserCustomTheme
	if err := h.db.Where("id = ? AND user_id = ?", themeID, userID).First(&theme).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到对应主题配置"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var data interface{}
	json.Unmarshal([]byte(theme.Data), &data)

	middleware.Success(c, "success", gin.H{
		"id":         theme.ID,
		"user_id":   theme.UserID,
		"name":      theme.Name,
		"data":      data,
		"status":    theme.Status,
		"is_default": theme.IsDefault,
		"sort_order": theme.SortOrder,
		"created_at": theme.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": theme.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteCustomTheme delete custom theme
func (h *UserCustomHandler) DeleteCustomTheme(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	themeIDStr := c.Param("id")

	themeID, err := strconv.Atoi(themeIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var theme models.UserCustomTheme
	if err := h.db.Where("id = ? AND user_id = ?", themeID, userID).First(&theme).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到主题配置，无法删除"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&theme).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// SetupUserCustomRoutes setup user custom routes
func SetupUserCustomRoutes(r *gin.Engine) {
	handler := NewUserCustomHandler()

	userCustomGroup := r.Group("/api/v1/user-custom")
	userCustomGroup.Use(middleware.JWTAuth())
	{
		// Size routes
		userCustomGroup.GET("/size/list", handler.GetCustomSizeList)
		userCustomGroup.POST("/size/save", handler.SaveCustomSize)
		userCustomGroup.PUT("/size/set-default/:id", handler.SetDefaultSize)
		userCustomGroup.GET("/size/:id", handler.GetCustomSizeDetail)
		userCustomGroup.DELETE("/size/:id", handler.DeleteCustomSize)

		// Theme routes
		userCustomGroup.GET("/theme/list", handler.GetCustomThemeList)
		userCustomGroup.POST("/theme/save", handler.SaveCustomTheme)
		userCustomGroup.PUT("/theme/set-default/:id", handler.SetDefaultTheme)
		userCustomGroup.GET("/theme/:id", handler.GetCustomThemeDetail)
		userCustomGroup.DELETE("/theme/:id", handler.DeleteCustomTheme)
	}
}


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

// PromptTemplateHandler prompt template handler
type PromptTemplateHandler struct {
	db *gorm.DB
}

// NewPromptTemplateHandler create prompt template handler
func NewPromptTemplateHandler() *PromptTemplateHandler {
	return &PromptTemplateHandler{
		db: repository.DB,
	}
}

// ========================= Request/Response Models =========================

// SavePromptTemplateParams save prompt template params
type SavePromptTemplateParams struct {
	ID          *int                   `json:"id"`
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Status      *int16                 `json:"status"`
	IsDefault   *bool                  `json:"is_default"`
	SortOrder   *int                   `json:"sort_order"`
}

// ========================= Prompt Template Handlers =========================

// GetPromptTemplateList get prompt template list
func (h *PromptTemplateHandler) GetPromptTemplateList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusStr := c.Query("status")
	name := c.Query("name")

	query := h.db.Model(&models.UserPromptTemplate{}).Where("user_id = ?", userID)

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
	var templates []models.UserPromptTemplate
	if err := query.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(templates))
	for _, template := range templates {
		var data interface{}
		json.Unmarshal([]byte(template.Data), &data)

		item := map[string]interface{}{
			"id":          template.ID,
			"user_id":    template.UserID,
			"name":       template.Name,
			"data":       data,
			"status":     template.Status,
			"is_default": template.IsDefault,
			"sort_order": template.SortOrder,
			"created_at": template.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": template.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if template.Description != nil {
			item["description"] = *template.Description
		} else {
			item["description"] = nil
		}

		items = append(items, item)
	}

	middleware.Success(c, "success", gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// SavePromptTemplate save prompt template
func (h *PromptTemplateHandler) SavePromptTemplate(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req SavePromptTemplateParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	var template models.UserPromptTemplate
	var isNew bool

	if req.ID != nil && *req.ID > 0 {
		if err := h.db.Where("id = ? AND user_id = ?", *req.ID, userID).First(&template).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "模板不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
			return
		}
		isNew = false
	} else {
		template.UserID = userID
		template.Name = ""
		if req.Name != nil {
			template.Name = *req.Name
		}
		template.Description = req.Description
		template.Status = models.PromptTemplateStatusActive
		if req.Status != nil {
			template.Status = models.PromptTemplateStatus(*req.Status)
		}
		template.IsDefault = false
		if req.IsDefault != nil {
			template.IsDefault = *req.IsDefault
		}
		template.SortOrder = 0
		if req.SortOrder != nil {
			template.SortOrder = *req.SortOrder
		}
		dataBytes, _ := json.Marshal(req.Data)
		template.Data = string(dataBytes)
		isNew = true
	}

	if req.Name != nil {
		if len(*req.Name) > 100 {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "模板名称不能超过100个字符"))
			return
		}
		template.Name = *req.Name
	}

	if req.Description != nil {
		template.Description = req.Description
	}

	if req.Data != nil {
		dataBytes, _ := json.Marshal(req.Data)
		template.Data = string(dataBytes)
	}

	if req.Status != nil {
		template.Status = models.PromptTemplateStatus(*req.Status)
	}

	if req.IsDefault != nil {
		template.IsDefault = *req.IsDefault
	}

	if req.SortOrder != nil {
		template.SortOrder = *req.SortOrder
	}

	if isNew {
		if err := h.db.Create(&template).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
			return
		}
	} else {
		if err := h.db.Save(&template).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
	}

	var data interface{}
	json.Unmarshal([]byte(template.Data), &data)

	response := map[string]interface{}{
		"id":          template.ID,
		"user_id":    template.UserID,
		"name":       template.Name,
		"data":       data,
		"status":     template.Status,
		"is_default": template.IsDefault,
		"sort_order": template.SortOrder,
		"created_at": template.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": template.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if template.Description != nil {
		response["description"] = *template.Description
	} else {
		response["description"] = nil
	}

	middleware.Success(c, "success", response)
}

// SetDefaultTemplate set default template
func (h *PromptTemplateHandler) SetDefaultTemplate(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	templateIDStr := c.Param("id")

	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var template models.UserPromptTemplate
	if err := h.db.Where("id = ? AND user_id = ?", templateID, userID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到模板"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 先取消其他默认模板
	h.db.Model(&models.UserPromptTemplate{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false)

	// 设置当前模板为默认
	template.IsDefault = true
	if err := h.db.Save(&template).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// GetPromptTemplateDetail get prompt template detail
func (h *PromptTemplateHandler) GetPromptTemplateDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	templateIDStr := c.Param("id")

	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var template models.UserPromptTemplate
	if err := h.db.Where("id = ? AND user_id = ?", templateID, userID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到对应模板"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var data interface{}
	json.Unmarshal([]byte(template.Data), &data)

	response := map[string]interface{}{
		"id":          template.ID,
		"user_id":    template.UserID,
		"name":       template.Name,
		"data":       data,
		"status":     template.Status,
		"is_default": template.IsDefault,
		"sort_order": template.SortOrder,
		"created_at": template.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": template.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if template.Description != nil {
		response["description"] = *template.Description
	} else {
		response["description"] = nil
	}

	middleware.Success(c, "success", response)
}

// DeletePromptTemplate delete prompt template
func (h *PromptTemplateHandler) DeletePromptTemplate(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	templateIDStr := c.Param("id")

	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var template models.UserPromptTemplate
	if err := h.db.Where("id = ? AND user_id = ?", templateID, userID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到模板，无法删除"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&template).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// SetupPromptTemplateRoutes setup prompt template routes
func SetupPromptTemplateRoutes(r *gin.Engine) {
	handler := NewPromptTemplateHandler()

	promptTemplateGroup := r.Group("/api/v1/prompt-template")
	promptTemplateGroup.Use(middleware.JWTAuth())
	{
		promptTemplateGroup.GET("/list", handler.GetPromptTemplateList)
		promptTemplateGroup.POST("/save", handler.SavePromptTemplate)
		promptTemplateGroup.PUT("/set-default/:id", handler.SetDefaultTemplate)
		promptTemplateGroup.GET("/:id", handler.GetPromptTemplateDetail)
		promptTemplateGroup.DELETE("/:id", handler.DeletePromptTemplate)
	}
}


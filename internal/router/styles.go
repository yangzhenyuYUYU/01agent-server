package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StylesHandler struct {
	markdownProcessor *tools.UnifiedMarkdownProcessor
}

// NewStylesHandler 创建样式处理器
func NewStylesHandler() *StylesHandler {
	return &StylesHandler{
		markdownProcessor: tools.NewUnifiedMarkdownProcessor(),
	}
}

// getThumbnailContent 获取缩略图预览内容
func getThumbnailContent() string {
	// 尝试从配置文件读取
	configPath := filepath.Join("configs", "docs", "theme_thumbnail_content.md")
	if content, err := os.ReadFile(configPath); err == nil {
		return string(content)
	}
	// 默认内容
	return "# 主题预览\n\n这是**主题效果**的展示。"
}

// parseTags 解析标签JSON字符串
func parseTags(tagsStr *string) []map[string]interface{} {
	if tagsStr == nil || *tagsStr == "" {
		return []map[string]interface{}{{"id": "user", "name": "用户模板"}}
	}

	var tags []map[string]interface{}
	if err := json.Unmarshal([]byte(*tagsStr), &tags); err != nil {
		return []map[string]interface{}{{"id": "user", "name": "用户模板"}}
	}

	if len(tags) == 0 {
		return []map[string]interface{}{{"id": "user", "name": "用户模板"}}
	}

	return tags
}

// GetAllThemes 获取所有主题列表
func (h *StylesHandler) GetAllThemes(c *gin.Context) {
	var req struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	// 限制最大页面大小
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 读取缩略图内容
	thumbnailContent := getThumbnailContent()

	// 构建查询
	query := repository.DB.Model(&models.PublicTemplate{}).
		Where("status = ? AND is_public = ?", models.TemplateStatusPublished, true)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		repository.Errorf("GetAllThemes: failed to count: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var templates []models.PublicTemplate
	if err := query.Order("is_featured DESC, sort_order DESC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&templates).Error; err != nil {
		repository.Errorf("GetAllThemes: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 处理模板列表
	themes := make([]gin.H, 0, len(templates))
	for _, template := range templates {
		// 生成缩略图HTML
		sectionHTML := ""
		if template.SectionHTML != nil && *template.SectionHTML != "" {
			sectionHTML = *template.SectionHTML
		} else {
			if processedHTML, err := h.markdownProcessor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
				sectionHTML = processedHTML
			}
		}

		// 解析标签
		var tags []map[string]interface{}
		if template.Tags != nil {
			tags = parseTags(template.Tags)
		}

		// 确定类型
		templateType := "unified"
		if template.TemplateType == models.TemplateTypeWechat {
			templateType = "wechat"
		}

		// 获取描述，如果没有则使用默认值
		description := tools.GetStringValue(template.Description)
		if description == "" {
			description = fmt.Sprintf("%s 主题", template.Name)
		}

		// 获取英文名称，如果没有则使用中文名称
		labelEn := tools.GetStringValue(template.NameEn)
		if labelEn == "" {
			labelEn = template.Name
		}

		themeInfo := gin.H{
			"id":            template.TemplateID,
			"label":         template.Name,
			"labelEn":       labelEn,
			"value":         template.TemplateID,
			"author":        template.Author,
			"type":          templateType,
			"previewUrl":    tools.GetStringValue(template.PreviewURL),
			"tags":          tags,
			"primary_color": template.PrimaryColor,
			"description":   description,
			"section_html":  sectionHTML,
		}

		themes = append(themes, themeInfo)
	}

	// 计算总页数
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	middleware.Success(c, "获取主题和装饰器列表成功", gin.H{
		"themes":            themes,
		"decorations":       []interface{}{},
		"total_themes":      total,
		"total_decorations": 0,
		"pagination": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"total":     total,
			"pages":     totalPages,
		},
	})
}

// GetThemeByName 根据主题ID获取主题配置
func (h *StylesHandler) GetThemeByName(c *gin.Context) {
	themeName := c.Param("theme_name")
	if themeName == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "主题ID不能为空"))
		return
	}

	// 按template_id查找
	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ? AND status = ? AND is_public = ?",
		themeName, models.TemplateStatusPublished, true).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("主题 '%s' 不存在", themeName)))
			return
		}
		repository.Errorf("GetThemeByName: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 生成缩略图HTML
	sectionHTML := ""
	if template.SectionHTML != nil && *template.SectionHTML != "" {
		sectionHTML = *template.SectionHTML
	} else {
		thumbnailContent := getThumbnailContent()
		if processedHTML, err := h.markdownProcessor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	// 解析标签
	var tags []map[string]interface{}
	if template.Tags != nil {
		tags = parseTags(template.Tags)
	}

	// 确定类型
	templateType := "unified"
	if template.TemplateType == models.TemplateTypeWechat {
		templateType = "wechat"
	}

	// 解析template_data
	var configData map[string]interface{}
	if template.TemplateData != nil && *template.TemplateData != "" {
		if err := json.Unmarshal([]byte(*template.TemplateData), &configData); err != nil {
			configData = make(map[string]interface{})
		}
	} else {
		configData = make(map[string]interface{})
	}

	// 获取描述，如果没有则使用默认值
	description := tools.GetStringValue(template.Description)
	if description == "" {
		description = fmt.Sprintf("%s 主题", template.Name)
	}

	// 获取英文名称，如果没有则使用中文名称
	labelEn := tools.GetStringValue(template.NameEn)
	if labelEn == "" {
		labelEn = template.Name
	}

	themeInfo := gin.H{
		"id":           template.TemplateID,
		"label":        template.Name,
		"labelEn":      labelEn,
		"value":        template.TemplateID,
		"author":       template.Author,
		"type":         templateType,
		"previewUrl":   tools.GetStringValue(template.PreviewURL),
		"tags":         tags,
		"section_html": sectionHTML,
		"description":  description,
		"config":       configData,
	}

	middleware.Success(c, "获取主题配置成功", themeInfo)
}

// GetUserThemes 获取用户自己的模板列表
func (h *StylesHandler) GetUserThemes(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	// 查询用户的模板
	var templates []models.UserTemplate
	if err := repository.DB.Where("user_id = ? AND status IN ?", userID, []models.TemplateStatus{
		models.TemplateStatusDraft,
		models.TemplateStatusPublished,
	}).
		Order("created_at DESC").
		Find(&templates).Error; err != nil {
		repository.Errorf("GetUserThemes: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 读取缩略图内容
	thumbnailContent := getThumbnailContent()

	// 获取用户信息
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		repository.Warnf("GetUserThemes: failed to get user: %v", err)
	}

	// 处理每个用户模板
	themes := make([]gin.H, 0, len(templates))
	for _, template := range templates {
		// 生成缩略图HTML
		sectionHTML := ""
		if template.SectionHTML != nil && *template.SectionHTML != "" {
			sectionHTML = *template.SectionHTML
		} else {
			if processedHTML, err := h.markdownProcessor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
				sectionHTML = processedHTML
			}
		}

		// 处理 tags 字段
		tags := parseTags(template.Tags)

		// 确定类型
		templateType := "custom"
		if template.TemplateType == models.TemplateTypeUnified {
			templateType = "unified"
		} else if template.TemplateType == models.TemplateTypeWechat {
			templateType = "wechat"
		}

		// 获取作者名称
		author := "用户"
		if user.Nickname != nil && *user.Nickname != "" {
			author = *user.Nickname
		} else if user.Username != nil && *user.Username != "" {
			author = *user.Username
		}

		// 获取描述，如果没有则使用默认值
		description := tools.GetStringValue(template.Description)
		if description == "" {
			description = fmt.Sprintf("%s 用户模板", template.Name)
		}

		themeInfo := gin.H{
			"id":             template.TemplateID,
			"label":          template.Name,
			"labelEn":        template.Name, // 用户模板暂时使用相同的名称
			"value":          template.TemplateID,
			"author":         author,
			"type":           templateType,
			"previewUrl":     tools.GetStringValue(template.PreviewURL),
			"tags":           tags,
			"primary_color":  template.PrimaryColor,
			"description":    description,
			"section_html":   sectionHTML,
			"status":         int(template.Status),
			"visibility":     int(template.Visibility),
			"created_at":     template.CreatedAt.Format(time.RFC3339),
			"updated_at":     template.UpdatedAt.Format(time.RFC3339),
			"use_count":      template.UseCount,
			"like_count":     template.LikeCount,
			"download_count": template.DownloadCount,
		}

		themes = append(themes, themeInfo)
	}

	middleware.Success(c, "获取用户模板列表成功", gin.H{
		"themes":       themes,
		"total_themes": len(themes),
	})
}

// AddOfficialThemeToUser 将官方模板添加到个人模板
func (h *StylesHandler) AddOfficialThemeToUser(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req struct {
		ThemeID    string `json:"theme_id" binding:"required"`
		CustomName string `json:"custom_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 获取官方模板
	var officialTemplate models.PublicTemplate
	if err := repository.DB.Where("template_id = ? AND status = ? AND is_public = ?",
		req.ThemeID, models.TemplateStatusPublished, true).
		First(&officialTemplate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "指定的官方模板不存在或不可用"))
			return
		}
		repository.Errorf("AddOfficialThemeToUser: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 检查用户是否已经添加过这个模板
	var existingTemplate models.UserTemplate
	if err := repository.DB.Where("user_id = ? AND base_template_id = ?", userID, officialTemplate.TemplateID).
		First(&existingTemplate).Error; err == nil {
		// 已经存在
		middleware.Success(c, "您已经添加过这个官方模板了", gin.H{})
		return
	}

	// 创建用户模板
	userTemplateName := req.CustomName
	if userTemplateName == "" {
		userTemplateName = fmt.Sprintf("%s (我的副本)", officialTemplate.Name)
	}

	// 生成新的template_id
	newTemplateID := fmt.Sprintf("user_%s_%d", userID[:8], time.Now().Unix())

	userTemplate := models.UserTemplate{
		TemplateID:     newTemplateID,
		UserID:         userID,
		Name:           userTemplateName,
		Description:    officialTemplate.Description,
		TemplateType:   officialTemplate.TemplateType,
		Status:         models.TemplateStatusPublished, // 直接设为已发布状态
		Visibility:     models.VisibilityTypePrivate,   // 默认私有
		TemplateData:   officialTemplate.TemplateData,
		BaseTemplateID: &officialTemplate.TemplateID,
		PreviewURL:     officialTemplate.PreviewURL,
		ThumbnailURL:   officialTemplate.ThumbnailURL,
		SectionHTML:    officialTemplate.SectionHTML,
		PrimaryColor:   officialTemplate.PrimaryColor,
		Tags:           officialTemplate.Tags,
		Category:       officialTemplate.Category,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := repository.DB.Create(&userTemplate).Error; err != nil {
		repository.Errorf("AddOfficialThemeToUser: failed to create: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "添加模板失败"))
		return
	}

	middleware.Success(c, "成功添加官方模板到个人模板", gin.H{
		"id":                 userTemplate.TemplateID,
		"name":               userTemplate.Name,
		"description":        tools.GetStringValue(userTemplate.Description),
		"base_template_id":   officialTemplate.TemplateID,
		"base_template_name": officialTemplate.Name,
		"created_at":         userTemplate.CreatedAt.Format(time.RFC3339),
	})
}

// UpdateUserTheme 修改用户模板信息
func (h *StylesHandler) UpdateUserTheme(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	themeID := c.Param("theme_id")
	if themeID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var req struct {
		Name        *string             `json:"name"`
		Description *string             `json:"description"`
		Visibility  *int                `json:"visibility"`
		Tags        []map[string]string `json:"tags"`
		Category    *string             `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 查询用户的模板
	var template models.UserTemplate
	if err := repository.DB.Where("template_id = ? AND user_id = ?", themeID, userID).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户模板不存在或无权限修改"))
			return
		}
		repository.Errorf("UpdateUserTheme: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	hasUpdate := false

	if req.Name != nil {
		name := *req.Name
		if name == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "模板名称不能为空"))
			return
		}
		updates["name"] = name
		hasUpdate = true
	}

	if req.Description != nil {
		updates["description"] = *req.Description
		hasUpdate = true
	}

	if req.Visibility != nil {
		visibility := models.VisibilityType(*req.Visibility)
		if visibility < models.VisibilityTypePrivate || visibility > models.VisibilityTypeShared {
			middleware.HandleError(c, middleware.NewBusinessError(400, "无效的可见性设置，必须是0（私有）、1（公开）或2（分享）"))
			return
		}
		updates["visibility"] = visibility
		hasUpdate = true
	}

	if req.Tags != nil {
		// 将tags转换为JSON字符串
		if tagsJSON, err := json.Marshal(req.Tags); err == nil {
			tagsStr := string(tagsJSON)
			updates["tags"] = &tagsStr
			hasUpdate = true
		}
	}

	if req.Category != nil {
		updates["category"] = *req.Category
		hasUpdate = true
	}

	if !hasUpdate {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	updates["updated_at"] = time.Now()

	// 执行更新
	if err := repository.DB.Model(&template).Updates(updates).Error; err != nil {
		repository.Errorf("UpdateUserTheme: failed to update: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败"))
		return
	}

	// 重新查询获取最新数据
	repository.DB.Where("template_id = ?", themeID).First(&template)

	middleware.Success(c, "模板信息更新成功", gin.H{
		"id":          template.TemplateID,
		"name":        template.Name,
		"description": tools.GetStringValue(template.Description),
		"visibility":  int(template.Visibility),
		"tags":        parseTags(template.Tags),
		"category":    tools.GetStringValue(template.Category),
		"updated_at":  template.UpdatedAt.Format(time.RFC3339),
	})
}

// DeleteUserTheme 删除用户模板
func (h *StylesHandler) DeleteUserTheme(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	themeID := c.Param("theme_id")
	if themeID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	// 查询用户的模板
	var template models.UserTemplate
	if err := repository.DB.Where("template_id = ? AND user_id = ?", themeID, userID).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户模板不存在或无权限删除"))
			return
		}
		repository.Errorf("DeleteUserTheme: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 删除模板
	if err := repository.DB.Delete(&template).Error; err != nil {
		repository.Errorf("DeleteUserTheme: failed to delete: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败"))
		return
	}

	middleware.Success(c, "模板删除成功", gin.H{
		"id":         themeID,
		"deleted_at": time.Now().Format(time.RFC3339),
	})
}

// GetAllTags 获取所有主题标签
func (h *StylesHandler) GetAllTags(c *gin.Context) {
	// 默认标签列表
	allTags := []map[string]string{
		{"id": "tag1", "name": "通用"},
		{"id": "tag2", "name": "时尚"},
		{"id": "tag3", "name": "科技"},
		{"id": "tag4", "name": "旅行"},
		{"id": "tag5", "name": "商务"},
		{"id": "tag6", "name": "复古"},
		{"id": "tag7", "name": "休闲"},
		{"id": "tag8", "name": "未来"},
		{"id": "tag9", "name": "暗黑"},
		{"id": "tag10", "name": "艺术"},
		{"id": "tag11", "name": "自然"},
		{"id": "tag12", "name": "文化"},
	}

	middleware.Success(c, "获取标签列表成功", gin.H{
		"tags":  allTags,
		"total": len(allTags),
	})
}

// SetupStylesRoutes 设置样式路由
func SetupStylesRoutes(r *gin.Engine) {
	handler := NewStylesHandler()

	styles := r.Group("/api/v1/styles")
	{
		// 公开接口
		styles.GET("/themes", handler.GetAllThemes)               // 获取所有主题列表
		styles.GET("/themes/:theme_name", handler.GetThemeByName) // 根据主题ID获取主题配置
		styles.GET("/tags", handler.GetAllTags)                   // 获取所有主题标签

		// 需要认证的接口
		stylesAuth := styles.Group("")
		stylesAuth.Use(middleware.JWTAuth())
		{
			stylesAuth.GET("/user-themes", handler.GetUserThemes)                        // 获取用户自己的模板列表
			stylesAuth.POST("/user-themes/add-official", handler.AddOfficialThemeToUser) // 将官方模板添加到个人模板
			stylesAuth.PUT("/user-themes/:theme_id", handler.UpdateUserTheme)            // 修改用户模板信息
			stylesAuth.DELETE("/user-themes/:theme_id", handler.DeleteUserTheme)         // 删除用户模板
		}
	}
}

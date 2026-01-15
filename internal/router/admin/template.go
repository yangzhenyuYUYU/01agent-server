package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *AdminHandler) GetTemplateList(c *gin.Context) {

}

func (h *AdminHandler) GetTemplateDetail(c *gin.Context) {

}

func (h *AdminHandler) CreateTemplate(c *gin.Context) {
}

// CreatePublicTemplate 创建官方模板（管理员接口）
func (h *AdminHandler) CreatePublicTemplate(c *gin.Context) {
	var req struct {
		TemplateID    string                 `json:"template_id" binding:"required"`
		Name          string                 `json:"name" binding:"required"`
		NameEn        *string                `json:"name_en"`
		Description   *string                `json:"description"`
		Author        string                 `json:"author"`
		TemplateType  *int                   `json:"template_type"`
		Status        *int                   `json:"status"`
		PriceType     *int                   `json:"price_type"`
		Price         *float64               `json:"price"`
		OriginalPrice *float64               `json:"original_price"`
		IsPublic      *bool                  `json:"is_public"`
		IsFeatured    *bool                  `json:"is_featured"`
		IsOfficial    *bool                  `json:"is_official"`
		PreviewURL    *string                `json:"preview_url"`
		ThumbnailURL  *string                `json:"thumbnail_url"`
		PrimaryColor  *string                `json:"primary_color"`
		Tags          interface{}            `json:"tags"`
		Category      *string                `json:"category"`
		SortOrder     *int                   `json:"sort_order"`
		TemplateData  map[string]interface{} `json:"template_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 验证必填字段
	templateID := strings.TrimSpace(req.TemplateID)
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板名称不能为空"))
		return
	}

	// 检查模板ID是否已存在
	var existingTemplate models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&existingTemplate).Error; err == nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, fmt.Sprintf("模板ID '%s' 已存在", templateID)))
		return
	} else if err != gorm.ErrRecordNotFound {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 设置默认值
	author := strings.TrimSpace(req.Author)
	if author == "" {
		author = "system"
	}

	templateType := models.TemplateTypeUnified
	if req.TemplateType != nil {
		templateType = models.TemplateType(*req.TemplateType)
	}

	status := models.TemplateStatusDraft
	if req.Status != nil {
		status = models.TemplateStatus(*req.Status)
	}

	priceType := models.PriceTypeFree
	if req.PriceType != nil {
		priceType = models.PriceType(*req.PriceType)
	}

	price := 0.0
	if req.Price != nil {
		price = *req.Price
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	isOfficial := true
	if req.IsOfficial != nil {
		isOfficial = *req.IsOfficial
	}

	primaryColor := "#000000"
	if req.PrimaryColor != nil && strings.TrimSpace(*req.PrimaryColor) != "" {
		primaryColor = strings.TrimSpace(*req.PrimaryColor)
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// 构建模板对象
	template := models.PublicTemplate{
		TemplateID:   templateID,
		Name:         name,
		NameEn:       req.NameEn,
		Description:  req.Description,
		Author:       author,
		TemplateType: templateType,
		Status:       status,
		PriceType:    priceType,
		Price:        price,
		IsPublic:     isPublic,
		IsFeatured:   isFeatured,
		IsOfficial:   isOfficial,
		PreviewURL:   req.PreviewURL,
		ThumbnailURL: req.ThumbnailURL,
		PrimaryColor: primaryColor,
		Category:     req.Category,
		SortOrder:    sortOrder,
	}

	// 设置原价
	if req.OriginalPrice != nil && *req.OriginalPrice > 0 {
		template.OriginalPrice = req.OriginalPrice
	}

	// 处理tags
	if req.Tags != nil {
		if tagsJSON, err := json.Marshal(req.Tags); err == nil {
			tagsStr := string(tagsJSON)
			template.Tags = &tagsStr
		}
	}

	// 处理template_data
	if req.TemplateData != nil {
		if templateDataJSON, err := json.Marshal(req.TemplateData); err == nil {
			templateDataStr := string(templateDataJSON)
			template.TemplateData = &templateDataStr
		}
	}

	// 如果状态为已发布，设置发布时间
	if status == models.TemplateStatusPublished {
		now := time.Now()
		template.PublishedAt = &now
	}

	// 创建模板
	if err := repository.DB.Create(&template).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	// 生成缩略图HTML
	sectionHTML := ""
	thumbnailContent := tools.GetThumbnailContent()
	processor := tools.NewUnifiedMarkdownProcessor()
	if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
		sectionHTML = processedHTML
	}

	nameEn := tools.GetStringValue(template.NameEn)
	if nameEn == "" {
		nameEn = template.Name
	}

	middleware.Success(c, "模板创建成功", gin.H{
		"id":             template.TemplateID,
		"template_id":    template.TemplateID,
		"name":           template.Name,
		"name_en":        nameEn,
		"description":    tools.GetStringValue(template.Description),
		"author":         template.Author,
		"template_type":  template.TemplateType,
		"status":         template.Status,
		"price_type":     template.PriceType,
		"price":          template.Price,
		"original_price": template.OriginalPrice,
		"is_public":      template.IsPublic,
		"is_featured":    template.IsFeatured,
		"is_official":    template.IsOfficial,
		"preview_url":    tools.GetStringValue(template.PreviewURL),
		"thumbnail_url":  tools.GetStringValue(template.ThumbnailURL),
		"primary_color":  template.PrimaryColor,
		"tags":           template.Tags,
		"category":       tools.GetStringValue(template.Category),
		"sort_order":     template.SortOrder,
		"section_html":   sectionHTML,
		"created_at":     template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":     template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetPublicTemplates 获取公开模板列表（管理员接口，包含所有状态）
func (h *AdminHandler) GetPublicTemplates(c *gin.Context) {
	var req struct {
		Page         int    `form:"page" binding:"min=1"`
		PageSize     int    `form:"page_size" binding:"min=1"`
		Search       string `form:"search"`
		Category     string `form:"category"`
		TemplateType *int   `form:"template_type"`
		Status       *int   `form:"status"`
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
		req.PageSize = 50
	}

	// 构建查询 - 管理员接口，不限制状态
	query := repository.DB.Model(&models.PublicTemplate{})

	// 搜索
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ?", searchPattern)
	}

	// 分类筛选
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	// 类型筛选
	if req.TemplateType != nil {
		query = query.Where("template_type = ?", *req.TemplateType)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询，按创建时间降序排序
	offset := (req.Page - 1) * req.PageSize
	var templates []models.PublicTemplate
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&templates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 获取缩略图内容
	thumbnailContent := tools.GetThumbnailContent()
	processor := tools.NewUnifiedMarkdownProcessor()

	// 构建返回数据
	result := make([]gin.H, 0, len(templates))
	for _, template := range templates {
		// 生成缩略图HTML（数据库未存储时动态生成）
		sectionHTML := tools.GetStringValue(template.SectionHTML)
		if sectionHTML == "" {
			// 使用UnifiedMarkdownProcessor处理markdown
			if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
				sectionHTML = processedHTML
			}
		}

		price := 0.0
		if template.Price > 0 {
			price = template.Price
		}

		var originalPrice *float64
		if template.OriginalPrice != nil && *template.OriginalPrice > 0 {
			originalPrice = template.OriginalPrice
		}

		templateType := "unified"
		if template.TemplateType == models.TemplateTypeWechat {
			templateType = "wechat"
		}

		nameEn := tools.GetStringValue(template.NameEn)
		if nameEn == "" {
			nameEn = template.Name
		}

		description := tools.GetStringValue(template.Description)
		if description == "" {
			description = fmt.Sprintf("%s 主题", template.Name)
		}

		result = append(result, gin.H{
			"id":             template.TemplateID,
			"template_id":    template.TemplateID,
			"label":          template.Name,
			"labelEn":        nameEn,
			"value":          template.TemplateID,
			"name":           template.Name,
			"name_en":        nameEn,
			"description":    description,
			"author":         template.Author,
			"template_type":  template.TemplateType,
			"type":           templateType,
			"status":         template.Status,
			"price_type":     template.PriceType,
			"price":          price,
			"original_price": originalPrice,
			"is_public":      template.IsPublic,
			"is_featured":    template.IsFeatured,
			"is_official":    template.IsOfficial,
			"preview_url":    tools.GetStringValue(template.PreviewURL),
			"previewUrl":     tools.GetStringValue(template.PreviewURL),
			"thumbnail_url":  tools.GetStringValue(template.ThumbnailURL),
			"primary_color":  template.PrimaryColor,
			"tags":           template.Tags,
			"category":       tools.GetStringValue(template.Category),
			"download_count": template.DownloadCount,
			"use_count":      template.UseCount,
			"like_count":     template.LikeCount,
			"view_count":     template.ViewCount,
			"sort_order":     template.SortOrder,
			"section_html":   sectionHTML,
			"created_at":     template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})

		if template.PublishedAt != nil {
			result[len(result)-1]["published_at"] = template.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		} else {
			result[len(result)-1]["published_at"] = nil
		}
	}

	middleware.Success(c, "获取成功", gin.H{
		"templates": result,
		"pagination": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"total":     total,
			"pages":     (int(total) + req.PageSize - 1) / req.PageSize,
		},
	})
}

// GetPublicTemplateDetail 获取公开模板详情（管理员接口，包含完整配置数据）
func (h *AdminHandler) GetPublicTemplateDetail(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("模板 '%s' 不存在", templateID)))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 生成缩略图HTML
	sectionHTML := tools.GetStringValue(template.SectionHTML)
	if sectionHTML == "" {
		thumbnailContent := tools.GetThumbnailContent()
		processor := tools.NewUnifiedMarkdownProcessor()
		if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	price := 0.0
	if template.Price > 0 {
		price = template.Price
	}

	var originalPrice *float64
	if template.OriginalPrice != nil && *template.OriginalPrice > 0 {
		originalPrice = template.OriginalPrice
	}

	templateType := "unified"
	if template.TemplateType == models.TemplateTypeWechat {
		templateType = "wechat"
	}

	nameEn := tools.GetStringValue(template.NameEn)
	if nameEn == "" {
		nameEn = template.Name
	}

	description := tools.GetStringValue(template.Description)
	if description == "" {
		description = fmt.Sprintf("%s 主题", template.Name)
	}

	result := gin.H{
		"id":             template.TemplateID,
		"template_id":    template.TemplateID,
		"label":          template.Name,
		"labelEn":        nameEn,
		"value":          template.TemplateID,
		"name":           template.Name,
		"name_en":        nameEn,
		"description":    description,
		"author":         template.Author,
		"template_type":  template.TemplateType,
		"type":           templateType,
		"status":         template.Status,
		"price_type":     template.PriceType,
		"price":          price,
		"original_price": originalPrice,
		"is_public":      template.IsPublic,
		"is_featured":    template.IsFeatured,
		"is_official":    template.IsOfficial,
		"preview_url":    tools.GetStringValue(template.PreviewURL),
		"previewUrl":     tools.GetStringValue(template.PreviewURL),
		"thumbnail_url":  tools.GetStringValue(template.ThumbnailURL),
		"primary_color":  template.PrimaryColor,
		"tags":           template.Tags,
		"category":       tools.GetStringValue(template.Category),
		"download_count": template.DownloadCount,
		"use_count":      template.UseCount,
		"like_count":     template.LikeCount,
		"view_count":     template.ViewCount,
		"sort_order":     template.SortOrder,
		"section_html":   sectionHTML,
		"config":         template.TemplateData,
		"template_data":  template.TemplateData,
		"created_at":     template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":     template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if template.PublishedAt != nil {
		result["published_at"] = template.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		result["published_at"] = nil
	}

	middleware.Success(c, "获取成功", result)
}

// GetThemePreview 获取主题预览HTML
func (h *AdminHandler) GetThemePreview(c *gin.Context) {
	themeName := c.Param("theme_name")
	if themeName == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "主题名称不能为空"))
		return
	}

	// 查找模板（theme_name可能是template_id或name_en）
	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ? OR name_en = ?", themeName, themeName).
		Where("status = ? AND is_public = ?", models.TemplateStatusPublished, true).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "主题不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 生成缩略图HTML
	sectionHTML := tools.GetStringValue(template.SectionHTML)
	if sectionHTML == "" {
		thumbnailContent := tools.GetThumbnailContent()
		processor := tools.NewUnifiedMarkdownProcessor()
		if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	// 构建完整预览HTML
	fullPreview := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 预览</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background-color: #f5f5f5;
        }
        .preview-container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border-radius: 8px;
        }
    </style>
</head>
<body>
    <div class="preview-container">
        %s
    </div>
</body>
</html>`, template.Name, sectionHTML)

	templateType := "unified"
	if template.TemplateType == models.TemplateTypeWechat {
		templateType = "wechat"
	}

	result := gin.H{
		"theme_name":   template.TemplateID,
		"chinese_name": template.Name,
		"type":         templateType,
		"section_html": sectionHTML,
		"full_preview": fullPreview,
	}

	// 如果有template_data，也返回配置信息
	if template.TemplateData != nil {
		result["meta"] = gin.H{} // 可以从template_data中提取meta信息
		result["config"] = template.TemplateData
	}

	middleware.Success(c, "获取成功", result)
}

// UpdatePublicTemplate 更新官方模板
func (h *AdminHandler) UpdatePublicTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("模板 '%s' 不存在", templateID)))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var req struct {
		Name          *string                `json:"name"`
		NameEn        *string                `json:"name_en"`
		Description   *string                `json:"description"`
		Author        *string                `json:"author"`
		TemplateType  *int                   `json:"template_type"`
		Status        *int                   `json:"status"`
		PriceType     *int                   `json:"price_type"`
		Price         *float64               `json:"price"`
		OriginalPrice *float64               `json:"original_price"`
		IsPublic      *bool                  `json:"is_public"`
		IsFeatured    *bool                  `json:"is_featured"`
		IsOfficial    *bool                  `json:"is_official"`
		PreviewURL    *string                `json:"preview_url"`
		ThumbnailURL  *string                `json:"thumbnail_url"`
		PrimaryColor  *string                `json:"primary_color"`
		Tags          interface{}            `json:"tags"`
		Category      *string                `json:"category"`
		SortOrder     *int                   `json:"sort_order"`
		TemplateData  map[string]interface{} `json:"template_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	hasUpdate := false

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "模板名称不能为空"))
			return
		}
		updates["name"] = name
		hasUpdate = true
	}

	if req.NameEn != nil {
		nameEn := strings.TrimSpace(*req.NameEn)
		if nameEn == "" {
			updates["name_en"] = nil
		} else {
			updates["name_en"] = nameEn
		}
		hasUpdate = true
	}

	if req.Description != nil {
		updates["description"] = *req.Description
		hasUpdate = true
	}

	if req.Author != nil {
		updates["author"] = *req.Author
		hasUpdate = true
	}

	if req.TemplateType != nil {
		updates["template_type"] = models.TemplateType(*req.TemplateType)
		hasUpdate = true
	}

	if req.Status != nil {
		newStatus := models.TemplateStatus(*req.Status)
		updates["status"] = newStatus
		// 如果状态改为已发布且之前没有发布时间，设置发布时间
		if newStatus == models.TemplateStatusPublished && template.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = &now
		}
		hasUpdate = true
	}

	if req.PriceType != nil {
		updates["price_type"] = models.PriceType(*req.PriceType)
		hasUpdate = true
	}

	if req.Price != nil {
		updates["price"] = *req.Price
		hasUpdate = true
	}

	if req.OriginalPrice != nil {
		if *req.OriginalPrice > 0 {
			updates["original_price"] = *req.OriginalPrice
		} else {
			updates["original_price"] = nil
		}
		hasUpdate = true
	}

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
		hasUpdate = true
	}

	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
		hasUpdate = true
	}

	if req.IsOfficial != nil {
		updates["is_official"] = *req.IsOfficial
		hasUpdate = true
	}

	if req.PreviewURL != nil {
		updates["preview_url"] = *req.PreviewURL
		hasUpdate = true
	}

	if req.ThumbnailURL != nil {
		updates["thumbnail_url"] = *req.ThumbnailURL
		hasUpdate = true
	}

	if req.PrimaryColor != nil {
		updates["primary_color"] = *req.PrimaryColor
		hasUpdate = true
	}

	if req.Tags != nil {
		// 将tags转换为JSON字符串
		if tagsJSON, err := json.Marshal(req.Tags); err == nil {
			tagsStr := string(tagsJSON)
			updates["tags"] = &tagsStr
		}
		hasUpdate = true
	}

	if req.Category != nil {
		updates["category"] = *req.Category
		hasUpdate = true
	}

	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
		hasUpdate = true
	}

	if req.TemplateData != nil {
		// 将template_data转换为JSON字符串
		if templateDataJSON, err := json.Marshal(req.TemplateData); err == nil {
			templateDataStr := string(templateDataJSON)
			updates["template_data"] = &templateDataStr
			// 如果更新了模板数据，清除section_html以便重新生成
			updates["section_html"] = nil
		}
		hasUpdate = true
	}

	if !hasUpdate {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	// 执行更新 - 使用 Select 确保所有字段都能更新
	updateFields := make([]string, 0, len(updates))
	for k := range updates {
		updateFields = append(updateFields, k)
	}

	if err := repository.DB.Model(&template).Select(updateFields).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 重新查询以获取最新数据
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询更新后的数据失败: "+err.Error()))
		return
	}

	// 生成新的section_html
	sectionHTML := tools.GetStringValue(template.SectionHTML)
	if sectionHTML == "" {
		thumbnailContent := tools.GetThumbnailContent()
		processor := tools.NewUnifiedMarkdownProcessor()
		if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	middleware.Success(c, "模板更新成功", gin.H{
		"id":           template.TemplateID,
		"template_id":  template.TemplateID,
		"name":         template.Name,
		"name_en":      tools.GetStringValue(template.NameEn),
		"description":  tools.GetStringValue(template.Description),
		"status":       template.Status,
		"is_public":    template.IsPublic,
		"is_featured":  template.IsFeatured,
		"sort_order":   template.SortOrder,
		"section_html": sectionHTML,
		"updated_at":   template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeletePublicTemplate 删除官方模板（硬删除）
func (h *AdminHandler) DeletePublicTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("模板 '%s' 不存在", templateID)))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	templateName := template.Name

	// 物理删除
	if err := repository.DB.Delete(&template).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "模板已删除", gin.H{
		"id":          templateID,
		"template_id": templateID,
		"name":        templateName,
		"deleted_at":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
	})
}

package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/models/short_post"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetImageExampleList 获取图文生成示例列表
func (h *AdminHandler) GetImageExampleList(c *gin.Context) {
	var req struct {
		Page      int                    `json:"page" binding:"min=1"`
		PageSize  int                    `json:"page_size" binding:"min=1"`
		Name      string                 `json:"name"`
		Tags      []string               `json:"tags"`
		ExtraData map[string]interface{} `json:"extra_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
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

	// 构建查询
	query := repository.DB.Model(&models.ImageExample{})

	// 名称搜索
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// Tags搜索（如果tags是JSON数组，需要特殊处理）
	// 这里简化处理，实际可能需要根据数据库类型调整

	// ExtraData搜索
	if req.ExtraData != nil {
		// 特殊处理：如果size是"1080×自适应"且width != height
		if size, ok := req.ExtraData["size"].(string); ok && size == "1080×自适应" {
			if width, wOk := req.ExtraData["width"]; wOk {
				if height, hOk := req.ExtraData["height"]; hOk && width != height {
					// 查询width=1080的记录
					query = query.Where("extra_data LIKE ?", "%\"width\":1080%")
				}
			}
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var examples []models.ImageExample
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&examples).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(examples))
	for _, item := range examples {
		images := []string{}
		// 如果是小红书类型，获取详情中的图片
		if item.ProjectType == "xiaohongshu" {
			var detail models.ImageExampleDetail
			if err := repository.DB.Where("example_id = ?", item.ID).First(&detail).Error; err == nil {
				if detail.Images != nil {
					// 解析JSON数组
					// 这里简化处理，实际需要解析JSON
				}
			}
		}

		result = append(result, gin.H{
			"id":           item.ID,
			"name":         item.Name,
			"prompt":       item.Prompt,
			"cover_url":    tools.GetStringValue(item.CoverURL),
			"tags":         item.Tags,
			"sort_order":   item.SortOrder,
			"is_visible":   item.IsVisible,
			"extra_data":   item.ExtraData,
			"images":       images,
			"project_type": item.ProjectType,
			"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items":     result,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// SaveImageExample 创建或更新图文生成示例
func (h *AdminHandler) SaveImageExample(c *gin.Context) {
	var req struct {
		ID          *int                   `json:"id"`
		Name        *string                `json:"name"`
		Prompt      *string                `json:"prompt"`
		CoverURL    *string                `json:"cover_url"`
		JsxCode     *string                `json:"jsx_code"`
		Tags        interface{}            `json:"tags"` // 可能是数组或JSON字符串
		ExtraData   map[string]interface{} `json:"extra_data"`
		NodeData    map[string]interface{} `json:"node_data"`
		Images      interface{}            `json:"images"` // 可能是数组或JSON字符串
		ProjectType *string                `json:"project_type"`
		SortOrder   *int                   `json:"sort_order"`
		IsVisible   *bool                  `json:"is_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	var example models.ImageExample
	var isNew bool

	if req.ID != nil && *req.ID > 0 {
		// 更新现有记录
		if err := repository.DB.Where("id = ?", *req.ID).First(&example).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(404, "示例不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}
		isNew = false
	} else {
		// 创建新记录
		if req.Name == nil || *req.Name == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "名称不能为空"))
			return
		}
		if req.Prompt == nil || *req.Prompt == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "提示词不能为空"))
			return
		}
		example = models.ImageExample{
			Name:        *req.Name,
			Prompt:      *req.Prompt,
			ProjectType: "other",
			IsVisible:   true,
			SortOrder:   0,
		}
		isNew = true
	}

	// 更新字段
	if req.Name != nil {
		example.Name = *req.Name
	}
	if req.Prompt != nil {
		example.Prompt = *req.Prompt
	}
	if req.CoverURL != nil {
		example.CoverURL = req.CoverURL
	}
	if req.SortOrder != nil {
		example.SortOrder = *req.SortOrder
	}
	if req.IsVisible != nil {
		example.IsVisible = *req.IsVisible
	}
	if req.ProjectType != nil {
		// 只有明确传入 "xiaohongshu" 时才使用，否则使用 "other"
		if *req.ProjectType == "xiaohongshu" {
			example.ProjectType = "xiaohongshu"
		} else {
			example.ProjectType = "other"
		}
	}

	// 处理 Tags（转换为JSON字符串）
	if req.Tags != nil {
		tagsJSON, err := json.Marshal(req.Tags)
		if err == nil {
			tagsStr := string(tagsJSON)
			example.Tags = &tagsStr
		}
	}

	// 处理 ExtraData（转换为JSON字符串）
	if req.ExtraData != nil {
		extraDataJSON, err := json.Marshal(req.ExtraData)
		if err == nil {
			extraDataStr := string(extraDataJSON)
			example.ExtraData = &extraDataStr
		}
	}

	// 保存主表
	if isNew {
		if err := repository.DB.Create(&example).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
			return
		}
	} else {
		if err := repository.DB.Save(&example).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
			return
		}
	}

	// 处理详情表（存储大字段）
	var detail models.ImageExampleDetail
	if err := repository.DB.Where("example_id = ?", example.ID).First(&detail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新详情记录
			detail = models.ImageExampleDetail{
				ExampleID: example.ID,
			}
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询详情失败: "+err.Error()))
			return
		}
	}

	// 更新详情字段
	if req.JsxCode != nil {
		detail.JsxCode = req.JsxCode
	}
	if req.NodeData != nil {
		nodeDataJSON, err := json.Marshal(req.NodeData)
		if err == nil {
			nodeDataStr := string(nodeDataJSON)
			detail.NodeData = &nodeDataStr
		}
	}
	if req.Images != nil {
		imagesJSON, err := json.Marshal(req.Images)
		if err == nil {
			imagesStr := string(imagesJSON)
			detail.Images = &imagesStr
		}
	}

	// 保存详情表
	if detail.ID == 0 {
		if err := repository.DB.Create(&detail).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建详情失败: "+err.Error()))
			return
		}
	} else {
		if err := repository.DB.Save(&detail).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新详情失败: "+err.Error()))
			return
		}
	}

	// 构建返回数据
	result := gin.H{
		"id":           example.ID,
		"name":         example.Name,
		"prompt":       example.Prompt,
		"cover_url":    tools.GetStringValue(example.CoverURL),
		"jsx_code":     tools.GetStringValue(detail.JsxCode),
		"tags":         example.Tags,
		"node_data":    detail.NodeData,
		"images":       detail.Images,
		"project_type": example.ProjectType,
		"sort_order":   example.SortOrder,
		"is_visible":   example.IsVisible,
		"extra_data":   example.ExtraData,
		"created_at":   example.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   example.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	middleware.Success(c, "success", result)
}

// GetImageExampleDetail 获取图文生成示例详情
func (h *AdminHandler) GetImageExampleDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	// 查询主表记录
	var example models.ImageExample
	if err := repository.DB.Where("id = ?", id).First(&example).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "未找到示例"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 关联查询详情表获取大字段数据
	var detail models.ImageExampleDetail
	repository.DB.Where("example_id = ?", id).First(&detail)

	// 构建返回数据
	result := gin.H{
		"id":           example.ID,
		"name":         example.Name,
		"prompt":       example.Prompt,
		"cover_url":    example.CoverURL,
		"jsx_code":     nil,
		"tags":         example.Tags,
		"node_data":    nil,
		"images":       nil,
		"project_type": example.ProjectType,
		"sort_order":   example.SortOrder,
		"is_visible":   example.IsVisible,
		"extra_data":   example.ExtraData,
		"created_at":   example.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   example.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 如果有详情记录，添加大字段数据
	if detail.ID > 0 {
		result["jsx_code"] = tools.GetStringValue(detail.JsxCode)
		result["node_data"] = detail.NodeData
		result["images"] = detail.Images
	}

	middleware.Success(c, "success", result)
}

// DeleteImageExample 删除图文生成示例
func (h *AdminHandler) DeleteImageExample(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	// 检查记录是否存在
	var example models.ImageExample
	if err := repository.DB.Where("id = ?", id).First(&example).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "示例不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 删除详情记录
	if err := repository.DB.Where("example_id = ?", id).Delete(&models.ImageExampleDetail{}).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除详情失败: "+err.Error()))
		return
	}

	// 删除主表记录
	if err := repository.DB.Delete(&example).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "success", gin.H{})
}

// GetAllContentList 获取全部用户的内容列表
func (h *AdminHandler) GetAllContentList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"min=1"`
		PageSize int    `form:"page_size" binding:"min=1,max=100"`
		Category string `form:"category"` // long_post/xiaohongshu/short_post/poster/other
		Keyword  string `form:"keyword"`
		UserID   string `form:"user_id"`
		OrderBy  string `form:"order_by"`
		Order    string `form:"order"` // asc/desc
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
		req.PageSize = 20
	}
	if req.OrderBy == "" {
		req.OrderBy = "updated_at"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	items := []gin.H{}
	var total int64

	// 查询长图文（当category为空或为long_post时）
	if req.Category == "" || req.Category == "long_post" {
		query := repository.DB.Model(&models.ArticleEditTask{})

		if req.UserID != "" {
			query = query.Where("user_id = ?", req.UserID)
		}

		if req.Keyword != "" {
			keywordPattern := "%" + req.Keyword + "%"
			query = query.Where("title LIKE ?", keywordPattern)
		}

		if req.Category == "long_post" {
			// 只查询长图文
			orderField := req.OrderBy
			if req.Order == "desc" {
				orderField = orderField + " DESC"
			} else {
				orderField = orderField + " ASC"
			}
			query = query.Order(orderField)

			if err := query.Count(&total).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
				return
			}

			offset := (req.Page - 1) * req.PageSize
			var longPosts []models.ArticleEditTask
			if err := query.Offset(offset).Limit(req.PageSize).Find(&longPosts).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
				return
			}

			for _, p := range longPosts {
				// 获取用户信息
				var user models.User
				repository.DB.Where("user_id = ?", p.UserID).First(&user)

				content := ""
				if p.Content != "" {
					if len(p.Content) > 100 {
						content = p.Content[:100] + "..."
					} else {
						content = p.Content
					}
				}

				items = append(items, gin.H{
					"id":              p.ID,
					"article_task_id": p.ArticleTaskID,
					"user_id":         p.UserID,
					"user_name":       tools.GetStringValue(user.Nickname),
					"title":           p.Title,
					"theme":           p.Theme,
					"content":         content,
					"section_html":    tools.GetStringValue(p.SectionHTML),
					"status":          p.Status,
					"is_public":       p.IsPublic,
					"tags":            p.Tags,
					"category":        "long_post",
					"created_at":      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"published_at": func() *string {
						if p.PublishedAt != nil {
							formatted := p.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
							return &formatted
						}
						return nil
					}(),
				})
			}

			middleware.Success(c, "success", gin.H{
				"items":     items,
				"total":     total,
				"page":      req.Page,
				"page_size": req.PageSize,
			})
			return
		}
	}

	// 查询ShortPostProject（当category为xiaohongshu/short_post/poster/other时）
	categoryToProjectType := map[string]string{
		"xiaohongshu": "xiaohongshu",
		"short_post":  "short_post",
		"poster":      "poster",
		"other":       "other",
	}

	if projectType, ok := categoryToProjectType[req.Category]; ok {
		query := repository.DB.Model(&short_post.ShortPostProject{})

		if req.UserID != "" {
			query = query.Where("user_id = ?", req.UserID)
		}

		if req.Keyword != "" {
			keywordPattern := "%" + req.Keyword + "%"
			query = query.Where("name LIKE ?", keywordPattern)
		}

		query = query.Where("project_type = ?", projectType)

		// 排序
		orderField := req.OrderBy
		if req.Order == "desc" {
			orderField = orderField + " DESC"
		} else {
			orderField = orderField + " ASC"
		}
		query = query.Order(orderField)

		if err := query.Count(&total).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}

		offset := (req.Page - 1) * req.PageSize
		var shortPosts []short_post.ShortPostProject
		if err := query.Offset(offset).Limit(req.PageSize).Find(&shortPosts).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}

		for _, p := range shortPosts {
			// 获取用户信息
			var user models.User
			repository.DB.Where("user_id = ?", p.UserID).First(&user)

			items = append(items, gin.H{
				"id":           p.ID,
				"user_id":      p.UserID,
				"user_name":    tools.GetStringValue(user.Nickname),
				"name":         p.Name,
				"description":  tools.GetStringValue(p.Description),
				"cover_image":  tools.GetStringValue(p.CoverImage),
				"thumbnail":    tools.GetStringValue(p.Thumbnail),
				"project_type": string(p.ProjectType),
				"status":       string(p.Status),
				"frame_count":  p.FrameCount,
				"category":     req.Category,
				"created_at":   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"updated_at":   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"saved_at": func() *string {
					if p.SavedAt != nil {
						formatted := p.SavedAt.Format("2006-01-02T15:04:05Z07:00")
						return &formatted
					}
					return nil
				}(),
			})
		}

		middleware.Success(c, "success", gin.H{
			"items":     items,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		})
		return
	}

	// 如果category不匹配任何已知类型，返回空结果
	middleware.Success(c, "success", gin.H{
		"items":     []gin.H{},
		"total":     0,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

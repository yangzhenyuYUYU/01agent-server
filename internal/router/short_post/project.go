package short_post

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/models/short_post"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectHandler short post project handler
type ProjectHandler struct {
	db *gorm.DB
}

// NewProjectHandler create project handler
func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		db: repository.DB,
	}
}

// ========================= Request/Response Models =========================

// CreateProjectRequest create project request
type CreateProjectRequest struct {
	Name        string                 `json:"name" binding:"required,max=200"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProjectType short_post.ProjectType `json:"project_type"`
	Description *string                `json:"description"`
	ThreadID    *string                `json:"thread_id"`
}

// UpdateProjectRequest update project request
type UpdateProjectRequest struct {
	Name        *string                   `json:"name"`
	Description *string                   `json:"description"`
	CoverImage  *string                   `json:"cover_image"`
	Thumbnail   *string                   `json:"thumbnail"`
	ProjectType *short_post.ProjectType   `json:"project_type"`
	Metadata    map[string]interface{}    `json:"metadata"`
	Status      *short_post.ProjectStatus `json:"status"`
}

// SaveProjectContentRequest save project content request
type SaveProjectContentRequest struct {
	CanvasConfig  map[string]interface{} `json:"canvas_config"`
	FramesData    []interface{}          `json:"frames_data"`
	ElementsData  []interface{}          `json:"elements_data"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreateVersion bool                   `json:"create_version"`
}

// SaveCopywritingRequest save copywriting request
type SaveCopywritingRequest struct {
	Title   *string  `json:"title"`
	Content *string  `json:"content"`
	Topics  []string `json:"topics"`
	Images  []string `json:"images"`
}

// ContentCategory content category
type ContentCategory string

const (
	ContentCategoryLongPost    ContentCategory = "long_post"
	ContentCategoryXiaohongshu ContentCategory = "xiaohongshu"
	ContentCategoryShortPost   ContentCategory = "short_post"
	ContentCategoryPoster      ContentCategory = "poster"
	ContentCategoryOther       ContentCategory = "other"
)

// ========================= Project Handlers =========================

// CreateProject create project
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	// 设置默认值
	if req.ProjectType == "" {
		req.ProjectType = short_post.ProjectTypeXiaohongshu
	}

	// 序列化 metadata
	var metadataJSON *string
	if req.Metadata != nil {
		metadataBytes, _ := json.Marshal(req.Metadata)
		metadataStr := string(metadataBytes)
		metadataJSON = &metadataStr
	}

	// 创建工程
	project := &short_post.ShortPostProject{
		ID:          uuid.New().String(),
		UserID:      userID,
		ThreadID:    req.ThreadID,
		Name:        req.Name,
		Description: req.Description,
		ProjectType: req.ProjectType,
		Metadata:    metadataJSON,
		Status:      short_post.ProjectStatusDraft,
	}

	if err := h.db.Create(project).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建工程失败: %v", err)))
		return
	}

	// 同时创建一个空的内容记录
	content := &short_post.ShortPostProjectContent{
		ID:        uuid.New().String(),
		ProjectID: project.ID,
		Version:   1,
		IsLatest:  true,
	}

	if err := h.db.Create(content).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建内容记录失败: %v", err)))
		return
	}

	// 解析 metadata 返回
	var metadata interface{}
	if metadataJSON != nil {
		json.Unmarshal([]byte(*metadataJSON), &metadata)
	}

	middleware.Success(c, "创建成功", gin.H{
		"id":           project.ID,
		"name":         project.Name,
		"description":  project.Description,
		"status":       string(project.Status),
		"project_type": string(project.ProjectType),
		"metadata":     metadata,
		"thread_id":    project.ThreadID,
		"created_at":   project.CreatedAt.Format(time.RFC3339),
	})
}

// GetProjectList get project list
func (h *ProjectHandler) GetProjectList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	statusStr := c.Query("status")
	projectTypeStr := c.Query("project_type")
	keyword := c.Query("keyword")
	orderBy := c.DefaultQuery("order_by", "updated_at")
	order := c.DefaultQuery("order", "desc")

	query := h.db.Model(&short_post.ShortPostProject{}).Where("user_id = ?", userID)

	if statusStr != "" {
		status := short_post.ProjectStatus(statusStr)
		query = query.Where("status = ?", status)
	}

	if projectTypeStr != "" {
		projectType := short_post.ProjectType(projectTypeStr)
		query = query.Where("project_type = ?", projectType)
	}

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	// 处理排序
	if order == "desc" {
		query = query.Order(fmt.Sprintf("%s DESC", orderBy))
	} else {
		query = query.Order(fmt.Sprintf("%s ASC", orderBy))
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	var projects []short_post.ShortPostProject
	if err := query.Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(projects))
	for _, p := range projects {
		// 解析 metadata
		var metadata interface{}
		if p.Metadata != nil {
			json.Unmarshal([]byte(*p.Metadata), &metadata)
		}

		item := map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"status":       nil,
			"frame_count":  p.FrameCount,
			"thread_id":    p.ThreadID,
			"project_type": nil,
			"metadata":     metadata,
			"created_at":   p.CreatedAt.Format(time.RFC3339),
			"updated_at":   p.UpdatedAt.Format(time.RFC3339),
			"saved_at":     nil,
		}

		if p.Description != nil {
			item["description"] = *p.Description
		}
		if p.CoverImage != nil {
			item["cover_image"] = *p.CoverImage
		}
		if p.Thumbnail != nil {
			item["thumbnail"] = *p.Thumbnail
		}
		if p.Status != "" {
			item["status"] = string(p.Status)
		}
		if p.ProjectType != "" {
			item["project_type"] = string(p.ProjectType)
		}
		if p.SavedAt != nil {
			item["saved_at"] = p.SavedAt.Format(time.RFC3339)
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

// GetAllContentList get all user content list (admin interface)
func (h *ProjectHandler) GetAllContentList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	// 检查是否为管理员（Role = 3）
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil || user.Role != 3 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusForbidden, "需要管理员权限"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryStr := c.Query("category")
	keyword := c.Query("keyword")
	filterUserID := c.Query("user_id")
	orderBy := c.DefaultQuery("order_by", "updated_at")
	order := c.DefaultQuery("order", "desc")

	var category ContentCategory
	if categoryStr != "" {
		category = ContentCategory(categoryStr)
	}

	items := make([]map[string]interface{}, 0)
	var total int64

	// 映射 ContentCategory 到 ProjectType
	categoryToProjectType := map[ContentCategory]short_post.ProjectType{
		ContentCategoryXiaohongshu: short_post.ProjectTypeXiaohongshu,
		ContentCategoryShortPost:   short_post.ProjectTypeShortPost,
		ContentCategoryPoster:      short_post.ProjectTypePoster,
		ContentCategoryOther:       short_post.ProjectTypeOther,
	}

	// 查询长图文（仅当 category 为 "" 或 long_post 时）
	if category == "" || category == ContentCategoryLongPost {
		longPostQuery := h.db.Model(&models.ArticleEditTask{})

		if filterUserID != "" {
			longPostQuery = longPostQuery.Where("user_id = ?", filterUserID)
		}

		if keyword != "" {
			longPostQuery = longPostQuery.Where("title LIKE ?", "%"+keyword+"%")
		}

		if category == ContentCategoryLongPost {
			// 只查询长图文
			if order == "desc" {
				longPostQuery = longPostQuery.Order(fmt.Sprintf("%s DESC", orderBy))
			} else {
				longPostQuery = longPostQuery.Order(fmt.Sprintf("%s ASC", orderBy))
			}

			longPostQuery.Count(&total)

			offset := (page - 1) * pageSize
			var longPosts []models.ArticleEditTask
			if err := longPostQuery.Offset(offset).Limit(pageSize).Find(&longPosts).Error; err == nil {
				for _, p := range longPosts {
					// 解析 tags
					var tags []string
					if p.Tags != nil && *p.Tags != "" {
						json.Unmarshal([]byte(*p.Tags), &tags)
					}

					// 截断 content
					content := p.Content
					if len(content) > 100 {
						content = content[:100] + "..."
					}

					sectionHTML := ""
					if p.SectionHTML != nil {
						sectionHTML = *p.SectionHTML
					}

					item := map[string]interface{}{
						"id":              p.ID,
						"article_task_id": p.ArticleTaskID,
						"user_id":         p.UserID,
						"title":           p.Title,
						"theme":           p.Theme,
						"content":         content,
						"section_html":    sectionHTML,
						"status":          p.Status,
						"is_public":       p.IsPublic,
						"tags":            tags,
						"category":        "long_post",
						"created_at":      p.CreatedAt.Format(time.RFC3339),
						"updated_at":      p.UpdatedAt.Format(time.RFC3339),
					}

					if p.PublishedAt != nil {
						item["published_at"] = p.PublishedAt.Format(time.RFC3339)
					}

					// 查询用户信息
					var user models.User
					if err := h.db.Where("user_id = ?", p.UserID).First(&user).Error; err == nil {
						if user.Nickname != nil {
							item["user_name"] = *user.Nickname
						}
					}

					items = append(items, item)
				}
			}

			middleware.Success(c, "success", gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
			return
		}
	}

	// 查询 ShortPostProject（当 category 为 xiaohongshu/short_post/poster/other 时）
	if category != "" && category != ContentCategoryLongPost {
		shortPostQuery := h.db.Model(&short_post.ShortPostProject{})

		if filterUserID != "" {
			shortPostQuery = shortPostQuery.Where("user_id = ?", filterUserID)
		}

		if keyword != "" {
			shortPostQuery = shortPostQuery.Where("name LIKE ?", "%"+keyword+"%")
		}

		// 根据 category 筛选 project_type
		projectType := categoryToProjectType[category]
		if projectType != "" {
			shortPostQuery = shortPostQuery.Where("project_type = ?", projectType)
		}

		// 处理排序
		if order == "desc" {
			shortPostQuery = shortPostQuery.Order(fmt.Sprintf("%s DESC", orderBy))
		} else {
			shortPostQuery = shortPostQuery.Order(fmt.Sprintf("%s ASC", orderBy))
		}

		shortPostQuery.Count(&total)

		offset := (page - 1) * pageSize
		var shortPosts []short_post.ShortPostProject
		if err := shortPostQuery.Offset(offset).Limit(pageSize).Find(&shortPosts).Error; err == nil {
			for _, p := range shortPosts {
				// 解析 metadata
				var metadata interface{}
				if p.Metadata != nil {
					json.Unmarshal([]byte(*p.Metadata), &metadata)
				}

				item := map[string]interface{}{
					"id":          p.ID,
					"user_id":     p.UserID,
					"name":        p.Name,
					"status":      nil,
					"category":    nil,
					"frame_count": p.FrameCount,
					"metadata":    metadata,
					"created_at":  p.CreatedAt.Format(time.RFC3339),
					"updated_at":  p.UpdatedAt.Format(time.RFC3339),
				}

				if p.Description != nil {
					item["description"] = *p.Description
				}
				if p.CoverImage != nil {
					item["cover_image"] = *p.CoverImage
				}
				if p.Thumbnail != nil {
					item["thumbnail"] = *p.Thumbnail
				}
				if p.Status != "" {
					item["status"] = string(p.Status)
				}
				if p.ProjectType != "" {
					item["category"] = string(p.ProjectType)
				}

				// 查询用户信息
				var user models.User
				if err := h.db.Where("user_id = ?", p.UserID).First(&user).Error; err == nil {
					if user.Nickname != nil {
						item["user_name"] = *user.Nickname
					}
				}

				items = append(items, item)
			}
		}

		middleware.Success(c, "success", gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
		return
	}

	// 如果没有指定类别（category is ""），合并查询所有数据
	longPostQuery := h.db.Model(&models.ArticleEditTask{})
	if filterUserID != "" {
		longPostQuery = longPostQuery.Where("user_id = ?", filterUserID)
	}
	if keyword != "" {
		longPostQuery = longPostQuery.Where("title LIKE ?", "%"+keyword+"%")
	}

	var longPosts []models.ArticleEditTask
	longPostQuery.Find(&longPosts)

	shortPostQuery := h.db.Model(&short_post.ShortPostProject{})
	if filterUserID != "" {
		shortPostQuery = shortPostQuery.Where("user_id = ?", filterUserID)
	}
	if keyword != "" {
		shortPostQuery = shortPostQuery.Where("name LIKE ?", "%"+keyword+"%")
	}

	var shortPosts []short_post.ShortPostProject
	shortPostQuery.Find(&shortPosts)

	total = int64(len(longPosts) + len(shortPosts))

	// 合并数据
	allItems := make([]map[string]interface{}, 0, len(longPosts)+len(shortPosts))

	for _, p := range longPosts {
		var tags []string
		if p.Tags != nil && *p.Tags != "" {
			json.Unmarshal([]byte(*p.Tags), &tags)
		}

		content := p.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		sectionHTML := ""
		if p.SectionHTML != nil {
			sectionHTML = *p.SectionHTML
		}

		item := map[string]interface{}{
			"id":              p.ID,
			"article_task_id": p.ArticleTaskID,
			"user_id":         p.UserID,
			"title":           p.Title,
			"theme":           p.Theme,
			"content":         content,
			"section_html":    sectionHTML,
			"status":          p.Status,
			"is_public":       p.IsPublic,
			"tags":            tags,
			"category":        "long_post",
			"created_at":      p.CreatedAt.Format(time.RFC3339),
			"updated_at":      p.UpdatedAt.Format(time.RFC3339),
			"_sort_time":      p.UpdatedAt,
		}

		if p.PublishedAt != nil {
			item["published_at"] = p.PublishedAt.Format(time.RFC3339)
		}

		var user models.User
		if err := h.db.Where("user_id = ?", p.UserID).First(&user).Error; err == nil {
			if user.Nickname != nil {
				item["user_name"] = *user.Nickname
			}
		}

		allItems = append(allItems, item)
	}

	for _, p := range shortPosts {
		var metadata interface{}
		if p.Metadata != nil {
			json.Unmarshal([]byte(*p.Metadata), &metadata)
		}

		item := map[string]interface{}{
			"id":          p.ID,
			"user_id":     p.UserID,
			"name":        p.Name,
			"status":      nil,
			"category":    nil,
			"frame_count": p.FrameCount,
			"metadata":    metadata,
			"created_at":  p.CreatedAt.Format(time.RFC3339),
			"updated_at":  p.UpdatedAt.Format(time.RFC3339),
			"_sort_time":  p.UpdatedAt,
		}

		if p.Description != nil {
			item["description"] = *p.Description
		}
		if p.CoverImage != nil {
			item["cover_image"] = *p.CoverImage
		}
		if p.Thumbnail != nil {
			item["thumbnail"] = *p.Thumbnail
		}
		if p.Status != "" {
			item["status"] = string(p.Status)
		}
		if p.ProjectType != "" {
			item["category"] = string(p.ProjectType)
		}

		var user models.User
		if err := h.db.Where("user_id = ?", p.UserID).First(&user).Error; err == nil {
			if user.Nickname != nil {
				item["user_name"] = *user.Nickname
			}
		}

		allItems = append(allItems, item)
	}

	// 排序
	reverse := order == "desc"
	sort.Slice(allItems, func(i, j int) bool {
		timeI := allItems[i]["_sort_time"].(time.Time)
		timeJ := allItems[j]["_sort_time"].(time.Time)
		if reverse {
			return timeI.After(timeJ)
		}
		return timeI.Before(timeJ)
	})

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(allItems) {
		end = len(allItems)
	}
	if start < len(allItems) {
		items = allItems[start:end]
	}

	// 移除排序辅助字段
	for _, item := range items {
		delete(item, "_sort_time")
	}

	middleware.Success(c, "success", gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetProjectDetail get project detail
func (h *ProjectHandler) GetProjectDetail(c *gin.Context) {
	projectID := c.Param("project_id")

	var project short_post.ShortPostProject
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 获取最新内容
	var content short_post.ShortPostProjectContent
	h.db.Where("project_id = ? AND is_latest = ?", projectID, true).First(&content)

	// 如果是小红书类型，获取文案信息
	var copywriting short_post.ShortPostProjectCopywriting
	if project.ProjectType == short_post.ProjectTypeXiaohongshu {
		h.db.Where("project_id = ?", projectID).First(&copywriting)
	}

	// 解析 JSON 字段
	var canvasConfig, framesData, elementsData, metadata interface{}
	if content.CanvasConfig != nil {
		json.Unmarshal([]byte(*content.CanvasConfig), &canvasConfig)
	}
	if content.FramesData != nil {
		json.Unmarshal([]byte(*content.FramesData), &framesData)
	}
	if content.ElementsData != nil {
		json.Unmarshal([]byte(*content.ElementsData), &elementsData)
	}
	if content.Metadata != nil {
		json.Unmarshal([]byte(*content.Metadata), &metadata)
	}

	projectMetadata := interface{}(nil)
	if project.Metadata != nil {
		json.Unmarshal([]byte(*project.Metadata), &projectMetadata)
	}

	contentData := map[string]interface{}{
		"id":            content.ID,
		"canvas_config": canvasConfig,
		"frames_data":   framesData,
		"elements_data": elementsData,
		"metadata":      metadata,
		"version":       content.Version,
		"updated_at":    content.UpdatedAt.Format(time.RFC3339),
	}

	copywritingData := map[string]interface{}{}
	if copywriting.ID != "" {
		var topics, images []string
		if copywriting.Topics != nil {
			json.Unmarshal([]byte(*copywriting.Topics), &topics)
		}
		if copywriting.Images != nil {
			json.Unmarshal([]byte(*copywriting.Images), &images)
		}

		copywritingData = map[string]interface{}{
			"id":         copywriting.ID,
			"title":      copywriting.Title,
			"content":    copywriting.Content,
			"topics":     topics,
			"images":     images,
			"updated_at": copywriting.UpdatedAt.Format(time.RFC3339),
		}
	}

	result := map[string]interface{}{
		"id":           project.ID,
		"user_id":      project.UserID,
		"name":         project.Name,
		"status":       string(project.Status),
		"frame_count":  project.FrameCount,
		"thread_id":    project.ThreadID,
		"project_type": string(project.ProjectType),
		"metadata":     projectMetadata,
		"created_at":   project.CreatedAt.Format(time.RFC3339),
		"updated_at":   project.UpdatedAt.Format(time.RFC3339),
		"content":      contentData,
		"copywriting":  copywritingData,
	}

	if project.Description != nil {
		result["description"] = *project.Description
	}
	if project.CoverImage != nil {
		result["cover_image"] = *project.CoverImage
	}
	if project.Thumbnail != nil {
		result["thumbnail"] = *project.Thumbnail
	}
	if project.SavedAt != nil {
		result["saved_at"] = project.SavedAt.Format(time.RFC3339)
	}

	middleware.Success(c, "success", result)
}

// CheckProjectHasContent check if project has content
func (h *ProjectHandler) CheckProjectHasContent(c *gin.Context) {
	projectID := c.Param("project_id")

	var project short_post.ShortPostProject
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var content short_post.ShortPostProjectContent
	h.db.Where("project_id = ? AND is_latest = ?", projectID, true).First(&content)

	// 不仅要存在内容记录，还要确保 metadata 不为空
	hasContent := content.ID != "" && content.Metadata != nil && *content.Metadata != "" && *content.Metadata != "{}"

	middleware.Success(c, "success", hasContent)
}

// UpdateProject update project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	projectID := c.Param("project_id")

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	// 检查是否为管理员或工程所有者
	var user models.User
	h.db.Where("user_id = ?", userID).First(&user)

	var project short_post.ShortPostProject
	query := h.db.Where("id = ?", projectID)
	if user.Role != 3 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}
	if req.Thumbnail != nil {
		updates["thumbnail"] = *req.Thumbnail
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ProjectType != nil {
		updates["project_type"] = *req.ProjectType
	}
	if req.Metadata != nil {
		metadataBytes, _ := json.Marshal(req.Metadata)
		updates["metadata"] = string(metadataBytes)
	}

	if len(updates) > 0 {
		if err := h.db.Model(&project).Updates(updates).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
		h.db.Where("id = ?", projectID).First(&project)
	}

	middleware.Success(c, "更新成功", gin.H{
		"id":         project.ID,
		"name":       project.Name,
		"status":     string(project.Status),
		"updated_at": project.UpdatedAt.Format(time.RFC3339),
	})
}

// SaveProjectContent save project content
func (h *ProjectHandler) SaveProjectContent(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	projectID := c.Param("project_id")

	var req SaveProjectContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	// 检查是否为管理员或工程所有者
	var user models.User
	h.db.Where("user_id = ?", userID).First(&user)

	var project short_post.ShortPostProject
	query := h.db.Where("id = ?", projectID)
	if user.Role != 3 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 序列化 JSON 字段
	var canvasConfigJSON, framesDataJSON, elementsDataJSON, metadataJSON *string
	if req.CanvasConfig != nil {
		bytes, _ := json.Marshal(req.CanvasConfig)
		str := string(bytes)
		canvasConfigJSON = &str
	}
	if req.FramesData != nil {
		bytes, _ := json.Marshal(req.FramesData)
		str := string(bytes)
		framesDataJSON = &str
	}
	if req.ElementsData != nil {
		bytes, _ := json.Marshal(req.ElementsData)
		str := string(bytes)
		elementsDataJSON = &str
	}
	if req.Metadata != nil {
		bytes, _ := json.Marshal(req.Metadata)
		str := string(bytes)
		metadataJSON = &str
	}

	var content short_post.ShortPostProjectContent
	h.db.Where("project_id = ? AND is_latest = ?", projectID, true).First(&content)

	if req.CreateVersion && content.ID != "" {
		// 创建新版本
		h.db.Model(&short_post.ShortPostProjectContent{}).
			Where("project_id = ? AND is_latest = ?", projectID, true).
			Update("is_latest", false)

		version := content.Version + 1
		content = short_post.ShortPostProjectContent{
			ID:           uuid.New().String(),
			ProjectID:    projectID,
			CanvasConfig: canvasConfigJSON,
			FramesData:   framesDataJSON,
			ElementsData: elementsDataJSON,
			Metadata:     metadataJSON,
			Version:      version,
			IsLatest:     true,
		}
		h.db.Create(&content)
	} else {
		if content.ID != "" {
			// 更新现有内容
			updates := make(map[string]interface{})
			if req.CanvasConfig != nil {
				updates["canvas_config"] = canvasConfigJSON
			}
			if req.FramesData != nil {
				updates["frames_data"] = framesDataJSON
			}
			if req.ElementsData != nil {
				updates["elements_data"] = elementsDataJSON
			}
			if req.Metadata != nil {
				updates["metadata"] = metadataJSON
			}
			if len(updates) > 0 {
				h.db.Model(&content).Updates(updates)
				h.db.Where("id = ?", content.ID).First(&content)
			}
		} else {
			// 创建新内容
			content = short_post.ShortPostProjectContent{
				ID:           uuid.New().String(),
				ProjectID:    projectID,
				CanvasConfig: canvasConfigJSON,
				FramesData:   framesDataJSON,
				ElementsData: elementsDataJSON,
				Metadata:     metadataJSON,
				Version:      1,
				IsLatest:     true,
			}
			h.db.Create(&content)
		}
	}

	// 更新工程信息
	frameCount := 0
	if req.FramesData != nil {
		frameCount = len(req.FramesData)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"saved_at":    now,
		"status":      short_post.ProjectStatusSaved,
		"frame_count": frameCount,
	}

	// 从 canvas_config 中提取 thumbnail
	if req.CanvasConfig != nil {
		if thumb, ok := req.CanvasConfig["thumbnail"].(string); ok && thumb != "" {
			updates["thumbnail"] = thumb
		}
	}

	h.db.Model(&project).Updates(updates)

	middleware.Success(c, "保存成功", gin.H{
		"content_id": content.ID,
		"version":    content.Version,
		"saved_at":   now.Format(time.RFC3339),
	})
}

// DeleteProject delete project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	projectID := c.Param("project_id")

	var project short_post.ShortPostProject
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&project).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "删除成功", nil)
}

// GetProjectVersions get project versions
func (h *ProjectHandler) GetProjectVersions(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	projectID := c.Param("project_id")

	var project short_post.ShortPostProject
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var contents []short_post.ShortPostProjectContent
	h.db.Where("project_id = ?", projectID).Order("version DESC").Find(&contents)

	items := make([]map[string]interface{}, 0, len(contents))
	for _, c := range contents {
		items = append(items, map[string]interface{}{
			"id":         c.ID,
			"version":    c.Version,
			"is_latest":  c.IsLatest,
			"created_at": c.CreatedAt.Format(time.RFC3339),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items": items,
		"total": len(items),
	})
}

// SaveCopywriting save copywriting
func (h *ProjectHandler) SaveCopywriting(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	projectID := c.Param("project_id")

	var req SaveCopywritingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	var project short_post.ShortPostProject
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 只有小红书类型的工程才能保存文案
	if project.ProjectType != short_post.ProjectTypeXiaohongshu {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "只有小红书类型的工程才能保存文案"))
		return
	}

	var copywriting short_post.ShortPostProjectCopywriting
	h.db.Where("project_id = ?", projectID).First(&copywriting)

	// 序列化 JSON 字段
	var topicsJSON, imagesJSON *string
	if req.Topics != nil {
		bytes, _ := json.Marshal(req.Topics)
		str := string(bytes)
		topicsJSON = &str
	}
	if req.Images != nil {
		bytes, _ := json.Marshal(req.Images)
		str := string(bytes)
		imagesJSON = &str
	}

	if copywriting.ID != "" {
		// 更新现有文案
		updates := make(map[string]interface{})
		if req.Title != nil {
			updates["title"] = *req.Title
		}
		if req.Content != nil {
			updates["content"] = *req.Content
		}
		if req.Topics != nil {
			updates["topics"] = topicsJSON
		}
		if req.Images != nil {
			updates["images"] = imagesJSON
		}
		if len(updates) > 0 {
			h.db.Model(&copywriting).Updates(updates)
			h.db.Where("id = ?", copywriting.ID).First(&copywriting)
		}
	} else {
		// 创建新文案
		copywriting = short_post.ShortPostProjectCopywriting{
			ID:        uuid.New().String(),
			ProjectID: projectID,
			Title:     req.Title,
			Content:   req.Content,
			Topics:    topicsJSON,
			Images:    imagesJSON,
		}
		h.db.Create(&copywriting)
	}

	var topics, images []string
	if copywriting.Topics != nil {
		json.Unmarshal([]byte(*copywriting.Topics), &topics)
	}
	if copywriting.Images != nil {
		json.Unmarshal([]byte(*copywriting.Images), &images)
	}

	middleware.Success(c, "保存成功", gin.H{
		"id":         copywriting.ID,
		"title":      copywriting.Title,
		"content":    copywriting.Content,
		"topics":     topics,
		"images":     images,
		"updated_at": copywriting.UpdatedAt.Format(time.RFC3339),
	})
}

// GetCopywriting get copywriting
func (h *ProjectHandler) GetCopywriting(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	projectID := c.Param("project_id")

	var project short_post.ShortPostProject
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "工程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 只有小红书类型的工程才有文案
	if project.ProjectType != short_post.ProjectTypeXiaohongshu {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "只有小红书类型的工程才有文案"))
		return
	}

	var copywriting short_post.ShortPostProjectCopywriting
	if err := h.db.Where("project_id = ?", projectID).First(&copywriting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.Success(c, "success", nil)
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var topics, images []string
	if copywriting.Topics != nil {
		json.Unmarshal([]byte(*copywriting.Topics), &topics)
	}
	if copywriting.Images != nil {
		json.Unmarshal([]byte(*copywriting.Images), &images)
	}

	middleware.Success(c, "success", gin.H{
		"id":         copywriting.ID,
		"title":      copywriting.Title,
		"content":    copywriting.Content,
		"topics":     topics,
		"images":     images,
		"created_at": copywriting.CreatedAt.Format(time.RFC3339),
		"updated_at": copywriting.UpdatedAt.Format(time.RFC3339),
	})
}

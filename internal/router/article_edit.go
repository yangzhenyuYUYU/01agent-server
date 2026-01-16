package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArticleEditHandler struct {
	db             *gorm.DB
	articleEditSvc *service.ArticleEditService
}

func NewArticleEditHandler() *ArticleEditHandler {
	return &ArticleEditHandler{
		db:             repository.DB,
		articleEditSvc: service.NewArticleEditService(),
	}
}

// Request models
type ArticleEditCreateRequest struct {
	ArticleTaskID uuid.UUID              `json:"article_task_id" binding:"required"`
	Title         *string                `json:"title"`
	Theme         *string                `json:"theme"`
	Content       *string                `json:"content"`
	Params        map[string]interface{} `json:"params"`
	IsPublic      bool                   `json:"is_public"`
	Tags          []string               `json:"tags"`
}

type ArticleEditUpdateRequest struct {
	Title    *string                `json:"title"`
	Theme    *string                `json:"theme"`
	Content  *string                `json:"content"`
	IsPublic *bool                  `json:"is_public"`
	Tags     []string               `json:"tags"`
	Status   *string                `json:"status"`
	Params   map[string]interface{} `json:"params"`
}

type PublishEditTaskRequest struct {
	ThumbURL           *string `json:"thumb_url"`
	Title              *string `json:"title"`
	Content            *string `json:"content"`
	Author             *string `json:"author"`
	Digest             *string `json:"digest"`
	NeedOpenComment    *int    `json:"need_open_comment"`
	OnlyFansCanComment *int    `json:"only_fans_can_comment"`
	SyncOnline         *bool   `json:"sync_online"`
	SectionHTML        *string `json:"section_html"`
}

type SavePublishConfigRequest struct {
	PublishTitle         string  `json:"publish_title" binding:"required"`
	AuthorName           string  `json:"author_name" binding:"required"`
	Summary              *string `json:"summary"`
	CoverImage           *string `json:"cover_image"`
	EnableComments       *bool   `json:"enable_comments"`
	FollowersOnlyComment *bool   `json:"followers_only_comment"`
}

type ArticleEditResponse struct {
	ID            string                 `json:"id"`
	ArticleTaskID string                 `json:"article_task_id"`
	Title         string                 `json:"title"`
	Theme         string                 `json:"theme"`
	Content       string                 `json:"content"`
	SectionHTML   string                 `json:"section_html"`
	Status        string                 `json:"status"`
	IsPublic      bool                   `json:"is_public"`
	Params        map[string]interface{} `json:"params"`
	Tags          []string               `json:"tags"`
	ThumbnailURL  string                 `json:"thumbnail_url"`
	Snippet       string                 `json:"snippet"`
	AuthorName    string                 `json:"author_name"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	PublishedAt   *string                `json:"published_at"`
}

func (h *ArticleEditHandler) convertToResponse(editTask *models.ArticleEditTask, previewMode bool, articleTask *models.ArticleTask) (*ArticleEditResponse, error) {
	var tags []string
	if editTask.Tags != nil && *editTask.Tags != "" {
		json.Unmarshal([]byte(*editTask.Tags), &tags)
	}

	if articleTask == nil && editTask.ArticleTaskID != nil {
		var task models.ArticleTask
		if err := h.db.Where("id = ?", *editTask.ArticleTaskID).First(&task).Error; err == nil {
			articleTask = &task
		}
	}

	authorName := "Author"
	if articleTask != nil && articleTask.AuthorName != "" {
		authorName = articleTask.AuthorName
	}

	thumbnailURL := "https://01agent.net/alllogo.png"
	if articleTask != nil && articleTask.Images != nil && editTask.Content != "" {
		re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
		if matches := re.FindStringSubmatch(editTask.Content); len(matches) > 1 {
			thumbnailURL = matches[1]
		}
	}

	sectionHTML := ""
	if editTask.SectionHTML != nil {
		sectionHTML = *editTask.SectionHTML
	}

	content := editTask.Content
	if !previewMode && len(content) > 100 {
		content = content[:100] + "..."
	}

	var params map[string]interface{}
	if editTask.Params != nil && *editTask.Params != "" {
		json.Unmarshal([]byte(*editTask.Params), &params)
	}

	snippet := ""
	if articleTask != nil && articleTask.Snippet != nil {
		snippet = *articleTask.Snippet
	}

	response := &ArticleEditResponse{
		ID:            editTask.ID,
		ArticleTaskID: *editTask.ArticleTaskID,
		Title:         editTask.Title,
		Theme:         editTask.Theme,
		Content:       content,
		SectionHTML:   sectionHTML,
		Status:        editTask.Status,
		IsPublic:      editTask.IsPublic,
		Params:        params,
		Tags:          tags,
		ThumbnailURL:  thumbnailURL,
		Snippet:       snippet,
		AuthorName:    authorName,
		CreatedAt:     editTask.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     editTask.UpdatedAt.Format(time.RFC3339),
	}

	if editTask.PublishedAt != nil {
		publishedAt := editTask.PublishedAt.Format(time.RFC3339)
		response.PublishedAt = &publishedAt
	}

	return response, nil
}

// simplifiedConvertToResponse 简化版转换（用于列表，避免N+1查询）
func (h *ArticleEditHandler) simplifiedConvertToResponse(editTask *models.ArticleEditTask) (*ArticleEditResponse, error) {
	var tags []string
	if editTask.Tags != nil && *editTask.Tags != "" {
		json.Unmarshal([]byte(*editTask.Tags), &tags)
	}

	sectionHTML := ""
	if editTask.SectionHTML != nil {
		sectionHTML = *editTask.SectionHTML
	}

	// 截断content为前100字符
	content := editTask.Content
	if len(content) > 100 {
		content = content[:100] + "..."
	}

	var params map[string]interface{}
	if editTask.Params != nil && *editTask.Params != "" {
		json.Unmarshal([]byte(*editTask.Params), &params)
	}

	response := &ArticleEditResponse{
		ID:          editTask.ID,
		Title:       editTask.Title,
		Theme:       editTask.Theme,
		Content:     content,
		SectionHTML: sectionHTML,
		Status:      editTask.Status,
		IsPublic:    editTask.IsPublic,
		Params:      params,
		Tags:        tags,
		CreatedAt:   editTask.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   editTask.UpdatedAt.Format(time.RFC3339),
	}

	if editTask.ArticleTaskID != nil {
		response.ArticleTaskID = *editTask.ArticleTaskID
	}

	if editTask.PublishedAt != nil {
		publishedAt := editTask.PublishedAt.Format(time.RFC3339)
		response.PublishedAt = &publishedAt
	}

	return response, nil
}

// simplifiedConvertToResponseFast 极速版转换（用于列表，最小化处理）
func (h *ArticleEditHandler) simplifiedConvertToResponseFast(editTask *models.ArticleEditTask) *ArticleEditResponse {
	// 处理 section_html
	sectionHTML := ""
	if editTask.SectionHTML != nil {
		html := *editTask.SectionHTML
		// 截断到 50000 字符（约 50KB），保证预览效果
		// 如果需要完整内容，前端应该调用详情接口
		const maxHTMLSize = 50000
		if len(html) > maxHTMLSize {
			sectionHTML = html[:maxHTMLSize]
		} else {
			sectionHTML = html
		}
	}

	response := &ArticleEditResponse{
		ID:          editTask.ID,
		Title:       editTask.Title,
		Theme:       editTask.Theme,
		Content:     "",          // 列表不返回content，减少数据传输
		SectionHTML: sectionHTML, // 返回保存的HTML（前端需要）
		Status:      editTask.Status,
		IsPublic:    editTask.IsPublic,
		Params:      nil,        // 列表不解析params
		Tags:        []string{}, // 列表不解析tags
		CreatedAt:   editTask.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   editTask.UpdatedAt.Format(time.RFC3339),
	}

	if editTask.ArticleTaskID != nil {
		response.ArticleTaskID = *editTask.ArticleTaskID
	}

	if editTask.PublishedAt != nil {
		publishedAt := editTask.PublishedAt.Format(time.RFC3339)
		response.PublishedAt = &publishedAt
	}

	return response
}

// GetEditTaskByArticleID GET /article-edit/edit_task/:article_task_id
func (h *ArticleEditHandler) GetEditTaskByArticleID(c *gin.Context) {
	articleTaskID := c.Param("article_task_id")

	var articleTask models.ArticleTask
	if err := h.db.Where("id = ?", articleTaskID).First(&articleTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Article task not found"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	var editTask models.ArticleEditTask
	if err := h.db.Where("article_task_id = ?", articleTaskID).
		Order("created_at DESC").First(&editTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	data, err := h.convertToResponse(&editTask, true, &articleTask)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Convert failed: %v", err)))
		return
	}

	// 将响应数据转换为 map 并添加 author_id 字段
	dataMap := make(map[string]interface{})
	dataBytes, _ := json.Marshal(data)
	json.Unmarshal(dataBytes, &dataMap)
	dataMap["author_id"] = articleTask.UserID

	middleware.Success(c, "Success", dataMap)
}

// CreateEditTask POST /article-edit/create
func (h *ArticleEditHandler) CreateEditTask(c *gin.Context) {
	var req ArticleEditCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("Invalid params: %v", err)))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	var articleTask models.ArticleTask
	if err := h.db.Where("id = ?", req.ArticleTaskID).First(&articleTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Article task not found"))
		return
	}

	// 从请求或 articleTask 中获取默认值
	title := ""
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	} else if articleTask.Title != nil {
		title = *articleTask.Title
	}

	theme := "default"
	if req.Theme != nil && *req.Theme != "" {
		theme = *req.Theme
	} else if articleTask.Theme != nil && *articleTask.Theme != "" {
		theme = *articleTask.Theme
	}

	content := ""
	if req.Content != nil && *req.Content != "" {
		content = *req.Content
	} else if articleTask.Content != nil {
		content = *articleTask.Content
	}

	var tagsJSON *string
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsStr := string(tagsBytes)
		tagsJSON = &tagsStr
	}

	var paramsJSON *string
	if req.Params != nil {
		paramsBytes, _ := json.Marshal(req.Params)
		paramsStr := string(paramsBytes)
		paramsJSON = &paramsStr
	}

	articleTaskIDStr := req.ArticleTaskID.String()
	var editTask models.ArticleEditTask
	created := h.db.Where("user_id = ? AND article_task_id = ?", userID, articleTaskIDStr).
		FirstOrCreate(&editTask, models.ArticleEditTask{
			ID:            uuid.New().String(),
			UserID:        userID,
			ArticleTaskID: &articleTaskIDStr,
			Title:         title,
			Theme:         theme,
			Content:       content,
			IsPublic:      req.IsPublic,
			Tags:          tagsJSON,
			Status:        models.ArticleEditStatusEditing,
			Params:        paramsJSON,
		}).Error

	if created != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Create failed: %v", created)))
		return
	}

	// 如果记录已存在且请求提供了新值，则更新
	updates := make(map[string]interface{})
	if req.Title != nil && *req.Title != "" {
		updates["title"] = *req.Title
	}
	if req.Theme != nil && *req.Theme != "" {
		updates["theme"] = *req.Theme
	}
	if req.Content != nil && *req.Content != "" {
		updates["content"] = *req.Content
	}
	if req.Params != nil {
		paramsBytes, _ := json.Marshal(req.Params)
		updates["params"] = string(paramsBytes)
	}
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		updates["tags"] = string(tagsBytes)
	}
	updates["is_public"] = req.IsPublic

	if len(updates) > 0 {
		if err := h.db.Model(&editTask).Updates(updates).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Update failed: %v", err)))
			return
		}
		// 重新查询以获取更新后的数据
		h.db.Where("id = ?", editTask.ID).First(&editTask)
	}

	data, err := h.convertToResponse(&editTask, true, &articleTask)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Convert failed: %v", err)))
		return
	}

	middleware.Success(c, "Success", data)
}

// GetEditDrafts GET /article-edit/drafts
func (h *ArticleEditHandler) GetEditDrafts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	status := c.Query("status")
	userID, _ := middleware.GetCurrentUserID(c)

	var pageInt, pageSizeInt int
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)

	if pageInt < 1 {
		pageInt = 1
	}
	if pageSizeInt < 1 {
		pageSizeInt = 10
	}

	query := h.db.Model(&models.ArticleEditTask{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	// 只查询必要的字段，不查询大字段 content，但保留 section_html（前端需要）
	var items []models.ArticleEditTask
	offset := (pageInt - 1) * pageSizeInt
	if err := query.Select("id", "article_task_id", "user_id", "title", "theme", "status", "is_public", "tags", "params", "section_html", "published_at", "created_at", "updated_at").
		Order("updated_at DESC").Offset(offset).Limit(pageSizeInt).Find(&items).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	resultItems := make([]*ArticleEditResponse, 0, len(items))
	for i := range items {
		item := h.simplifiedConvertToResponseFast(&items[i])
		resultItems = append(resultItems, item)
	}

	middleware.Success(c, "Success", gin.H{
		"items":     resultItems,
		"total":     total,
		"page":      pageInt,
		"page_size": pageSizeInt,
	})
}

// GetEditTask GET /article-edit/:edit_task_id
func (h *ArticleEditHandler) GetEditTask(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := h.db.Where("article_task_id = ? AND user_id = ?", editTaskID, userID).
				Order("created_at DESC").First(&editTask).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
				return
			}
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
			return
		}
	}

	data, err := h.convertToResponse(&editTask, true, nil)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Convert failed: %v", err)))
		return
	}

	middleware.Success(c, "Success", data)
}

// UpdateEditTask PUT /article-edit/:edit_task_id
func (h *ArticleEditHandler) UpdateEditTask(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var req ArticleEditUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("Invalid params: %v", err)))
		return
	}

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
		return
	}

	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Theme != nil {
		updates["theme"] = *req.Theme
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.Tags != nil {
		tagsBytes, _ := json.Marshal(req.Tags)
		updates["tags"] = string(tagsBytes)
	}
	if req.Params != nil {
		paramsBytes, _ := json.Marshal(req.Params)
		updates["params"] = string(paramsBytes)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == models.ArticleEditStatusPublished {
			updates["published_at"] = time.Now()
		}
	}

	if err := h.db.Model(&editTask).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Update failed: %v", err)))
		return
	}

	h.db.Where("id = ?", editTaskID).First(&editTask)

	data, err := h.convertToResponse(&editTask, false, nil)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Convert failed: %v", err)))
		return
	}

	middleware.Success(c, "Success", data)
}

// DeleteEditTask DELETE /article-edit/:edit_task_id
func (h *ArticleEditHandler) DeleteEditTask(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
		return
	}

	if err := h.db.Delete(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Delete failed: %v", err)))
		return
	}

	middleware.Success(c, "Success", nil)
}

// PublishEditTask PUT /article-edit/:edit_task_id/publish
func (h *ArticleEditHandler) PublishEditTask(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var req PublishEditTaskRequest
	c.ShouldBindJSON(&req)

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
		return
	}

	go h.articleEditSvc.ProcessPublishTask(editTaskID, (*service.PublishEditTaskRequest)(&req), userID)

	middleware.Success(c, "Success", gin.H{
		"task_id": editTaskID,
		"status":  "processing",
		"message": "Publish task submitted",
	})
}

// GetPublishStatus GET /article-edit/:edit_task_id/publish-status
func (h *ArticleEditHandler) GetPublishStatus(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
		return
	}

	statusInfo := gin.H{
		"task_id": editTaskID,
		"status":  editTask.Status,
	}

	if editTask.PublishedAt != nil {
		statusInfo["published_at"] = editTask.PublishedAt.Format(time.RFC3339)
	}

	switch editTask.Status {
	case models.ArticleEditStatusPublished:
		statusInfo["message"] = "Published"
	case models.ArticleEditStatusPending:
		statusInfo["message"] = "Pending"
	case models.ArticleEditStatusDraft:
		statusInfo["message"] = "Draft"
	case models.ArticleEditStatusEditing:
		statusInfo["message"] = "Editing"
	default:
		statusInfo["message"] = fmt.Sprintf("Status: %s", editTask.Status)
	}

	middleware.Success(c, "Success", statusInfo)
}

// SavePublishConfig POST /article-edit/:edit_task_id/publish-config
func (h *ArticleEditHandler) SavePublishConfig(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var req SavePublishConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("Invalid params: %v", err)))
		return
	}

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
		return
	}

	var publishConfig models.ArticlePublishConfig
	err := h.db.Where("edit_task_id = ?", editTaskID).First(&publishConfig).Error

	if err == gorm.ErrRecordNotFound {
		publishConfig = models.ArticlePublishConfig{
			ID:                   uuid.New().String(),
			EditTaskID:           editTaskID,
			PublishTitle:         req.PublishTitle,
			AuthorName:           req.AuthorName,
			Summary:              req.Summary,
			CoverImage:           req.CoverImage,
			EnableComments:       req.EnableComments != nil && *req.EnableComments,
			FollowersOnlyComment: req.FollowersOnlyComment != nil && *req.FollowersOnlyComment,
		}
		h.db.Create(&publishConfig)
	} else if err == nil {
		updates := map[string]interface{}{
			"publish_title": req.PublishTitle,
			"author_name":   req.AuthorName,
		}
		if req.Summary != nil {
			updates["summary"] = *req.Summary
		}
		if req.CoverImage != nil {
			updates["cover_image"] = *req.CoverImage
		}
		if req.EnableComments != nil {
			updates["enable_comments"] = *req.EnableComments
		}
		if req.FollowersOnlyComment != nil {
			updates["followers_only_comment"] = *req.FollowersOnlyComment
		}
		h.db.Model(&publishConfig).Updates(updates)
		h.db.Where("id = ?", publishConfig.ID).First(&publishConfig)
	}

	middleware.Success(c, "Success", gin.H{
		"id":                     publishConfig.ID,
		"edit_task_id":           publishConfig.EditTaskID,
		"publish_title":          publishConfig.PublishTitle,
		"author_name":            publishConfig.AuthorName,
		"summary":                publishConfig.Summary,
		"cover_image":            publishConfig.CoverImage,
		"enable_comments":        publishConfig.EnableComments,
		"followers_only_comment": publishConfig.FollowersOnlyComment,
		"created_at":             publishConfig.CreatedAt.Format(time.RFC3339),
		"updated_at":             publishConfig.UpdatedAt.Format(time.RFC3339),
	})
}

// GetPublishConfig GET /article-edit/:edit_task_id/publish-config
func (h *ArticleEditHandler) GetPublishConfig(c *gin.Context) {
	editTaskID := c.Param("edit_task_id")
	userID, _ := middleware.GetCurrentUserID(c)

	var editTask models.ArticleEditTask
	if err := h.db.Where("id = ? AND user_id = ?", editTaskID, userID).First(&editTask).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Edit task not found"))
		return
	}

	var publishConfig models.ArticlePublishConfig
	if err := h.db.Where("edit_task_id = ?", editTaskID).First(&publishConfig).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.Success(c, "No config found", nil)
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	middleware.Success(c, "Success", gin.H{
		"id":                     publishConfig.ID,
		"edit_task_id":           publishConfig.EditTaskID,
		"publish_title":          publishConfig.PublishTitle,
		"author_name":            publishConfig.AuthorName,
		"summary":                publishConfig.Summary,
		"cover_image":            publishConfig.CoverImage,
		"enable_comments":        publishConfig.EnableComments,
		"followers_only_comment": publishConfig.FollowersOnlyComment,
		"created_at":             publishConfig.CreatedAt.Format(time.RFC3339),
		"updated_at":             publishConfig.UpdatedAt.Format(time.RFC3339),
	})
}

func SetupArticleEditRoutes(r *gin.Engine) {
	handler := NewArticleEditHandler()
	articleEdit := r.Group("/api/v1/article-edit")

	// 不需要认证的接口
	articleEdit.GET("/edit_task/:article_task_id", handler.GetEditTaskByArticleID)

	// 需要认证的接口
	articleEditWithAuth := articleEdit.Group("")
	articleEditWithAuth.Use(middleware.JWTAuth())
	{
		articleEditWithAuth.POST("/create", handler.CreateEditTask)
		articleEditWithAuth.GET("/drafts", handler.GetEditDrafts)
		articleEditWithAuth.GET("/:edit_task_id", handler.GetEditTask)
		articleEditWithAuth.PUT("/:edit_task_id", handler.UpdateEditTask)
		articleEditWithAuth.DELETE("/:edit_task_id", handler.DeleteEditTask)
		articleEditWithAuth.PUT("/:edit_task_id/publish", handler.PublishEditTask)
		articleEditWithAuth.GET("/:edit_task_id/publish-status", handler.GetPublishStatus)
		articleEditWithAuth.POST("/:edit_task_id/publish-config", handler.SavePublishConfig)
		articleEditWithAuth.GET("/:edit_task_id/publish-config", handler.GetPublishConfig)
	}
}

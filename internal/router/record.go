package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RecordHandler record handler
type RecordHandler struct {
	db                *gorm.DB
	markdownProcessor *tools.UnifiedMarkdownProcessor
}

// NewRecordHandler create record handler
func NewRecordHandler() *RecordHandler {
	return &RecordHandler{
		db:                repository.DB,
		markdownProcessor: tools.NewUnifiedMarkdownProcessor(),
	}
}

// SetupRecordRoutes setup record routes
func SetupRecordRoutes(r *gin.Engine) {
	handler := NewRecordHandler()

	group := r.Group("/api/v1/agent/record")
	{
		// Task routes
		group.GET("/task/public/list", middleware.JWTAuth(), handler.GetPublicTaskList)
		group.GET("/task/list", middleware.JWTAuth(), handler.GetTaskList)
		group.DELETE("/task/:task_id", middleware.JWTAuth(), handler.DeleteTask)
		group.GET("/task/:task_id", middleware.JWTAuth(), handler.GetTaskDetail)

		// Topic routes (skip create/update)
		group.GET("/topic/list", middleware.JWTAuth(), handler.GetTopicList)
		group.DELETE("/topic/:id", middleware.JWTAuth(), handler.DeleteTopic)
		group.POST("/topics/delete", middleware.JWTAuth(), handler.DeleteTopicsBatch)
		group.POST("/topics/clear", middleware.JWTAuth(), handler.ClearTopics)
		group.GET("/topic/analysis_task_data", middleware.JWTAuth(), handler.GetAnalysisTaskData)

		// Hot topics
		group.POST("/hot_topics", handler.GetHotTopics)

		// Copilot chat
		group.GET("/copilot/chat/:thread_id", middleware.JWTAuth(), handler.GetCopilotChatHistory)
		group.GET("/copilot/sessions", middleware.JWTAuth(), handler.GetCopilotSessions)
	}
}

// TaskListItem task list item (simplified for performance)
type TaskListItem struct {
	ID        string  `json:"id"`
	Title     *string `json:"title"`
	Status    string  `json:"status"`
	Snippet   *string `json:"snippet"`
	CreatedAt string  `json:"created_at"`
	Thumbnail string  `json:"thumbnail"`
	WordCount *int    `json:"word_count"`
	UpdatedAt string  `json:"updated_at"`
}

// GetPublicTaskList get public task list
func (h *RecordHandler) GetPublicTaskList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.ArticleTask{}).Where("is_public = ?", true)

	if startDate != "" {
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if endDate != "" {
		if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", endTime)
		}
	}

	var total int64
	query.Count(&total)

	var tasks []models.ArticleTask
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	filteredTasks := make([]models.ArticleTask, 0)
	for _, task := range tasks {
		if task.Content != nil && !tools.IsDefaultWelcomeContent(*task.Content) {
			filteredTasks = append(filteredTasks, task)
		} else if task.Content == nil {
			filteredTasks = append(filteredTasks, task)
		}
	}

	items := make([]TaskListItem, 0, len(filteredTasks))
	for _, task := range filteredTasks {
		thumbnail := ""
		if task.Content != nil {
			thumbnail = tools.ProcessSingleRecordThumbnail(*task.Content, task.Images)
		}

		items = append(items, TaskListItem{
			ID:        task.ID,
			Title:     task.Title,
			Snippet:   task.Snippet,
			CreatedAt: task.CreatedAt.Format("2006-01-02 15:04:05"),
			Thumbnail: thumbnail,
			WordCount: task.WordCount,
			UpdatedAt: task.CreatedAt.Format("2006-01-02 15:04:05"), // No edit task for public list
		})
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Success",
		"data": gin.H{
			"items":       items,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"start_date":  startDate,
			"end_date":    endDate,
		},
	})
}

// GetTaskList get user task list
func (h *RecordHandler) GetTaskList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.ArticleTask{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if startDate != "" {
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if endDate != "" {
		if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", endTime)
		}
	}

	var total int64
	query.Count(&total)

	// 优化：不查询 content 大字段，只查询必要字段和 images（用于提取封面）
	startQuery := time.Now()
	var tasks []models.ArticleTask
	if err := query.Select("id", "title", "status", "snippet", "word_count", "images", "created_at", "updated_at").
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}
	fmt.Printf("[PERF] Main query took: %v\n", time.Since(startQuery))

	// 优化：批量更新超时任务状态
	startUpdate := time.Now()
	currentTime := time.Now()
	timeoutThreshold := currentTime.Add(-24 * time.Hour)
	h.db.Model(&models.ArticleTask{}).
		Where("user_id = ? AND created_at < ? AND content IS NULL AND status != ?",
			userID, timeoutThreshold, "failed").
		Update("status", "failed")
	fmt.Printf("[PERF] Status update took: %v\n", time.Since(startUpdate))

	// 收集任务ID并一次性查询编辑任务
	taskIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	// 优化：只查询编辑任务的 updated_at（不查询 content 大字段）
	startEditQuery := time.Now()
	editTasksMap := make(map[string]*models.ArticleEditTask)
	if len(taskIDs) > 0 {
		var editTasks []models.ArticleEditTask
		h.db.Select("id", "article_task_id", "updated_at").
			Where("article_task_id IN ?", taskIDs).
			Order("created_at DESC").
			Find(&editTasks)

		for i := range editTasks {
			et := &editTasks[i]
			if et.ArticleTaskID != nil {
				if _, exists := editTasksMap[*et.ArticleTaskID]; !exists {
					editTasksMap[*et.ArticleTaskID] = et
				}
			}
		}
	}
	fmt.Printf("[PERF] Edit tasks query took: %v\n", time.Since(startEditQuery))

	// 构建响应
	startBuild := time.Now()

	// 构建响应（只从 images JSON 提取缩略图，不读取 content 大字段）
	items := make([]TaskListItem, 0, len(tasks))
	for _, task := range tasks {
		editTask := editTasksMap[task.ID]

		// 直接从 images JSON 字段提取缩略图（快速，不涉及大字段）
		thumbnail := ""
		if task.Images != nil && *task.Images != "" {
			// images 是 JSON 字符串，需要先解析
			var imagesData interface{}
			if err := json.Unmarshal([]byte(*task.Images), &imagesData); err == nil {
				thumbnail = tools.ExtractThumbnailFromImages(imagesData)
			}
		}

		updatedAt := task.UpdatedAt.Format("2006-01-02 15:04:05")
		if editTask != nil {
			updatedAt = editTask.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		items = append(items, TaskListItem{
			ID:        task.ID,
			Title:     task.Title,
			Status:    task.Status,
			Snippet:   task.Snippet,
			CreatedAt: task.CreatedAt.Format("2006-01-02 15:04:05"),
			Thumbnail: thumbnail,
			WordCount: task.WordCount,
			UpdatedAt: updatedAt,
		})
	}
	fmt.Printf("[PERF] Build response took: %v, items: %d\n", time.Since(startBuild), len(items))

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Success",
		"data": gin.H{
			"items":       items,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"start_date":  startDate,
			"end_date":    endDate,
		},
	})
}

// DeleteTask delete task
func (h *RecordHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("task_id")

	var task models.ArticleTask
	if err := h.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "Task not found"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	if err := h.db.Delete(&task).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Delete failed: %v", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Task deleted successfully",
	})
}

// GetTaskDetail get task detail
func (h *RecordHandler) GetTaskDetail(c *gin.Context) {
	taskID := c.Param("task_id")

	var task models.ArticleTask
	if err := h.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": 404,
				"msg":  "Task not found",
				"data": nil,
			})
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	var editTask models.ArticleEditTask
	err := h.db.Where("article_task_id = ?", task.ID).Order("created_at DESC").First(&editTask).Error

	content := task.Content
	theme := "default"
	if task.Theme != nil {
		theme = *task.Theme
	}

	if err == nil {
		content = &editTask.Content
		theme = editTask.Theme
	} else if err == gorm.ErrRecordNotFound && task.Content != nil {
		editTask = models.ArticleEditTask{
			ID:            generateID(),
			ArticleTaskID: &task.ID,
			UserID:        task.UserID,
			Title:         *task.Title,
			Content:       *task.Content,
			Theme:         theme,
			Status:        "editing",
		}
		h.db.Create(&editTask)
	}

	sectionHTML := ""
	if content != nil {
		sectionHTML, _ = h.markdownProcessor.ProcessMarkdown(*content, theme)
	}

	thumbnailURL := ""
	if content != nil {
		thumbnailURL = tools.ExtractFirstImageFromMarkdown(*content)
	}
	if thumbnailURL == "" && task.Images != nil {
		var imagesMap map[string]interface{}
		if err := json.Unmarshal([]byte(*task.Images), &imagesMap); err == nil {
			for _, value := range imagesMap {
				if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
					if dict, ok := arr[0].(map[string]interface{}); ok {
						if imageURL, exists := dict["imageUrl"]; exists {
							if url, ok := imageURL.(string); ok {
								thumbnailURL = url
								break
							}
						}
					}
				}
			}
		}
	}

	endTime := ""
	if task.EndTime != nil {
		endTime = task.EndTime.Format("2006-01-02 15:04:05")
	}

	var editTaskID *string
	if editTask.ID != "" {
		editTaskID = &editTask.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Success",
		"data": gin.H{
			"id":                task.ID,
			"client_id":         task.ClientID,
			"user_id":           task.UserID,
			"topic":             task.Topic,
			"author_name":       task.AuthorName,
			"status":            task.Status,
			"current_step":      task.CurrentStep,
			"steps":             task.Steps,
			"title":             task.Title,
			"snippet":           task.Snippet,
			"section_html":      sectionHTML,
			"content":           content,
			"theme":             theme,
			"thumbnail_url":     thumbnailURL,
			"edit_task_id":      editTaskID,
			"edit_theme":        editTask.Theme,
			"word_count":        task.WordCount,
			"images":            task.Images,
			"is_published":      task.IsPublished,
			"publish_url":       task.PublishURL,
			"start_time":        task.StartTime.Format("2006-01-02 15:04:05"),
			"end_time":          endTime,
			"created_at":        task.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":        task.UpdatedAt.Format("2006-01-02 15:04:05"),
			"total_duration":    task.TotalDuration,
			"search_duration":   task.SearchDuration,
			"parse_duration":    task.ParseDuration,
			"generate_duration": task.GenerateDuration,
			"complete_duration": task.CompleteDuration,
		},
	})
}

// ======= Topic related handlers =======

// TopicListItem topic list item
type TopicListItem struct {
	ID           string  `json:"id"`
	RelatedTask  *string `json:"related_task"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	AuthorName   *string `json:"author_name"`
	Category     *string `json:"category"`
	Tags         *string `json:"tags"`
	Status       *string `json:"status"`
	PublishTime  string  `json:"publish_time"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	ViewCount    int     `json:"view_count"`
	LikeCount    int     `json:"like_count"`
	CommentCount int     `json:"comment_count"`
}

// GetTopicList get topic list
func (h *RecordHandler) GetTopicList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.ArticleTopic{}).Where("user_id = ?", userID)

	if startDate != "" && endDate != "" {
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
				query = query.Where("publish_time >= ? AND publish_time <= ?", startTime, endTime)
			}
		}
	}

	if status == "expired" {
		query = query.Where("status = ? OR (publish_time < ? AND status NOT IN ?)",
			"expired", time.Now(), []string{"expired", "created", "pushed", "published"})
	} else if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var topics []models.ArticleTopic
	if err := query.Order("publish_time ASC").Offset(offset).Limit(pageSize).Find(&topics).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	currentTime := time.Now()
	type topicWithDiff struct {
		topic    models.ArticleTopic
		timeDiff float64
	}
	topicsWithDiff := make([]topicWithDiff, 0, len(topics))

	for i := range topics {
		topic := &topics[i]

		if topic.Status != nil && *topic.Status == "scheduled" {
			if topic.PublishTime != nil && topic.PublishTime.Add(24*time.Hour).Before(currentTime) {
				expired := "expired"
				topic.Status = &expired
				h.db.Model(topic).Update("status", "expired")
			}
		}

		timeDiff := float64(999999999)
		if topic.PublishTime != nil {
			duration := currentTime.Sub(*topic.PublishTime)
			if duration < 0 {
				duration = -duration
			}
			timeDiff = duration.Seconds()
		}

		topicsWithDiff = append(topicsWithDiff, topicWithDiff{
			topic:    *topic,
			timeDiff: timeDiff,
		})
	}

	// Sort by time diff
	for i := 0; i < len(topicsWithDiff)-1; i++ {
		for j := i + 1; j < len(topicsWithDiff); j++ {
			if topicsWithDiff[i].timeDiff > topicsWithDiff[j].timeDiff {
				topicsWithDiff[i], topicsWithDiff[j] = topicsWithDiff[j], topicsWithDiff[i]
			}
		}
	}

	items := make([]TopicListItem, 0, len(topicsWithDiff))
	for _, td := range topicsWithDiff {
		topic := td.topic
		publishTime := ""
		if topic.PublishTime != nil {
			publishTime = topic.PublishTime.Format(time.RFC3339)
		}

		items = append(items, TopicListItem{
			ID:           topic.ID,
			RelatedTask:  topic.RelatedTask,
			Title:        topic.Title,
			Description:  topic.Description,
			AuthorName:   topic.AuthorName,
			Category:     topic.Category,
			Tags:         topic.Tags,
			Status:       topic.Status,
			PublishTime:  publishTime,
			CreatedAt:    topic.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    topic.UpdatedAt.Format(time.RFC3339),
			ViewCount:    topic.ViewCount,
			LikeCount:    topic.LikeCount,
			CommentCount: topic.CommentCount,
		})
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Success",
		"data": gin.H{
			"items":       items,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// DeleteTopic delete topic
func (h *RecordHandler) DeleteTopic(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Where("id = ?", id).Delete(&models.ArticleTopic{}).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Delete failed: %v", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "delete success",
		"data": id,
	})
}

// DeleteTopicsRequest batch delete topics request
type DeleteTopicsRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// DeleteTopicsBatch batch delete topics
func (h *RecordHandler) DeleteTopicsBatch(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req DeleteTopicsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "Invalid parameters"))
		return
	}

	if err := h.db.Where("id IN ? AND user_id = ?", req.IDs, userID).Delete(&models.ArticleTopic{}).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Delete failed: %v", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  fmt.Sprintf("Successfully removed %d topics", len(req.IDs)),
	})
}

// ClearTopicsRequest clear topics request
type ClearTopicsRequest struct {
	Status string `json:"status" binding:"required"`
}

// ClearTopics clear topics by status
func (h *RecordHandler) ClearTopics(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req ClearTopicsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "Invalid parameters"))
		return
	}

	statusMap := map[string]string{
		"expired":   "Expired",
		"scheduled": "Scheduled",
		"pushed":    "Pushed to draft",
		"failed":    "Failed",
		"created":   "To be published",
		"published": "Published",
	}

	if err := h.db.Where("status = ? AND user_id = ?", req.Status, userID).Delete(&models.ArticleTopic{}).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Delete failed: %v", err)))
		return
	}

	statusText := statusMap[req.Status]
	if statusText == "" {
		statusText = req.Status
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  fmt.Sprintf("Successfully removed all %s topics", statusText),
	})
}

// GetAnalysisTaskData get task analysis data
func (h *RecordHandler) GetAnalysisTaskData(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	isFilter := c.Query("is_Filter")

	query := h.db.Model(&models.ArticleTopic{}).Where("user_id = ?", userID)

	if isFilter == "true" {
		if startDate != "" {
			if start, err := time.Parse("2006-01-02", startDate); err == nil {
				query = query.Where("publish_date >= ?", start)
			}
		}
		if endDate != "" {
			if end, err := time.Parse("2006-01-02", endDate); err == nil {
				query = query.Where("publish_date <= ?", end)
			}
		}
	}

	var total int64
	query.Count(&total)

	var expiredCount int64
	h.db.Model(&models.ArticleTopic{}).Where("user_id = ?", userID).
		Where("status = ? OR (publish_time < ? AND status NOT IN ?)",
			"expired", time.Now(), []string{"expired", "created", "pushed", "published"}).
		Count(&expiredCount)

	var scheduledCount, failedCount, completedCount, publishedCount int64
	query.Where("status = ?", "scheduled").Count(&scheduledCount)
	query.Where("status = ?", "failed").Count(&failedCount)
	query.Where("status = ?", "created").Count(&completedCount)
	query.Where("status = ?", "published").Count(&publishedCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"total":           total,
			"expired_count":   expiredCount,
			"scheduled_count": scheduledCount,
			"failed_count":    failedCount,
			"completed_count": completedCount,
			"published_count": publishedCount,
		},
	})
}

// ======= Hot Topics handlers =======

// HotTopicsRequest hot topics request
type HotTopicsRequest struct {
	SourceNames    []string `json:"source_names"`
	LimitPerSource int      `json:"limit_per_source"`
}

// GetHotTopics get hot topics
func (h *RecordHandler) GetHotTopics(c *gin.Context) {
	var req HotTopicsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.SourceNames = []string{}
		req.LimitPerSource = 10
	}

	// TODO: Implement Redis hot topics logic
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Get hot topics from Redis successfully",
		"data": []interface{}{},
	})
}

// ======= Copilot Chat handlers =======

// CopilotMessage copilot message
type CopilotMessage struct {
	Role             string  `json:"role"`
	Content          string  `json:"content"`
	SessionID        string  `json:"session_id"`
	WorkflowID       string  `json:"workflow_id"`
	SessionCreatedAt *string `json:"session_created_at"`
	SessionStatus    string  `json:"session_status"`
}

// GetCopilotChatHistory get copilot chat history
func (h *RecordHandler) GetCopilotChatHistory(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	threadID := c.Param("thread_id")

	var sessions []models.CopilotChatSession
	if err := h.db.Where("thread_id = ? AND user_id = ?", threadID, userID).
		Order("created_at ASC").
		Find(&sessions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	if len(sessions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "No chat history found",
			"data": gin.H{
				"thread_id":      threadID,
				"messages":       []interface{}{},
				"total_sessions": 0,
			},
		})
		return
	}

	allMessages := make([]CopilotMessage, 0)

	for _, session := range sessions {
		if session.AiResponse != nil {
			var sessionState map[string]interface{}
			if err := json.Unmarshal([]byte(*session.AiResponse), &sessionState); err == nil {
				if messages, ok := sessionState["messages"].([]interface{}); ok {
					for _, msg := range messages {
						if msgMap, ok := msg.(map[string]interface{}); ok {
							role, _ := msgMap["role"].(string)
							content, _ := msgMap["content"].(string)

							// Filter out assistant messages starting with #
							if role == "assistant" && len(content) > 0 && content[0] == '#' {
								continue
							}

							createdAtStr := ""
							if !session.CreatedAt.IsZero() {
								createdAtStr = session.CreatedAt.Format(time.RFC3339)
							}

							allMessages = append(allMessages, CopilotMessage{
								Role:             role,
								Content:          content,
								SessionID:        session.ID,
								WorkflowID:       session.WorkflowID,
								SessionCreatedAt: &createdAtStr,
								SessionStatus:    string(session.Status),
							})
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Get chat history successfully",
		"data": gin.H{
			"thread_id":      threadID,
			"messages":       allMessages,
			"total_sessions": len(sessions),
			"session_count":  len(sessions),
		},
	})
}

// CopilotSessionItem copilot session item
type CopilotSessionItem struct {
	ID              string  `json:"id"`
	ThreadID        string  `json:"thread_id"`
	WorkflowID      string  `json:"workflow_id"`
	ClientID        *string `json:"client_id"`
	UserQuery       string  `json:"user_query"`
	AiResponse      *string `json:"ai_response"`
	Status          string  `json:"status"`
	Feedback        int     `json:"feedback"`
	FeedbackContent *string `json:"feedback_content"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	CompletedAt     *string `json:"completed_at"`
}

// GetCopilotSessions get copilot sessions
func (h *RecordHandler) GetCopilotSessions(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.CopilotChatSession{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var sessions []models.CopilotChatSession
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err)))
		return
	}

	items := make([]CopilotSessionItem, 0, len(sessions))
	for _, session := range sessions {
		var completedAtPtr *string
		if session.CompletedAt != nil {
			completedAt := session.CompletedAt.Format(time.RFC3339)
			completedAtPtr = &completedAt
		}

		items = append(items, CopilotSessionItem{
			ID:              session.ID,
			ThreadID:        session.ThreadID,
			WorkflowID:      session.WorkflowID,
			ClientID:        session.ClientID,
			UserQuery:       session.UserQuery,
			AiResponse:      session.AiResponse,
			Status:          string(session.Status),
			Feedback:        session.Feedback,
			FeedbackContent: session.FeedbackContent,
			CreatedAt:       session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
			CompletedAt:     completedAtPtr,
		})
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Get sessions successfully",
		"data": gin.H{
			"items":       items,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// generateID generate ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

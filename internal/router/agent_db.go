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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentDBHandler agent database handler
type AgentDBHandler struct {
	db *gorm.DB
}

// NewAgentDBHandler create agent database handler
func NewAgentDBHandler() *AgentDBHandler {
	return &AgentDBHandler{
		db: repository.DB,
	}
}

// SetupAgentDBRoutes setup agent database routes
func SetupAgentDBRoutes(r *gin.Engine) {
	handler := NewAgentDBHandler()

	group := r.Group("/api/v1/agent/db")
	{
		// Thread routes
		group.POST("/threads", middleware.JWTAuth(), handler.CreateThread)
		group.GET("/threads/list", middleware.JWTAuth(), handler.GetThreadList)
		group.PUT("/threads/:thread_id", middleware.JWTAuth(), handler.UpdateThread)
		group.DELETE("/threads/:thread_id", middleware.JWTAuth(), handler.DeleteThread)

		// Session routes
		group.GET("/sessions/thread/:thread_id", middleware.JWTAuth(), handler.GetSessionsByThread)
		group.DELETE("/sessions/:workflow_id", middleware.JWTAuth(), handler.DeleteSession)

		// Workflow record routes
		group.GET("/workflow-records/thread/:thread_id/articles", middleware.JWTAuth(), handler.GetWorkflowArticlesByThread)

		// Feedback routes
		group.GET("/sessions/:workflow_id/feedback", middleware.JWTAuth(), handler.GetSessionFeedback)
		group.POST("/sessions/:workflow_id/feedback", middleware.JWTAuth(), handler.SubmitSessionFeedback)
	}
}

// ========================= Request/Response Models =========================

// CreateThreadRequest create thread request
type CreateThreadRequest struct {
	Label *string              `json:"label"`
	Scene *models.CopilotScene `json:"scene"`
}

// UpdateThreadRequest update thread request
type UpdateThreadRequest struct {
	Label string `json:"label" binding:"required"`
}

// SubmitFeedbackRequest submit feedback request
type SubmitFeedbackRequest struct {
	Feedback        int     `json:"feedback" binding:"required"` // 1-赞，-1-踩，0-取消反馈
	FeedbackContent *string `json:"feedback_content"`
}

// ThreadResponse thread response
type ThreadResponse struct {
	ID             string               `json:"id"`
	ThreadID       string               `json:"thread_id"`
	Label          string               `json:"label"`
	Scene          *models.CopilotScene `json:"scene"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
	CleanedThreads int                  `json:"cleaned_threads,omitempty"`
	CleanedOld     int                  `json:"cleaned_old,omitempty"`
	CleanedToday   int                  `json:"cleaned_today,omitempty"`
}

// SessionResponse session response
type SessionResponse struct {
	ID              string                  `json:"id"`
	UserID          string                  `json:"user_id"`
	ThreadID        string                  `json:"thread_id"`
	WorkflowID      string                  `json:"workflow_id"`
	ClientID        *string                 `json:"client_id"`
	UserQuery       string                  `json:"user_query"`
	AiResponse      interface{}             `json:"ai_response,omitempty"`
	Status          string                  `json:"status"`
	Feedback        int                     `json:"feedback"`
	FeedbackContent *string                 `json:"feedback_content"`
	WorkflowRecord  *WorkflowRecordResponse `json:"workflow_record,omitempty"`
	TokenUsage      *TokenUsageResponse     `json:"token_usage,omitempty"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at"`
	CompletedAt     *string                 `json:"completed_at"`
}

// WorkflowRecordResponse workflow record response
type WorkflowRecordResponse struct {
	Config         interface{} `json:"config,omitempty"`
	WorkflowData   interface{} `json:"workflow_data,omitempty"`
	TopicContent   interface{} `json:"topic_content,omitempty"`
	ArticleContent string      `json:"article_content"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
}

// TokenUsageResponse token usage response
type TokenUsageResponse struct {
	TotalInputTokens  int                    `json:"total_input_tokens"`
	TotalOutputTokens int                    `json:"total_output_tokens"`
	TotalTokens       int                    `json:"total_tokens"`
	TotalCost         float64                `json:"total_cost"`
	ModelCount        int                    `json:"model_count"`
	SessionCount      int                    `json:"session_count"`
	PrimaryModel      *string                `json:"primary_model"`
	ModelBreakdown    map[string]interface{} `json:"model_breakdown"`
	CreatedAt         string                 `json:"created_at"`
}

// ArticleRecordResponse article record response
type ArticleRecordResponse struct {
	ThreadID       string `json:"thread_id"`
	WorkflowID     string `json:"workflow_id"`
	ArticleContent string `json:"article_content"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// FeedbackResponse feedback response
type FeedbackResponse struct {
	SessionID       string  `json:"session_id"`
	WorkflowID      string  `json:"workflow_id"`
	Feedback        int     `json:"feedback"`
	FeedbackContent *string `json:"feedback_content"`
	UpdatedAt       string  `json:"updated_at"`
}

// ========================= Thread Handlers =========================

// CreateThread create new thread
func (h *AgentDBHandler) CreateThread(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供body，使用默认值
		req = CreateThreadRequest{}
	}

	// 获取今天0点的时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 清理规则1：删除昨天及更早的所有空对话线程（看到就清空）
	var oldEmptyThreads []models.CopilotChatThread
	deletedCountOld := 0
	if err := h.db.Where("user_id = ? AND label LIKE ? AND created_at < ?",
		userID, "新对话%", todayStart).Find(&oldEmptyThreads).Error; err == nil {
		for _, thread := range oldEmptyThreads {
			if err := h.db.Delete(&thread).Error; err != nil {
				fmt.Printf("删除历史空对话线程失败 %s: %v\n", thread.ThreadID, err)
			} else {
				deletedCountOld++
			}
		}
	}

	// 清理规则2：当今天的空对话达到或超过10个时，清空所有今天的空对话
	var todayEmptyThreads []models.CopilotChatThread
	deletedCountToday := 0
	if err := h.db.Where("user_id = ? AND label LIKE ? AND created_at >= ?",
		userID, "新对话%", todayStart).Find(&todayEmptyThreads).Error; err == nil {
		// 当今天的空对话达到或超过10个时，清空所有今天的空对话
		if len(todayEmptyThreads) >= 10 {
			for _, thread := range todayEmptyThreads {
				if err := h.db.Delete(&thread).Error; err != nil {
					fmt.Printf("删除今天的空对话线程失败 %s: %v\n", thread.ThreadID, err)
				} else {
					deletedCountToday++
				}
			}
		}
	}

	totalDeleted := deletedCountOld + deletedCountToday

	// 生成唯一的thread_id
	threadID := uuid.New().String()

	// 使用传入的label，如果没有则使用默认名称
	label := req.Label
	if label == nil || *label == "" {
		defaultLabel := fmt.Sprintf("新对话 %s", threadID)
		label = &defaultLabel
	}

	// 设置默认scene
	scene := req.Scene
	if scene == nil {
		defaultScene := models.CopilotSceneContext
		scene = &defaultScene
	}

	// 创建线程记录
	thread := &models.CopilotChatThread{
		ID:       uuid.New().String(),
		UserID:   userID,
		ThreadID: threadID,
		Label:    label,
		Scene:    scene,
	}

	if err := h.db.Create(thread).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("创建线程失败: %v", err)))
		return
	}

	middleware.Success(c, "创建线程成功", gin.H{
		"id":              thread.ID,
		"thread_id":       thread.ThreadID,
		"label":           *thread.Label,
		"scene":           *thread.Scene,
		"created_at":      thread.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":      thread.UpdatedAt.Format("2006-01-02 15:04:05"),
		"cleaned_threads": totalDeleted,
		"cleaned_old":     deletedCountOld,
		"cleaned_today":   deletedCountToday,
	})
}

// GetThreadList get thread list with pagination
func (h *AgentDBHandler) GetThreadList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	sceneStr := c.Query("scene")

	// 构建查询条件
	query := h.db.Model(&models.CopilotChatThread{}).Where("user_id = ?", userID)
	if sceneStr != "" {
		scene := models.CopilotScene(sceneStr)
		query = query.Where("scene = ?", scene)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	var threads []models.CopilotChatThread
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&threads).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 格式化数据列表
	items := make([]map[string]interface{}, 0, len(threads))
	for _, thread := range threads {
		item := map[string]interface{}{
			"id":         thread.ID,
			"thread_id":  thread.ThreadID,
			"label":      "",
			"scene":      nil,
			"created_at": thread.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": thread.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if thread.Label != nil {
			item["label"] = *thread.Label
		}
		if thread.Scene != nil {
			item["scene"] = *thread.Scene
		}
		items = append(items, item)
	}

	middleware.Success(c, "获取线程列表成功", gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateThread update thread name
func (h *AgentDBHandler) UpdateThread(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	threadID := c.Param("thread_id")

	var req UpdateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "参数错误"))
		return
	}

	// 查询线程记录
	var thread models.CopilotChatThread
	if err := h.db.Where("thread_id = ? AND user_id = ?", threadID, userID).First(&thread).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "线程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 更新名称
	thread.Label = &req.Label
	if err := h.db.Save(&thread).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	middleware.Success(c, "更新成功", gin.H{
		"id":         thread.ID,
		"thread_id":  thread.ThreadID,
		"label":      *thread.Label,
		"updated_at": thread.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteThread delete thread record
func (h *AgentDBHandler) DeleteThread(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	threadID := c.Param("thread_id")

	// 查询线程记录
	var thread models.CopilotChatThread
	if err := h.db.Where("thread_id = ? AND user_id = ?", threadID, userID).First(&thread).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "线程不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 删除关联的会话记录
	h.db.Where("thread_id = ? AND user_id = ?", threadID, userID).Delete(&models.CopilotChatSession{})
	// 删除关联的工作流记录 - 根据Python代码逻辑，使用workflow_id=thread_id
	h.db.Where("workflow_id = ? AND user_id = ?", threadID, userID).Delete(&models.CopilotWorkflowRecord{})
	// 删除关联的Token使用记录 - 根据Python代码逻辑，使用workflow_id=thread_id
	h.db.Where("workflow_id = ? AND user_id = ?", threadID, userID).Delete(&models.TokenUsageRecord{})

	// 删除线程记录
	if err := h.db.Delete(&thread).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "删除成功", nil)
}

// ========================= Session Handlers =========================

// GetSessionsByThread get sessions by thread_id with pagination
func (h *AgentDBHandler) GetSessionsByThread(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	threadID := c.Param("thread_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 基础查询条件
	query := h.db.Model(&models.CopilotChatSession{}).
		Where("thread_id = ? AND user_id = ? AND status = ?", threadID, userID, models.WorkflowStatusCompleted)

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	var sessions []models.CopilotChatSession
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 格式化响应
	items := make([]SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		// 获取工作流记录
		var workflowRecord models.CopilotWorkflowRecord
		var workflowData *WorkflowRecordResponse
		if err := h.db.Where("workflow_id = ? AND user_id = ?", session.WorkflowID, userID).First(&workflowRecord).Error; err == nil {
			workflowData = &WorkflowRecordResponse{
				ArticleContent: "",
				CreatedAt:      workflowRecord.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:      workflowRecord.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
			if workflowRecord.ArticleContent != nil {
				workflowData.ArticleContent = *workflowRecord.ArticleContent
			}
			if workflowRecord.Config != nil {
				json.Unmarshal([]byte(*workflowRecord.Config), &workflowData.Config)
			}
			if workflowRecord.WorkflowData != nil {
				json.Unmarshal([]byte(*workflowRecord.WorkflowData), &workflowData.WorkflowData)
			}
			if workflowRecord.TopicContent != nil {
				json.Unmarshal([]byte(*workflowRecord.TopicContent), &workflowData.TopicContent)
			}
		}

		// 获取Token使用记录
		var tokenRecord models.TokenUsageRecord
		var tokenData *TokenUsageResponse
		if err := h.db.Where("workflow_id = ? AND user_id = ?", session.WorkflowID, userID).First(&tokenRecord).Error; err == nil {
			modelBreakdown := make(map[string]interface{})
			if tokenRecord.ModelBreakdown != nil {
				json.Unmarshal([]byte(*tokenRecord.ModelBreakdown), &modelBreakdown)
			}
			tokenData = &TokenUsageResponse{
				TotalInputTokens:  tokenRecord.TotalInputTokens,
				TotalOutputTokens: tokenRecord.TotalOutputTokens,
				TotalTokens:       tokenRecord.TotalTokens,
				TotalCost:         tokenRecord.TotalCost,
				ModelCount:        tokenRecord.ModelCount,
				SessionCount:      tokenRecord.SessionCount,
				PrimaryModel:      tokenRecord.PrimaryModel,
				ModelBreakdown:    modelBreakdown,
				CreatedAt:         tokenRecord.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		// 解析AI响应
		var aiResponse interface{}
		if session.AiResponse != nil {
			json.Unmarshal([]byte(*session.AiResponse), &aiResponse)
		}

		completedAt := (*string)(nil)
		if session.CompletedAt != nil {
			completedAtStr := session.CompletedAt.Format("2006-01-02 15:04:05")
			completedAt = &completedAtStr
		}

		item := SessionResponse{
			ID:              session.ID,
			UserID:          session.UserID,
			ThreadID:        session.ThreadID,
			WorkflowID:      session.WorkflowID,
			ClientID:        session.ClientID,
			UserQuery:       session.UserQuery,
			AiResponse:      aiResponse,
			Status:          string(session.Status),
			Feedback:        session.Feedback,
			FeedbackContent: session.FeedbackContent,
			WorkflowRecord:  workflowData,
			TokenUsage:      tokenData,
			CreatedAt:       session.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       session.UpdatedAt.Format("2006-01-02 15:04:05"),
			CompletedAt:     completedAt,
		}
		items = append(items, item)
	}

	middleware.Success(c, "获取会话列表成功", gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteSession delete session record
func (h *AgentDBHandler) DeleteSession(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	workflowID := c.Param("workflow_id")

	// 查询会话记录
	var session models.CopilotChatSession
	if err := h.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "会话不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 删除会话记录
	if err := h.db.Delete(&session).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "删除会话记录成功", gin.H{
		"data": workflowID,
	})
}

// ========================= Workflow Record Handlers =========================

// GetWorkflowArticlesByThread get workflow articles by thread_id with pagination
func (h *AgentDBHandler) GetWorkflowArticlesByThread(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	threadID := c.Param("thread_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	// 限制page_size最大值，避免一次性返回太多大字段数据
	if pageSize > 50 {
		pageSize = 50
	}
	if page < 1 {
		page = 1
	}

	// 基础查询条件 - 只查询必要的字段，避免加载大的JSON字段（config, workflow_data, topic_content）
	baseQuery := h.db.Model(&models.CopilotWorkflowRecord{}).
		Where("thread_id = ? AND user_id = ?", threadID, userID).
		Where("article_content IS NOT NULL AND article_content != ?", "")

	// 获取总数 - 使用独立的查询，避免状态污染
	var total int64
	baseQuery.Count(&total)

	// 分页查询 - 只选择必要的字段，大幅提升性能
	offset := (page - 1) * pageSize
	var records []models.CopilotWorkflowRecord
	if err := baseQuery.
		Select("thread_id", "workflow_id", "article_content", "created_at", "updated_at").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 格式化响应数据
	items := make([]ArticleRecordResponse, 0, len(records))
	for _, record := range records {
		articleContent := ""
		if record.ArticleContent != nil {
			articleContent = *record.ArticleContent
		}
		items = append(items, ArticleRecordResponse{
			ThreadID:       record.ThreadID,
			WorkflowID:     record.WorkflowID,
			ArticleContent: articleContent,
			CreatedAt:      record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      record.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "获取工作流文章记录成功", gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ========================= Feedback Handlers =========================

// GetSessionFeedback get session feedback
func (h *AgentDBHandler) GetSessionFeedback(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	workflowID := c.Param("workflow_id")

	// 查询会话记录
	var session models.CopilotChatSession
	if err := h.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "会话不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	middleware.Success(c, "获取反馈信息成功", gin.H{
		"session_id":       session.ID,
		"workflow_id":      session.WorkflowID,
		"feedback":         session.Feedback,
		"feedback_content": session.FeedbackContent,
		"updated_at":       session.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// SubmitSessionFeedback submit session feedback
func (h *AgentDBHandler) SubmitSessionFeedback(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	workflowID := c.Param("workflow_id")

	var req SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "参数错误"))
		return
	}

	// 查询会话记录
	var session models.CopilotChatSession
	if err := h.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "会话不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 更新反馈信息
	session.Feedback = req.Feedback

	// 如果有新的反馈内容，采用追加模式
	if req.FeedbackContent != nil && *req.FeedbackContent != "" {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		newContent := fmt.Sprintf("--- %s ---\n%s", timestamp, *req.FeedbackContent)
		if session.FeedbackContent != nil && *session.FeedbackContent != "" {
			// 追加新的反馈内容，用分隔符分开
			*session.FeedbackContent = *session.FeedbackContent + "\n" + newContent
		} else {
			// 第一次添加反馈内容
			session.FeedbackContent = &newContent
		}
	}

	if err := h.db.Save(&session).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("保存失败: %v", err)))
		return
	}

	middleware.Success(c, "反馈提交成功", gin.H{
		"session_id":       session.ID,
		"workflow_id":      session.WorkflowID,
		"feedback":         session.Feedback,
		"feedback_content": session.FeedbackContent,
		"updated_at":       session.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

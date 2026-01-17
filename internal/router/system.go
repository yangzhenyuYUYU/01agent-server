package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemHandler system handler
type SystemHandler struct {
	db *gorm.DB
}

// NewSystemHandler create system handler
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		db: repository.DB,
	}
}

// ========================= Notification Constants =========================

const (
	NotificationStatusUnpublished = "unpublished"
	NotificationStatusUnread     = "unread"
	NotificationStatusRead       = "read"
	NotificationStatusDeleted    = "deleted"
)

const (
	NotificationTypeSystem   = "system"
	NotificationTypeUpdate   = "update"
	NotificationTypeActivity = "activity"
	NotificationTypeFeedback = "feedback"
	NotificationTypePayment  = "payment"
	NotificationTypeOther    = "other"
)

const (
	FeedbackStatusPending = "pending"
	FeedbackStatusProcessing = "processing"
	FeedbackStatusResolved = "resolved"
	FeedbackStatusClosed   = "closed"
)

// ========================= Helper Functions =========================

// verifyAdmin 验证管理员权限
func (h *SystemHandler) verifyAdmin(c *gin.Context) bool {
	userID, _ := middleware.GetCurrentUserID(c)
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return false
	}
	return user.Role == 3
}

// ========================= Feedback Handlers =========================

// CreateFeedback create feedback
func (h *SystemHandler) CreateFeedback(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req struct {
		Title       string   `json:"title" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		Type        string   `json:"type"`
		ContactInfo *string  `json:"contact_info"`
		Images      []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	if req.Type == "" {
		req.Type = "other"
	}

	feedbackID := uuid.New().String()
	imagesJSON, _ := json.Marshal(req.Images)
	imagesStr := string(imagesJSON)

	feedback := models.Feedback{
		FeedbackID:  feedbackID,
		UserID:      &userID,
		Title:       &req.Title,
		Content:     req.Content,
		Type:        req.Type,
		ContactInfo: req.ContactInfo,
		Images:      &imagesStr,
		Status:      FeedbackStatusPending,
	}

	if err := h.db.Create(&feedback).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
		return
	}

	middleware.Success(c, "反馈提交成功", gin.H{
		"feedback_id": feedbackID,
	})
}

// GetFeedbackList get feedback list
func (h *SystemHandler) GetFeedbackList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var total int64
	h.db.Model(&models.Feedback{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * limit
	var feedbackList []models.Feedback
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&feedbackList).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	data := make([]map[string]interface{}, 0, len(feedbackList))
	for _, item := range feedbackList {
		data = append(data, map[string]interface{}{
			"feedback_id": item.FeedbackID,
			"title":       item.Title,
			"content":     item.Content,
			"type":        item.Type,
			"status":      item.Status,
			"created_at":  item.CreatedAt.Format(time.RFC3339),
			"updated_at":   item.UpdatedAt.Format(time.RFC3339),
		})
	}

	middleware.Success(c, "获取反馈列表成功", gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"list":  data,
	})
}

// GetFeedbackDetail get feedback detail
func (h *SystemHandler) GetFeedbackDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	feedbackID := c.Param("feedback_id")

	var feedback models.Feedback
	if err := h.db.Where("feedback_id = ? AND user_id = ?", feedbackID, userID).First(&feedback).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "反馈不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	var images []string
	if feedback.Images != nil {
		json.Unmarshal([]byte(*feedback.Images), &images)
	}

	middleware.Success(c, "获取反馈详情成功", gin.H{
		"feedback_id": feedback.FeedbackID,
		"title":       feedback.Title,
		"content":     feedback.Content,
		"type":        feedback.Type,
		"contact_info": feedback.ContactInfo,
		"images":      images,
		"status":      feedback.Status,
		"created_at":  feedback.CreatedAt.Format(time.RFC3339),
		"updated_at":   feedback.UpdatedAt.Format(time.RFC3339),
	})
}

// ========================= Notification Handlers =========================

// CreateNotification create notification (admin only)
func (h *SystemHandler) CreateNotification(c *gin.Context) {
	if !h.verifyAdmin(c) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusForbidden, "需要管理员权限"))
		return
	}

	var req struct {
		Title       string     `json:"title" binding:"required"`
		Content     string     `json:"content" binding:"required"`
		Type        string     `json:"type"`
		UserID      *string    `json:"user_id"`
		Link        *string    `json:"link"`
		IsImportant bool       `json:"is_important"`
		ExpireTime  *string    `json:"expire_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	if req.Type == "" {
		req.Type = NotificationTypeSystem
	}

	var expireTime *time.Time
	if req.ExpireTime != nil {
		expireTimeStr := strings.Replace(*req.ExpireTime, "Z", "+00:00", 1)
		if t, err := time.Parse(time.RFC3339, expireTimeStr); err == nil {
			expireTime = &t
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "过期时间格式错误，请使用ISO格式"))
			return
		}
	}

	notificationID := uuid.New().String()
	notification := models.SystemNotification{
		NotificationID: notificationID,
		UserID:          req.UserID,
		Type:            req.Type,
		Title:           req.Title,
		Content:         req.Content,
		Link:            req.Link,
		IsImportant:     req.IsImportant,
		Status:          NotificationStatusUnpublished,
		ExpireTime:      expireTime,
	}

	if err := h.db.Create(&notification).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
		return
	}

	middleware.Success(c, "通知创建成功", gin.H{
		"notification_id": notificationID,
		"status":          notification.Status,
	})
}

// GetNotificationList get notification list
func (h *SystemHandler) GetNotificationList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	typeFilter := c.Query("type")
	statusFilter := c.Query("status") // read/unread
	isImportantStr := c.Query("is_important")

	// 构建基础查询条件
	// 对于指定用户的通知，排除已删除的；对于全体用户的通知，排除未发布的
	query := h.db.Model(&models.SystemNotification{}).
		Where("((user_id = ? AND status != ?) OR (user_id IS NULL AND status != ?))",
			userID, NotificationStatusDeleted, NotificationStatusUnpublished)

	if typeFilter != "" {
		validTypes := []string{NotificationTypeSystem, NotificationTypeUpdate, NotificationTypeActivity,
			NotificationTypeFeedback, NotificationTypePayment, NotificationTypeOther}
		isValid := false
		for _, vt := range validTypes {
			if typeFilter == vt {
				isValid = true
				break
			}
		}
		if !isValid {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest,
				fmt.Sprintf("无效的通知类型，支持的类型：%s", strings.Join(validTypes, ", "))))
			return
		}
		query = query.Where("type = ?", typeFilter)
	}

	if isImportantStr != "" {
		if isImportant, err := strconv.ParseBool(isImportantStr); err == nil {
			query = query.Where("is_important = ?", isImportant)
		}
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	var notificationList []models.SystemNotification
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notificationList).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 获取全体通知的用户状态记录
	globalNotificationIDs := make([]string, 0)
	personalNotificationIDs := make([]string, 0)
	for _, item := range notificationList {
		if item.UserID == nil {
			globalNotificationIDs = append(globalNotificationIDs, item.NotificationID)
		} else {
			personalNotificationIDs = append(personalNotificationIDs, item.NotificationID)
		}
	}

	userRecords := make(map[string]*models.NotificationUserRecord)
	if len(globalNotificationIDs) > 0 {
		var records []models.NotificationUserRecord
		h.db.Where("notification_id IN ? AND user_id = ?", globalNotificationIDs, userID).Find(&records)
		for i := range records {
			userRecords[records[i].NotificationID] = &records[i]
		}
	}

	// 批量获取指定用户通知的最新状态
	personalNotifications := make(map[string]*models.SystemNotification)
	if len(personalNotificationIDs) > 0 {
		var freshNotifications []models.SystemNotification
		h.db.Where("notification_id IN ?", personalNotificationIDs).Find(&freshNotifications)
		for i := range freshNotifications {
			personalNotifications[freshNotifications[i].NotificationID] = &freshNotifications[i]
		}
	}

	data := make([]map[string]interface{}, 0)
	for _, item := range notificationList {
		isRead := false
		var readTime *time.Time

		if item.UserID != nil {
			// 指定用户的通知
			freshNotif := personalNotifications[item.NotificationID]
			if freshNotif != nil {
				if freshNotif.Status == NotificationStatusDeleted {
					continue // 跳过已删除的
				}
				isRead = freshNotif.Status == NotificationStatusRead
				readTime = freshNotif.ReadTime
			} else {
				if item.Status == NotificationStatusDeleted {
					continue
				}
				isRead = item.Status == NotificationStatusRead
				readTime = item.ReadTime
			}
		} else {
			// 全体用户的通知
			userRecord := userRecords[item.NotificationID]
			if userRecord != nil {
				if userRecord.Status == models.NotificationUserStatusDeleted {
					continue // 跳过已删除的
				}
				if userRecord.Status == models.NotificationUserStatusRead {
					isRead = true
					readTime = userRecord.ReadTime
				}
			}
		}

		// 应用status筛选
		if statusFilter != "" {
			if statusFilter == "read" && !isRead {
				continue
			}
			if statusFilter == "unread" && isRead {
				continue
			}
		}

		readTimeStr := ""
		if readTime != nil {
			readTimeStr = readTime.Format(time.RFC3339)
		}

		data = append(data, map[string]interface{}{
			"notification_id": item.NotificationID,
			"title":           item.Title,
			"content":         item.Content,
			"type":            item.Type,
			"link":            item.Link,
			"is_important":    item.IsImportant,
			"status":          "read",
			"is_read":         isRead,
			"read_time":       readTimeStr,
			"created_at":      item.CreatedAt.Format(time.RFC3339),
		})
	}

	actualCount := len(data)

	middleware.Success(c, "获取通知列表成功", gin.H{
		"total":          total,
		"actual_count":   actualCount,
		"page":           page,
		"limit":          limit,
		"type_filter":    typeFilter,
		"status_filter":  statusFilter,
		"is_important_filter": isImportantStr,
		"list":           data,
	})
}

// GetNotificationListAdmin get notification list (admin)
func (h *SystemHandler) GetNotificationListAdmin(c *gin.Context) {
	if !h.verifyAdmin(c) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusForbidden, "需要管理员权限"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	statusFilter := c.Query("status")
	typeFilter := c.Query("type")

	query := h.db.Model(&models.SystemNotification{})

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if typeFilter != "" {
		validTypes := []string{NotificationTypeSystem, NotificationTypeUpdate, NotificationTypeActivity,
			NotificationTypeFeedback, NotificationTypePayment, NotificationTypeOther}
		isValid := false
		for _, vt := range validTypes {
			if typeFilter == vt {
				isValid = true
				break
			}
		}
		if !isValid {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest,
				fmt.Sprintf("无效的通知类型，支持的类型：%s", strings.Join(validTypes, ", "))))
			return
		}
		query = query.Where("type = ?", typeFilter)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	var notificationList []models.SystemNotification
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notificationList).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	data := make([]map[string]interface{}, 0, len(notificationList))
	for _, item := range notificationList {
		readTimeStr := ""
		if item.ReadTime != nil {
			readTimeStr = item.ReadTime.Format(time.RFC3339)
		}
		expireTimeStr := ""
		if item.ExpireTime != nil {
			expireTimeStr = item.ExpireTime.Format(time.RFC3339)
		}

		userIDStr := ""
		if item.UserID != nil {
			userIDStr = *item.UserID
		}

		data = append(data, map[string]interface{}{
			"notification_id": item.NotificationID,
			"user_id":         userIDStr,
			"title":           item.Title,
			"content":         item.Content,
			"type":            item.Type,
			"link":            item.Link,
			"is_important":    item.IsImportant,
			"status":          item.Status,
			"read_time":       readTimeStr,
			"expire_time":     expireTimeStr,
			"created_at":      item.CreatedAt.Format(time.RFC3339),
			"updated_at":       item.UpdatedAt.Format(time.RFC3339),
		})
	}

	middleware.Success(c, "获取通知列表成功", gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"list":  data,
	})
}

// PublishNotification publish notification (admin only)
func (h *SystemHandler) PublishNotification(c *gin.Context) {
	if !h.verifyAdmin(c) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusForbidden, "需要管理员权限"))
		return
	}

	notificationID := c.Param("notification_id")

	var notification models.SystemNotification
	if err := h.db.Where("notification_id = ?", notificationID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if notification.Status == NotificationStatusUnpublished {
		notification.Status = NotificationStatusUnread
		if err := h.db.Save(&notification).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
		middleware.Success(c, "通知发布成功", gin.H{
			"notification_id": notificationID,
			"status":          notification.Status,
		})
	} else if notification.Status == NotificationStatusUnread {
		middleware.Success(c, "通知已经是未读状态，无需发布", gin.H{
			"notification_id": notificationID,
			"status":          notification.Status,
		})
	} else {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest,
			fmt.Sprintf("当前通知状态为%s，无法发布", notification.Status)))
	}
}

// DeleteNotification delete notification (admin only)
func (h *SystemHandler) DeleteNotification(c *gin.Context) {
	if !h.verifyAdmin(c) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusForbidden, "需要管理员权限"))
		return
	}

	notificationID := c.Param("notification_id")

	var notification models.SystemNotification
	if err := h.db.Where("notification_id = ?", notificationID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&notification).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "通知删除成功", gin.H{
		"notification_id": notificationID,
		"deleted":         true,
	})
}

// GetNotificationDetail get notification detail
func (h *SystemHandler) GetNotificationDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	notificationID := c.Param("notification_id")

	var notification models.SystemNotification
	if err := h.db.Where("notification_id = ? AND status NOT IN ?",
		notificationID, []string{NotificationStatusUnpublished, NotificationStatusDeleted}).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 检查权限：如果是指定用户的通知，必须是当前用户
	if notification.UserID != nil && *notification.UserID != userID {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
		return
	}

	isRead := false
	var readTime *time.Time

	if notification.UserID != nil {
		// 指定用户的通知，更新父表状态
		if notification.Status == NotificationStatusUnread {
			notification.Status = NotificationStatusRead
			now := time.Now()
			notification.ReadTime = &now
			h.db.Save(&notification)
		}
		isRead = notification.Status == NotificationStatusRead
		readTime = notification.ReadTime
	} else {
		// 全体用户的通知，处理用户状态记录表
		var userRecord models.NotificationUserRecord
		if err := h.db.Where("notification_id = ? AND user_id = ?", notificationID, userID).First(&userRecord).Error; err != nil {
			// 创建已读记录
			recordID := uuid.New().String()
			now := time.Now()
			userRecord = models.NotificationUserRecord{
				RecordID:       recordID,
				NotificationID: notificationID,
				UserID:         userID,
				Status:         models.NotificationUserStatusRead,
				ReadTime:       &now,
			}
			h.db.Create(&userRecord)
		} else if userRecord.Status != models.NotificationUserStatusRead {
			// 更新为已读状态
			now := time.Now()
			userRecord.Status = models.NotificationUserStatusRead
			userRecord.ReadTime = &now
			h.db.Save(&userRecord)
		}
		isRead = true
		readTime = userRecord.ReadTime
	}

	readTimeStr := ""
	if readTime != nil {
		readTimeStr = readTime.Format(time.RFC3339)
	}

	middleware.Success(c, "获取通知详情成功", gin.H{
		"notification_id": notification.NotificationID,
		"title":           notification.Title,
		"content":         notification.Content,
		"type":            notification.Type,
		"link":            notification.Link,
		"is_important":    notification.IsImportant,
		"status":          "read",
		"is_read":         isRead,
		"created_at":      notification.CreatedAt.Format(time.RFC3339),
		"read_time":       readTimeStr,
	})
}

// MarkAllNotificationRead mark all notifications as read
func (h *SystemHandler) MarkAllNotificationRead(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	// 1. 标记指定用户的未读通知为已读
	now := time.Now()
	h.db.Model(&models.SystemNotification{}).
		Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":    NotificationStatusRead,
			"read_time": now,
		})

	// 2. 处理全体用户的通知
	var globalNotifications []models.SystemNotification
	h.db.Where("user_id IS NULL AND status NOT IN ?",
		[]string{NotificationStatusUnpublished, NotificationStatusDeleted}).Find(&globalNotifications)

	// 获取当前用户已有的状态记录
	var existingRecords []models.NotificationUserRecord
	h.db.Where("user_id = ?", userID).Find(&existingRecords)
	existingMap := make(map[string]bool)
	for _, record := range existingRecords {
		existingMap[record.NotificationID] = true
	}

	// 为未有记录的全体通知创建已读记录
	newRecords := make([]models.NotificationUserRecord, 0)
	for _, notification := range globalNotifications {
		if !existingMap[notification.NotificationID] {
			recordID := uuid.New().String()
			newRecords = append(newRecords, models.NotificationUserRecord{
				RecordID:       recordID,
				NotificationID: notification.NotificationID,
				UserID:         userID,
				Status:         models.NotificationUserStatusRead,
				ReadTime:       &now,
			})
		}
	}

	if len(newRecords) > 0 {
		h.db.Create(&newRecords)
	}

	middleware.Success(c, "标记所有通知为已读成功", nil)
}

// MarkNotificationRead mark notification as read
func (h *SystemHandler) MarkNotificationRead(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	notificationID := c.Param("notification_id")

	var notification models.SystemNotification
	if err := h.db.Where("notification_id = ? AND status NOT IN ?",
		notificationID, []string{NotificationStatusUnpublished, NotificationStatusDeleted}).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 检查权限：如果是指定用户的通知，必须是当前用户
	if notification.UserID != nil && *notification.UserID != userID {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
		return
	}

	if notification.UserID != nil {
		// 指定用户的通知，更新父表状态
		now := time.Now()
		notification.Status = NotificationStatusRead
		notification.ReadTime = &now
		h.db.Save(&notification)
	} else {
		// 全体用户的通知，创建或更新用户状态记录
		var userRecord models.NotificationUserRecord
		if err := h.db.Where("notification_id = ? AND user_id = ?", notificationID, userID).First(&userRecord).Error; err != nil {
			recordID := uuid.New().String()
			now := time.Now()
			userRecord = models.NotificationUserRecord{
				RecordID:       recordID,
				NotificationID: notificationID,
				UserID:         userID,
				Status:         models.NotificationUserStatusRead,
				ReadTime:       &now,
			}
			h.db.Create(&userRecord)
		} else {
			now := time.Now()
			userRecord.Status = models.NotificationUserStatusRead
			userRecord.ReadTime = &now
			h.db.Save(&userRecord)
		}
	}

	middleware.Success(c, "标记通知为已读成功", nil)
}

// DeleteUserNotification delete user notification
func (h *SystemHandler) DeleteUserNotification(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	notificationID := c.Param("notification_id")

	var notification models.SystemNotification
	if err := h.db.Where("notification_id = ? AND status NOT IN ?",
		notificationID, []string{NotificationStatusUnpublished, NotificationStatusDeleted}).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 检查权限：如果是指定用户的通知，必须是当前用户
	if notification.UserID != nil && *notification.UserID != userID {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "通知不存在"))
		return
	}

	if notification.UserID != nil {
		// 真删除
		h.db.Delete(&notification)
		h.db.Where("notification_id = ? AND user_id = ?", notificationID, userID).
			Delete(&models.NotificationUserRecord{})
	} else {
		// 全体用户的通知，在用户状态记录表中标记为DELETED
		var userRecord models.NotificationUserRecord
		if err := h.db.Where("notification_id = ? AND user_id = ?", notificationID, userID).First(&userRecord).Error; err != nil {
			recordID := uuid.New().String()
			now := time.Now()
			userRecord = models.NotificationUserRecord{
				RecordID:       recordID,
				NotificationID: notificationID,
				UserID:         userID,
				Status:         models.NotificationUserStatusDeleted,
				DeletedTime:    &now,
			}
			h.db.Create(&userRecord)
		} else {
			now := time.Now()
			userRecord.Status = models.NotificationUserStatusDeleted
			userRecord.DeletedTime = &now
			h.db.Save(&userRecord)
		}
	}

	middleware.Success(c, "通知删除成功", nil)
}

// BatchDeleteUserNotifications batch delete user notifications
func (h *SystemHandler) BatchDeleteUserNotifications(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	if len(req.NotificationIDs) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "通知ID列表不能为空"))
		return
	}

	// 先检查通知是否存在
	var allNotifications []models.SystemNotification
	h.db.Where("notification_id IN ?", req.NotificationIDs).Find(&allNotifications)

	if len(allNotifications) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest,
			fmt.Sprintf("通知不存在，请检查ID: %v", req.NotificationIDs)))
		return
	}

	// 筛选有效的通知
	validNotifications := make([]models.SystemNotification, 0)
	for _, notif := range allNotifications {
		if notif.Status != NotificationStatusUnpublished && notif.Status != NotificationStatusDeleted {
			validNotifications = append(validNotifications, notif)
		}
	}

	deletedCount := 0
	successIDs := make([]string, 0)
	failedIDs := make([]string, 0)

	for _, notification := range validNotifications {
		notificationIDStr := notification.NotificationID

		if notification.UserID != nil {
			// 指定用户的通知，检查权限
			if *notification.UserID == userID {
				h.db.Delete(&notification)
				h.db.Where("notification_id = ? AND user_id = ?", notificationIDStr, userID).
					Delete(&models.NotificationUserRecord{})
				deletedCount++
				successIDs = append(successIDs, notificationIDStr)
			} else {
				failedIDs = append(failedIDs, notificationIDStr)
			}
		} else {
			// 全体用户的通知，处理用户状态记录
			var userRecord models.NotificationUserRecord
			if err := h.db.Where("notification_id = ? AND user_id = ?", notificationIDStr, userID).First(&userRecord).Error; err != nil {
				recordID := uuid.New().String()
				now := time.Now()
				userRecord = models.NotificationUserRecord{
					RecordID:       recordID,
					NotificationID: notificationIDStr,
					UserID:         userID,
					Status:         models.NotificationUserStatusDeleted,
					DeletedTime:    &now,
				}
				h.db.Create(&userRecord)
			} else {
				now := time.Now()
				userRecord.Status = models.NotificationUserStatusDeleted
				userRecord.DeletedTime = &now
				h.db.Save(&userRecord)
			}
			deletedCount++
			successIDs = append(successIDs, notificationIDStr)
		}
	}

	// 检查哪些ID完全没找到
	foundIDs := make(map[string]bool)
	for _, notif := range allNotifications {
		foundIDs[notif.NotificationID] = true
	}
	for _, nid := range req.NotificationIDs {
		if !foundIDs[nid] {
			failedIDs = append(failedIDs, nid)
		}
	}

	successRate := "0%"
	if len(req.NotificationIDs) > 0 {
		successRate = fmt.Sprintf("%.1f%%", float64(deletedCount)/float64(len(req.NotificationIDs))*100)
	}

	middleware.Success(c, "批量删除处理完成", gin.H{
		"total_requested": len(req.NotificationIDs),
		"deleted_count":   deletedCount,
		"success_ids":     successIDs,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
		"success_rate":    successRate,
	})
}

// ClearAllReadNotifications clear all read notifications
func (h *SystemHandler) ClearAllReadNotifications(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	deletedCount := 0

	// 1. 处理指定用户的已读通知
	var personalReadNotifications []models.SystemNotification
	h.db.Where("user_id = ? AND status = ?", userID, NotificationStatusRead).Find(&personalReadNotifications)

	if len(personalReadNotifications) > 0 {
		personalIDs := make([]string, 0, len(personalReadNotifications))
		for _, notif := range personalReadNotifications {
			personalIDs = append(personalIDs, notif.NotificationID)
		}
		h.db.Model(&models.SystemNotification{}).
			Where("notification_id IN ?", personalIDs).
			Update("status", NotificationStatusDeleted)
		deletedCount += len(personalIDs)
	}

	// 2. 处理全体用户通知中用户已读的
	var userReadRecords []models.NotificationUserRecord
	h.db.Where("user_id = ? AND status = ?", userID, models.NotificationUserStatusRead).Find(&userReadRecords)

	if len(userReadRecords) > 0 {
		now := time.Now()
		for i := range userReadRecords {
			userReadRecords[i].Status = models.NotificationUserStatusDeleted
			userReadRecords[i].DeletedTime = &now
			h.db.Save(&userReadRecords[i])
		}
		deletedCount += len(userReadRecords)
	}

	middleware.Success(c, "清除所有已读通知成功", gin.H{
		"deleted_count":   deletedCount,
		"personal_deleted": len(personalReadNotifications),
		"global_deleted":  len(userReadRecords),
	})
}

// GetUnreadNotificationCount get unread notification count
func (h *SystemHandler) GetUnreadNotificationCount(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	// 1. 统计指定用户的未读通知数量
	var personalUnreadNotifications []models.SystemNotification
	h.db.Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Find(&personalUnreadNotifications)

	// 2. 获取所有全体用户的通知
	var globalNotifications []models.SystemNotification
	h.db.Where("user_id IS NULL AND status NOT IN ?",
		[]string{NotificationStatusUnpublished, NotificationStatusDeleted}).Find(&globalNotifications)

	// 获取当前用户对全体通知的状态记录
	var userGlobalRecords []models.NotificationUserRecord
	h.db.Where("user_id = ?", userID).Find(&userGlobalRecords)

	// 分别获取已读和已删除的通知ID
	readGlobalNotificationIDs := make(map[string]bool)
	deletedGlobalNotificationIDs := make(map[string]bool)
	for _, record := range userGlobalRecords {
		if record.Status == models.NotificationUserStatusRead {
			readGlobalNotificationIDs[record.NotificationID] = true
		}
		if record.Status == models.NotificationUserStatusDeleted {
			deletedGlobalNotificationIDs[record.NotificationID] = true
		}
	}

	// 按类型统计未读数量
	typeCounts := make(map[string]int)
	allTypes := []string{NotificationTypeSystem, NotificationTypeUpdate, NotificationTypeActivity,
		NotificationTypeFeedback, NotificationTypePayment, NotificationTypeOther}
	for _, notificationType := range allTypes {
		typeCounts[notificationType] = 0
	}

	totalCount := 0

	// 统计指定用户的未读通知
	for _, notification := range personalUnreadNotifications {
		typeCounts[notification.Type]++
		totalCount++
	}

	// 统计全体用户的未读通知（排除已读和已删除的）
	for _, notification := range globalNotifications {
		notificationIDStr := notification.NotificationID
		if !readGlobalNotificationIDs[notificationIDStr] && !deletedGlobalNotificationIDs[notificationIDStr] {
			typeCounts[notification.Type]++
			totalCount++
		}
	}

	typeDetails := map[string]int{
		"system":   typeCounts[NotificationTypeSystem],
		"update":   typeCounts[NotificationTypeUpdate],
		"activity": typeCounts[NotificationTypeActivity],
		"feedback": typeCounts[NotificationTypeFeedback],
		"payment":  typeCounts[NotificationTypePayment],
		"other":    typeCounts[NotificationTypeOther],
	}

	middleware.Success(c, "获取未读通知数量成功", gin.H{
		"total_count":  totalCount,
		"type_counts":  typeCounts,
		"type_details": typeDetails,
	})
}

// ========================= Contact QRCode Handlers =========================

// UploadContactQRCode upload contact qrcode
func (h *SystemHandler) UploadContactQRCode(c *gin.Context) {
	qrType := c.PostForm("type")
	if qrType == "" {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "缺少type参数"))
		return
	}

	if qrType != "personal" && qrType != "group" {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "类型参数错误，仅支持：personal(个人) 或 group(群聊)"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("获取文件失败: %v", err)))
		return
	}

	// 检查文件类型
	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/bmp", "image/webp", "image/svg+xml", "image/tiff"}
	contentType := file.Header.Get("Content-Type")
	isAllowed := false
	for _, at := range allowedTypes {
		if contentType == at {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "文件格式不支持，仅支持jpg、jpeg、png、gif、bmp、webp、svg、tiff格式"))
		return
	}

	// 获取文件扩展名
	ext := filepath.Ext(file.Filename)
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tif", ".tiff"}
	isExtAllowed := false
	for _, ae := range allowedExts {
		if strings.ToLower(ext) == ae {
			isExtAllowed = true
			break
		}
	}
	if !isExtAllowed {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "文件扩展名不支持，仅支持jpg、jpeg、png、gif、bmp、webp、svg、tif、tiff格式"))
		return
	}

	// 创建目录
	contactDir := "config/contact"
	if err := os.MkdirAll(contactDir, 0755); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建目录失败: %v", err)))
		return
	}

	// 根据类型生成固定文件名
	filename := qrType + ".png"
	filePath := filepath.Join(contactDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("保存文件失败: %v", err)))
		return
	}

	middleware.Success(c, "二维码图片上传成功", gin.H{
		"filename": filename,
		"type":     qrType,
		"path":     filePath,
	})
}

// GetContactQRCodeList get contact qrcode list
func (h *SystemHandler) GetContactQRCodeList(c *gin.Context) {
	contactDir := "config/contact"

	// 确保目录存在
	if err := os.MkdirAll(contactDir, 0755); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建目录失败: %v", err)))
		return
	}

	imageList := make([]map[string]interface{}, 0)

	// 固定检查两个文件：personal.png 和 group.png
	for _, qrType := range []string{"personal", "group"} {
		filePath := filepath.Join(contactDir, qrType+".png")

		item := map[string]interface{}{
			"type":     qrType,
			"name":     qrType,
			"filename": qrType + ".png",
			"path":     filePath,
			"url":      fmt.Sprintf("https://api.01editor.com/%s", filePath),
			"exists":   false,
		}

		if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
			item["exists"] = true
			item["size"] = fileInfo.Size()
			item["created_time"] = fileInfo.ModTime().Unix()
		} else {
			item["size"] = 0
			item["created_time"] = nil
		}

		imageList = append(imageList, item)
	}

	middleware.Success(c, "获取二维码图片列表成功", gin.H{
		"list": imageList,
	})
}

// SetupSystemRoutes setup system routes
func SetupSystemRoutes(r *gin.Engine) {
	handler := NewSystemHandler()

	systemGroup := r.Group("/api/v1/system")
	{
		// Feedback routes
		systemGroup.POST("/feedback", middleware.JWTAuth(), handler.CreateFeedback)
		systemGroup.GET("/feedback", middleware.JWTAuth(), handler.GetFeedbackList)
		systemGroup.GET("/feedback/:feedback_id", middleware.JWTAuth(), handler.GetFeedbackDetail)

		// Notification routes
		systemGroup.POST("/notification", middleware.JWTAuth(), handler.CreateNotification)
		systemGroup.GET("/notification", middleware.JWTAuth(), handler.GetNotificationList)
		systemGroup.GET("/notification/list", middleware.JWTAuth(), handler.GetNotificationListAdmin)
		systemGroup.PUT("/notification/:notification_id/publish", middleware.JWTAuth(), handler.PublishNotification)
		systemGroup.DELETE("/notification/:notification_id", middleware.JWTAuth(), handler.DeleteNotification)
		systemGroup.GET("/notification/:notification_id", middleware.JWTAuth(), handler.GetNotificationDetail)
		systemGroup.PUT("/notification/read/all", middleware.JWTAuth(), handler.MarkAllNotificationRead)
		systemGroup.PUT("/notification/read/:notification_id", middleware.JWTAuth(), handler.MarkNotificationRead)
		systemGroup.DELETE("/notification/user/:notification_id", middleware.JWTAuth(), handler.DeleteUserNotification)
		systemGroup.POST("/notification/batch-delete", middleware.JWTAuth(), handler.BatchDeleteUserNotifications)
		systemGroup.DELETE("/notification/read/clear", middleware.JWTAuth(), handler.ClearAllReadNotifications)
		systemGroup.GET("/notification/unread/count", middleware.JWTAuth(), handler.GetUnreadNotificationCount)

		// Contact QRCode routes
		systemGroup.POST("/contact/qrcode/upload", middleware.JWTAuth(), handler.UploadContactQRCode)
		systemGroup.GET("/contact/qrcode/list", handler.GetContactQRCodeList)
	}
}


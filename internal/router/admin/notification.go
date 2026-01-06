package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetNotificationList 获取通知列表
func (h *AdminHandler) GetNotificationList(c *gin.Context) {
	var req struct {
		Page   int    `form:"page"`
		Limit  int    `form:"limit"`
		Status string `form:"status"`
		Type   string `form:"type"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	// 构建查询
	query := repository.DB.Model(&models.SystemNotification{})

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 类型筛选
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.Limit
	var notifications []models.SystemNotification
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&notifications).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(notifications))
	for _, notification := range notifications {
		item := gin.H{
			"notification_id": notification.NotificationID,
			"user_id":         notification.UserID,
			"title":           notification.Title,
			"content":         notification.Content,
			"type":            notification.Type,
			"link":            notification.Link,
			"is_important":    notification.IsImportant,
			"status":          notification.Status,
			"created_at":      notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      notification.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if notification.ReadTime != nil {
			item["read_time"] = notification.ReadTime.Format("2006-01-02T15:04:05Z07:00")
		}
		if notification.ExpireTime != nil {
			item["expire_time"] = notification.ExpireTime.Format("2006-01-02T15:04:05Z07:00")
		}

		result = append(result, item)
	}

	middleware.Success(c, "获取通知列表成功", gin.H{
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
		"list":  result,
	})
}

// GetNotificationUserRecordList 获取用户通知记录列表
func (h *AdminHandler) GetNotificationUserRecordList(c *gin.Context) {
	var req struct {
		Page           int    `form:"page"`
		PageSize       int    `form:"page_size"`
		OrderBy        string `form:"order_by"`
		OrderDirection string `form:"order_direction"`
		Relations      string `form:"relations"`
		RelationDepth  int    `form:"relation_depth"`
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
		req.PageSize = 10
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}

	// 构建查询
	query := repository.DB.Model(&models.NotificationUserRecord{})

	// 排序
	orderField := req.OrderBy
	if req.OrderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var records []models.NotificationUserRecord
	if err := query.Offset(offset).Limit(req.PageSize).Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 加载关联数据
	if req.Relations != "" {
		relations := strings.Split(req.Relations, ",")
		loadUser := false
		loadNotification := false
		for _, relation := range relations {
			relation = strings.TrimSpace(relation)
			if relation == "user" {
				loadUser = true
			}
			if relation == "notification" {
				loadNotification = true
			}
		}

		// 批量加载用户信息
		if loadUser {
			userIDs := make([]string, 0)
			userIDMap := make(map[string]bool)
			for _, record := range records {
				if !userIDMap[record.UserID] {
					userIDs = append(userIDs, record.UserID)
					userIDMap[record.UserID] = true
				}
			}
			if len(userIDs) > 0 {
				var users []models.User
				repository.DB.Where("user_id IN ?", userIDs).Find(&users)
				userMap := make(map[string]*models.User)
				for i := range users {
					userMap[users[i].UserID] = &users[i]
				}
				for i := range records {
					if user, ok := userMap[records[i].UserID]; ok {
						records[i].User = user
					}
				}
			}
		}

		// 批量加载通知信息
		if loadNotification {
			notificationIDs := make([]string, 0)
			notificationIDMap := make(map[string]bool)
			for _, record := range records {
				if !notificationIDMap[record.NotificationID] {
					notificationIDs = append(notificationIDs, record.NotificationID)
					notificationIDMap[record.NotificationID] = true
				}
			}
			if len(notificationIDs) > 0 {
				var notifications []models.SystemNotification
				repository.DB.Where("notification_id IN ?", notificationIDs).Find(&notifications)
				notificationMap := make(map[string]*models.SystemNotification)
				for i := range notifications {
					notificationMap[notifications[i].NotificationID] = &notifications[i]
				}
				for i := range records {
					if notification, ok := notificationMap[records[i].NotificationID]; ok {
						records[i].Notification = notification
					}
				}
			}
		}
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(records))
	for _, record := range records {
		item := gin.H{
			"record_id":       record.RecordID,
			"notification_id": record.NotificationID,
			"user_id":         record.UserID,
			"status":          string(record.Status),
			"created_at":      record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if record.ReadTime != nil {
			item["read_time"] = record.ReadTime.Format("2006-01-02T15:04:05Z07:00")
		}
		if record.DeletedTime != nil {
			item["deleted_time"] = record.DeletedTime.Format("2006-01-02T15:04:05Z07:00")
		}

		// 添加关联数据
		if record.User != nil {
			item["user"] = gin.H{
				"user_id":  record.User.UserID,
				"username": record.User.Username,
				"nickname": record.User.Nickname,
			}
		}
		if record.Notification != nil {
			item["notification"] = gin.H{
				"notification_id": record.Notification.NotificationID,
				"title":           record.Notification.Title,
				"content":         record.Notification.Content,
				"type":            record.Notification.Type,
			}
		}

		result = append(result, item)
	}

	middleware.Success(c, "获取成功", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      result,
	})
}

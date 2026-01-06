package admin

import (
	"fmt"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetFeedbackStats 获取系统反馈统计
func (h *AdminHandler) GetFeedbackStats(c *gin.Context) {
	// 统计各状态反馈的数量
	var pendingCount, processingCount, resolvedCount, totalCount int64

	// 统计待处理（pending）
	if err := repository.DB.Model(&models.Feedback{}).
		Where("status = ?", "pending").
		Count(&pendingCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计处理中（processing）
	if err := repository.DB.Model(&models.Feedback{}).
		Where("status = ?", "processing").
		Count(&processingCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计已解决（resolved）
	if err := repository.DB.Model(&models.Feedback{}).
		Where("status = ?", "resolved").
		Count(&resolvedCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计总数
	if err := repository.DB.Model(&models.Feedback{}).
		Count(&totalCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取反馈统计成功", gin.H{
		"pending_count":    pendingCount,
		"processing_count": processingCount,
		"resolved_count":   resolvedCount,
		"total_count":      totalCount,
	})
}

// ReplyFeedback 管理员回复用户反馈
// @Summary 管理员回复用户反馈
// @Description 管理员对用户反馈进行回复，并可更新反馈状态
// @Tags admin-feedback
// @Accept json
// @Produce json
// @Param id path string true "反馈ID"
// @Param body body ReplyFeedbackRequest true "回复内容和状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/feedback/{id}/reply [put]
func (h *AdminHandler) ReplyFeedback(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "缺少ID参数"))
		return
	}

	var req struct {
		AdminReply string `json:"admin_reply" binding:"required"`
		Status     int    `json:"status"` // 0-待处理，1-处理中，2-已解决
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	var feedback models.Feedback
	if err := repository.DB.Where("feedback_id = ?", id).First(&feedback).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "反馈记录不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	updateData := make(map[string]interface{})
	updateData["admin_reply"] = req.AdminReply

	var statusStr string
	switch req.Status {
	case 0:
		statusStr = "pending"
	case 1:
		statusStr = "processing"
	case 2:
		statusStr = "resolved"
	default:
		statusStr = feedback.Status
	}
	updateData["status"] = statusStr

	if err := repository.DB.Model(&feedback).Updates(updateData).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	if err := repository.DB.Where("feedback_id = ?", id).First(&feedback).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 如果状态为"已解决"且管理员回复不为空，且有用户ID，则发送系统通知
	if statusStr == "resolved" && req.AdminReply != "" && feedback.UserID != nil && *feedback.UserID != "" {
		title := "反馈已处理"
		if feedback.Title != nil && *feedback.Title != "" {
			title = fmt.Sprintf("反馈已处理：%s", *feedback.Title)
		}

		content := fmt.Sprintf("管理员已回复您的反馈：\n\n%s", req.AdminReply)

		notification := &models.SystemNotification{
			NotificationID: fmt.Sprintf("notif_%s_%s", time.Now().Format("20060102150405"), id[:8]), // 使用时间戳和反馈ID前缀
			UserID:         feedback.UserID,
			Type:           "feedback",
			Title:          title,
			Content:        content,
			IsImportant:    true,
			Status:         "unread",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := repository.DB.Create(notification).Error; err != nil {
			repository.Warnf("发送反馈回复通知失败: %v", err)
		}
	}

	middleware.Success(c, "回复成功", gin.H{
		"feedback_id": feedback.FeedbackID,
		"admin_reply": feedback.AdminReply,
		"status":      feedback.Status,
		"updated_at":  feedback.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

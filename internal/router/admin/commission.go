package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCommissionOverview 获取佣金概览统计
func (h *AdminHandler) GetCommissionOverview(c *gin.Context) {
	// 总佣金记录数
	var totalRecords int64
	repository.DB.Model(&models.CommissionRecord{}).Count(&totalRecords)

	// 按状态统计
	statusStats := make(map[string]gin.H)
	statuses := []models.CommissionStatus{
		models.CommissionPending,
		models.CommissionIssued,
		models.CommissionWithdrawn,
		models.CommissionRejected,
		models.CommissionApplying,
	}

	for _, status := range statuses {
		var count int64
		var totalAmount float64

		repository.DB.Model(&models.CommissionRecord{}).
			Where("status = ?", int(status)).
			Count(&count)

		repository.DB.Model(&models.CommissionRecord{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("status = ?", int(status)).
			Scan(&totalAmount)

		statusName := ""
		switch status {
		case models.CommissionPending:
			statusName = "PENDING"
		case models.CommissionIssued:
			statusName = "ISSUED"
		case models.CommissionWithdrawn:
			statusName = "WITHDRAWN"
		case models.CommissionRejected:
			statusName = "REJECTED"
		case models.CommissionApplying:
			statusName = "APPLYING"
		}

		statusStats[statusName] = gin.H{
			"count":  count,
			"amount": totalAmount,
		}
	}

	// 总佣金金额（所有状态）
	var totalAmount float64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)

	// 已发放和已提现的佣金总额
	var issuedAmount float64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status IN ?", []int{int(models.CommissionIssued), int(models.CommissionWithdrawn)}).
		Scan(&issuedAmount)

	// 待发放佣金总额
	var pendingAmount float64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status = ?", int(models.CommissionPending)).
		Scan(&pendingAmount)

	// 唯一用户数
	var uniqueUsers int64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COUNT(DISTINCT user_id)").
		Scan(&uniqueUsers)

	middleware.Success(c, "获取成功", gin.H{
		"total_records":  totalRecords,
		"total_amount":   totalAmount,
		"issued_amount":  issuedAmount,
		"pending_amount": pendingAmount,
		"unique_users":   uniqueUsers,
		"status_stats":   statusStats,
	})
}

// GetCommissionList 获取佣金列表
func (h *AdminHandler) GetCommissionList(c *gin.Context) {
	var req struct {
		Page      int    `form:"page" binding:"min=1"`
		PageSize  int    `form:"page_size" binding:"min=1"`
		Status    *int   `form:"status"`
		UserID    string `form:"user_id"`
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
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

	// 构建查询
	query := repository.DB.Model(&models.CommissionRecord{})

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 用户筛选
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 日期范围筛选
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// 设置为当天的最后一刻
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", endTime)
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
	var records []models.CommissionRecord
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(records))
	for _, record := range records {
		// 加载用户信息
		var user models.User
		repository.DB.Where("user_id = ?", record.UserID).First(&user)

		// 加载邀请关系信息
		var relation models.InvitationRelation
		var invitee models.User
		if err := repository.DB.Where("id = ?", record.RelationID).First(&relation).Error; err == nil {
			repository.DB.Where("user_id = ?", relation.InviteeID).First(&invitee)
		}

		// 加载订单信息（如果有）
		var order models.Trade
		if record.OrderID != nil {
			repository.DB.Where("id = ?", *record.OrderID).First(&order)
		}

		item := gin.H{
			"id":      record.ID,
			"user_id": record.UserID,
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"username": user.Username,
				"phone":    user.Phone,
			},
			"amount":      record.Amount,
			"status":      int(record.Status),
			"status_text": record.Status.String(),
			"description": record.Description,
			"created_at":  record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if record.IssueTime != nil {
			item["issue_time"] = record.IssueTime.Format("2006-01-02T15:04:05Z07:00")
		}
		if record.WithdrawalTime != nil {
			item["withdrawal_time"] = record.WithdrawalTime.Format("2006-01-02T15:04:05Z07:00")
		}

		if invitee.UserID != "" {
			item["invitee"] = gin.H{
				"user_id":  invitee.UserID,
				"nickname": invitee.Nickname,
				"username": invitee.Username,
			}
		}

		if order.ID != 0 {
			item["order"] = gin.H{
				"trade_no": order.TradeNo,
				"amount":   order.Amount,
				"title":    order.Title,
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

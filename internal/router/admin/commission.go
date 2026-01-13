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

// GetUserCommissionDistribution 获取用户佣金分布统计（不包含被邀请人详细列表）
func (h *AdminHandler) GetUserCommissionDistribution(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	// 验证用户是否存在
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
		return
	}

	// 1. 按状态统计佣金分布
	type StatusDistribution struct {
		Status     int     `json:"status"`
		StatusText string  `json:"status_text"`
		Count      int64   `json:"count"`
		Amount     float64 `json:"amount"`
	}

	statusDistribution := []StatusDistribution{}
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
			Where("user_id = ? AND status = ?", userID, int(status)).
			Count(&count)

		repository.DB.Model(&models.CommissionRecord{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("user_id = ? AND status = ?", userID, int(status)).
			Scan(&totalAmount)

		statusDistribution = append(statusDistribution, StatusDistribution{
			Status:     int(status),
			StatusText: status.String(),
			Count:      count,
			Amount:     totalAmount,
		})
	}

	// 2. 按月统计佣金趋势（最近12个月）
	type MonthlyTrend struct {
		Month  string  `json:"month"`
		Count  int64   `json:"count"`
		Amount float64 `json:"amount"`
	}

	monthlyTrend := []MonthlyTrend{}
	now := time.Now()
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var count int64
		var totalAmount float64

		repository.DB.Model(&models.CommissionRecord{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, monthStart, monthEnd).
			Count(&count)

		repository.DB.Model(&models.CommissionRecord{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, monthStart, monthEnd).
			Scan(&totalAmount)

		monthlyTrend = append(monthlyTrend, MonthlyTrend{
			Month:  monthStart.Format("2006-01"),
			Count:  count,
			Amount: totalAmount,
		})
	}

	// 3. 统计该用户邀请的被邀请人的会员类型分布和付费金额（并发优化）
	type MembershipDistribution struct {
		ProductName string  `json:"product_name"`
		VipLevel    int     `json:"vip_level"`
		Count       int64   `json:"count"`
		TotalAmount float64 `json:"total_amount"`
		Percentage  float64 `json:"percentage"`
	}

	// 获取该用户所有的邀请关系（只获取被邀请人ID）
	var inviteeIDs []string
	if err := repository.DB.Model(&models.InvitationRelation{}).
		Where("inviter_id = ?", userID).
		Pluck("invitee_id", &inviteeIDs).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询邀请关系失败: "+err.Error()))
		return
	}

	membershipStats := make(map[string]*MembershipDistribution)
	var totalPaymentAmount float64

	if len(inviteeIDs) > 0 {
		// 批量查询所有被邀请人的成功支付订单
		var trades []models.Trade
		repository.DB.Select("title, amount").
			Where("user_id IN ? AND payment_status = ?", inviteeIDs, "success").
			Find(&trades)

		// 按产品名称（title）分组统计
		for _, trade := range trades {
			productName := trade.Title
			if productName == "" {
				productName = "未知产品"
			}

			if _, exists := membershipStats[productName]; !exists {
				membershipStats[productName] = &MembershipDistribution{
					ProductName: productName,
					Count:       0,
					TotalAmount: 0,
				}
			}

			membershipStats[productName].Count++
			membershipStats[productName].TotalAmount += trade.Amount
			totalPaymentAmount += trade.Amount
		}
	}

	// 转换为数组并计算百分比
	membershipDistribution := make([]MembershipDistribution, 0, len(membershipStats))
	for _, stat := range membershipStats {
		if totalPaymentAmount > 0 {
			stat.Percentage = (stat.TotalAmount / totalPaymentAmount) * 100
		}
		membershipDistribution = append(membershipDistribution, *stat)
	}

	// 4. 统计总览数据
	var totalCommissionCount int64
	var totalCommissionAmount float64
	var totalInviteeCount int64

	repository.DB.Model(&models.CommissionRecord{}).
		Where("user_id = ?", userID).
		Count(&totalCommissionCount)

	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ?", userID).
		Scan(&totalCommissionAmount)

	totalInviteeCount = int64(len(inviteeIDs))

	// 返回统计数据（不包含被邀请人详细列表）
	middleware.Success(c, "获取成功", gin.H{
		"user": gin.H{
			"user_id":  user.UserID,
			"nickname": user.Nickname,
			"username": user.Username,
			"phone":    user.Phone,
		},
		"overview": gin.H{
			"total_commission_count":  totalCommissionCount,
			"total_commission_amount": totalCommissionAmount,
			"total_invitee_count":     totalInviteeCount,
			"total_payment_amount":    totalPaymentAmount,
		},
		"status_distribution":     statusDistribution,
		"monthly_trend":           monthlyTrend,
		"membership_distribution": membershipDistribution,
	})
}

// GetUserInviteeList 获取用户的被邀请人详细列表（批量查询优化）
func (h *AdminHandler) GetUserInviteeList(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	// 获取分页和排序参数
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		OrderBy  string `form:"order_by"` // commission_amount, total_payment, invited_at
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
		req.PageSize = 20 // 默认每页20条
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 最大每页100条
	}
	if req.OrderBy == "" {
		req.OrderBy = "invited_at" // 默认按邀请时间排序
	}

	// 验证用户是否存在
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
		return
	}

	// 获取该用户所有的邀请关系
	var relations []models.InvitationRelation
	if err := repository.DB.Where("inviter_id = ?", userID).Find(&relations).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询邀请关系失败: "+err.Error()))
		return
	}

	if len(relations) == 0 {
		// 如果没有邀请任何人，返回空列表
		middleware.Success(c, "获取成功", gin.H{
			"list": []interface{}{},
			"pagination": gin.H{
				"page":       req.Page,
				"page_size":  req.PageSize,
				"total":      0,
				"total_page": 0,
			},
			"order_by": req.OrderBy,
		})
		return
	}

	// 定义数据结构
	type InviteeStats struct {
		RelationID       int
		InviteeID        string
		InvitedAt        time.Time
		TotalPayment     float64
		CommissionAmount float64
		OrderCount       int
		CommissionCount  int
	}

	// 创建关系映射
	relationMap := make(map[string]*InviteeStats)
	inviteeIDs := make([]string, 0, len(relations))
	relationIDs := make([]int, 0, len(relations))

	for _, rel := range relations {
		stats := &InviteeStats{
			RelationID: rel.ID,
			InviteeID:  rel.InviteeID,
			InvitedAt:  rel.CreatedAt,
		}
		relationMap[rel.InviteeID] = stats
		inviteeIDs = append(inviteeIDs, rel.InviteeID)
		relationIDs = append(relationIDs, rel.ID)
	}

	// 批量查询所有被邀请人的成功订单
	var trades []models.Trade
	repository.DB.Select("user_id, amount").
		Where("user_id IN ? AND payment_status = ?", inviteeIDs, "success").
		Find(&trades)

	// 统计每个被邀请人的订单数和付费金额
	for _, trade := range trades {
		if stats, exists := relationMap[trade.UserID]; exists {
			stats.OrderCount++
			stats.TotalPayment += trade.Amount
		}
	}

	// 批量查询所有关系的佣金记录
	var commissions []models.CommissionRecord
	repository.DB.Select("relation_id, amount").
		Where("user_id = ? AND relation_id IN ?", userID, relationIDs).
		Find(&commissions)

	// 创建relation_id到invitee_id的映射
	relationToInviteeMap := make(map[int]string)
	for _, rel := range relations {
		relationToInviteeMap[rel.ID] = rel.InviteeID
	}

	// 统计每个被邀请人的佣金
	for _, commission := range commissions {
		if inviteeID, exists := relationToInviteeMap[commission.RelationID]; exists {
			if stats, exists := relationMap[inviteeID]; exists {
				stats.CommissionCount++
				stats.CommissionAmount += commission.Amount
			}
		}
	}

	// 转换为数组
	inviteeStatsList := make([]*InviteeStats, 0, len(relationMap))
	for _, stats := range relationMap {
		inviteeStatsList = append(inviteeStatsList, stats)
	}

	// 根据排序字段排序
	switch req.OrderBy {
	case "commission_amount":
		// 按佣金金额降序
		for i := 0; i < len(inviteeStatsList); i++ {
			for j := i + 1; j < len(inviteeStatsList); j++ {
				if inviteeStatsList[i].CommissionAmount < inviteeStatsList[j].CommissionAmount {
					inviteeStatsList[i], inviteeStatsList[j] = inviteeStatsList[j], inviteeStatsList[i]
				}
			}
		}
	case "total_payment":
		// 按付费金额降序
		for i := 0; i < len(inviteeStatsList); i++ {
			for j := i + 1; j < len(inviteeStatsList); j++ {
				if inviteeStatsList[i].TotalPayment < inviteeStatsList[j].TotalPayment {
					inviteeStatsList[i], inviteeStatsList[j] = inviteeStatsList[j], inviteeStatsList[i]
				}
			}
		}
	default: // invited_at
		// 按邀请时间降序（最新的在前面）
		for i := 0; i < len(inviteeStatsList); i++ {
			for j := i + 1; j < len(inviteeStatsList); j++ {
				if inviteeStatsList[i].InvitedAt.Before(inviteeStatsList[j].InvitedAt) {
					inviteeStatsList[i], inviteeStatsList[j] = inviteeStatsList[j], inviteeStatsList[i]
				}
			}
		}
	}

	// 计算分页
	totalCount := len(inviteeStatsList)
	offset := (req.Page - 1) * req.PageSize
	end := offset + req.PageSize
	if offset > totalCount {
		offset = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	// 获取当前页的数据
	pagedStats := inviteeStatsList[offset:end]

	// 定义详细信息结构
	type InviteeDetail struct {
		UserID            string   `json:"user_id"`
		Nickname          *string  `json:"nickname"`
		Username          *string  `json:"username"`
		Phone             *string  `json:"phone"`
		VipLevel          int      `json:"vip_level"`
		InvitedAt         string   `json:"invited_at"`
		OrderCount        int      `json:"order_count"`
		TotalPayment      float64  `json:"total_payment"`
		CommissionAmount  float64  `json:"commission_amount"`
		CommissionCount   int      `json:"commission_count"`
		LatestOrderTime   *string  `json:"latest_order_time"`
		ProductsPurchased []string `json:"products_purchased"`
	}

	// 批量查询当前页被邀请人的用户信息
	pagedInviteeIDs := make([]string, len(pagedStats))
	for i, stats := range pagedStats {
		pagedInviteeIDs[i] = stats.InviteeID
	}

	var users []models.User
	repository.DB.Select("user_id, nickname, username, phone, vip_level").
		Where("user_id IN ?", pagedInviteeIDs).
		Find(&users)

	// 创建用户映射
	userMap := make(map[string]models.User)
	for _, user := range users {
		userMap[user.UserID] = user
	}

	// 批量查询当前页被邀请人的订单（用于获取产品列表和最新订单时间）
	var pagedTrades []models.Trade
	repository.DB.Select("user_id, title, paid_at").
		Where("user_id IN ? AND payment_status = ?", pagedInviteeIDs, "success").
		Order("paid_at DESC").
		Find(&pagedTrades)

	// 按用户ID组织订单数据
	type UserTrades struct {
		LatestPaidAt      *time.Time
		ProductsPurchased []string
	}
	tradesMap := make(map[string]*UserTrades)
	for _, trade := range pagedTrades {
		if _, exists := tradesMap[trade.UserID]; !exists {
			tradesMap[trade.UserID] = &UserTrades{
				ProductsPurchased: []string{},
			}
		}
		userTrades := tradesMap[trade.UserID]

		// 记录最新订单时间
		if trade.PaidAt != nil && (userTrades.LatestPaidAt == nil || trade.PaidAt.After(*userTrades.LatestPaidAt)) {
			userTrades.LatestPaidAt = trade.PaidAt
		}

		// 收集产品（去重）
		productName := trade.Title
		if productName != "" {
			// 简单去重检查
			found := false
			for _, p := range userTrades.ProductsPurchased {
				if p == productName {
					found = true
					break
				}
			}
			if !found {
				userTrades.ProductsPurchased = append(userTrades.ProductsPurchased, productName)
			}
		}
	}

	// 组装详细信息
	inviteeDetails := make([]InviteeDetail, 0, len(pagedStats))
	for _, stats := range pagedStats {
		user, userExists := userMap[stats.InviteeID]
		if !userExists {
			continue
		}

		detail := InviteeDetail{
			UserID:            user.UserID,
			Nickname:          user.Nickname,
			Username:          user.Username,
			Phone:             user.Phone,
			VipLevel:          user.VipLevel,
			InvitedAt:         stats.InvitedAt.Format("2006-01-02 15:04:05"),
			OrderCount:        stats.OrderCount,
			TotalPayment:      stats.TotalPayment,
			CommissionAmount:  stats.CommissionAmount,
			CommissionCount:   stats.CommissionCount,
			ProductsPurchased: []string{},
		}

		// 添加订单详情
		if userTrades, exists := tradesMap[stats.InviteeID]; exists {
			if userTrades.LatestPaidAt != nil {
				paidAtStr := userTrades.LatestPaidAt.Format("2006-01-02 15:04:05")
				detail.LatestOrderTime = &paidAtStr
			}
			detail.ProductsPurchased = userTrades.ProductsPurchased
		}

		inviteeDetails = append(inviteeDetails, detail)
	}

	// 返回被邀请人列表数据
	middleware.Success(c, "获取成功", gin.H{
		"list": inviteeDetails,
		"pagination": gin.H{
			"page":       req.Page,
			"page_size":  req.PageSize,
			"total":      totalCount,
			"total_page": (totalCount + req.PageSize - 1) / req.PageSize,
		},
		"order_by": req.OrderBy,
	})
}

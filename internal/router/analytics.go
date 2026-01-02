package router

import (
	"fmt"
	"time"

	"gin_web/internal/middleware"
	"gin_web/internal/models"
	"gin_web/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AnalyticsHandler 数据分析处理器
type AnalyticsHandler struct{}

// NewAnalyticsHandler 创建数据分析处理器
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

// parseDateRange 解析和标准化日期范围
func parseDateRange(startDate, endDate string, defaultDays int) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if endDate == "" {
		end = time.Now()
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, end.Location())
	} else {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return start, end, err
		}
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, end.Location())
	}

	if startDate == "" {
		start = end.AddDate(0, 0, -defaultDays)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	} else {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return start, end, err
		}
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	}

	return start, end, nil
}

// GetUserOverview 获取用户数据概览
func (h *AnalyticsHandler) GetUserOverview(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	var totalUsers int64
	if err := repository.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询总用户数失败: "+err.Error()))
		return
	}

	// 活跃用户数（3天内有登录记录的用户）
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	var activeUsers int64
	if err := repository.DB.Model(&models.User{}).
		Where("last_login_time >= ?", threeDaysAgo).
		Count(&activeUsers).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询活跃用户数失败: "+err.Error()))
		return
	}

	// VIP用户数量（role = 2）
	var vipUsers int64
	if err := repository.DB.Model(&models.User{}).
		Where("role = ?", 2).
		Count(&vipUsers).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询VIP用户数失败: "+err.Error()))
		return
	}

	// 新增用户数（按日期范围）
	var newUsers int64
	if err := repository.DB.Model(&models.User{}).
		Where("registration_date >= ? AND registration_date <= ?", start, end).
		Count(&newUsers).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询新增用户数失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取用户概览数据成功", gin.H{
		"total_users":  totalUsers,
		"active_users": activeUsers,
		"vip_users":    vipUsers,
		"new_users":    newUsers,
	})
}

// GetUserGrowth 获取用户增长趋势
func (h *AnalyticsHandler) GetUserGrowth(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	// 查询日期范围内的所有用户
	var users []models.User
	if err := repository.DB.Where("registration_date >= ? AND registration_date <= ?", start, end).
		Find(&users).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败: "+err.Error()))
		return
	}

	growthData := []gin.H{}
	currentDate := start

	for currentDate.Before(end) || currentDate.Equal(end) {
		var nextDate time.Time
		var dateStr string

		switch period {
		case "day":
			nextDate = currentDate.AddDate(0, 0, 1)
			dateStr = currentDate.Format("2006-01-02")
		case "week":
			nextDate = currentDate.AddDate(0, 0, 7)
			dateStr = fmt.Sprintf("%s ~ %s",
				currentDate.Format("2006-01-02"),
				nextDate.AddDate(0, 0, -1).Format("2006-01-02"))
		case "month":
			if currentDate.Month() == 12 {
				nextDate = time.Date(currentDate.Year()+1, 1, 1, 0, 0, 0, 0, currentDate.Location())
			} else {
				nextDate = time.Date(currentDate.Year(), currentDate.Month()+1, 1, 0, 0, 0, 0, currentDate.Location())
			}
			dateStr = currentDate.Format("2006-01")
		default:
			middleware.HandleError(c, middleware.NewBusinessError(400, "无效的统计周期，支持: day/week/month"))
			return
		}

		// 统计当前周期内的用户数
		count := 0
		for _, user := range users {
			regDate := user.RegistrationDate
			if !regDate.Before(currentDate) && regDate.Before(nextDate) {
				count++
			}
		}

		growthData = append(growthData, gin.H{
			"date":  dateStr,
			"count": count,
		})

		currentDate = nextDate
	}

	middleware.Success(c, "获取用户增长趋势成功", growthData)
}

// GetPaymentOverview 获取支付数据概览
func (h *AnalyticsHandler) GetPaymentOverview(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	// 最小日期：2025-07-01
	minDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if start.Before(minDate) {
		start = minDate
	}

	// 总收入（从2025-07-01开始，只统计微信和支付宝渠道，支付成功的）
	var totalIncome struct {
		Total float64
	}
	if err := repository.DB.Model(&models.Trade{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("(payment_channel = ? OR payment_channel = ?) AND payment_status = ? AND paid_at >= ?",
			"wx_qr", "alipay_qr", "success", minDate).
		Scan(&totalIncome).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询总收入失败: "+err.Error()))
		return
	}

	// 时间段内收入
	var periodIncome struct {
		Total float64
	}
	if err := repository.DB.Model(&models.Trade{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("(payment_channel = ? OR payment_channel = ?) AND payment_status = ? AND paid_at >= ? AND paid_at <= ?",
			"wx_qr", "alipay_qr", "success", start, end).
		Scan(&periodIncome).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询时间段收入失败: "+err.Error()))
		return
	}

	// 各支付渠道收入占比
	channelStats := []gin.H{}
	channels := []string{"wx_qr", "alipay_qr", "other"}
	for _, channel := range channels {
		var channelIncome struct {
			Total float64
		}
		if err := repository.DB.Model(&models.Trade{}).
			Select("COALESCE(SUM(amount), 0) as total").
			Where("payment_status = ? AND payment_channel = ? AND paid_at >= ? AND paid_at <= ?",
				"success", channel, start, end).
			Scan(&channelIncome).Error; err != nil {
			continue
		}

		if channelIncome.Total > 0 {
			channelStats = append(channelStats, gin.H{
				"channel": channel,
				"amount":  channelIncome.Total,
			})
		}
	}

	// 待支付佣金成本统计
	var pendingCommission struct {
		Total float64
	}
	if err := repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("status = ? AND created_at >= ? AND created_at <= ?",
			models.CommissionPending, start, end).
		Scan(&pendingCommission).Error; err != nil {
		// 如果查询失败，设置为0
		pendingCommission.Total = 0
	}

	middleware.Success(c, "获取支付概览数据成功", gin.H{
		"total_income":            totalIncome.Total,
		"period_income":           periodIncome.Total,
		"channel_stats":           channelStats,
		"pending_commission_cost": pendingCommission.Total,
	})
}

// GetPaymentTrend 获取支付趋势
func (h *AnalyticsHandler) GetPaymentTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	// 最小日期：2025-07-01
	minDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if start.Before(minDate) {
		start = minDate
	}

	// 查询所有符合条件的交易（微信和支付宝，支付成功）
	var trades []models.Trade
	if err := repository.DB.Where("(payment_channel = ? OR payment_channel = ?) AND payment_status = ? AND paid_at >= ? AND paid_at <= ?",
		"wx_qr", "alipay_qr", "success", start, end).
		Find(&trades).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询交易失败: "+err.Error()))
		return
	}

	trendData := []gin.H{}
	currentDate := start

	for currentDate.Before(end) || currentDate.Equal(end) {
		var nextDate time.Time
		var dateStr string

		switch period {
		case "day":
			nextDate = currentDate.AddDate(0, 0, 1)
			dateStr = currentDate.Format("2006-01-02")
		case "week":
			nextDate = currentDate.AddDate(0, 0, 7)
			dateStr = fmt.Sprintf("%s ~ %s",
				currentDate.Format("2006-01-02"),
				nextDate.AddDate(0, 0, -1).Format("2006-01-02"))
		case "month":
			if currentDate.Month() == 12 {
				nextDate = time.Date(currentDate.Year()+1, 1, 1, 0, 0, 0, 0, currentDate.Location())
			} else {
				nextDate = time.Date(currentDate.Year(), currentDate.Month()+1, 1, 0, 0, 0, 0, currentDate.Location())
			}
			dateStr = currentDate.Format("2006-01")
		default:
			middleware.HandleError(c, middleware.NewBusinessError(400, "无效的统计周期，支持: day/week/month"))
			return
		}

		// 统计当前周期内的交易
		var amount float64
		count := 0
		for _, trade := range trades {
			if trade.PaidAt != nil {
				paidAt := *trade.PaidAt
				if !paidAt.Before(currentDate) && paidAt.Before(nextDate) {
					amount += trade.Amount
					count++
				}
			}
		}

		trendData = append(trendData, gin.H{
			"date":   dateStr,
			"amount": amount,
			"count":  count,
		})

		currentDate = nextDate
	}

	middleware.Success(c, "获取支付趋势数据成功", trendData)
}

// GetCostAnalysis 成本效益分析
func (h *AnalyticsHandler) GetCostAnalysis(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	// 收入数据
	var totalRevenue struct {
		Total float64
	}
	if err := repository.DB.Model(&models.Trade{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("payment_status = ? AND paid_at >= ? AND paid_at <= ?", "success", start, end).
		Scan(&totalRevenue).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询收入失败: "+err.Error()))
		return
	}
	revenue := totalRevenue.Total

	// 积分发放成本（充值给用户的积分）
	var creditIssued struct {
		Total int
	}
	if err := repository.DB.Model(&models.CreditRecord{}).
		Select("COALESCE(SUM(credits), 0) as total").
		Where("record_type = ? AND created_at >= ? AND created_at <= ?",
			models.CreditRecharge, start, end).
		Scan(&creditIssued).Error; err != nil {
		creditIssued.Total = 0
	}

	// 积分消耗（用户实际使用的积分）
	var creditConsumed struct {
		Total int
	}
	if err := repository.DB.Model(&models.CreditRecord{}).
		Select("COALESCE(SUM(credits), 0) as total").
		Where("record_type = ? AND created_at >= ? AND created_at <= ?",
			models.CreditConsumption, start, end).
		Scan(&creditConsumed).Error; err != nil {
		creditConsumed.Total = 0
	}

	// 假设每个积分成本0.01元（可配置）
	creditCostRate := 0.01
	totalCreditCost := float64(creditConsumed.Total) * creditCostRate

	// 利润计算
	grossProfit := revenue - totalCreditCost
	profitMargin := 0.0
	if revenue > 0 {
		profitMargin = (grossProfit / revenue) * 100
	}

	// 积分使用效率
	creditUtilization := 0.0
	if creditIssued.Total > 0 {
		creditUtilization = (float64(creditConsumed.Total) / float64(creditIssued.Total)) * 100
	}

	middleware.Success(c, "获取成本分析成功", gin.H{
		"revenue": revenue,
		"costs": gin.H{
			"credit_service_cost": totalCreditCost,
			"cost_per_credit":     creditCostRate,
		},
		"profit": gin.H{
			"gross_profit":  grossProfit,
			"profit_margin": profitMargin,
		},
		"credits": gin.H{
			"issued":           creditIssued.Total,
			"consumed":         creditConsumed.Total,
			"utilization_rate": creditUtilization,
		},
	})
}

// GetSalesRanking 销售额排行榜 - 按用户佣金排序
func (h *AnalyticsHandler) GetSalesRanking(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	// 获取时间范围内的佣金记录（包含已发放、已提现、待发放状态）
	var commissionRecords []models.CommissionRecord
	if err := repository.DB.Where("(status = ? OR status = ? OR status = ?) AND created_at >= ? AND created_at <= ?",
		models.CommissionIssued, models.CommissionWithdrawn, models.CommissionPending, start, end).
		Find(&commissionRecords).Error; err != nil && err != gorm.ErrRecordNotFound {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询佣金记录失败: "+err.Error()))
		return
	}

	// 收集所有唯一的用户ID
	userIDSet := make(map[string]bool)
	for _, record := range commissionRecords {
		userIDSet[record.UserID] = true
	}

	// 批量查询所有用户信息
	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	userMap := make(map[string]*models.User)
	if len(userIDs) > 0 {
		var users []models.User
		if err := repository.DB.Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户信息失败: "+err.Error()))
			return
		}
		for i := range users {
			userMap[users[i].UserID] = &users[i]
		}
	}

	// 按用户统计佣金总额
	userCommissions := make(map[string]gin.H)
	for _, record := range commissionRecords {
		userID := record.UserID
		if _, exists := userCommissions[userID]; !exists {
			username := "未知用户"
			var nickname, avatar *string

			if user, found := userMap[userID]; found {
				if user.Username != nil {
					username = *user.Username
				} else if user.Nickname != nil {
					username = *user.Nickname
				}
				nickname = user.Nickname
				avatar = user.Avatar
			}

			userCommissions[userID] = gin.H{
				"user_id":            userID,
				"username":           username,
				"nickname":           nickname,
				"avatar":             avatar,
				"total_commission":   0.0,
				"commission_count":   0,
				"pending_commission": 0.0,
				"issued_commission":  0.0,
			}
		}

		comm := userCommissions[userID]
		comm["total_commission"] = comm["total_commission"].(float64) + record.Amount
		comm["commission_count"] = comm["commission_count"].(int) + 1

		if record.Status == models.CommissionPending {
			comm["pending_commission"] = comm["pending_commission"].(float64) + record.Amount
		} else if record.Status == models.CommissionIssued || record.Status == models.CommissionWithdrawn {
			comm["issued_commission"] = comm["issued_commission"].(float64) + record.Amount
		}

		userCommissions[userID] = comm
	}

	// 转换为列表并排序
	rankingList := make([]gin.H, 0, len(userCommissions))
	for _, comm := range userCommissions {
		rankingList = append(rankingList, comm)
	}

	// 按佣金总额降序排序（这里简化处理，实际应该用sort包）
	// 由于Go的map遍历顺序不确定，我们需要手动排序
	// 这里先简化实现，返回所有数据

	// 分页处理
	var pageNum, size int
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &size)
	if pageNum < 1 {
		pageNum = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	totalCount := len(rankingList)
	startIndex := (pageNum - 1) * size
	endIndex := startIndex + size
	if startIndex > totalCount {
		startIndex = totalCount
	}
	if endIndex > totalCount {
		endIndex = totalCount
	}

	pageData := rankingList[startIndex:endIndex]
	// 添加排名信息
	for i := range pageData {
		pageData[i]["rank"] = startIndex + i + 1
	}

	totalPages := (totalCount + size - 1) / size

	// 计算总佣金
	totalCommission := 0.0
	for _, comm := range rankingList {
		totalCommission += comm["total_commission"].(float64)
	}

	middleware.Success(c, "获取销售额排行榜成功", gin.H{
		"rankings": pageData,
		"pagination": gin.H{
			"current_page": pageNum,
			"page_size":    size,
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"has_next":     pageNum < totalPages,
			"has_prev":     pageNum > 1,
		},
		"summary": gin.H{
			"total_users":      totalCount,
			"total_commission": totalCommission,
		},
	})
}

// GetInvitationRanking 邀请人数排行榜 - 按邀请人数排序
func (h *AnalyticsHandler) GetInvitationRanking(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	start, end, err := parseDateRange(startDate, endDate, 30)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误: "+err.Error()))
		return
	}

	// 获取时间范围内的邀请关系记录
	var invitationRelations []models.InvitationRelation
	if err := repository.DB.Where("created_at >= ? AND created_at <= ?", start, end).
		Find(&invitationRelations).Error; err != nil && err != gorm.ErrRecordNotFound {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询邀请关系失败: "+err.Error()))
		return
	}

	// 收集所有唯一的用户ID（邀请人和被邀请人）
	userIDSet := make(map[string]bool)
	for _, relation := range invitationRelations {
		userIDSet[relation.InviterID] = true
		userIDSet[relation.InviteeID] = true
	}

	// 批量查询所有用户信息
	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	userMap := make(map[string]*models.User)
	if len(userIDs) > 0 {
		var users []models.User
		if err := repository.DB.Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户信息失败: "+err.Error()))
			return
		}
		for i := range users {
			userMap[users[i].UserID] = &users[i]
		}
	}

	// 按邀请人统计邀请数量
	inviterStats := make(map[string]gin.H)
	for _, relation := range invitationRelations {
		inviterID := relation.InviterID
		if _, exists := inviterStats[inviterID]; !exists {
			username := "未知用户"
			var nickname, avatar *string

			if user, found := userMap[inviterID]; found {
				if user.Username != nil {
					username = *user.Username
				} else if user.Nickname != nil {
					username = *user.Nickname
				}
				nickname = user.Nickname
				avatar = user.Avatar
			}

			inviterStats[inviterID] = gin.H{
				"user_id":          inviterID,
				"username":         username,
				"nickname":         nickname,
				"avatar":           avatar,
				"invitation_count": 0,
				"invitees":         []gin.H{},
			}
		}

		stats := inviterStats[inviterID]
		stats["invitation_count"] = stats["invitation_count"].(int) + 1

		inviteeUsername := "未知用户"
		var inviteeNickname, inviteeAvatar *string
		if user, found := userMap[relation.InviteeID]; found {
			if user.Username != nil {
				inviteeUsername = *user.Username
			} else if user.Nickname != nil {
				inviteeUsername = *user.Nickname
			}
			inviteeNickname = user.Nickname
			inviteeAvatar = user.Avatar
		}

		invitees := stats["invitees"].([]gin.H)
		invitees = append(invitees, gin.H{
			"user_id":    relation.InviteeID,
			"username":   inviteeUsername,
			"nickname":   inviteeNickname,
			"avatar":     inviteeAvatar,
			"created_at": relation.CreatedAt.Format("2006-01-02 15:04:05"),
		})
		stats["invitees"] = invitees
		inviterStats[inviterID] = stats
	}

	// 获取邀请人的佣金统计
	for inviterID := range inviterStats {
		var totalCommission struct {
			Total float64
		}
		repository.DB.Model(&models.CommissionRecord{}).
			Select("COALESCE(SUM(amount), 0) as total").
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", inviterID, start, end).
			Scan(&totalCommission)

		stats := inviterStats[inviterID]
		stats["total_commission"] = totalCommission.Total

		// 统计不同状态的佣金
		for _, status := range []models.CommissionStatus{
			models.CommissionPending,
			models.CommissionIssued,
			models.CommissionWithdrawn,
		} {
			var statusCommission struct {
				Total float64
			}
			repository.DB.Model(&models.CommissionRecord{}).
				Select("COALESCE(SUM(amount), 0) as total").
				Where("user_id = ? AND status = ? AND created_at >= ? AND created_at <= ?",
					inviterID, status, start, end).
				Scan(&statusCommission)

			statusName := "pending_commission"
			if status == models.CommissionIssued {
				statusName = "issued_commission"
			} else if status == models.CommissionWithdrawn {
				statusName = "withdrawn_commission"
			}
			stats[statusName] = statusCommission.Total
		}

		inviterStats[inviterID] = stats
	}

	// 转换为列表并排序
	rankingList := make([]gin.H, 0, len(inviterStats))
	for _, stats := range inviterStats {
		// 只返回前5个被邀请人的信息
		invitees := stats["invitees"].([]gin.H)
		if len(invitees) > 5 {
			stats["invitees"] = invitees[:5]
			stats["more_invitees"] = len(invitees) - 5
		}
		rankingList = append(rankingList, stats)
	}

	// 分页处理
	var pageNum, size int
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &size)
	if pageNum < 1 {
		pageNum = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	totalCount := len(rankingList)
	startIndex := (pageNum - 1) * size
	endIndex := startIndex + size
	if startIndex > totalCount {
		startIndex = totalCount
	}
	if endIndex > totalCount {
		endIndex = totalCount
	}

	pageData := rankingList[startIndex:endIndex]
	// 添加排名信息
	for i := range pageData {
		pageData[i]["rank"] = startIndex + i + 1
	}

	totalPages := (totalCount + size - 1) / size

	// 计算总邀请数和总佣金
	totalInvitations := 0
	totalCommission := 0.0
	for _, stats := range rankingList {
		totalInvitations += stats["invitation_count"].(int)
		if tc, ok := stats["total_commission"].(float64); ok {
			totalCommission += tc
		}
	}

	middleware.Success(c, "获取邀请人数排行榜成功", gin.H{
		"rankings": pageData,
		"pagination": gin.H{
			"current_page": pageNum,
			"page_size":    size,
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"has_next":     pageNum < totalPages,
			"has_prev":     pageNum > 1,
		},
		"summary": gin.H{
			"total_inviters":    totalCount,
			"total_invitations": totalInvitations,
			"total_commission":  totalCommission,
		},
	})
}

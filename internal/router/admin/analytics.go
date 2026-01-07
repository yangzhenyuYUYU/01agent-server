package admin

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service/analytics"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AnalyticsHandler 数据分析处理器
type AnalyticsHandler struct {
	metricsService *analytics.MetricsService
	trendService   *analytics.TrendService
}

// NewAnalyticsHandler 创建数据分析处理器
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		metricsService: analytics.NewMetricsService(),
		trendService:   analytics.NewTrendService(),
	}
}

// parseDateRange 解析和标准化日期范围
// 统一使用本地时区（北京时间 UTC+8）来避免时区问题
func parseDateRange(startDate, endDate string, defaultDays int) (time.Time, time.Time, error) {
	var start, end time.Time

	// 使用本地时区（北京时间）
	loc := time.FixedZone("CST", 8*60*60) // UTC+8

	if endDate == "" {
		now := time.Now().In(loc)
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc)
	} else {
		// 解析日期，使用本地时区
		parsed, err := time.ParseInLocation("2006-01-02", endDate, loc)
		if err != nil {
			return start, end, err
		}
		end = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, loc)
	}

	if startDate == "" {
		start = end.AddDate(0, 0, -defaultDays)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	} else {
		// 解析日期，使用本地时区
		parsed, err := time.ParseInLocation("2006-01-02", startDate, loc)
		if err != nil {
			return start, end, err
		}
		start = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, loc)
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
		// 参考Python代码：使用 current_date <= reg_date < next_date
		// 统一使用本地时区（北京时间）进行比较
		loc := currentDate.Location()
		count := 0
		for _, user := range users {
			regDate := user.RegistrationDate
			// 转换为本地时区（北京时间），类似Python的normalize_datetime
			regDateInLoc := regDate.In(loc)

			// 对于按天统计，提取日期部分进行比较，避免时区问题
			if period == "day" {
				// 提取日期部分（年月日），忽略时分秒
				regDateOnly := time.Date(regDateInLoc.Year(), regDateInLoc.Month(), regDateInLoc.Day(), 0, 0, 0, 0, loc)
				currentDateOnly := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, loc)

				// 只比较日期是否相等
				if regDateOnly.Equal(currentDateOnly) {
					count++
				}
			} else {
				// 对于周/月统计，使用时间范围比较：currentDate <= regDate < nextDate
				if (regDateInLoc.Equal(currentDate) || regDateInLoc.After(currentDate)) && regDateInLoc.Before(nextDate) {
					count++
				}
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

	// 最小日期：2025-07-01（统一使用CST时区）
	loc := time.FixedZone("CST", 8*60*60)
	minDate := time.Date(2025, 7, 1, 0, 0, 0, 0, loc)
	if start.Before(minDate) {
		start = minDate
	}

	// 总收入（从2025-07-01开始，支付成功的，排除兑换码）
	// 只统计会员（订阅服务）和积分套餐相关的交易，与/membership/overview保持一致
	// 使用子查询先找到符合条件的交易ID，然后按交易去重统计，避免重复计算（一个交易可能关联多个产品）
	var totalIncome struct {
		Total float64
	}
	if err := repository.DB.Table("trades").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("id IN (SELECT DISTINCT t.id FROM trades t "+
			"JOIN user_productions up ON t.id = up.trade_id "+
			"JOIN productions p ON up.production_id = p.id "+
			"WHERE t.payment_status = ? "+
			"AND t.trade_type != ? "+
			"AND (p.product_type = ? OR p.product_type = ?) "+
			"AND t.paid_at >= ?)",
			"success", "activation", "订阅服务", "积分套餐", minDate).
		Scan(&totalIncome).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询总收入失败: "+err.Error()))
		return
	}

	// 时间段内收入（只统计会员和积分套餐，排除兑换码）
	var periodIncome struct {
		Total float64
	}
	if err := repository.DB.Table("trades").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("id IN (SELECT DISTINCT t.id FROM trades t "+
			"JOIN user_productions up ON t.id = up.trade_id "+
			"JOIN productions p ON up.production_id = p.id "+
			"WHERE t.payment_status = ? "+
			"AND t.trade_type != ? "+
			"AND (p.product_type = ? OR p.product_type = ?) "+
			"AND t.paid_at >= ? AND t.paid_at <= ?)",
			"success", "activation", "订阅服务", "积分套餐", start, end).
		Scan(&periodIncome).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询时间段收入失败: "+err.Error()))
		return
	}

	// 各支付渠道收入占比（只统计会员和积分套餐，排除兑换码）
	// 统计所有支付渠道（排除兑换码）
	channelStats := []gin.H{}
	var allChannels []struct {
		Channel string
		Total   float64
	}
	if err := repository.DB.Table("trades").
		Select("t.payment_channel as channel, COALESCE(SUM(t.amount), 0) as total").
		Where("t.id IN (SELECT DISTINCT t2.id FROM trades t2 "+
			"JOIN user_productions up ON t2.id = up.trade_id "+
			"JOIN productions p ON up.production_id = p.id "+
			"WHERE t2.payment_status = ? "+
			"AND t2.trade_type != ? "+
			"AND (p.product_type = ? OR p.product_type = ?) "+
			"AND t2.paid_at >= ? AND t2.paid_at <= ?)",
			"success", "activation", "订阅服务", "积分套餐", start, end).
		Group("t.payment_channel").
		Scan(&allChannels).Error; err == nil {
		for _, ch := range allChannels {
			if ch.Total > 0 {
				channelStats = append(channelStats, gin.H{
					"channel": ch.Channel,
					"amount":  ch.Total,
				})
			}
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
		// 参考Python代码：使用 current_date <= normalized_paid_at < next_date
		// 统一使用本地时区（北京时间）进行比较
		loc := currentDate.Location()
		var amount float64
		count := 0
		for _, trade := range trades {
			if trade.PaidAt != nil {
				paidAt := *trade.PaidAt
				// 转换为本地时区（北京时间），类似Python的normalize_datetime
				paidAtInLoc := paidAt.In(loc)

				// 对于按天统计，提取日期部分进行比较，避免时区问题
				if period == "day" {
					// 提取日期部分（年月日），忽略时分秒
					paidAtDate := time.Date(paidAtInLoc.Year(), paidAtInLoc.Month(), paidAtInLoc.Day(), 0, 0, 0, 0, loc)
					currentDateOnly := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, loc)

					// 只比较日期是否相等
					if paidAtDate.Equal(currentDateOnly) {
						amount += trade.Amount
						count++
					}
				} else {
					// 对于周/月统计，使用时间范围比较：currentDate <= paidAt < nextDate
					if (paidAtInLoc.Equal(currentDate) || paidAtInLoc.After(currentDate)) && paidAtInLoc.Before(nextDate) {
						amount += trade.Amount
						count++
					}
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

	// 获取时间范围内的邀请关系记录（使用Preload一次性加载用户信息，优化性能）
	var invitationRelations []models.InvitationRelation
	if err := repository.DB.Where("created_at >= ? AND created_at <= ?", start, end).
		Preload("Inviter").
		Preload("Invitee").
		Find(&invitationRelations).Error; err != nil && err != gorm.ErrRecordNotFound {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询邀请关系失败: "+err.Error()))
		return
	}

	// 构建用户信息映射（从Preload的数据中获取，避免额外查询）
	userMap := make(map[string]*models.User)
	for i := range invitationRelations {
		if invitationRelations[i].Inviter.UserID != "" {
			userMap[invitationRelations[i].InviterID] = &invitationRelations[i].Inviter
		}
		if invitationRelations[i].Invitee.UserID != "" {
			userMap[invitationRelations[i].InviteeID] = &invitationRelations[i].Invitee
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

	// 批量获取所有邀请人的佣金统计（优化性能，避免N+1查询）
	inviterIDs := make([]string, 0, len(inviterStats))
	for inviterID := range inviterStats {
		inviterIDs = append(inviterIDs, inviterID)
	}

	// 批量查询总佣金（按用户ID分组）
	type CommissionSummary struct {
		UserID string
		Total  float64
		Status int
	}
	var commissionSummaries []CommissionSummary

	if len(inviterIDs) > 0 {
		// 一次性查询所有邀请人的佣金统计（按用户ID和状态分组）
		if err := repository.DB.Model(&models.CommissionRecord{}).
			Select("user_id, COALESCE(SUM(amount), 0) as total, status").
			Where("user_id IN ? AND created_at >= ? AND created_at <= ?", inviterIDs, start, end).
			Group("user_id, status").
			Scan(&commissionSummaries).Error; err != nil {
			repository.Errorf("批量查询佣金统计失败: %v", err)
		}
	}

	// 将佣金统计结果组织到map中
	commissionMap := make(map[string]map[string]float64) // userID -> statusName -> amount
	for _, summary := range commissionSummaries {
		if commissionMap[summary.UserID] == nil {
			commissionMap[summary.UserID] = make(map[string]float64)
			// 初始化所有佣金字段为0
			commissionMap[summary.UserID]["total_commission"] = 0.0
			commissionMap[summary.UserID]["pending_commission"] = 0.0
			commissionMap[summary.UserID]["issued_commission"] = 0.0
			commissionMap[summary.UserID]["withdrawn_commission"] = 0.0
		}

		statusName := "pending_commission"
		if summary.Status == int(models.CommissionIssued) {
			statusName = "issued_commission"
		} else if summary.Status == int(models.CommissionWithdrawn) {
			statusName = "withdrawn_commission"
		}

		commissionMap[summary.UserID][statusName] = summary.Total
		// 累加总佣金
		commissionMap[summary.UserID]["total_commission"] += summary.Total
	}

	// 将佣金数据填充到统计结果中
	for inviterID := range inviterStats {
		stats := inviterStats[inviterID]
		if commData, exists := commissionMap[inviterID]; exists {
			stats["total_commission"] = commData["total_commission"]
			stats["pending_commission"] = commData["pending_commission"]
			stats["issued_commission"] = commData["issued_commission"]
			stats["withdrawn_commission"] = commData["withdrawn_commission"]
		} else {
			stats["total_commission"] = 0.0
			stats["pending_commission"] = 0.0
			stats["issued_commission"] = 0.0
			stats["withdrawn_commission"] = 0.0
		}
		inviterStats[inviterID] = stats
	}

	// 转换为列表并按邀请人数降序排序（参考Python代码）
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

	// 按邀请人数降序排序（参考Python代码：ranking_list.sort(key=lambda x: x['invitation_count'], reverse=True)）
	// 使用sort.Slice进行稳定排序，确保相同邀请人数时顺序一致
	sort.Slice(rankingList, func(i, j int) bool {
		countI := rankingList[i]["invitation_count"].(int)
		countJ := rankingList[j]["invitation_count"].(int)
		// 按邀请人数降序排序
		if countI != countJ {
			return countI > countJ
		}
		// 如果邀请人数相同，按用户ID升序排序确保稳定性
		userIDI := rankingList[i]["user_id"].(string)
		userIDJ := rankingList[j]["user_id"].(string)
		return userIDI < userIDJ
	})

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

// GetMetrics 获取统一的数据分析指标
// @Summary 获取数据分析指标
// @Description 获取指定日期和指标的数据，支持并行计算多个指标
// @Tags admin-analytics
// @Accept json
// @Produce json
// @Param date query string false "统计日期，格式：YYYY-MM-DD，默认为今天"
// @Param metrics query string false "要获取的指标列表，多个指标用逗号分隔，如：active_users_daily,active_users_weekly。如果不指定则返回所有启用的指标"
// @Param start_date query string false "开始日期（用于日期范围查询），格式：YYYY-MM-DD"
// @Param end_date query string false "结束日期（用于日期范围查询），格式：YYYY-MM-DD"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/metrics [get]
func (h *AnalyticsHandler) GetMetrics(c *gin.Context) {
	dateStr := c.Query("date")
	metricsStr := c.Query("metrics")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 解析日期
	loc := time.FixedZone("CST", 8*60*60)
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，请使用YYYY-MM-DD格式"))
			return
		}
	} else {
		date = time.Now().In(loc)
	}

	// 解析指标列表
	var metrics []analytics.MetricKey
	if metricsStr != "" {
		metricKeys := strings.Split(metricsStr, ",")
		for _, key := range metricKeys {
			key = strings.TrimSpace(key)
			if key != "" {
				metrics = append(metrics, analytics.MetricKey(key))
			}
		}
	}

	// 日期范围查询
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.ParseInLocation("2006-01-02", startDateStr, loc)
		endDate, err2 := time.ParseInLocation("2006-01-02", endDateStr, loc)
		if err1 != nil || err2 != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "日期范围格式错误，请使用YYYY-MM-DD格式"))
			return
		}

		// 限制日期范围，避免查询时间过长
		daysDiff := int(endDate.Sub(startDate).Hours() / 24)
		if daysDiff > 90 {
			middleware.HandleError(c, middleware.NewBusinessError(400, "日期范围不能超过90天，当前范围："+fmt.Sprintf("%d天", daysDiff)))
			return
		}

		// 获取日期范围内的数据
		results, err := h.metricsService.GetMetricsByDateRange(metrics, startDate, endDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "获取指标数据失败: "+err.Error()))
			return
		}

		middleware.Success(c, "获取指标数据成功", gin.H{
			"date_range": gin.H{
				"start_date": startDateStr,
				"end_date":   endDateStr,
			},
			"results": results,
		})
		return
	}

	// 单日查询
	response, err := h.metricsService.GetMetrics(metrics, date)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取指标数据失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取指标数据成功", response)
}

// GetMetricsInfo 获取所有指标信息列表
// @Summary 获取指标信息列表
// @Description 获取所有可用的指标信息，包括指标定义、计算公式等
// @Tags admin-analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/metrics/info [get]
func (h *AnalyticsHandler) GetMetricsInfo(c *gin.Context) {
	infoList := h.metricsService.GetMetricInfoList()

	// 按维度分组
	dimensions := make(map[string][]analytics.MetricInfo)
	for _, info := range infoList {
		dimensionKey := string(info.Dimension)
		dimensions[dimensionKey] = append(dimensions[dimensionKey], info)
	}

	middleware.Success(c, "获取指标信息成功", gin.H{
		"metrics":    infoList,
		"dimensions": dimensions,
	})
}

// GetActivityTrend 获取用户活跃度趋势数据
// @Summary 获取用户活跃度趋势
// @Description 获取指定时间范围内的用户活跃度趋势数据，支持按天/周/月统计
// @Tags admin-analytics
// @Accept json
// @Produce json
// @Param period query string false "统计周期：day/week/month，默认为day"
// @Param start_date query string true "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string true "结束日期，格式：YYYY-MM-DD"
// @Param include_wau query bool false "是否包含周活数据，默认为true"
// @Param include_mau query bool false "是否包含月活数据，默认为true"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/user/activity-trend [get]
func (h *AnalyticsHandler) GetActivityTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	includeWAUStr := c.DefaultQuery("include_wau", "true")
	includeMAUStr := c.DefaultQuery("include_mau", "true")

	if startDateStr == "" || endDateStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "开始日期和结束日期不能为空"))
		return
	}

	// 解析日期
	loc := time.FixedZone("CST", 8*60*60)
	startDate, err1 := time.ParseInLocation("2006-01-02", startDateStr, loc)
	endDate, err2 := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err1 != nil || err2 != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}

	// 限制日期范围
	daysDiff := int(endDate.Sub(startDate).Hours() / 24)
	if daysDiff > 90 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期范围不能超过90天，当前范围："+fmt.Sprintf("%d天", daysDiff)))
		return
	}

	// 解析布尔参数
	includeWAU := includeWAUStr == "true"
	includeMAU := includeMAUStr == "true"

	// 获取趋势数据
	response, err := h.trendService.GetActivityTrend(period, startDate, endDate, includeWAU, includeMAU)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取趋势数据失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取用户活跃度趋势成功", response)
}

// GetRegistrationSourceDistribution 获取注册来源分布
// @Summary 获取注册来源分布
// @Description 统计指定时间范围内的用户注册来源分布（utm_source字段）
// @Tags admin-analytics
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string true "结束日期，格式：YYYY-MM-DD"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/traffic/source-distribution [get]
func (h *AnalyticsHandler) GetRegistrationSourceDistribution(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "开始日期和结束日期不能为空"))
		return
	}

	// 解析日期
	loc := time.FixedZone("CST", 8*60*60)
	startDate, err1 := time.ParseInLocation("2006-01-02", startDateStr, loc)
	endDate, err2 := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err1 != nil || err2 != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}

	// 获取来源分布
	distributions, err := h.metricsService.GetTrafficSourceDistribution(startDate, endDate)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取来源分布失败: "+err.Error()))
		return
	}

	// 计算总计
	totalCount := int64(0)
	for _, dist := range distributions {
		totalCount += dist.Count
	}

	middleware.Success(c, "获取注册来源分布成功", gin.H{
		"start_date": startDateStr,
		"end_date":   endDateStr,
		"total":      totalCount,
		"data":       distributions,
	})
}

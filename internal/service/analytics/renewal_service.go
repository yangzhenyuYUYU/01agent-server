package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"fmt"
	"time"
)

// RenewalService 续费用户统计服务
type RenewalService struct{}

// NewRenewalService 创建续费用户统计服务
func NewRenewalService() *RenewalService {
	return &RenewalService{}
}

// RenewalUserRanking 续费用户排行榜数据
type RenewalUserRanking struct {
	UserID               string                   `json:"user_id"`                 // 用户ID
	Nickname             *string                  `json:"nickname"`                // 昵称
	Avatar               *string                  `json:"avatar"`                  // 头像
	Phone                *string                  `json:"phone"`                   // 手机号
	Email                *string                  `json:"email"`                   // 邮箱
	RenewalCount         int                      `json:"renewal_count"`           // 续费次数（总购买次数）
	TotalRenewalAmount   float64                  `json:"total_renewal_amount"`    // 总续费金额
	FirstPurchaseTime    time.Time                `json:"first_purchase_time"`     // 首次购买时间
	LastRenewalTime      time.Time                `json:"last_renewal_time"`       // 最后续费时间
	RenewalProducts      []RenewalProductInfo     `json:"renewal_products"`        // 续费产品列表
	AvgRenewalIntervalDays float64                `json:"avg_renewal_interval_days"` // 平均续费间隔（天）
	LifetimeValueRank    int                      `json:"lifetime_value_rank"`     // 生命周期价值排名
}

// RenewalProductInfo 续费产品信息
type RenewalProductInfo struct {
	ProductName  string    `json:"product_name"`  // 产品名称
	ProductType  string    `json:"product_type"`  // 产品类型
	PurchaseCount int      `json:"purchase_count"` // 购买次数
	TotalAmount  float64   `json:"total_amount"`  // 总金额
	LastPurchase time.Time `json:"last_purchase"` // 最后购买时间
}

// RenewalSummary 续费统计汇总
type RenewalSummary struct {
	// 基础指标
	TotalUsers           int     `json:"total_users"`            // 总用户数
	FirstTimePayingUsers int     `json:"first_time_paying_users"` // 首次付费用户数
	RenewalUsers         int     `json:"renewal_users"`          // 续费用户数（购买≥2次）
	RenewalRate          float64 `json:"renewal_rate"`           // 续费率（%）
	
	// 续费指标
	AvgRenewalCount      float64 `json:"avg_renewal_count"`      // 平均续费次数
	AvgRenewalAmount     float64 `json:"avg_renewal_amount"`     // 平均续费金额
	TotalRenewalAmount   float64 `json:"total_renewal_amount"`   // 总续费金额
	
	// 分布统计
	RenewalCountDistribution []RenewalCountStats `json:"renewal_count_distribution"` // 续费次数分布
	RenewalProductDistribution []ProductStats    `json:"renewal_product_distribution"` // 续费产品分布
	RenewalIntervalDistribution []IntervalStats  `json:"renewal_interval_distribution"` // 续费间隔分布
}

// RenewalCountStats 续费次数统计
type RenewalCountStats struct {
	RenewalCount int     `json:"renewal_count"` // 续费次数
	UserCount    int     `json:"user_count"`    // 用户数
	Percentage   float64 `json:"percentage"`    // 占比
	TotalAmount  float64 `json:"total_amount"`  // 该档位总金额
}

// IntervalStats 续费间隔统计
type IntervalStats struct {
	IntervalDays int     `json:"interval_days"` // 间隔天数范围（起始）
	IntervalEnd  int     `json:"interval_end"`  // 间隔天数范围（结束）
	UserCount    int     `json:"user_count"`    // 用户数
	Percentage   float64 `json:"percentage"`    // 占比
}

// UserRenewalDetail 用户续费详情
type UserRenewalDetail struct {
	UserInfo         RenewalUserRanking    `json:"user_info"`          // 用户基本信息
	RenewalHistory   []RenewalHistoryItem  `json:"renewal_history"`    // 续费历史
	RenewalTimeline  []time.Time           `json:"renewal_timeline"`   // 续费时间线
}

// RenewalHistoryItem 续费历史记录
type RenewalHistoryItem struct {
	TradeNo       string    `json:"trade_no"`       // 交易流水号
	ProductName   string    `json:"product_name"`   // 产品名称
	ProductType   string    `json:"product_type"`   // 产品类型
	Amount        float64   `json:"amount"`         // 金额
	PaymentTime   time.Time `json:"payment_time"`   // 支付时间
	DaysSinceLast int       `json:"days_since_last"` // 距离上次续费天数
}

// GetRenewalRanking 获取续费用户排行榜
// sortBy: count（按续费次数）, amount（按续费金额）
// limit: 返回数量限制，默认100
func (s *RenewalService) GetRenewalRanking(startDate, endDate *time.Time, sortBy string, limit int) ([]RenewalUserRanking, error) {
	loc := time.FixedZone("CST", 8*60*60)
	
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	
	// 1. 查询所有购买过订阅服务的用户及其购买记录
	baseQuery := repository.DB.Table("trades as t").
		Select(`
			t.user_id,
			COUNT(*) as renewal_count,
			COALESCE(SUM(t.amount), 0) as total_amount,
			MIN(t.paid_at) as first_purchase_time,
			MAX(t.paid_at) as last_renewal_time
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.payment_channel != ?", "activation").
		Where("t.trade_type = ?", "recharge").
		Where("p.product_type = ?", "订阅服务").
		Where("t.paid_at IS NOT NULL")
	
	// 日期范围过滤
	if startDate != nil {
		startInLoc := startDate.In(loc)
		startOfDay := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		baseQuery = baseQuery.Where("t.paid_at >= ?", startOfDay)
	}
	if endDate != nil {
		endInLoc := endDate.In(loc)
		endOfDay := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		baseQuery = baseQuery.Where("t.paid_at <= ?", endOfDay)
	}
	
	// 只统计续费用户（购买次数≥2次）
	baseQuery = baseQuery.Group("t.user_id").Having("COUNT(*) >= 2")
	
	// 排序
	switch sortBy {
	case "amount":
		baseQuery = baseQuery.Order("total_amount DESC")
	default: // count
		baseQuery = baseQuery.Order("renewal_count DESC")
	}
	
	baseQuery = baseQuery.Limit(limit)
	
	var userStats []struct {
		UserID            string    `gorm:"column:user_id"`
		RenewalCount      int       `gorm:"column:renewal_count"`
		TotalAmount       float64   `gorm:"column:total_amount"`
		FirstPurchaseTime time.Time `gorm:"column:first_purchase_time"`
		LastRenewalTime   time.Time `gorm:"column:last_renewal_time"`
	}
	
	err := baseQuery.Scan(&userStats).Error
	if err != nil {
		return nil, fmt.Errorf("查询续费用户统计失败: %w", err)
	}
	
	if len(userStats) == 0 {
		return []RenewalUserRanking{}, nil
	}
	
	// 2. 批量查询用户信息
	userIDs := make([]string, len(userStats))
	for i, stat := range userStats {
		userIDs[i] = stat.UserID
	}
	
	var users []models.User
	userInfoMap := make(map[string]*models.User)
	err = repository.DB.Where("user_id IN ?", userIDs).Find(&users).Error
	if err == nil {
		for i := range users {
			userInfoMap[users[i].UserID] = &users[i]
		}
	} else {
		repository.Warnf("批量查询用户信息失败: %v", err)
	}
	
	// 3. 批量查询每个用户的产品购买详情
	var productDetails []struct {
		UserID        string    `gorm:"column:user_id"`
		ProductName   string    `gorm:"column:product_name"`
		ProductType   string    `gorm:"column:product_type"`
		PurchaseCount int       `gorm:"column:purchase_count"`
		TotalAmount   float64   `gorm:"column:total_amount"`
		LastPurchase  time.Time `gorm:"column:last_purchase"`
	}
	
	productQuery := repository.DB.Table("trades as t").
		Select(`
			t.user_id,
			p.name as product_name,
			p.product_type,
			COUNT(*) as purchase_count,
			COALESCE(SUM(t.amount), 0) as total_amount,
			MAX(t.paid_at) as last_purchase
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.payment_channel != ?", "activation").
		Where("t.trade_type = ?", "recharge").
		Where("p.product_type = ?", "订阅服务").
		Where("t.paid_at IS NOT NULL").
		Where("t.user_id IN ?", userIDs)
	
	if startDate != nil {
		startInLoc := startDate.In(loc)
		startOfDay := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		productQuery = productQuery.Where("t.paid_at >= ?", startOfDay)
	}
	if endDate != nil {
		endInLoc := endDate.In(loc)
		endOfDay := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		productQuery = productQuery.Where("t.paid_at <= ?", endOfDay)
	}
	
	err = productQuery.Group("t.user_id, p.name, p.product_type").Scan(&productDetails).Error
	if err != nil {
		repository.Warnf("批量查询产品详情失败: %v", err)
	}
	
	// 组织产品信息
	userProductMap := make(map[string][]RenewalProductInfo)
	for _, detail := range productDetails {
		userProductMap[detail.UserID] = append(userProductMap[detail.UserID], RenewalProductInfo{
			ProductName:   detail.ProductName,
			ProductType:   detail.ProductType,
			PurchaseCount: detail.PurchaseCount,
			TotalAmount:   detail.TotalAmount,
			LastPurchase:  detail.LastPurchase,
		})
	}
	
	// 4. 构建排行榜数据
	rankings := make([]RenewalUserRanking, 0, len(userStats))
	for rank, stat := range userStats {
		user := userInfoMap[stat.UserID]
		
		// 计算平均续费间隔
		avgInterval := 0.0
		if stat.RenewalCount > 1 {
			totalDays := stat.LastRenewalTime.Sub(stat.FirstPurchaseTime).Hours() / 24
			avgInterval = totalDays / float64(stat.RenewalCount-1)
		}
		
		ranking := RenewalUserRanking{
			UserID:                 stat.UserID,
			RenewalCount:           stat.RenewalCount,
			TotalRenewalAmount:     stat.TotalAmount,
			FirstPurchaseTime:      stat.FirstPurchaseTime,
			LastRenewalTime:        stat.LastRenewalTime,
			RenewalProducts:        userProductMap[stat.UserID],
			AvgRenewalIntervalDays: avgInterval,
			LifetimeValueRank:      rank + 1,
		}
		
		if user != nil {
			ranking.Nickname = user.Nickname
			ranking.Avatar = user.Avatar
			ranking.Phone = user.Phone
			ranking.Email = user.Email
		}
		
		rankings = append(rankings, ranking)
	}
	
	return rankings, nil
}

// GetRenewalSummary 获取续费统计汇总
func (s *RenewalService) GetRenewalSummary(startDate, endDate *time.Time) (*RenewalSummary, error) {
	loc := time.FixedZone("CST", 8*60*60)
	
	// 1. 查询所有付费用户的购买次数
	baseQuery := repository.DB.Table("trades as t").
		Select(`
			t.user_id,
			COUNT(*) as purchase_count,
			COALESCE(SUM(t.amount), 0) as total_amount,
			MIN(t.paid_at) as first_purchase_time,
			MAX(t.paid_at) as last_purchase_time
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.payment_channel != ?", "activation").
		Where("t.trade_type = ?", "recharge").
		Where("p.product_type = ?", "订阅服务").
		Where("t.paid_at IS NOT NULL")
	
	// 日期范围过滤
	if startDate != nil {
		startInLoc := startDate.In(loc)
		startOfDay := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		baseQuery = baseQuery.Where("t.paid_at >= ?", startOfDay)
	}
	if endDate != nil {
		endInLoc := endDate.In(loc)
		endOfDay := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		baseQuery = baseQuery.Where("t.paid_at <= ?", endOfDay)
	}
	
	var userPurchases []struct {
		UserID            string    `gorm:"column:user_id"`
		PurchaseCount     int       `gorm:"column:purchase_count"`
		TotalAmount       float64   `gorm:"column:total_amount"`
		FirstPurchaseTime time.Time `gorm:"column:first_purchase_time"`
		LastPurchaseTime  time.Time `gorm:"column:last_purchase_time"`
	}
	
	err := baseQuery.Group("t.user_id").Scan(&userPurchases).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户购买记录失败: %w", err)
	}
	
	if len(userPurchases) == 0 {
		return &RenewalSummary{
			RenewalCountDistribution:   []RenewalCountStats{},
			RenewalProductDistribution: []ProductStats{},
			RenewalIntervalDistribution: []IntervalStats{},
		}, nil
	}
	
	// 2. 统计基础指标
	summary := &RenewalSummary{}
	summary.TotalUsers = len(userPurchases)
	
	renewalUsers := 0
	totalRenewalCount := 0
	totalRenewalAmount := 0.0
	
	// 续费次数分布统计
	countMap := make(map[int]*RenewalCountStats)
	// 续费间隔分布统计
	var intervalDays []float64
	
	for _, purchase := range userPurchases {
		if purchase.PurchaseCount == 1 {
			summary.FirstTimePayingUsers++
		} else {
			// 续费用户（购买≥2次）
			renewalUsers++
			totalRenewalCount += purchase.PurchaseCount
			totalRenewalAmount += purchase.TotalAmount
			
			// 统计续费次数分布
			count := purchase.PurchaseCount
			if count > 5 {
				count = 6 // 5次以上归为一类
			}
			if countMap[count] == nil {
				countMap[count] = &RenewalCountStats{
					RenewalCount: count,
				}
			}
			countMap[count].UserCount++
			countMap[count].TotalAmount += purchase.TotalAmount
			
			// 计算续费间隔
			if purchase.PurchaseCount > 1 {
				totalDays := purchase.LastPurchaseTime.Sub(purchase.FirstPurchaseTime).Hours() / 24
				intervalDays = append(intervalDays, totalDays)
			}
		}
	}
	
	summary.RenewalUsers = renewalUsers
	if summary.TotalUsers > 0 {
		summary.RenewalRate = float64(renewalUsers) / float64(summary.TotalUsers) * 100
	}
	if renewalUsers > 0 {
		summary.AvgRenewalCount = float64(totalRenewalCount) / float64(renewalUsers)
		summary.AvgRenewalAmount = totalRenewalAmount / float64(renewalUsers)
	}
	summary.TotalRenewalAmount = totalRenewalAmount
	
	// 3. 填充续费次数分布
	summary.RenewalCountDistribution = make([]RenewalCountStats, 0)
	for count := 2; count <= 6; count++ {
		if stat, exists := countMap[count]; exists {
			if renewalUsers > 0 {
				stat.Percentage = float64(stat.UserCount) / float64(renewalUsers) * 100
			}
			summary.RenewalCountDistribution = append(summary.RenewalCountDistribution, *stat)
		}
	}
	
	// 4. 查询续费产品分布
	var productStats []struct {
		ProductName string  `gorm:"column:product_name"`
		ProductType string  `gorm:"column:product_type"`
		UserCount   int     `gorm:"column:user_count"`
		TotalAmount float64 `gorm:"column:total_amount"`
	}
	
	productQuery := repository.DB.Table("trades as t").
		Select(`
			p.name as product_name,
			p.product_type,
			COUNT(DISTINCT t.user_id) as user_count,
			COALESCE(SUM(t.amount), 0) as total_amount
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.payment_channel != ?", "activation").
		Where("t.trade_type = ?", "recharge").
		Where("p.product_type = ?", "订阅服务").
		Where("t.paid_at IS NOT NULL")
	
	if startDate != nil {
		startInLoc := startDate.In(loc)
		startOfDay := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		productQuery = productQuery.Where("t.paid_at >= ?", startOfDay)
	}
	if endDate != nil {
		endInLoc := endDate.In(loc)
		endOfDay := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		productQuery = productQuery.Where("t.paid_at <= ?", endOfDay)
	}
	
	err = productQuery.Group("p.name, p.product_type").Order("user_count DESC").Scan(&productStats).Error
	if err == nil {
		summary.RenewalProductDistribution = make([]ProductStats, 0, len(productStats))
		for _, stat := range productStats {
			percentage := 0.0
			if renewalUsers > 0 {
				percentage = float64(stat.UserCount) / float64(renewalUsers) * 100
			}
			summary.RenewalProductDistribution = append(summary.RenewalProductDistribution, ProductStats{
				ProductName: stat.ProductName,
				ProductType: stat.ProductType,
				UserCount:   stat.UserCount,
				Percentage:  percentage,
				AvgPrice:    stat.TotalAmount / float64(stat.UserCount),
			})
		}
	}
	
	// 5. 计算续费间隔分布
	summary.RenewalIntervalDistribution = calculateIntervalDistribution(intervalDays, renewalUsers)
	
	return summary, nil
}

// GetUserRenewalDetail 获取单个用户的续费详情
func (s *RenewalService) GetUserRenewalDetail(userID string) (*UserRenewalDetail, error) {
	// 1. 查询用户的所有订阅服务购买记录
	var trades []struct {
		TradeNo     string    `gorm:"column:trade_no"`
		ProductName string    `gorm:"column:product_name"`
		ProductType string    `gorm:"column:product_type"`
		Amount      float64   `gorm:"column:amount"`
		PaidAt      time.Time `gorm:"column:paid_at"`
	}
	
	err := repository.DB.Table("trades as t").
		Select(`
			t.trade_no,
			p.name as product_name,
			p.product_type,
			t.amount,
			t.paid_at
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.user_id = ?", userID).
		Where("t.payment_status = ?", "success").
		Where("t.payment_channel != ?", "activation").
		Where("t.trade_type = ?", "recharge").
		Where("p.product_type = ?", "订阅服务").
		Where("t.paid_at IS NOT NULL").
		Order("t.paid_at ASC").
		Scan(&trades).Error
	
	if err != nil {
		return nil, fmt.Errorf("查询用户续费记录失败: %w", err)
	}
	
	if len(trades) == 0 {
		return nil, fmt.Errorf("用户未找到订阅服务购买记录")
	}
	
	// 2. 查询用户基本信息
	var user models.User
	err = repository.DB.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		repository.Warnf("查询用户信息失败: %v", err)
	}
	
	// 3. 构建续费历史
	renewalHistory := make([]RenewalHistoryItem, 0, len(trades))
	renewalTimeline := make([]time.Time, 0, len(trades))
	productMap := make(map[string]*RenewalProductInfo)
	
	var lastPaidAt time.Time
	totalAmount := 0.0
	
	for _, trade := range trades {
		daysSinceLast := 0
		if !lastPaidAt.IsZero() {
			daysSinceLast = int(trade.PaidAt.Sub(lastPaidAt).Hours() / 24)
		}
		
		renewalHistory = append(renewalHistory, RenewalHistoryItem{
			TradeNo:       trade.TradeNo,
			ProductName:   trade.ProductName,
			ProductType:   trade.ProductType,
			Amount:        trade.Amount,
			PaymentTime:   trade.PaidAt,
			DaysSinceLast: daysSinceLast,
		})
		
		renewalTimeline = append(renewalTimeline, trade.PaidAt)
		lastPaidAt = trade.PaidAt
		totalAmount += trade.Amount
		
		// 统计产品购买次数
		key := fmt.Sprintf("%s|%s", trade.ProductName, trade.ProductType)
		if prod, exists := productMap[key]; exists {
			prod.PurchaseCount++
			prod.TotalAmount += trade.Amount
			if trade.PaidAt.After(prod.LastPurchase) {
				prod.LastPurchase = trade.PaidAt
			}
		} else {
			productMap[key] = &RenewalProductInfo{
				ProductName:   trade.ProductName,
				ProductType:   trade.ProductType,
				PurchaseCount: 1,
				TotalAmount:   trade.Amount,
				LastPurchase:  trade.PaidAt,
			}
		}
	}
	
	// 4. 构建产品列表
	renewalProducts := make([]RenewalProductInfo, 0, len(productMap))
	for _, prod := range productMap {
		renewalProducts = append(renewalProducts, *prod)
	}
	
	// 计算平均续费间隔
	avgInterval := 0.0
	if len(trades) > 1 {
		totalDays := trades[len(trades)-1].PaidAt.Sub(trades[0].PaidAt).Hours() / 24
		avgInterval = totalDays / float64(len(trades)-1)
	}
	
	// 5. 构建用户信息
	userInfo := RenewalUserRanking{
		UserID:                 userID,
		Nickname:               user.Nickname,
		Avatar:                 user.Avatar,
		Phone:                  user.Phone,
		Email:                  user.Email,
		RenewalCount:           len(trades),
		TotalRenewalAmount:     totalAmount,
		FirstPurchaseTime:      trades[0].PaidAt,
		LastRenewalTime:        trades[len(trades)-1].PaidAt,
		RenewalProducts:        renewalProducts,
		AvgRenewalIntervalDays: avgInterval,
	}
	
	return &UserRenewalDetail{
		UserInfo:        userInfo,
		RenewalHistory:  renewalHistory,
		RenewalTimeline: renewalTimeline,
	}, nil
}

// calculateIntervalDistribution 计算续费间隔分布
func calculateIntervalDistribution(intervals []float64, totalUsers int) []IntervalStats {
	if len(intervals) == 0 {
		return []IntervalStats{}
	}
	
	// 定义间隔区间：0-7天, 8-30天, 31-90天, 91-180天, 180天以上
	ranges := []struct {
		start int
		end   int
		label string
	}{
		{0, 7, "一周内"},
		{8, 30, "1个月内"},
		{31, 90, "3个月内"},
		{91, 180, "半年内"},
		{181, 99999, "半年以上"},
	}
	
	rangeCounts := make([]int, len(ranges))
	for _, interval := range intervals {
		days := int(interval)
		for i, r := range ranges {
			if days >= r.start && days <= r.end {
				rangeCounts[i]++
				break
			}
		}
	}
	
	result := make([]IntervalStats, 0, len(ranges))
	for i, r := range ranges {
		if rangeCounts[i] > 0 || true { // 显示所有区间，包括0的
			percentage := 0.0
			if totalUsers > 0 {
				percentage = float64(rangeCounts[i]) / float64(totalUsers) * 100
			}
			result = append(result, IntervalStats{
				IntervalDays: r.start,
				IntervalEnd:  r.end,
				UserCount:    rangeCounts[i],
				Percentage:   percentage,
			})
		}
	}
	
	return result
}

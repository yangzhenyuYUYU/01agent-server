package analytics

import (
	"01agent_server/internal/repository"
	"sync"
	"time"
)

// MembershipService 会员统计服务
type MembershipService struct{}

// NewMembershipService 创建会员统计服务
func NewMembershipService() *MembershipService {
	return &MembershipService{}
}

// MembershipCategory 会员分类
type MembershipCategory string

const (
	CategoryFree         MembershipCategory = "免费版"    // 免费版
	CategoryLight        MembershipCategory = "轻量版"    // 轻量版系列
	CategoryProfessional MembershipCategory = "专业版"    // 专业版系列
	CategoryLifetime     MembershipCategory = "种子终身会员" // 种子终身会员
)

// MembershipStats 会员统计数据
type MembershipStats struct {
	Category    MembershipCategory `json:"category"`     // 会员分类
	ProductName string             `json:"product_name"` // 产品名称
	Count       int64              `json:"count"`        // 购买数量
	Revenue     float64            `json:"revenue"`      // 收入金额
	Percentage  float64            `json:"percentage"`   // 占比（%）
	UniqueUsers int64              `json:"unique_users"` // 去重用户数
}

// MembershipOverview 会员概览数据
type MembershipOverview struct {
	// 会员统计
	MembershipCount   int64   `json:"membership_count"`   // 会员购买数
	MembershipRevenue float64 `json:"membership_revenue"` // 会员收入
	MembershipUsers   int64   `json:"membership_users"`   // 会员用户数

	// 积分套餐统计
	CreditPackageCount   int64   `json:"credit_package_count"`   // 积分套餐购买数
	CreditPackageRevenue float64 `json:"credit_package_revenue"` // 积分套餐收入
	CreditPackageUsers   int64   `json:"credit_package_users"`   // 积分套餐用户数

	// 总计
	TotalCount   int64   `json:"total_count"`   // 总购买数
	TotalRevenue float64 `json:"total_revenue"` // 总收入
	TotalUsers   int64   `json:"total_users"`   // 总用户数

	// 分类统计
	CategoryStats       []MembershipStats `json:"category_stats"`        // 按分类统计（仅会员，用于饼状图）
	CreditCategoryStats []MembershipStats `json:"credit_category_stats"` // 按积分套餐分类统计（用于饼状图）
	ProductStats        []MembershipStats `json:"product_stats"`         // 按产品统计（会员+积分套餐）
}

// ProductTrendData 产品趋势数据（用于折线图）
type ProductTrendData struct {
	ProductName string              `json:"product_name"` // 产品名称
	Data        []ProductTrendPoint `json:"data"`         // 趋势数据点
}

// ProductTrendPoint 产品趋势数据点
type ProductTrendPoint struct {
	Date    string  `json:"date"`    // 日期
	Count   int64   `json:"count"`   // 购买数量
	Revenue float64 `json:"revenue"` // 收入金额
}

// MembershipTrendPoint 会员趋势数据点
type MembershipTrendPoint struct {
	Date          string            `json:"date"`           // 日期
	CategoryStats []MembershipStats `json:"category_stats"` // 各分类统计
	TotalCount    int64             `json:"total_count"`    // 当日总购买数
	TotalRevenue  float64           `json:"total_revenue"`  // 当日总收入
}

// MembershipTrend 会员趋势数据
type MembershipTrend struct {
	StartDate string                 `json:"start_date"` // 开始日期
	EndDate   string                 `json:"end_date"`   // 结束日期
	Period    string                 `json:"period"`     // 统计周期：day/week/month
	Data      []MembershipTrendPoint `json:"data"`       // 趋势数据点
	Summary   MembershipOverview     `json:"summary"`    // 汇总数据
}

// ProductSalesTrend 产品销售趋势数据（用于折线图）
type ProductSalesTrend struct {
	StartDate string             `json:"start_date"` // 开始日期
	EndDate   string             `json:"end_date"`   // 结束日期
	Period    string             `json:"period"`     // 统计周期：day/week/month
	Products  []ProductTrendData `json:"products"`   // 各产品趋势数据
	Summary   MembershipOverview `json:"summary"`    // 汇总数据
}

// GetMembershipOverview 获取会员购买概览（包含会员和积分套餐）
// startDate和endDate为可选，如果为空则统计全部数据
func (s *MembershipService) GetMembershipOverview(startDate, endDate *time.Time) (*MembershipOverview, error) {
	loc := time.FixedZone("CST", 8*60*60)

	// 使用goroutine并行查询会员和积分套餐数据
	var wg sync.WaitGroup
	var mu sync.Mutex
	var membershipStats []MembershipStats
	var creditPackageStats []MembershipStats
	var membershipErr, creditErr error

	// 1. 查询会员统计（订阅服务）
	wg.Add(1)
	go func() {
		defer wg.Done()
		stats, err := s.getMembershipStats(startDate, endDate, loc)
		mu.Lock()
		membershipStats = stats
		membershipErr = err
		mu.Unlock()
	}()

	// 2. 查询积分套餐统计（从Trade表，trade_type = 'recharge'）
	wg.Add(1)
	go func() {
		defer wg.Done()
		stats, err := s.getCreditPackageStats(startDate, endDate, loc)
		mu.Lock()
		creditPackageStats = stats
		creditErr = err
		mu.Unlock()
	}()

	wg.Wait()

	if membershipErr != nil {
		return nil, membershipErr
	}
	if creditErr != nil {
		return nil, creditErr
	}

	// 合并统计数据
	allProductStats := make([]MembershipStats, 0, len(membershipStats)+len(creditPackageStats))
	allProductStats = append(allProductStats, membershipStats...)
	allProductStats = append(allProductStats, creditPackageStats...)

	// 计算会员总计
	var membershipCount, membershipUsers int64
	var membershipRevenue float64
	for _, stat := range membershipStats {
		membershipCount += stat.Count
		membershipRevenue += stat.Revenue
		membershipUsers += stat.UniqueUsers
	}

	// 计算积分套餐总计
	var creditCount, creditUsers int64
	var creditRevenue float64
	for _, stat := range creditPackageStats {
		creditCount += stat.Count
		creditRevenue += stat.Revenue
		creditUsers += stat.UniqueUsers
	}

	// 计算总计数
	totalCount := membershipCount + creditCount

	// 总收入需要按交易去重统计，避免一个交易关联多个产品时重复计算
	// 使用子查询先找到所有符合条件的交易ID，然后统计金额
	var totalRevenueResult struct {
		Total float64
	}
	minDate := time.Date(2025, 7, 1, 0, 0, 0, 0, loc)

	// 构建日期条件
	var dateCondition string
	var dateArgs []interface{}
	var startOfDay, endOfDay time.Time
	if startDate != nil && endDate != nil {
		startInLoc := startDate.In(loc)
		startOfDay = time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		endInLoc := endDate.In(loc)
		endOfDay = time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		dateCondition = "t.paid_at >= ? AND t.paid_at <= ?"
		dateArgs = []interface{}{startOfDay, endOfDay}
	} else if startDate != nil {
		startInLoc := startDate.In(loc)
		startOfDay = time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		dateCondition = "t.paid_at >= ?"
		dateArgs = []interface{}{startOfDay}
	} else if endDate != nil {
		endInLoc := endDate.In(loc)
		endOfDay = time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		dateCondition = "t.paid_at <= ?"
		dateArgs = []interface{}{endOfDay}
	} else {
		dateCondition = "t.paid_at >= ?"
		dateArgs = []interface{}{minDate}
	}

	totalRevenueErr := repository.DB.Table("trades").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("id IN (SELECT DISTINCT t.id FROM trades t "+
			"JOIN user_productions up ON t.id = up.trade_id "+
			"JOIN productions p ON up.production_id = p.id "+
			"WHERE t.payment_status = ? "+
			"AND t.trade_type != ? "+
			"AND (p.product_type = ? OR p.product_type = ?) "+
			"AND "+dateCondition+")",
			append([]interface{}{"success", "activation", "订阅服务", "积分套餐"}, dateArgs...)...).
		Scan(&totalRevenueResult).Error

	var totalRevenue float64
	if totalRevenueErr == nil {
		totalRevenue = totalRevenueResult.Total
	} else {
		// 如果查询失败，使用累加的方式（可能有重复计算）
		totalRevenue = membershipRevenue + creditRevenue
	}

	// 计算占比（基于总数）
	for i := range allProductStats {
		if totalCount > 0 {
			allProductStats[i].Percentage = float64(allProductStats[i].Count) / float64(totalCount) * 100
		}
	}

	// 扁平化处理：按产品名称展示，不再按分类聚合（仅会员）
	// 这样轻量版、专业版的各个套餐都能单独显示在饼状图中
	categoryStats := make([]MembershipStats, 0, len(membershipStats))
	for _, stat := range membershipStats {
		category := s.getCategoryByProductName(stat.ProductName)
		percentage := 0.0
		if membershipCount > 0 {
			percentage = float64(stat.Count) / float64(membershipCount) * 100
		}
		categoryStats = append(categoryStats, MembershipStats{
			Category:    category,
			ProductName: stat.ProductName,
			Count:       stat.Count,
			Revenue:     stat.Revenue,
			Percentage:  percentage,
			UniqueUsers: stat.UniqueUsers,
		})
	}

	// 按积分套餐分类聚合（用于饼状图）
	// 积分套餐本身就是按产品名称分类的，所以直接使用产品名称作为分类
	creditCategoryStats := make([]MembershipStats, 0, len(creditPackageStats))
	for _, stat := range creditPackageStats {
		percentage := 0.0
		if creditCount > 0 {
			percentage = float64(stat.Count) / float64(creditCount) * 100
		}
		creditCategoryStats = append(creditCategoryStats, MembershipStats{
			Category:    MembershipCategory(stat.ProductName), // 使用产品名称作为分类
			ProductName: stat.ProductName,
			Count:       stat.Count,
			Revenue:     stat.Revenue,
			Percentage:  percentage,
			UniqueUsers: stat.UniqueUsers,
		})
	}

	// 获取总用户数（去重）
	totalUsers := membershipUsers + creditUsers

	return &MembershipOverview{
		MembershipCount:      membershipCount,
		MembershipRevenue:    membershipRevenue,
		MembershipUsers:      membershipUsers,
		CreditPackageCount:   creditCount,
		CreditPackageRevenue: creditRevenue,
		CreditPackageUsers:   creditUsers,
		TotalCount:           totalCount,
		TotalRevenue:         totalRevenue,
		TotalUsers:           totalUsers,
		CategoryStats:        categoryStats,
		CreditCategoryStats:  creditCategoryStats,
		ProductStats:         allProductStats,
	}, nil
}

// getMembershipStats 获取会员统计数据
// 参考GetPaymentOverview的逻辑：统一查询条件和时区处理
func (s *MembershipService) getMembershipStats(startDate, endDate *time.Time, loc *time.Location) ([]MembershipStats, error) {
	// 参考GetPaymentOverview的逻辑：
	// 1. 直接从Trade表查询（只统计微信和支付宝渠道，支付成功的）
	// 2. 关联UserProduction和Production表，筛选product_type = "订阅服务"
	// 3. 使用paid_at进行日期过滤，统一使用CST时区
	// 注意：与GetPaymentOverview保持一致，不排除activation类型
	baseQuery := repository.DB.Table("trades as t").
		Select(`
			p.name as product_name,
			COUNT(*) as count,
			COALESCE(SUM(t.amount), 0) as revenue,
			COUNT(DISTINCT t.user_id) as unique_users
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.trade_type != ?", "activation").
		Where("p.product_type = ?", "订阅服务")

	// 日期范围过滤（使用paid_at，统一使用CST时区，与GetPaymentOverview保持一致）
	if startDate != nil {
		// 确保使用CST时区
		startInLoc := startDate.In(loc)
		startOfDay := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		baseQuery = baseQuery.Where("t.paid_at >= ?", startOfDay)
	}
	if endDate != nil {
		// 确保使用CST时区，结束时间设置为当天的23:59:59.999999999
		endInLoc := endDate.In(loc)
		endOfDay := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		baseQuery = baseQuery.Where("t.paid_at <= ?", endOfDay)
	}

	var productStats []struct {
		ProductName string  `gorm:"column:product_name"`
		Count       int64   `gorm:"column:count"`
		Revenue     float64 `gorm:"column:revenue"`
		UniqueUsers int64   `gorm:"column:unique_users"`
	}

	err := baseQuery.
		Group("p.name").
		Order("count DESC").
		Scan(&productStats).Error

	if err != nil {
		return nil, err
	}

	stats := make([]MembershipStats, 0, len(productStats))
	for _, stat := range productStats {
		stats = append(stats, MembershipStats{
			ProductName: stat.ProductName,
			Count:       stat.Count,
			Revenue:     stat.Revenue,
			UniqueUsers: stat.UniqueUsers,
		})
	}

	return stats, nil
}

// getCreditPackageStats 获取积分套餐统计数据
// 参考GetPaymentOverview的逻辑：统一查询条件和时区处理
func (s *MembershipService) getCreditPackageStats(startDate, endDate *time.Time, loc *time.Location) ([]MembershipStats, error) {
	// 参考GetPaymentOverview的逻辑：
	// 1. 直接从Trade表查询（只统计微信和支付宝渠道，支付成功的）
	// 2. 关联UserProduction和Production表，筛选product_type = "积分套餐"
	// 3. 产品名称匹配products.py中的定义（600积分、1500积分、3000积分）
	// 4. 使用paid_at进行日期过滤，统一使用CST时区
	// 注意：与GetPaymentOverview保持一致，不排除activation类型
	creditPackageNames := []string{"600积分", "1500积分", "3000积分"}

	baseQuery := repository.DB.Table("trades as t").
		Select(`
			p.name as product_name,
			COUNT(*) as count,
			COALESCE(SUM(t.amount), 0) as revenue,
			COUNT(DISTINCT t.user_id) as unique_users
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.trade_type != ?", "activation").
		Where("p.product_type = ?", "积分套餐").
		Where("p.name IN ?", creditPackageNames)

	// 日期范围过滤（使用paid_at，统一使用CST时区，与GetPaymentOverview保持一致）
	if startDate != nil {
		// 确保使用CST时区
		startInLoc := startDate.In(loc)
		startOfDay := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		baseQuery = baseQuery.Where("t.paid_at >= ?", startOfDay)
	}
	if endDate != nil {
		// 确保使用CST时区，结束时间设置为当天的23:59:59.999999999
		endInLoc := endDate.In(loc)
		endOfDay := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)
		baseQuery = baseQuery.Where("t.paid_at <= ?", endOfDay)
	}

	var productStats []struct {
		ProductName string  `gorm:"column:product_name"`
		Count       int64   `gorm:"column:count"`
		Revenue     float64 `gorm:"column:revenue"`
		UniqueUsers int64   `gorm:"column:unique_users"`
	}

	err := baseQuery.
		Group("p.name").
		Order("count DESC").
		Scan(&productStats).Error

	if err != nil {
		return nil, err
	}

	stats := make([]MembershipStats, 0, len(productStats))
	for _, stat := range productStats {
		stats = append(stats, MembershipStats{
			ProductName: stat.ProductName,
			Count:       stat.Count,
			Revenue:     stat.Revenue,
			UniqueUsers: stat.UniqueUsers,
		})
	}

	return stats, nil
}

// GetMembershipTrend 获取会员购买趋势
// period: day/week/month
// 最大跨度限制为1年
func (s *MembershipService) GetMembershipTrend(startDate, endDate time.Time, period string) (*MembershipTrend, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	// 限制最大跨度为1年
	maxDuration := 365 * 24 * time.Hour
	if endInLoc.Sub(startInLoc) > maxDuration {
		endInLoc = startInLoc.Add(maxDuration)
	}

	// 根据period确定日期分组方式（使用paid_at）
	var dateGroupExpr string
	switch period {
	case "week":
		dateGroupExpr = "DATE_FORMAT(t.paid_at, '%Y-%u')"
	case "month":
		dateGroupExpr = "DATE_FORMAT(t.paid_at, '%Y-%m')"
	default: // day
		dateGroupExpr = "DATE(t.paid_at)"
	}

	// 构建查询：按日期和产品分组
	// 参考GetPaymentOverview的逻辑：直接从Trade表查询
	var trendData []struct {
		Date        string  `gorm:"column:date"`
		ProductName string  `gorm:"column:product_name"`
		Count       int64   `gorm:"column:count"`
		Revenue     float64 `gorm:"column:revenue"`
		UniqueUsers int64   `gorm:"column:unique_users"`
	}

	startOfPeriod := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
	endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

	err := repository.DB.Table("trades as t").
		Select(`
			`+dateGroupExpr+` as date,
			p.name as product_name,
			COUNT(*) as count,
			COALESCE(SUM(t.amount), 0) as revenue,
			COUNT(DISTINCT t.user_id) as unique_users
		`).
		Joins("JOIN user_productions up ON t.id = up.trade_id").
		Joins("JOIN productions p ON up.production_id = p.id").
		Where("t.payment_status = ?", "success").
		Where("t.trade_type != ?", "activation").
		Where("p.product_type = ?", "订阅服务").
		Where("t.paid_at >= ? AND t.paid_at <= ?", startOfPeriod, endOfPeriod).
		Group("date, p.name").
		Order("date ASC, count DESC").
		Scan(&trendData).Error

	if err != nil {
		return nil, err
	}

	// 按日期聚合数据
	dateMap := make(map[string]*MembershipTrendPoint)
	for _, data := range trendData {
		if dateMap[data.Date] == nil {
			dateMap[data.Date] = &MembershipTrendPoint{
				Date:          data.Date,
				CategoryStats: make([]MembershipStats, 0),
				TotalCount:    0,
				TotalRevenue:  0,
			}
		}

		point := dateMap[data.Date]
		point.TotalCount += data.Count
		point.TotalRevenue += data.Revenue

		// 添加到产品统计
		category := s.getCategoryByProductName(data.ProductName)
		point.CategoryStats = append(point.CategoryStats, MembershipStats{
			Category:    category,
			ProductName: data.ProductName,
			Count:       data.Count,
			Revenue:     data.Revenue,
			UniqueUsers: data.UniqueUsers,
		})
	}

	// 转换为列表并排序
	trendPoints := make([]MembershipTrendPoint, 0, len(dateMap))
	for _, point := range dateMap {
		// 计算各分类占比
		for i := range point.CategoryStats {
			if point.TotalCount > 0 {
				point.CategoryStats[i].Percentage = float64(point.CategoryStats[i].Count) / float64(point.TotalCount) * 100
			}
		}
		trendPoints = append(trendPoints, *point)
	}

	// 按日期排序（这里简化处理，实际应该按日期字符串排序）
	// 为了性能，使用简单的字符串排序
	for i := 0; i < len(trendPoints)-1; i++ {
		for j := i + 1; j < len(trendPoints); j++ {
			if trendPoints[i].Date > trendPoints[j].Date {
				trendPoints[i], trendPoints[j] = trendPoints[j], trendPoints[i]
			}
		}
	}

	// 计算汇总数据
	summary, err := s.GetMembershipOverview(&startDate, &endDate)
	if err != nil {
		return nil, err
	}

	return &MembershipTrend{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Period:    period,
		Data:      trendPoints,
		Summary:   *summary,
	}, nil
}

// getCategoryByProductName 根据产品名称获取分类
func (s *MembershipService) getCategoryByProductName(productName string) MembershipCategory {
	// 根据products.py中的定义进行分类
	switch {
	case productName == "免费版":
		return CategoryFree
	case productName == "轻量版" || productName == "轻量版体验" || productName == "轻量版年度会员":
		return CategoryLight
	case productName == "专业版" || productName == "专业版半年订阅升级套餐" ||
		productName == "专业版年度会员" || productName == "专业版体验" ||
		productName == "专业版周体验" || productName == "专业版开通测试":
		return CategoryProfessional
	case productName == "种子终身会员":
		return CategoryLifetime
	default:
		// 默认归类为专业版
		return CategoryProfessional
	}
}

// GetProductSalesTrend 获取产品销售趋势（用于折线图）
// 每个产品一条折线，支持会员和积分套餐
// period: day/week/month
// 最大跨度限制为1年
func (s *MembershipService) GetProductSalesTrend(startDate, endDate time.Time, period string) (*ProductSalesTrend, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	// 限制最大跨度为1年
	maxDuration := 365 * 24 * time.Hour
	if endInLoc.Sub(startInLoc) > maxDuration {
		endInLoc = startInLoc.Add(maxDuration)
	}

	// 使用goroutine并行查询会员和积分套餐趋势
	var wg sync.WaitGroup
	var mu sync.Mutex
	membershipTrendData := make(map[string]map[string]ProductTrendPoint) // product -> date -> point
	creditTrendData := make(map[string]map[string]ProductTrendPoint)
	var membershipErr, creditErr error

	// 1. 查询会员趋势
	wg.Add(1)
	go func() {
		defer wg.Done()
		startOfPeriod := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

		var trendData []struct {
			Date        string  `gorm:"column:date"`
			ProductName string  `gorm:"column:product_name"`
			Count       int64   `gorm:"column:count"`
			Revenue     float64 `gorm:"column:revenue"`
		}

		// 根据period调整dateGroupExpr（使用paid_at）
		var groupExpr string
		switch period {
		case "week":
			groupExpr = "DATE_FORMAT(t.paid_at, '%Y-%u')"
		case "month":
			groupExpr = "DATE_FORMAT(t.paid_at, '%Y-%m')"
		default:
			groupExpr = "DATE(t.paid_at)"
		}

		// 参考GetPaymentOverview的逻辑：直接从Trade表查询
		err := repository.DB.Table("trades as t").
			Select(`
				`+groupExpr+` as date,
				p.name as product_name,
				COUNT(*) as count,
				COALESCE(SUM(t.amount), 0) as revenue
			`).
			Joins("JOIN user_productions up ON t.id = up.trade_id").
			Joins("JOIN productions p ON up.production_id = p.id").
			Where("t.payment_status = ?", "success").
			Where("t.trade_type != ?", "activation").
			Where("p.product_type = ?", "订阅服务").
			Where("t.paid_at >= ? AND t.paid_at <= ?", startOfPeriod, endOfPeriod).
			Group("date, p.name").
			Order("date ASC, count DESC").
			Scan(&trendData).Error

		if err != nil {
			mu.Lock()
			membershipErr = err
			mu.Unlock()
			return
		}

		mu.Lock()
		for _, data := range trendData {
			if membershipTrendData[data.ProductName] == nil {
				membershipTrendData[data.ProductName] = make(map[string]ProductTrendPoint)
			}
			membershipTrendData[data.ProductName][data.Date] = ProductTrendPoint{
				Date:    data.Date,
				Count:   data.Count,
				Revenue: data.Revenue,
			}
		}
		mu.Unlock()
	}()

	// 2. 查询积分套餐趋势
	wg.Add(1)
	go func() {
		defer wg.Done()
		startOfPeriod := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
		endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

		var trendData []struct {
			Date        string  `gorm:"column:date"`
			ProductName string  `gorm:"column:product_name"`
			Count       int64   `gorm:"column:count"`
			Revenue     float64 `gorm:"column:revenue"`
		}

		// 积分套餐产品名称列表（从products.py）
		creditPackageNames := []string{"600积分", "1500积分", "3000积分"}

		// 根据period调整dateGroupExpr（使用paid_at）
		var groupExpr string
		switch period {
		case "week":
			groupExpr = "DATE_FORMAT(t.paid_at, '%Y-%u')"
		case "month":
			groupExpr = "DATE_FORMAT(t.paid_at, '%Y-%m')"
		default:
			groupExpr = "DATE(t.paid_at)"
		}

		// 参考GetPaymentOverview的逻辑：直接从Trade表查询
		err := repository.DB.Table("trades as t").
			Select(`
				`+groupExpr+` as date,
				p.name as product_name,
				COUNT(*) as count,
				COALESCE(SUM(t.amount), 0) as revenue
			`).
			Joins("JOIN user_productions up ON t.id = up.trade_id").
			Joins("JOIN productions p ON up.production_id = p.id").
			Where("t.payment_status = ?", "success").
			Where("t.trade_type != ?", "activation").
			Where("p.product_type = ?", "积分套餐").
			Where("p.name IN ?", creditPackageNames).
			Where("t.paid_at >= ? AND t.paid_at <= ?", startOfPeriod, endOfPeriod).
			Group("date, p.name").
			Order("date ASC, count DESC").
			Scan(&trendData).Error

		if err != nil {
			mu.Lock()
			creditErr = err
			mu.Unlock()
			return
		}

		mu.Lock()
		for _, data := range trendData {
			if creditTrendData[data.ProductName] == nil {
				creditTrendData[data.ProductName] = make(map[string]ProductTrendPoint)
			}
			creditTrendData[data.ProductName][data.Date] = ProductTrendPoint{
				Date:    data.Date,
				Count:   data.Count,
				Revenue: data.Revenue,
			}
		}
		mu.Unlock()
	}()

	wg.Wait()

	if membershipErr != nil {
		return nil, membershipErr
	}
	if creditErr != nil {
		return nil, creditErr
	}

	// 生成所有日期列表
	allDates := make(map[string]bool)
	for _, productData := range membershipTrendData {
		for date := range productData {
			allDates[date] = true
		}
	}
	for _, productData := range creditTrendData {
		for date := range productData {
			allDates[date] = true
		}
	}

	// 构建产品趋势数据列表
	products := make([]ProductTrendData, 0)

	// 添加会员产品
	for productName, dateMap := range membershipTrendData {
		points := make([]ProductTrendPoint, 0)
		for date := range allDates {
			if point, exists := dateMap[date]; exists {
				points = append(points, point)
			} else {
				points = append(points, ProductTrendPoint{
					Date:    date,
					Count:   0,
					Revenue: 0,
				})
			}
		}
		// 按日期排序
		for i := 0; i < len(points)-1; i++ {
			for j := i + 1; j < len(points); j++ {
				if points[i].Date > points[j].Date {
					points[i], points[j] = points[j], points[i]
				}
			}
		}
		products = append(products, ProductTrendData{
			ProductName: productName,
			Data:        points,
		})
	}

	// 添加积分套餐产品
	for productName, dateMap := range creditTrendData {
		points := make([]ProductTrendPoint, 0)
		for date := range allDates {
			if point, exists := dateMap[date]; exists {
				points = append(points, point)
			} else {
				points = append(points, ProductTrendPoint{
					Date:    date,
					Count:   0,
					Revenue: 0,
				})
			}
		}
		// 按日期排序
		for i := 0; i < len(points)-1; i++ {
			for j := i + 1; j < len(points); j++ {
				if points[i].Date > points[j].Date {
					points[i], points[j] = points[j], points[i]
				}
			}
		}
		products = append(products, ProductTrendData{
			ProductName: productName,
			Data:        points,
		})
	}

	// 获取汇总数据
	summary, err := s.GetMembershipOverview(&startDate, &endDate)
	if err != nil {
		return nil, err
	}

	return &ProductSalesTrend{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Period:    period,
		Products:  products,
		Summary:   *summary,
	}, nil
}

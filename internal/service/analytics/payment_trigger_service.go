package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"fmt"
	"time"
)

// PaymentTriggerService 首充触发点分析服务
type PaymentTriggerService struct{}

// NewPaymentTriggerService 创建首充触发点分析服务
func NewPaymentTriggerService() *PaymentTriggerService {
	return &PaymentTriggerService{}
}

// FirstPaymentAnalysis 首次充值分析结果
type FirstPaymentAnalysis struct {
	UserID              string                       `json:"user_id"`                // 用户ID
	Phone               *string                      `json:"phone"`                  // 手机号
	Avatar              *string                      `json:"avatar"`                 // 头像
	FirstPaymentTime    time.Time                    `json:"first_payment_time"`     // 首次充值时间
	CreditsBeforePayment int                          `json:"credits_before_payment"` // 首充前消耗的积分总数
	ProductName         string                       `json:"product_name"`           // 购买的产品名称
	ProductType         string                       `json:"product_type"`           // 产品类型
	ProductPrice        float64                      `json:"product_price"`          // 产品价格
	SceneConsumption    map[string]int               `json:"scene_consumption"`      // 按场景分组的积分消耗
}

// FirstPaymentSummary 首充分析汇总
type FirstPaymentSummary struct {
	TotalPayingUsers        int                       `json:"total_paying_users"`         // 总付费用户数
	AvgCreditsBeforePayment float64                   `json:"avg_credits_before_payment"` // 平均首充前消耗积分
	MedianCreditsBeforePayment int                    `json:"median_credits_before_payment"` // 中位数首充前消耗积分
	TopScenes               []SceneConsumption        `json:"top_scenes"`                 // Top场景消耗排名
	ProductDistribution     []ProductStats            `json:"product_distribution"`       // 产品购买分布
	CreditRangeDistribution []CreditRangeStats        `json:"credit_range_distribution"`  // 积分消耗区间分布
}

// SceneConsumption 场景消耗统计
type SceneConsumption struct {
	ServiceCode  string  `json:"service_code"`  // 服务代码
	ServiceName  string  `json:"service_name"`  // 服务名称
	TotalCredits int     `json:"total_credits"` // 总消耗积分
	UserCount    int     `json:"user_count"`    // 使用用户数
	AvgCredits   float64 `json:"avg_credits"`   // 平均消耗
}

// ProductStats 产品统计
type ProductStats struct {
	ProductName  string  `json:"product_name"`  // 产品名称
	ProductType  string  `json:"product_type"`  // 产品类型
	UserCount    int     `json:"user_count"`    // 购买用户数
	Percentage   float64 `json:"percentage"`    // 占比
	AvgPrice     float64 `json:"avg_price"`     // 平均价格
}

// CreditRangeStats 积分消耗区间统计
type CreditRangeStats struct {
	RangeStart int     `json:"range_start"` // 区间开始
	RangeEnd   int     `json:"range_end"`   // 区间结束
	UserCount  int     `json:"user_count"`  // 用户数
	Percentage float64 `json:"percentage"`  // 占比
}

// GetFirstPaymentAnalysis 获取首充触发点分析
// startDate, endDate: 分析首次充值发生在这个时间范围内的用户
func (s *PaymentTriggerService) GetFirstPaymentAnalysis(startDate, endDate time.Time) (*FirstPaymentSummary, []FirstPaymentAnalysis, error) {
	// 1. 查询所有付费用户及其首次充值信息
	var firstPayments []struct {
		UserID           string    `gorm:"column:user_id"`
		FirstPaymentTime time.Time `gorm:"column:first_payment_time"`
		TradeID          int       `gorm:"column:trade_id"`
	}

	// 查询每个用户的首次充值记录（只统计成功的充值）
	err := repository.DB.Raw(`
		SELECT 
			user_id,
			MIN(paid_at) as first_payment_time,
			(SELECT id FROM trades t2 WHERE t2.user_id = t1.user_id AND t2.paid_at = MIN(t1.paid_at) LIMIT 1) as trade_id
		FROM trades t1
		WHERE trade_type = 'recharge'
			AND payment_status = 'success'
			AND paid_at IS NOT NULL
			AND paid_at >= ?
			AND paid_at <= ?
		GROUP BY user_id
	`, startDate, endDate).Scan(&firstPayments).Error

	if err != nil {
		return nil, nil, fmt.Errorf("查询首次充值记录失败: %w", err)
	}

	if len(firstPayments) == 0 {
		return &FirstPaymentSummary{
			TotalPayingUsers: 0,
			TopScenes:        []SceneConsumption{},
			ProductDistribution: []ProductStats{},
			CreditRangeDistribution: []CreditRangeStats{},
		}, []FirstPaymentAnalysis{}, nil
	}

	// 2. 批量查询所有用户的基本信息（手机号、头像）
	userIDs := make([]string, len(firstPayments))
	tradeIDs := make([]int, len(firstPayments))
	userPaymentTimeMap := make(map[string]time.Time)
	userTradeIDMap := make(map[string]int)
	
	for i, fp := range firstPayments {
		userIDs[i] = fp.UserID
		tradeIDs[i] = fp.TradeID
		userPaymentTimeMap[fp.UserID] = fp.FirstPaymentTime
		userTradeIDMap[fp.UserID] = fp.TradeID
	}

	// 批量查询用户信息
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

	// 批量查询所有积分消耗记录
	var allCreditRecords []models.CreditRecord
	err = repository.DB.Where(
		"user_id IN ? AND record_type = ?",
		userIDs,
		models.CreditConsumption,
	).Find(&allCreditRecords).Error

	if err != nil {
		return nil, nil, fmt.Errorf("批量查询积分消耗记录失败: %w", err)
	}

	// 批量查询产品信息
	var userProductions []models.UserProduction
	err = repository.DB.Where("trade_id IN ?", tradeIDs).Find(&userProductions).Error
	if err != nil {
		repository.Warnf("批量查询用户产品信息失败: %v", err)
	}

	// 获取所有产品ID
	productIDs := make([]int, 0, len(userProductions))
	for _, up := range userProductions {
		productIDs = append(productIDs, up.ProductionID)
	}

	// 批量查询产品详情
	var productions []models.Production
	productMap := make(map[int]*models.Production)
	if len(productIDs) > 0 {
		err = repository.DB.Where("id IN ?", productIDs).Find(&productions).Error
		if err == nil {
			for i := range productions {
				productMap[productions[i].ID] = &productions[i]
			}
		}
	}

	// 建立trade_id到产品的映射
	tradeProductMap := make(map[int]*models.Production)
	for _, up := range userProductions {
		if prod, exists := productMap[up.ProductionID]; exists {
			tradeProductMap[up.TradeID] = prod
		}
	}

	// 3. 组织数据：按用户分组积分消耗记录
	userCreditMap := make(map[string][]models.CreditRecord)
	for _, record := range allCreditRecords {
		// 只统计首充前的记录
		if paymentTime, exists := userPaymentTimeMap[record.UserID]; exists {
			if record.CreatedAt.Before(paymentTime) {
				userCreditMap[record.UserID] = append(userCreditMap[record.UserID], record)
			}
		}
	}

	// 4. 计算每个用户的分析结果
	analyses := make([]FirstPaymentAnalysis, 0, len(firstPayments))
	totalCredits := 0
	creditsSlice := make([]int, 0, len(firstPayments))
	
	// 用于汇总统计
	sceneMap := make(map[string]*SceneConsumption)
	prodStatsMap := make(map[string]*ProductStats)

	for _, fp := range firstPayments {
		// 获取该用户的积分消耗记录
		creditRecords := userCreditMap[fp.UserID]

		// 统计总消耗和场景消耗
		totalConsumed := 0
		sceneConsumption := make(map[string]int)
		for _, record := range creditRecords {
			if record.Credits != nil {
				credits := *record.Credits
				// Credits在消费时是负数，取绝对值
				if credits < 0 {
					credits = -credits
				}
				totalConsumed += credits

				if record.ServiceCode != nil && *record.ServiceCode != "" {
					sceneConsumption[*record.ServiceCode] += credits
				}
			}
		}

		// 获取产品信息
		productName := "未知产品"
		productType := "未知"
		productPrice := 0.0

		if prod, exists := tradeProductMap[fp.TradeID]; exists {
			productName = prod.Name
			productType = prod.ProductType
			productPrice = prod.Price
		}

		// 获取用户信息（手机号、头像）
		var phone *string
		var avatar *string
		if user, exists := userInfoMap[fp.UserID]; exists {
			phone = user.Phone
			avatar = user.Avatar
		}

		// 添加到分析结果
		analyses = append(analyses, FirstPaymentAnalysis{
			UserID:              fp.UserID,
			Phone:               phone,
			Avatar:              avatar,
			FirstPaymentTime:    fp.FirstPaymentTime,
			CreditsBeforePayment: totalConsumed,
			ProductName:         productName,
			ProductType:         productType,
			ProductPrice:        productPrice,
			SceneConsumption:    sceneConsumption,
		})

		totalCredits += totalConsumed
		creditsSlice = append(creditsSlice, totalConsumed)

		// 汇总场景统计
		for serviceCode, credits := range sceneConsumption {
			if scene, exists := sceneMap[serviceCode]; exists {
				scene.TotalCredits += credits
				scene.UserCount++
			} else {
				sceneMap[serviceCode] = &SceneConsumption{
					ServiceCode:  serviceCode,
					TotalCredits: credits,
					UserCount:    1,
				}
			}
		}

		// 汇总产品统计
		key := fmt.Sprintf("%s|%s", productName, productType)
		if prod, exists := prodStatsMap[key]; exists {
			prod.UserCount++
		} else {
			prodStatsMap[key] = &ProductStats{
				ProductName: productName,
				ProductType: productType,
				UserCount:   1,
				AvgPrice:    productPrice,
			}
		}
	}

	// 3. 计算汇总统计
	summary := &FirstPaymentSummary{
		TotalPayingUsers: len(analyses),
	}

	// 计算平均值
	if len(analyses) > 0 {
		summary.AvgCreditsBeforePayment = float64(totalCredits) / float64(len(analyses))
	}

	// 计算中位数
	if len(creditsSlice) > 0 {
		summary.MedianCreditsBeforePayment = calculateMedian(creditsSlice)
	}

	// 批量查询所有场景的服务名称（性能优化）
	serviceCodes := make([]string, 0, len(sceneMap))
	for serviceCode := range sceneMap {
		serviceCodes = append(serviceCodes, serviceCode)
	}
	
	var servicePrices []models.CreditServicePrice
	serviceNameMap := make(map[string]string)
	if len(serviceCodes) > 0 {
		err := repository.DB.Where("service_code IN ?", serviceCodes).Find(&servicePrices).Error
		if err == nil {
			for _, sp := range servicePrices {
				if sp.Name != nil {
					serviceNameMap[sp.ServiceCode] = *sp.Name
				}
			}
		}
	}

	// 填充场景统计
	summary.TopScenes = make([]SceneConsumption, 0, len(sceneMap))
	for _, scene := range sceneMap {
		scene.AvgCredits = float64(scene.TotalCredits) / float64(scene.UserCount)
		
		// 从map中获取服务名称
		if name, exists := serviceNameMap[scene.ServiceCode]; exists {
			scene.ServiceName = name
		} else {
			scene.ServiceName = scene.ServiceCode
		}
		
		summary.TopScenes = append(summary.TopScenes, *scene)
	}

	// 按总消耗排序场景
	sortScenesByCredits(summary.TopScenes)

	// 填充产品统计
	summary.ProductDistribution = make([]ProductStats, 0, len(prodStatsMap))
	for _, prod := range prodStatsMap {
		prod.Percentage = float64(prod.UserCount) / float64(len(analyses)) * 100
		summary.ProductDistribution = append(summary.ProductDistribution, *prod)
	}

	// 按用户数排序产品
	sortProductsByUserCount(summary.ProductDistribution)

	// 计算积分消耗区间分布
	summary.CreditRangeDistribution = calculateCreditRangeDistribution(creditsSlice)

	return summary, analyses, nil
}

// GetUserFirstPaymentTrigger 获取单个用户的首充触发点分析
func (s *PaymentTriggerService) GetUserFirstPaymentTrigger(userID string) (*FirstPaymentAnalysis, error) {
	// 查询用户首次充值信息
	var firstPayment struct {
		FirstPaymentTime time.Time `gorm:"column:first_payment_time"`
		TradeID          int       `gorm:"column:trade_id"`
	}

	err := repository.DB.Raw(`
		SELECT 
			MIN(paid_at) as first_payment_time,
			(SELECT id FROM trades t2 WHERE t2.user_id = ? AND t2.paid_at = MIN(t1.paid_at) LIMIT 1) as trade_id
		FROM trades t1
		WHERE user_id = ?
			AND trade_type = 'recharge'
			AND payment_status = 'success'
			AND paid_at IS NOT NULL
	`, userID, userID).Scan(&firstPayment).Error

	if err != nil || firstPayment.FirstPaymentTime.IsZero() {
		return nil, fmt.Errorf("用户未找到首次充值记录")
	}

	// 查询首充前的积分消耗记录
	var creditRecords []models.CreditRecord
	err = repository.DB.Where(
		"user_id = ? AND record_type = ? AND created_at < ?",
		userID,
		models.CreditConsumption,
		firstPayment.FirstPaymentTime,
	).Find(&creditRecords).Error

	if err != nil {
		return nil, fmt.Errorf("查询积分消耗记录失败: %w", err)
	}

	// 统计总消耗和场景消耗
	totalConsumed := 0
	sceneConsumption := make(map[string]int)
	for _, record := range creditRecords {
		if record.Credits != nil {
			credits := *record.Credits
			if credits < 0 {
				credits = -credits
			}
			totalConsumed += credits

			if record.ServiceCode != nil && *record.ServiceCode != "" {
				sceneConsumption[*record.ServiceCode] += credits
			}
		}
	}

	// 查询用户信息（手机号、头像）
	var user models.User
	var phone *string
	var avatar *string
	err = repository.DB.Where("user_id = ?", userID).First(&user).Error
	if err == nil {
		phone = user.Phone
		avatar = user.Avatar
	}

	// 查询首次购买的产品信息
	var userProduction models.UserProduction
	var production models.Production
	productName := "未知产品"
	productType := "未知"
	productPrice := 0.0

	err = repository.DB.Where("trade_id = ?", firstPayment.TradeID).First(&userProduction).Error
	if err == nil {
		err = repository.DB.Where("id = ?", userProduction.ProductionID).First(&production).Error
		if err == nil {
			productName = production.Name
			productType = production.ProductType
			productPrice = production.Price
		}
	}

	return &FirstPaymentAnalysis{
		UserID:              userID,
		Phone:               phone,
		Avatar:              avatar,
		FirstPaymentTime:    firstPayment.FirstPaymentTime,
		CreditsBeforePayment: totalConsumed,
		ProductName:         productName,
		ProductType:         productType,
		ProductPrice:        productPrice,
		SceneConsumption:    sceneConsumption,
	}, nil
}

// 辅助函数

// calculateMedian 计算中位数
func calculateMedian(numbers []int) int {
	if len(numbers) == 0 {
		return 0
	}

	// 复制切片避免修改原数据
	sorted := make([]int, len(numbers))
	copy(sorted, numbers)

	// 简单冒泡排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// sortScenesByCredits 按总积分消耗降序排序
func sortScenesByCredits(scenes []SceneConsumption) {
	for i := 0; i < len(scenes); i++ {
		for j := i + 1; j < len(scenes); j++ {
			if scenes[i].TotalCredits < scenes[j].TotalCredits {
				scenes[i], scenes[j] = scenes[j], scenes[i]
			}
		}
	}
}

// sortProductsByUserCount 按用户数降序排序
func sortProductsByUserCount(products []ProductStats) {
	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			if products[i].UserCount < products[j].UserCount {
				products[i], products[j] = products[j], products[i]
			}
		}
	}
}

// calculateCreditRangeDistribution 计算积分消耗区间分布
func calculateCreditRangeDistribution(credits []int) []CreditRangeStats {
	if len(credits) == 0 {
		return []CreditRangeStats{}
	}

	// 定义区间：0-50, 51-100, 101-200, 201-500, 501-1000, 1000+
	ranges := []struct {
		start int
		end   int
	}{
		{0, 50},
		{51, 100},
		{101, 200},
		{201, 500},
		{501, 1000},
		{1001, 999999},
	}

	rangeCounts := make([]int, len(ranges))
	for _, credit := range credits {
		for i, r := range ranges {
			if credit >= r.start && credit <= r.end {
				rangeCounts[i]++
				break
			}
		}
	}

	result := make([]CreditRangeStats, 0, len(ranges))
	total := len(credits)
	for i, r := range ranges {
		if rangeCounts[i] > 0 {
			result = append(result, CreditRangeStats{
				RangeStart: r.start,
				RangeEnd:   r.end,
				UserCount:  rangeCounts[i],
				Percentage: float64(rangeCounts[i]) / float64(total) * 100,
			})
		}
	}

	return result
}

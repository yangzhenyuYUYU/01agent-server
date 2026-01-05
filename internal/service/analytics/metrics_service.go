package analytics

import (
	"sync"
	"time"
)


// MetricsService 统一的数据分析服务
type MetricsService struct {
	userActivity *UserActivityService
	revenue      *RevenueService
	retention    *RetentionService
	traffic      *TrafficService
	productValue *ProductValueService
	cost         *CostService
}

// NewMetricsService 创建数据分析服务
func NewMetricsService() *MetricsService {
	return &MetricsService{
		userActivity: NewUserActivityService(),
		revenue:      NewRevenueService(),
		retention:    NewRetentionService(),
		traffic:      NewTrafficService(),
		productValue: NewProductValueService(),
		cost:         NewCostService(),
	}
}

// GetMetrics 获取指定指标的数据
// metrics: 要获取的指标列表，如果为空则返回所有启用的指标
// date: 统计日期，如果为空则使用当前日期
func (s *MetricsService) GetMetrics(metrics []MetricKey, date time.Time) (*MetricsResponse, error) {
	if date.IsZero() {
		date = time.Now()
	}

	// 获取所有指标信息
	allMetrics := GetMetricInfo()

	// 确定要计算的指标
	metricsToCalculate := make(map[MetricKey]bool)
	if len(metrics) == 0 {
		// 如果没有指定，则计算所有启用的指标
		for key, info := range allMetrics {
			if info.Enabled {
				metricsToCalculate[key] = true
			}
		}
	} else {
		// 只计算指定的指标
		for _, key := range metrics {
			if info, exists := allMetrics[key]; exists && info.Enabled {
				metricsToCalculate[key] = true
			}
		}
	}

	// 使用goroutine并行计算各个指标
	var wg sync.WaitGroup
	var mu sync.Mutex
	metricsMap := make(map[MetricKey]MetricValue)
	errors := make([]error, 0)

	// 计算用户活跃度指标
	if metricsToCalculate[MetricDAU] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.userActivity.GetDAU(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricDAU] = MetricValue{
				Key:       MetricDAU,
				Value:     value,
				Unit:      "人",
				Timestamp: date,
				Dimension: DimensionUserActivity,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricWAU] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.userActivity.GetWAU(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricWAU] = MetricValue{
				Key:       MetricWAU,
				Value:     value,
				Unit:      "人",
				Timestamp: date,
				Dimension: DimensionUserActivity,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricMAU] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.userActivity.GetMAU(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricMAU] = MetricValue{
				Key:       MetricMAU,
				Value:     value,
				Unit:      "人",
				Timestamp: date,
				Dimension: DimensionUserActivity,
			}
			mu.Unlock()
		}()
	}

	// 计算核心营收指标
	if metricsToCalculate[MetricMRR] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.revenue.GetMRR(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricMRR] = MetricValue{
				Key:       MetricMRR,
				Value:     value,
				Unit:      "元",
				Timestamp: date,
				Dimension: DimensionCoreRevenue,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricNewPayingUsers] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.revenue.GetNewPayingUsers(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricNewPayingUsers] = MetricValue{
				Key:       MetricNewPayingUsers,
				Value:     value,
				Unit:      "人",
				Timestamp: date,
				Dimension: DimensionCoreRevenue,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricPaymentConversion] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.revenue.GetPaymentConversionRate(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricPaymentConversion] = MetricValue{
				Key:       MetricPaymentConversion,
				Value:     value,
				Unit:      "%",
				Timestamp: date,
				Dimension: DimensionCoreRevenue,
			}
			mu.Unlock()
		}()
	}

	// 计算留存指标
	if metricsToCalculate[MetricRetentionDay1] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.retention.GetDay1Retention(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricRetentionDay1] = MetricValue{
				Key:       MetricRetentionDay1,
				Value:     value,
				Unit:      "%",
				Timestamp: date,
				Dimension: DimensionRetention,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricRetentionWeek1] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.retention.GetWeek1Retention(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricRetentionWeek1] = MetricValue{
				Key:       MetricRetentionWeek1,
				Value:     value,
				Unit:      "%",
				Timestamp: date,
				Dimension: DimensionRetention,
			}
			mu.Unlock()
		}()
	}

	// 计算产品价值指标
	if metricsToCalculate[MetricTotalGenerations] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.productValue.GetTotalGenerations(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricTotalGenerations] = MetricValue{
				Key:       MetricTotalGenerations,
				Value:     value,
				Unit:      "次",
				Timestamp: date,
				Dimension: DimensionUserActivity,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricAvgTasksPerUser] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.productValue.GetAvgTasksPerUser(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricAvgTasksPerUser] = MetricValue{
				Key:       MetricAvgTasksPerUser,
				Value:     value,
				Unit:      "次/人",
				Timestamp: date,
				Dimension: DimensionProductValue,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricAdoptionRate] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.productValue.GetAdoptionRate(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricAdoptionRate] = MetricValue{
				Key:       MetricAdoptionRate,
				Value:     value,
				Unit:      "%",
				Timestamp: date,
				Dimension: DimensionProductValue,
			}
			mu.Unlock()
		}()
	}

	// 计算成本监控指标
	if metricsToCalculate[MetricCostPerUserToken] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.cost.GetCostPerUserToken(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricCostPerUserToken] = MetricValue{
				Key:       MetricCostPerUserToken,
				Value:     value,
				Unit:      "积分",
				Timestamp: date,
				Dimension: DimensionCostMonitoring,
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricTaskErrorRate] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := s.cost.GetTaskErrorRate(date)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			metricsMap[MetricTaskErrorRate] = MetricValue{
				Key:       MetricTaskErrorRate,
				Value:     value,
				Unit:      "%",
				Timestamp: date,
				Dimension: DimensionCostMonitoring,
			}
			mu.Unlock()
		}()
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return nil, errors[0]
	}

	// 按维度分组（可选）
	dimensions := make(map[MetricDimension][]MetricValue)
	for _, metricValue := range metricsMap {
		dimensions[metricValue.Dimension] = append(dimensions[metricValue.Dimension], metricValue)
	}

	return &MetricsResponse{
		Date:       date.Format("2006-01-02"),
		Metrics:    metricsMap,
		Dimensions: dimensions,
	}, nil
}

// GetMetricsByDateRange 获取指定日期范围内的指标数据
// 优化：使用批量查询替代逐日查询，大幅提升性能
func (s *MetricsService) GetMetricsByDateRange(metrics []MetricKey, startDate, endDate time.Time) ([]*MetricsResponse, error) {
	// 计算日期范围
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	// 获取所有指标信息
	allMetrics := GetMetricInfo()

	// 确定要计算的指标
	metricsToCalculate := make(map[MetricKey]bool)
	if len(metrics) == 0 {
		// 如果没有指定，则计算所有启用的指标
		for key, info := range allMetrics {
			if info.Enabled {
				metricsToCalculate[key] = true
			}
		}
	} else {
		// 只计算指定的指标
		for _, key := range metrics {
			if info, exists := allMetrics[key]; exists && info.Enabled {
				metricsToCalculate[key] = true
			}
		}
	}

	// 生成所有日期列表
	var dates []time.Time
	currentDate := startInLoc
	for !currentDate.After(endInLoc) {
		dates = append(dates, currentDate)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// 使用批量查询方法获取数据（大幅减少数据库查询次数）
	var wg sync.WaitGroup
	var mu sync.Mutex
	dateMetricsMap := make(map[string]map[MetricKey]MetricValue) // date -> metrics map
	errors := make([]error, 0)

	// 初始化所有日期的map
	for _, date := range dates {
		dateStr := date.Format("2006-01-02")
		dateMetricsMap[dateStr] = make(map[MetricKey]MetricValue)
	}

	// 批量获取用户活跃度指标（使用批量查询）
	if metricsToCalculate[MetricDAU] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dauMap, err := s.userActivity.GetDAUByDateRange(startInLoc, endInLoc)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			for dateStr, value := range dauMap {
				if metricsMap, exists := dateMetricsMap[dateStr]; exists {
					// 解析日期字符串获取正确的时间戳
					if date, err := time.ParseInLocation("2006-01-02", dateStr, loc); err == nil {
						metricsMap[MetricDAU] = MetricValue{
							Key:       MetricDAU,
							Value:     value,
							Unit:      "人",
							Timestamp: date,
							Dimension: DimensionUserActivity,
						}
					}
				}
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricWAU] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wauMap, err := s.userActivity.GetWAUByDateRange(startInLoc, endInLoc)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			for dateStr, value := range wauMap {
				if metricsMap, exists := dateMetricsMap[dateStr]; exists {
					if date, err := time.ParseInLocation("2006-01-02", dateStr, loc); err == nil {
						metricsMap[MetricWAU] = MetricValue{
							Key:       MetricWAU,
							Value:     value,
							Unit:      "人",
							Timestamp: date,
							Dimension: DimensionUserActivity,
						}
					}
				}
			}
			mu.Unlock()
		}()
	}

	if metricsToCalculate[MetricMAU] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mauMap, err := s.userActivity.GetMAUByDateRange(startInLoc, endInLoc)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			for dateStr, value := range mauMap {
				if metricsMap, exists := dateMetricsMap[dateStr]; exists {
					if date, err := time.ParseInLocation("2006-01-02", dateStr, loc); err == nil {
						metricsMap[MetricMAU] = MetricValue{
							Key:       MetricMAU,
							Value:     value,
							Unit:      "人",
							Timestamp: date,
							Dimension: DimensionUserActivity,
						}
					}
				}
			}
			mu.Unlock()
		}()
	}

	// 对于其他指标，使用并发逐日查询（但限制并发数以提高性能）
	// 只查询非用户活跃度指标
	nonActivityMetrics := make([]MetricKey, 0)
	for key := range metricsToCalculate {
		if key != MetricDAU && key != MetricWAU && key != MetricMAU {
			nonActivityMetrics = append(nonActivityMetrics, key)
		}
	}
	
	if len(nonActivityMetrics) > 0 {
		// 限制并发数，避免数据库连接耗尽
		maxConcurrency := 3
		if len(dates) > 30 {
			maxConcurrency = 2 // 大范围时进一步降低并发
		}
		semaphore := make(chan struct{}, maxConcurrency)
		
		for _, date := range dates {
			wg.Add(1)
			go func(d time.Time) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				response, err := s.GetMetrics(nonActivityMetrics, d)
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					return
				}
				
				mu.Lock()
				dateStr := d.Format("2006-01-02")
				if metricsMap, exists := dateMetricsMap[dateStr]; exists {
					for key, value := range response.Metrics {
						value.Timestamp = d
						metricsMap[key] = value
					}
				}
				mu.Unlock()
			}(date)
		}
	}

	wg.Wait()

	// 构建结果列表
	results := make([]*MetricsResponse, 0, len(dates))
	for _, date := range dates {
		dateStr := date.Format("2006-01-02")
		metricsMap := dateMetricsMap[dateStr]
		
		// 按维度分组
		dimensions := make(map[MetricDimension][]MetricValue)
		for _, metricValue := range metricsMap {
			dimensions[metricValue.Dimension] = append(dimensions[metricValue.Dimension], metricValue)
		}
		
		results = append(results, &MetricsResponse{
			Date:       dateStr,
			Metrics:    metricsMap,
			Dimensions: dimensions,
		})
	}

	// 如果有错误，但至少有一些成功的结果，返回部分结果而不是错误
	if len(errors) > 0 && len(results) == 0 {
		return nil, errors[0]
	}

	return results, nil
}

// GetMetricInfoList 获取所有指标信息列表
func (s *MetricsService) GetMetricInfoList() []MetricInfo {
	allMetrics := GetMetricInfo()
	infoList := make([]MetricInfo, 0, len(allMetrics))
	for _, info := range allMetrics {
		infoList = append(infoList, info)
	}
	return infoList
}

// GetTrafficSourceDistribution 获取注册来源分布
func (s *MetricsService) GetTrafficSourceDistribution(startDate, endDate time.Time) ([]SourceDistribution, error) {
	return s.traffic.GetRegistrationSourceDistribution(startDate, endDate)
}

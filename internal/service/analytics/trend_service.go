package analytics

import (
	"sort"
	"sync"
	"time"
)

// TrendService 趋势数据服务
type TrendService struct {
	userActivity *UserActivityService
}

// NewTrendService 创建趋势数据服务
func NewTrendService() *TrendService {
	return &TrendService{
		userActivity: NewUserActivityService(),
	}
}

// TrendDataPoint 趋势数据点
type TrendDataPoint struct {
	Date       string   `json:"date"`                 // 日期字符串，格式：YYYY-MM-DD
	DAU        int64    `json:"dau"`                  // 日活跃用户数
	WAU        *int64   `json:"wau,omitempty"`        // 周活跃用户数（可选）
	MAU        *int64   `json:"mau,omitempty"`        // 月活跃用户数（可选）
	Stickiness *float64 `json:"stickiness,omitempty"` // 用户粘性 (DAU/MAU * 100%)
}

// ActivityTrendResponse 活跃度趋势响应
type ActivityTrendResponse struct {
	Period    string           `json:"period"`     // 统计周期：day/week/month
	StartDate string           `json:"start_date"` // 开始日期
	EndDate   string           `json:"end_date"`   // 结束日期
	Data      []TrendDataPoint `json:"data"`       // 趋势数据点列表
	Summary   TrendSummary     `json:"summary"`    // 汇总数据
}

// TrendSummary 趋势汇总数据
type TrendSummary struct {
	AvgDAU     float64  `json:"avg_dau"`               // 平均日活
	MaxDAU     int64    `json:"max_dau"`               // 最大日活
	MinDAU     int64    `json:"min_dau"`               // 最小日活
	CurrentDAU int64    `json:"current_dau"`           // 当前日活
	CurrentMAU int64    `json:"current_mau"`           // 当前月活
	Stickiness float64  `json:"stickiness"`            // 当前粘性
	GrowthRate *float64 `json:"growth_rate,omitempty"` // 增长率（相比第一个数据点）
}

// GetActivityTrend 获取用户活跃度趋势数据
// period: day/week/month
// startDate, endDate: 日期范围
// includeWAU: 是否包含周活数据
// includeMAU: 是否包含月活数据
func (s *TrendService) GetActivityTrend(period string, startDate, endDate time.Time, includeWAU, includeMAU bool) (*ActivityTrendResponse, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	// 生成日期列表
	var dates []time.Time
	currentDate := startInLoc

	switch period {
	case "day":
		// 按天统计
		for !currentDate.After(endInLoc) {
			dates = append(dates, currentDate)
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	case "week":
		// 按周统计（每周的第一天）
		for !currentDate.After(endInLoc) {
			dates = append(dates, currentDate)
			currentDate = currentDate.AddDate(0, 0, 7)
		}
	case "month":
		// 按月统计（每月的第一天）
		for !currentDate.After(endInLoc) {
			dates = append(dates, currentDate)
			// 计算下个月的第一天
			if currentDate.Month() == 12 {
				currentDate = time.Date(currentDate.Year()+1, 1, 1, 0, 0, 0, 0, loc)
			} else {
				currentDate = time.Date(currentDate.Year(), currentDate.Month()+1, 1, 0, 0, 0, 0, loc)
			}
		}
	default:
		period = "day"
		// 默认按天
		for !currentDate.After(endInLoc) {
			dates = append(dates, currentDate)
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	// 使用批量查询优化性能，减少数据库查询次数
	// 并行查询DAU、WAU、MAU的批量数据
	var dauMap, wauMap, mauMap map[string]int64
	var dauErr, wauErr, mauErr error

	var wg sync.WaitGroup

	// 并行查询DAU
	wg.Add(1)
	go func() {
		defer wg.Done()
		dauMap, dauErr = s.userActivity.GetDAUByDateRange(startInLoc, endInLoc)
	}()

	// 并行查询WAU（如果需要）
	if includeWAU {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wauMap, wauErr = s.userActivity.GetWAUByDateRange(startInLoc, endInLoc)
		}()
	}

	// 并行查询MAU（如果需要）
	if includeMAU {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mauMap, mauErr = s.userActivity.GetMAUByDateRange(startInLoc, endInLoc)
		}()
	}

	wg.Wait()

	// 检查错误
	if dauErr != nil {
		return nil, dauErr
	}
	if includeWAU && wauErr != nil {
		return nil, wauErr
	}
	if includeMAU && mauErr != nil {
		return nil, mauErr
	}

	// 构建数据点（确保所有日期都有数据，即使为0）
	dataPoints := make([]TrendDataPoint, 0, len(dates))
	for _, date := range dates {
		dateStr := date.Format("2006-01-02")

		// 获取DAU（如果不存在则为0）
		dau := int64(0)
		if dauMap != nil {
			if val, exists := dauMap[dateStr]; exists {
				dau = val
			}
		}

		point := TrendDataPoint{
			Date: dateStr,
			DAU:  dau,
		}

		// 添加WAU
		if includeWAU && wauMap != nil {
			if wau, exists := wauMap[dateStr]; exists {
				point.WAU = &wau
			} else {
				// 如果没有数据，设置为0
				zeroWAU := int64(0)
				point.WAU = &zeroWAU
			}
		}

		// 添加MAU
		if includeMAU && mauMap != nil {
			if mau, exists := mauMap[dateStr]; exists {
				point.MAU = &mau
				// 计算粘性 (DAU/MAU * 100%)
				if mau > 0 {
					stickiness := float64(point.DAU) / float64(mau) * 100
					point.Stickiness = &stickiness
				}
			} else {
				// 如果没有数据，设置为0
				zeroMAU := int64(0)
				point.MAU = &zeroMAU
			}
		}

		dataPoints = append(dataPoints, point)
	}

	// 按日期排序
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Date < dataPoints[j].Date
	})

	// 计算汇总数据
	summary := s.calculateSummary(dataPoints)

	return &ActivityTrendResponse{
		Period:    period,
		StartDate: startInLoc.Format("2006-01-02"),
		EndDate:   endInLoc.Format("2006-01-02"),
		Data:      dataPoints,
		Summary:   summary,
	}, nil
}

// calculateSummary 计算汇总数据
func (s *TrendService) calculateSummary(dataPoints []TrendDataPoint) TrendSummary {
	if len(dataPoints) == 0 {
		return TrendSummary{}
	}

	var totalDAU int64
	var maxDAU, minDAU int64 = 0, 999999999
	var currentDAU, currentMAU int64
	var currentStickiness float64

	for i, point := range dataPoints {
		totalDAU += point.DAU
		if point.DAU > maxDAU {
			maxDAU = point.DAU
		}
		if point.DAU < minDAU {
			minDAU = point.DAU
		}

		// 最后一个数据点作为当前值
		if i == len(dataPoints)-1 {
			currentDAU = point.DAU
			if point.MAU != nil {
				currentMAU = *point.MAU
			}
			if point.Stickiness != nil {
				currentStickiness = *point.Stickiness
			}
		}
	}

	avgDAU := float64(totalDAU) / float64(len(dataPoints))

	// 计算增长率（相比第一个数据点）
	var growthRate *float64
	if len(dataPoints) > 1 && dataPoints[0].DAU > 0 {
		firstDAU := float64(dataPoints[0].DAU)
		lastDAU := float64(dataPoints[len(dataPoints)-1].DAU)
		rate := (lastDAU - firstDAU) / firstDAU * 100
		growthRate = &rate
	}

	return TrendSummary{
		AvgDAU:     avgDAU,
		MaxDAU:     maxDAU,
		MinDAU:     minDAU,
		CurrentDAU: currentDAU,
		CurrentMAU: currentMAU,
		Stickiness: currentStickiness,
		GrowthRate: growthRate,
	}
}

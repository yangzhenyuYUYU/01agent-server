package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"
)

// TrafficService 流量来源统计服务
type TrafficService struct{}

// NewTrafficService 创建流量来源统计服务
func NewTrafficService() *TrafficService {
	return &TrafficService{}
}

// SourceDistribution 来源分布数据
type SourceDistribution struct {
	Source string `json:"source"` // 来源渠道
	Count  int64  `json:"count"`  // 用户数
}

// GetRegistrationSourceDistribution 获取注册来源分布
// 统计 utm_source 字段分布
func (s *TrafficService) GetRegistrationSourceDistribution(startDate, endDate time.Time) ([]SourceDistribution, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	startOfPeriod := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
	endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

	var results []struct {
		Source string `gorm:"column:source"`
		Count  int64  `gorm:"column:count"`
	}

	err := repository.DB.Model(&models.User{}).
		Select("COALESCE(utm_source, 'direct') as source, COUNT(*) as count").
		Where("registration_date >= ? AND registration_date <= ?", startOfPeriod, endOfPeriod).
		Group("COALESCE(utm_source, 'direct')").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为返回格式
	distributions := make([]SourceDistribution, 0, len(results))
	for _, result := range results {
		distributions = append(distributions, SourceDistribution{
			Source: result.Source,
			Count:  result.Count,
		})
	}

	return distributions, nil
}

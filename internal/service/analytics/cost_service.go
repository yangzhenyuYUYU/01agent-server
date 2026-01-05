package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"
)

// CostService 成本监控统计服务
type CostService struct{}

// NewCostService 创建成本监控统计服务
func NewCostService() *CostService {
	return &CostService{}
}

// GetCostPerUserToken 获取单用户 Token 消耗
// 基于CreditRecord中消费类型的积分消耗来近似（实际应该是task_generate_log表的token_consumed字段）
func (s *CostService) GetCostPerUserToken(date time.Time) (float64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)
	
	// 获取当天的开始和结束时间
	startOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 23, 59, 59, 999999999, loc)

	// 获取当日消费的总积分（取绝对值）
	var totalCredits float64
	err := repository.DB.Model(&models.CreditRecord{}).
		Select("COALESCE(SUM(ABS(credits)), 0) as total").
		Where("record_type = ?", models.CreditConsumption).
		Where("created_at >= ? AND created_at <= ?", startOfDay, endOfDay).
		Scan(&totalCredits).Error

	if err != nil {
		return 0, err
	}

	// 获取当日有消费的用户数
	var userCount int64
	err = repository.DB.Model(&models.CreditRecord{}).
		Select("COUNT(DISTINCT user_id) as count").
		Where("record_type = ?", models.CreditConsumption).
		Where("created_at >= ? AND created_at <= ?", startOfDay, endOfDay).
		Scan(&userCount).Error

	if err != nil {
		return 0, err
	}

	// 计算人均消耗
	if userCount == 0 {
		return 0, nil
	}

	avgCost := totalCredits / float64(userCount)
	return avgCost, nil
}

// GetTaskErrorRate 获取任务失败率
// 注意：此指标需要task_generate_log表，当前返回0表示未实现
func (s *CostService) GetTaskErrorRate(date time.Time) (float64, error) {
	// TODO: 需要task_generate_log表来实现
	// (生成失败或报错的请求数 / 总请求数) * 100%
	return 0, nil
}


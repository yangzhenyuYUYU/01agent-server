package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"
)

// ProductValueService 产品价值统计服务
type ProductValueService struct {
	userActivity *UserActivityService
}

// NewProductValueService 创建产品价值统计服务
func NewProductValueService() *ProductValueService {
	return &ProductValueService{
		userActivity: NewUserActivityService(),
	}
}

// GetTotalGenerations 获取核心动作执行次数
// 基于CreditRecord中消费类型的记录数来近似（实际应该是task_generate_log表）
func (s *ProductValueService) GetTotalGenerations(date time.Time) (int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)
	
	// 获取当天的开始和结束时间
	startOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 23, 59, 59, 999999999, loc)

	var count int64
	// 统计当日消费类型的积分记录数（近似为生成任务数）
	err := repository.DB.Model(&models.CreditRecord{}).
		Where("record_type = ?", models.CreditConsumption).
		Where("created_at >= ? AND created_at <= ?", startOfDay, endOfDay).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetAvgTasksPerUser 获取人均生成任务数
// 当日总生成次数 / DAU
func (s *ProductValueService) GetAvgTasksPerUser(date time.Time) (float64, error) {
	// 获取总生成次数
	totalGenerations, err := s.GetTotalGenerations(date)
	if err != nil {
		return 0, err
	}

	// 获取DAU
	dau, err := s.userActivity.GetDAU(date)
	if err != nil {
		return 0, err
	}

	// 计算人均任务数
	if dau == 0 {
		return 0, nil
	}

	avgTasks := float64(totalGenerations) / float64(dau)
	return avgTasks, nil
}

// GetAdoptionRate 获取生成结果采纳率
// 注意：此指标需要user_action_log表，当前返回0表示未实现
func (s *ProductValueService) GetAdoptionRate(date time.Time) (float64, error) {
	// TODO: 需要user_action_log表来实现
	// (用户点击复制、下载、保存的次数 / 总生成次数) * 100%
	return 0, nil
}


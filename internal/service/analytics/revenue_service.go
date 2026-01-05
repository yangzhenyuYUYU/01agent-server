package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"
)

// RevenueService 核心营收统计服务
type RevenueService struct{}

// NewRevenueService 创建核心营收统计服务
func NewRevenueService() *RevenueService {
	return &RevenueService{}
}

// GetMRR 获取MRR (月经常性收入)
// 计算当前生效中所有订阅用户的月费总和
func (s *RevenueService) GetMRR(date time.Time) (float64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)

	// 获取当天的结束时间，用于判断订阅是否生效
	endOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 23, 59, 59, 999999999, loc)

	var total float64
	// 查询所有生效中的订阅（通过UserProduction和Production关联）
	// 只统计订阅服务类型，排除单次充值
	// 注意：这里假设订阅在创建时生效，实际可能需要根据订阅的起止时间判断
	err := repository.DB.Model(&models.UserProduction{}).
		Select("COALESCE(SUM(p.price), 0) as total").
		Joins("JOIN productions p ON user_productions.production_id = p.id").
		Where("user_productions.status = ? AND p.product_type = ?", "active", "订阅服务").
		Where("user_productions.created_at <= ?", endOfDay).
		Scan(&total).Error

	if err != nil {
		return 0, err
	}

	return total, nil
}

// GetNewPayingUsers 获取新增付费用户数
// 统计当日首次完成付费(订阅或充值)的用户数
func (s *RevenueService) GetNewPayingUsers(date time.Time) (int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)

	// 获取当天的开始时间
	startOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 0, 0, 0, 0, loc)

	var count int64
	// 查询当日支付成功的交易，且是该用户首次付费
	// 使用子查询找出每个用户首次付费的时间，然后统计当日首次付费的用户数
	err := repository.DB.Raw(`
		SELECT COUNT(DISTINCT first_pay.user_id) as count
		FROM (
			SELECT 
				user_id,
				MIN(paid_at) as first_paid_at
			FROM trades
			WHERE payment_status = 'success' 
				AND paid_at IS NOT NULL
			GROUP BY user_id
		) as first_pay
		WHERE DATE(first_pay.first_paid_at) = DATE(?)
	`, startOfDay).Scan(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetPaymentConversionRate 获取付费转化率
// (当日新增付费用户数 / 当日新增注册用户数) * 100%
func (s *RevenueService) GetPaymentConversionRate(date time.Time) (float64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)

	// 获取当日新增付费用户数
	newPayingUsers, err := s.GetNewPayingUsers(date)
	if err != nil {
		return 0, err
	}

	// 获取当日新增注册用户数
	startOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 0, 0, 0, 0, loc)

	var newRegisteredUsers int64
	err = repository.DB.Model(&models.User{}).
		Where("DATE(registration_date) = DATE(?)", startOfDay).
		Count(&newRegisteredUsers).Error

	if err != nil {
		return 0, err
	}

	// 计算转化率
	if newRegisteredUsers == 0 {
		return 0, nil
	}

	conversionRate := float64(newPayingUsers) / float64(newRegisteredUsers) * 100
	return conversionRate, nil
}

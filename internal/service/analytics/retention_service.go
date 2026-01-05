package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"
)

// RetentionService 留存率统计服务
type RetentionService struct{}

// NewRetentionService 创建留存率统计服务
func NewRetentionService() *RetentionService {
	return &RetentionService{}
}

// GetDay1Retention 获取次日留存率
// (昨日注册且今日再次访问并操作的用户数 / 昨日新增注册数) * 100%
func (s *RetentionService) GetDay1Retention(date time.Time) (float64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)

	// 昨日日期
	yesterday := dateInLoc.AddDate(0, 0, -1)

	// 昨日开始时间
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, loc)

	// 今日开始和结束时间
	todayStart := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 0, 0, 0, 0, loc)
	todayEnd := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 23, 59, 59, 999999999, loc)

	// 获取昨日新增注册用户数
	var yesterdayNewUsers int64
	err := repository.DB.Model(&models.User{}).
		Where("DATE(registration_date) = DATE(?)", yesterdayStart).
		Count(&yesterdayNewUsers).Error

	if err != nil {
		return 0, err
	}

	if yesterdayNewUsers == 0 {
		return 0, nil
	}

	// 获取昨日注册且今日有登录的用户数
	var retainedUsers int64
	err = repository.DB.Model(&models.User{}).
		Where("DATE(registration_date) = DATE(?)", yesterdayStart).
		Where("last_login_time >= ? AND last_login_time <= ?", todayStart, todayEnd).
		Where("last_login_time IS NOT NULL").
		Count(&retainedUsers).Error

	if err != nil {
		return 0, err
	}

	// 计算留存率
	retentionRate := float64(retainedUsers) / float64(yesterdayNewUsers) * 100
	return retentionRate, nil
}

// GetWeek1Retention 获取周留存率
// (上周注册且本周再次访问并操作的用户数 / 上周新增注册数) * 100%
func (s *RetentionService) GetWeek1Retention(date time.Time) (float64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)

	// 计算上周的开始和结束日期（上周一至上周日）
	// 获取本周一
	weekday := int(dateInLoc.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日转换为7
	}
	thisMonday := dateInLoc.AddDate(0, 0, -(weekday - 1))

	// 上周一和上周日
	lastMonday := thisMonday.AddDate(0, 0, -7)
	lastSunday := thisMonday.AddDate(0, 0, -1)

	lastWeekStart := time.Date(lastMonday.Year(), lastMonday.Month(), lastMonday.Day(), 0, 0, 0, 0, loc)
	lastWeekEnd := time.Date(lastSunday.Year(), lastSunday.Month(), lastSunday.Day(), 23, 59, 59, 999999999, loc)

	// 本周开始和结束时间
	thisWeekStart := time.Date(thisMonday.Year(), thisMonday.Month(), thisMonday.Day(), 0, 0, 0, 0, loc)
	thisWeekEnd := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 23, 59, 59, 999999999, loc)

	// 获取上周新增注册用户数
	var lastWeekNewUsers int64
	err := repository.DB.Model(&models.User{}).
		Where("registration_date >= ? AND registration_date <= ?", lastWeekStart, lastWeekEnd).
		Count(&lastWeekNewUsers).Error

	if err != nil {
		return 0, err
	}

	if lastWeekNewUsers == 0 {
		return 0, nil
	}

	// 获取上周注册且本周有登录的用户数
	var retainedUsers int64
	err = repository.DB.Model(&models.User{}).
		Where("registration_date >= ? AND registration_date <= ?", lastWeekStart, lastWeekEnd).
		Where("last_login_time >= ? AND last_login_time <= ?", thisWeekStart, thisWeekEnd).
		Where("last_login_time IS NOT NULL").
		Count(&retainedUsers).Error

	if err != nil {
		return 0, err
	}

	// 计算留存率
	retentionRate := float64(retainedUsers) / float64(lastWeekNewUsers) * 100
	return retentionRate, nil
}

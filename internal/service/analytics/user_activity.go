package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"time"
)

// UserActivityService 用户活跃度统计服务
type UserActivityService struct{}

// NewUserActivityService 创建用户活跃度统计服务
func NewUserActivityService() *UserActivityService {
	return &UserActivityService{}
}

// GetDAU 获取日活跃用户数
// 基于last_login_time统计（当前实现）
// 未来可基于task_generate_log表统计有效生成操作
func (s *UserActivityService) GetDAU(date time.Time) (int64, error) {
	// 使用本地时区（北京时间）
	loc := time.FixedZone("CST", 8*60*60)
	dateInLoc := date.In(loc)

	// 获取当天的开始和结束时间
	startOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(dateInLoc.Year(), dateInLoc.Month(), dateInLoc.Day(), 23, 59, 59, 999999999, loc)

	var count int64
	err := repository.DB.Model(&models.User{}).
		Where("last_login_time >= ? AND last_login_time <= ?", startOfDay, endOfDay).
		Count(&count).Error

	return count, err
}

// GetWAU 获取周活跃用户数
// 统计过去7天内有登录的用户数
func (s *UserActivityService) GetWAU(endDate time.Time) (int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	endDateInLoc := endDate.In(loc)
	startDate := endDateInLoc.AddDate(0, 0, -7) // 过去7天

	// 获取开始日期的开始时间
	startOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	// 获取结束日期的结束时间
	endOfPeriod := time.Date(endDateInLoc.Year(), endDateInLoc.Month(), endDateInLoc.Day(), 23, 59, 59, 999999999, loc)

	var count int64
	err := repository.DB.Model(&models.User{}).
		Where("last_login_time >= ? AND last_login_time <= ?", startOfPeriod, endOfPeriod).
		Count(&count).Error

	return count, err
}

// GetMAU 获取月活跃用户数
// 统计过去30天内有登录的用户数
func (s *UserActivityService) GetMAU(endDate time.Time) (int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	endDateInLoc := endDate.In(loc)
	startDate := endDateInLoc.AddDate(0, 0, -30) // 过去30天

	// 获取开始日期的开始时间
	startOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	// 获取结束日期的结束时间
	endOfPeriod := time.Date(endDateInLoc.Year(), endDateInLoc.Month(), endDateInLoc.Day(), 23, 59, 59, 999999999, loc)

	var count int64
	err := repository.DB.Model(&models.User{}).
		Where("last_login_time >= ? AND last_login_time <= ?", startOfPeriod, endOfPeriod).
		Count(&count).Error

	return count, err
}

// GetActiveUsersByDateRange 获取指定日期范围内的活跃用户数
func (s *UserActivityService) GetActiveUsersByDateRange(startDate, endDate time.Time) (int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	startOfPeriod := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
	endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

	var count int64
	err := repository.DB.Model(&models.User{}).
		Where("last_login_time >= ? AND last_login_time <= ?", startOfPeriod, endOfPeriod).
		Count(&count).Error

	return count, err
}

// DailyActiveUserCount 每日活跃用户数统计结果
type DailyActiveUserCount struct {
	Date  string `json:"date" gorm:"column:date"`
	Count int64  `json:"count" gorm:"column:count"`
}

// GetDAUByDateRange 批量获取日期范围内每天的DAU（使用SQL GROUP BY优化性能）
func (s *UserActivityService) GetDAUByDateRange(startDate, endDate time.Time) (map[string]int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	startOfPeriod := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)
	endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

	var results []DailyActiveUserCount
	// 使用SQL的DATE函数按日期分组，一次性查询所有日期的DAU
	// 注意：如果数据库存储的是UTC时间，需要转换；如果是本地时间，直接使用DATE函数
	err := repository.DB.Raw(`
		SELECT 
			DATE(last_login_time) as date, 
			COUNT(DISTINCT user_id) as count
		FROM user
		WHERE last_login_time IS NOT NULL 
			AND last_login_time >= ? 
			AND last_login_time <= ?
		GROUP BY DATE(last_login_time)
	`, startOfPeriod, endOfPeriod).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为map，key为日期字符串
	// 确保日期格式为 YYYY-MM-DD
	resultMap := make(map[string]int64)
	for _, result := range results {
		// DATE()函数返回的格式可能是 "2006-01-02" 或 "2006-01-02 00:00:00"
		dateStr := result.Date
		// 如果dateStr包含时间部分，只取日期部分
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		// 确保格式正确
		if len(dateStr) == 10 {
			resultMap[dateStr] = result.Count
		}
	}

	// 调试：打印查询结果
	repository.Infof("DAU查询结果: %+v, 结果数量: %d", resultMap, len(resultMap))

	return resultMap, nil
}

// GetWAUByDateRange 批量获取日期范围内每天的WAU（过去7天）
// 使用SQL窗口函数优化性能
func (s *UserActivityService) GetWAUByDateRange(startDate, endDate time.Time) (map[string]int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	// 为了计算WAU，需要从更早的日期开始（提前7天）
	wauStartDate := startInLoc.AddDate(0, 0, -7)
	startOfPeriod := time.Date(wauStartDate.Year(), wauStartDate.Month(), wauStartDate.Day(), 0, 0, 0, 0, loc)
	endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

	// 一次性查询所有登录记录（去重），然后在内存中高效计算
	var loginRecords []struct {
		LoginDate string `gorm:"column:login_date"`
		UserID    string `gorm:"column:user_id"`
	}

	err := repository.DB.Raw(`
		SELECT DISTINCT 
			DATE(last_login_time) as login_date, 
			user_id
		FROM user
		WHERE last_login_time IS NOT NULL
			AND last_login_time >= ? 
			AND last_login_time <= ?
	`, startOfPeriod, endOfPeriod).Scan(&loginRecords).Error

	if err != nil {
		return nil, err
	}

	// 在内存中高效计算每天的WAU
	// 使用map按日期组织数据
	dateUserMap := make(map[string]map[string]bool)
	for _, record := range loginRecords {
		// 确保日期格式为 YYYY-MM-DD
		dateStr := record.LoginDate
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		if dateUserMap[dateStr] == nil {
			dateUserMap[dateStr] = make(map[string]bool)
		}
		dateUserMap[dateStr][record.UserID] = true
	}

	// 计算每天的WAU（过去7天内的去重用户数）
	resultMap := make(map[string]int64)
	currentDate := startInLoc
	for !currentDate.After(endInLoc) {
		dateStr := currentDate.Format("2006-01-02")
		sevenDaysAgo := currentDate.AddDate(0, 0, -7)

		// 统计过去7天内的去重用户数
		userSet := make(map[string]bool)
		checkDate := sevenDaysAgo
		for !checkDate.After(currentDate) {
			checkDateStr := checkDate.Format("2006-01-02")
			if users, exists := dateUserMap[checkDateStr]; exists {
				for userID := range users {
					userSet[userID] = true
				}
			}
			checkDate = checkDate.AddDate(0, 0, 1)
		}

		resultMap[dateStr] = int64(len(userSet))
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return resultMap, nil
}

// GetMAUByDateRange 批量获取日期范围内每天的MAU（过去30天）
// 使用与WAU相同的优化策略
func (s *UserActivityService) GetMAUByDateRange(startDate, endDate time.Time) (map[string]int64, error) {
	loc := time.FixedZone("CST", 8*60*60)
	startInLoc := startDate.In(loc)
	endInLoc := endDate.In(loc)

	// 为了计算MAU，需要从更早的日期开始（提前30天）
	mauStartDate := startInLoc.AddDate(0, 0, -30)
	startOfPeriod := time.Date(mauStartDate.Year(), mauStartDate.Month(), mauStartDate.Day(), 0, 0, 0, 0, loc)
	endOfPeriod := time.Date(endInLoc.Year(), endInLoc.Month(), endInLoc.Day(), 23, 59, 59, 999999999, loc)

	// 一次性查询所有登录记录（去重）
	var loginRecords []struct {
		LoginDate string `gorm:"column:login_date"`
		UserID    string `gorm:"column:user_id"`
	}

	err := repository.DB.Raw(`
		SELECT DISTINCT 
			DATE(last_login_time) as login_date, 
			user_id
		FROM user
		WHERE last_login_time IS NOT NULL
			AND last_login_time >= ? 
			AND last_login_time <= ?
	`, startOfPeriod, endOfPeriod).Scan(&loginRecords).Error

	if err != nil {
		return nil, err
	}

	// 在内存中高效计算每天的MAU
	// 使用map按日期组织数据
	dateUserMap := make(map[string]map[string]bool)
	for _, record := range loginRecords {
		// 确保日期格式为 YYYY-MM-DD
		dateStr := record.LoginDate
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		if dateUserMap[dateStr] == nil {
			dateUserMap[dateStr] = make(map[string]bool)
		}
		dateUserMap[dateStr][record.UserID] = true
	}

	// 计算每天的MAU（过去30天内的去重用户数）
	resultMap := make(map[string]int64)
	currentDate := startInLoc
	for !currentDate.After(endInLoc) {
		dateStr := currentDate.Format("2006-01-02")
		thirtyDaysAgo := currentDate.AddDate(0, 0, -30)

		// 统计过去30天内的去重用户数
		userSet := make(map[string]bool)
		checkDate := thirtyDaysAgo
		for !checkDate.After(currentDate) {
			checkDateStr := checkDate.Format("2006-01-02")
			if users, exists := dateUserMap[checkDateStr]; exists {
				for userID := range users {
					userSet[userID] = true
				}
			}
			checkDate = checkDate.AddDate(0, 0, 1)
		}

		resultMap[dateStr] = int64(len(userSet))
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return resultMap, nil
}

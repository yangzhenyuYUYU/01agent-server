package analytics

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// SceneUsageService 场景使用分析服务
type SceneUsageService struct {
	db *gorm.DB
}

// NewSceneUsageService 创建场景使用分析服务实例
func NewSceneUsageService(db *gorm.DB) *SceneUsageService {
	return &SceneUsageService{db: db}
}

// SceneUsageStats 场景使用统计
type SceneUsageStats struct {
	SceneType    string  `json:"scene_type"`     // 场景类型
	UsageCount   int64   `json:"usage_count"`    // 使用次数
	UserCount    int64   `json:"user_count"`     // 使用人数
	AvgPerUser   float64 `json:"avg_per_user"`   // 人均使用次数
	Percentage   float64 `json:"percentage"`     // 占比
	CompletedCount int64 `json:"completed_count"` // 完成数量
	CompletionRate float64 `json:"completion_rate"` // 完成率
}

// UserTypeSceneStats 用户类型场景统计
type UserTypeSceneStats struct {
	UserType   string  `json:"user_type"`    // 用户类型
	SceneType  string  `json:"scene_type"`   // 场景类型
	UsageCount int64   `json:"usage_count"`  // 使用次数
	UserCount  int64   `json:"user_count"`   // 使用人数
	AvgPerUser float64 `json:"avg_per_user"` // 人均使用次数
}

// ProductSceneStats 产品套餐场景统计
type ProductSceneStats struct {
	ProductName string  `json:"product_name"` // 产品名称
	SceneType   string  `json:"scene_type"`   // 场景类型
	UsageCount  int64   `json:"usage_count"`  // 使用次数
	UserCount   int64   `json:"user_count"`   // 使用人数
	AvgPerUser  float64 `json:"avg_per_user"` // 人均使用次数
}

// ExportFormatStats 导出格式统计
type ExportFormatStats struct {
	ExportFormat string  `json:"export_format"` // 导出格式
	ExportCount  int64   `json:"export_count"`  // 导出次数
	UserCount    int64   `json:"user_count"`    // 使用人数
	Percentage   float64 `json:"percentage"`    // 占比
}

// AIFeatureStats AI能力使用统计
type AIFeatureStats struct {
	FeatureType string  `json:"feature_type"`  // 功能类型
	UsageCount  int64   `json:"usage_count"`   // 使用次数
	UserCount   int64   `json:"user_count"`    // 使用人数
	SuccessRate float64 `json:"success_rate"`  // 成功率
}

// SceneUsageReport 场景使用综合报告
type SceneUsageReport struct {
	ReportDate         string                   `json:"report_date"`          // 报告日期
	StartDate          string                   `json:"start_date"`           // 开始日期
	EndDate            string                   `json:"end_date"`             // 结束日期
	TotalUsers         int64                    `json:"total_users"`          // 总用户数
	FreeUsers          int64                    `json:"free_users"`           // 免费用户数
	PaidUsers          int64                    `json:"paid_users"`           // 付费用户数
	TotalProjects      int64                    `json:"total_projects"`       // 总项目数
	TotalExports       int64                    `json:"total_exports"`        // 总导出数
	SceneStats         []SceneUsageStats        `json:"scene_stats"`          // 场景统计
	UserTypeComparison []UserTypeSceneStats     `json:"user_type_comparison"` // 用户类型对比
	ProductComparison  []ProductSceneStats      `json:"product_comparison"`   // 产品套餐对比
	ExportStats        []ExportFormatStats      `json:"export_stats"`         // 导出统计
	AIFeatureStats     []AIFeatureStats         `json:"ai_feature_stats"`     // AI功能统计
}

// GetSceneUsageReport 获取场景使用综合报告
func (s *SceneUsageService) GetSceneUsageReport(days int) (*SceneUsageReport, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	report := &SceneUsageReport{
		ReportDate: time.Now().Format("2006-01-02 15:04:05"),
		StartDate:  startDate.Format("2006-01-02"),
		EndDate:    endDate.Format("2006-01-02"),
	}

	// 1. 获取用户总览统计
	if err := s.getUserOverview(report, startDate, endDate); err != nil {
		return nil, fmt.Errorf("获取用户总览失败: %w", err)
	}

	// 2. 获取场景使用统计
	sceneStats, err := s.getSceneUsageStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取场景统计失败: %w", err)
	}
	report.SceneStats = sceneStats

	// 3. 获取用户类型对比
	userTypeStats, err := s.getUserTypeSceneStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取用户类型对比失败: %w", err)
	}
	report.UserTypeComparison = userTypeStats

	// 4. 获取产品套餐对比
	productStats, err := s.getProductSceneStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取产品对比失败: %w", err)
	}
	report.ProductComparison = productStats

	// 5. 获取导出统计
	exportStats, err := s.getExportFormatStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取导出统计失败: %w", err)
	}
	report.ExportStats = exportStats

	// 6. 获取AI功能统计
	aiStats, err := s.getAIFeatureStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取AI功能统计失败: %w", err)
	}
	report.AIFeatureStats = aiStats

	return report, nil
}

// getUserOverview 获取用户总览
func (s *SceneUsageService) getUserOverview(report *SceneUsageReport, startDate, endDate time.Time) error {
	// 统计总用户数
	if err := s.db.Table("user").
		Where("registration_date <= ?", endDate).
		Count(&report.TotalUsers).Error; err != nil {
		return err
	}

	// 统计免费用户
	if err := s.db.Table("user").
		Where("registration_date <= ? AND vip_level = 0", endDate).
		Count(&report.FreeUsers).Error; err != nil {
		return err
	}

	// 统计付费用户
	if err := s.db.Table("user").
		Where("registration_date <= ? AND vip_level > 0", endDate).
		Count(&report.PaidUsers).Error; err != nil {
		return err
	}

	// 统计时间范围内的项目数（短文项目 + 长文章）
	var shortPostCount, articleCount int64
	if err := s.db.Table("short_post_projects").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&shortPostCount).Error; err != nil {
		return err
	}
	if err := s.db.Table("article_edit_tasks").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&articleCount).Error; err != nil {
		return err
	}
	report.TotalProjects = shortPostCount + articleCount

	// 统计时间范围内的导出数
	if err := s.db.Table("short_post_export_records").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&report.TotalExports).Error; err != nil {
		return err
	}

	return nil
}

// getSceneUsageStats 获取场景使用统计（合并短文项目和长文章）
func (s *SceneUsageService) getSceneUsageStats(startDate, endDate time.Time) ([]SceneUsageStats, error) {
	var stats []SceneUsageStats

	// 合并短文项目和长文章的统计数据
	query := `
		SELECT 
			scene_type,
			SUM(usage_count) as usage_count,
			SUM(user_count) as user_count,
			ROUND(SUM(usage_count) * 1.0 / SUM(user_count), 2) as avg_per_user,
			SUM(completed_count) as completed_count
		FROM (
			-- 短文项目统计
			SELECT 
				project_type as scene_type,
				COUNT(*) as usage_count,
				COUNT(DISTINCT user_id) as user_count,
				SUM(CASE WHEN status = 'saved' THEN 1 ELSE 0 END) as completed_count
			FROM short_post_projects
			WHERE created_at BETWEEN ? AND ?
			GROUP BY project_type
			
			UNION ALL
			
			-- 长文章统计
			SELECT 
				CONCAT('article_', IFNULL(scene_type, 'other')) as scene_type,
				COUNT(*) as usage_count,
				COUNT(DISTINCT user_id) as user_count,
				SUM(CASE WHEN status IN ('pending', 'published') THEN 1 ELSE 0 END) as completed_count
			FROM article_edit_tasks
			WHERE created_at BETWEEN ? AND ?
			GROUP BY scene_type
		) combined
		GROUP BY scene_type
		ORDER BY usage_count DESC
	`

	if err := s.db.Raw(query, startDate, endDate, startDate, endDate).Scan(&stats).Error; err != nil {
		return nil, err
	}

	// 计算总数和百分比、完成率
	var total int64
	for _, stat := range stats {
		total += stat.UsageCount
	}

	for i := range stats {
		if total > 0 {
			stats[i].Percentage = float64(stats[i].UsageCount) * 100 / float64(total)
		}
		if stats[i].UsageCount > 0 {
			stats[i].CompletionRate = float64(stats[i].CompletedCount) * 100 / float64(stats[i].UsageCount)
		}
	}

	return stats, nil
}

// getUserTypeSceneStats 获取用户类型场景统计（合并短文项目和长文章）
func (s *SceneUsageService) getUserTypeSceneStats(startDate, endDate time.Time) ([]UserTypeSceneStats, error) {
	var stats []UserTypeSceneStats

	query := `
		SELECT 
			user_type,
			scene_type,
			SUM(usage_count) as usage_count,
			SUM(user_count) as user_count,
			ROUND(SUM(usage_count) * 1.0 / SUM(user_count), 2) as avg_per_user
		FROM (
			-- 短文项目统计
			SELECT 
				CASE WHEN IFNULL(u.vip_level, 0) = 0 THEN '免费用户' ELSE '付费用户' END as user_type,
				spp.project_type as scene_type,
				COUNT(spp.id) as usage_count,
				COUNT(DISTINCT spp.user_id) as user_count
			FROM short_post_projects spp
			INNER JOIN user u ON spp.user_id = u.user_id
			WHERE spp.created_at BETWEEN ? AND ?
			GROUP BY user_type, scene_type
			
			UNION ALL
			
			-- 长文章统计
			SELECT 
				CASE WHEN IFNULL(u.vip_level, 0) = 0 THEN '免费用户' ELSE '付费用户' END as user_type,
				CONCAT('article_', IFNULL(aet.scene_type, 'other')) as scene_type,
				COUNT(aet.id) as usage_count,
				COUNT(DISTINCT aet.user_id) as user_count
			FROM article_edit_tasks aet
			INNER JOIN user u ON aet.user_id = u.user_id
			WHERE aet.created_at BETWEEN ? AND ?
			GROUP BY user_type, scene_type
		) combined
		GROUP BY user_type, scene_type
		ORDER BY user_type DESC, usage_count DESC
	`

	if err := s.db.Raw(query, startDate, endDate, startDate, endDate).Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// getProductSceneStats 获取产品套餐场景统计（合并短文项目和长文章）
func (s *SceneUsageService) getProductSceneStats(startDate, endDate time.Time) ([]ProductSceneStats, error) {
	// 归类产品到主要会员类型的SQL CASE表达式
	categorySQL := `
		CASE 
			WHEN p.name IS NULL THEN '免费用户'
			WHEN p.name = '种子终身会员' THEN '种子终身会员'
			WHEN p.name = '专业版年度会员' THEN '专业版年度'
			WHEN p.name = '轻量版年度会员' THEN '轻量版年度'
			WHEN p.name IN ('专业版', '专业版周体验', '专业版体验', '专业版开通测试') THEN '专业版月度'
			WHEN p.name = '轻量版' THEN '轻量版月度'
			ELSE '其他付费'
		END
	`

	query := fmt.Sprintf(`
		SELECT 
			product_name,
			scene_type,
			SUM(usage_count) as usage_count,
			SUM(user_count) as user_count,
			ROUND(SUM(usage_count) * 1.0 / SUM(user_count), 2) as avg_per_user
		FROM (
			-- 短文项目统计
			SELECT 
				%s as product_name,
				spp.project_type as scene_type,
				COUNT(spp.id) as usage_count,
				COUNT(DISTINCT spp.user_id) as user_count
			FROM short_post_projects spp
			LEFT JOIN user u ON spp.user_id = u.user_id
			LEFT JOIN user_productions up ON u.user_id = up.user_id AND up.status = 'active'
			LEFT JOIN productions p ON up.production_id = p.id
			WHERE spp.created_at BETWEEN ? AND ?
			GROUP BY product_name, scene_type
			
			UNION ALL
			
			-- 长文章统计
			SELECT 
				%s as product_name,
				CONCAT('article_', IFNULL(aet.scene_type, 'other')) as scene_type,
				COUNT(aet.id) as usage_count,
				COUNT(DISTINCT aet.user_id) as user_count
			FROM article_edit_tasks aet
			LEFT JOIN user u ON aet.user_id = u.user_id
			LEFT JOIN user_productions up ON u.user_id = up.user_id AND up.status = 'active'
			LEFT JOIN productions p ON up.production_id = p.id
			WHERE aet.created_at BETWEEN ? AND ?
			GROUP BY product_name, scene_type
		) combined
		GROUP BY product_name, scene_type
		ORDER BY usage_count DESC
	`, categorySQL, categorySQL)

	var stats []ProductSceneStats
	if err := s.db.Raw(query, startDate, endDate, startDate, endDate).Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// getExportFormatStats 获取导出格式统计
func (s *SceneUsageService) getExportFormatStats(startDate, endDate time.Time) ([]ExportFormatStats, error) {
	var stats []ExportFormatStats

	query := `
		SELECT 
			export_format,
			COUNT(*) as export_count,
			COUNT(DISTINCT user_id) as user_count
		FROM short_post_export_records
		WHERE created_at BETWEEN ? AND ?
		GROUP BY export_format
		ORDER BY export_count DESC
	`

	if err := s.db.Raw(query, startDate, endDate).Scan(&stats).Error; err != nil {
		return nil, err
	}

	// 计算百分比
	var total int64
	for _, stat := range stats {
		total += stat.ExportCount
	}

	for i := range stats {
		if total > 0 {
			stats[i].Percentage = float64(stats[i].ExportCount) * 100 / float64(total)
		}
	}

	return stats, nil
}

// getAIFeatureStats 获取AI功能统计
func (s *SceneUsageService) getAIFeatureStats(startDate, endDate time.Time) ([]AIFeatureStats, error) {
	var stats []AIFeatureStats

	// AI排版统计
	var formatStats AIFeatureStats
	formatQuery := `
		SELECT 
			'AI排版' as feature_type,
			COUNT(*) as usage_count,
			COUNT(DISTINCT user_id) as user_count,
			ROUND(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as success_rate
		FROM ai_format_records
		WHERE created_at BETWEEN ? AND ?
	`
	if err := s.db.Raw(formatQuery, startDate, endDate).Scan(&formatStats).Error; err == nil && formatStats.UsageCount > 0 {
		stats = append(stats, formatStats)
	}

	// AI改写统计
	var rewriteStats AIFeatureStats
	rewriteQuery := `
		SELECT 
			'AI改写' as feature_type,
			COUNT(*) as usage_count,
			COUNT(DISTINCT user_id) as user_count,
			ROUND(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as success_rate
		FROM ai_rewrite_records
		WHERE created_at BETWEEN ? AND ?
	`
	if err := s.db.Raw(rewriteQuery, startDate, endDate).Scan(&rewriteStats).Error; err == nil && rewriteStats.UsageCount > 0 {
		stats = append(stats, rewriteStats)
	}

	// AI润色统计
	var polishStats AIFeatureStats
	polishQuery := `
		SELECT 
			'AI润色' as feature_type,
			COUNT(*) as usage_count,
			COUNT(DISTINCT user_id) as user_count,
			ROUND(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as success_rate
		FROM ai_topic_polish_records
		WHERE created_at BETWEEN ? AND ?
	`
	if err := s.db.Raw(polishQuery, startDate, endDate).Scan(&polishStats).Error; err == nil && polishStats.UsageCount > 0 {
		stats = append(stats, polishStats)
	}

	return stats, nil
}

// SceneRankingItem 场景排名项
type SceneRankingItem struct {
	Rank         int     `json:"rank"`           // 排名
	SceneType    string  `json:"scene_type"`     // 场景类型
	SceneName    string  `json:"scene_name"`     // 场景名称
	UsageCount   int64   `json:"usage_count"`    // 使用次数
	UserCount    int64   `json:"user_count"`     // 使用人数
	GrowthRate   float64 `json:"growth_rate"`    // 增长率（对比上一周期）
	Percentage   float64 `json:"percentage"`     // 占比
}

// rawSceneData 原始场景数据（用于批量查询）
type rawSceneData struct {
	CreatedAt  time.Time `gorm:"column:created_at"`
	SceneType  string    `gorm:"column:scene_type"`
	UserID     string    `gorm:"column:user_id"`
	VipLevel   int       `gorm:"column:vip_level"`
}

// PeriodSceneRanking 时期场景排名
type PeriodSceneRanking struct {
	Period      string             `json:"period"`       // 时期（日期或周/月标识）
	PeriodType  string             `json:"period_type"`  // 时期类型：daily/weekly/monthly
	PaidUsers   []SceneRankingItem `json:"paid_users"`   // 付费用户排名
	FreeUsers   []SceneRankingItem `json:"free_users"`   // 免费用户排名
	AllUsers    []SceneRankingItem `json:"all_users"`    // 所有用户排名
}

// SceneRankingResponse 场景排名响应
type SceneRankingResponse struct {
	PeriodType string               `json:"period_type"` // daily/weekly/monthly
	StartDate  string               `json:"start_date"`  // 开始日期
	EndDate    string               `json:"end_date"`    // 结束日期
	Rankings   []PeriodSceneRanking `json:"rankings"`    // 各时期排名列表
}

// GetSceneRanking 获取场景排名（支持每日/每周/每月）- 优化版本
func (s *SceneUsageService) GetSceneRanking(periodType string, days int) (*SceneRankingResponse, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	response := &SceneRankingResponse{
		PeriodType: periodType,
		StartDate:  startDate.Format("2006-01-02"),
		EndDate:    endDate.Format("2006-01-02"),
		Rankings:   []PeriodSceneRanking{},
	}

	// 根据周期类型获取时间段列表
	periods := s.generatePeriods(startDate, endDate, periodType)

	// 优化：批量查询所有时期的数据
	allData, err := s.batchGetSceneRankingData(startDate, endDate, periods)
	if err != nil {
		return nil, fmt.Errorf("批量获取排名数据失败: %w", err)
	}

	// 为每个时间段生成排名（使用已查询的数据）
	for _, period := range periods {
		ranking := s.buildRankingFromData(period, allData)
		response.Rankings = append(response.Rankings, ranking)
	}

	return response, nil
}

// generatePeriods 生成时间段列表
func (s *SceneUsageService) generatePeriods(startDate, endDate time.Time, periodType string) []map[string]interface{} {
	periods := []map[string]interface{}{}
	current := startDate

	switch periodType {
	case "daily":
		// 按天生成
		for current.Before(endDate) || current.Equal(endDate) {
			nextDay := current.AddDate(0, 0, 1)
			periods = append(periods, map[string]interface{}{
				"label": current.Format("2006-01-02"),
				"start": current,
				"end":   nextDay.Add(-time.Second),
				"type":  periodType,
			})
			current = nextDay
		}
	case "weekly":
		// 按周生成（每周一开始）
		// 调整到本周一
		weekday := int(current.Weekday())
		if weekday == 0 {
			weekday = 7 // 周日转换为7
		}
		current = current.AddDate(0, 0, -(weekday - 1))

		for current.Before(endDate) {
			weekEnd := current.AddDate(0, 0, 7).Add(-time.Second)
			if weekEnd.After(endDate) {
				weekEnd = endDate
			}
			periods = append(periods, map[string]interface{}{
				"label": fmt.Sprintf("%s~%s", current.Format("01-02"), weekEnd.Format("01-02")),
				"start": current,
				"end":   weekEnd,
				"type":  periodType,
			})
			current = current.AddDate(0, 0, 7)
		}
	case "monthly":
		// 按月生成
		for current.Before(endDate) {
			year, month, _ := current.Date()
			monthStart := time.Date(year, month, 1, 0, 0, 0, 0, current.Location())
			monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
			if monthEnd.After(endDate) {
				monthEnd = endDate
			}
			periods = append(periods, map[string]interface{}{
				"label": current.Format("2006-01"),
				"start": monthStart,
				"end":   monthEnd,
				"type":  periodType,
			})
			current = monthStart.AddDate(0, 1, 0)
		}
	}

	return periods
}

// getSceneRankingForPeriod 获取指定时期的场景排名
func (s *SceneUsageService) getSceneRankingForPeriod(period map[string]interface{}, periodType string) (*PeriodSceneRanking, error) {
	startDate := period["start"].(time.Time)
	endDate := period["end"].(time.Time)
	label := period["label"].(string)

	ranking := &PeriodSceneRanking{
		Period:     label,
		PeriodType: periodType,
	}

	// 获取付费用户排名
	paidRanking, err := s.getSceneRankingByUserType(startDate, endDate, "paid")
	if err != nil {
		return nil, err
	}
	ranking.PaidUsers = paidRanking

	// 获取免费用户排名
	freeRanking, err := s.getSceneRankingByUserType(startDate, endDate, "free")
	if err != nil {
		return nil, err
	}
	ranking.FreeUsers = freeRanking

	// 获取所有用户排名
	allRanking, err := s.getSceneRankingByUserType(startDate, endDate, "all")
	if err != nil {
		return nil, err
	}
	ranking.AllUsers = allRanking

	return ranking, nil
}

// getSceneRankingByUserType 按用户类型获取场景排名
func (s *SceneUsageService) getSceneRankingByUserType(startDate, endDate time.Time, userType string) ([]SceneRankingItem, error) {
	// 构建用户类型过滤条件
	userFilter := ""
	switch userType {
	case "paid":
		userFilter = "AND IFNULL(u.vip_level, 0) > 0"
	case "free":
		userFilter = "AND IFNULL(u.vip_level, 0) = 0"
	case "all":
		userFilter = ""
	}

	// 查询场景使用统计（合并短文项目和长文章）
	query := fmt.Sprintf(`
		SELECT 
			scene_type,
			SUM(usage_count) as usage_count,
			SUM(user_count) as user_count
		FROM (
			-- 短文项目统计
			SELECT 
				spp.project_type as scene_type,
				COUNT(*) as usage_count,
				COUNT(DISTINCT spp.user_id) as user_count
			FROM short_post_projects spp
			INNER JOIN user u ON spp.user_id = u.user_id
			WHERE spp.created_at BETWEEN ? AND ?
			%s
			GROUP BY spp.project_type
			
			UNION ALL
			
			-- 长文章统计
			SELECT 
				IFNULL(aet.scene_type, 'other') as scene_type,
				COUNT(*) as usage_count,
				COUNT(DISTINCT aet.user_id) as user_count
			FROM article_edit_tasks aet
			INNER JOIN user u ON aet.user_id = u.user_id
			WHERE aet.created_at BETWEEN ? AND ?
			%s
			GROUP BY scene_type
		) combined
		GROUP BY scene_type
		ORDER BY usage_count DESC
	`, userFilter, userFilter)

	var rawStats []struct {
		SceneType  string `json:"scene_type"`
		UsageCount int64  `json:"usage_count"`
		UserCount  int64  `json:"user_count"`
	}

	if err := s.db.Raw(query, startDate, endDate, startDate, endDate).Scan(&rawStats).Error; err != nil {
		return nil, err
	}

	// 计算总数和百分比
	var total int64
	for _, stat := range rawStats {
		total += stat.UsageCount
	}

	// 场景名称映射
	sceneNameMap := map[string]string{
		"xiaohongshu":      "小红书",
		"poster":           "海报",
		"long_post":        "长图文",
		"short_post":       "短图文",
		"other":            "其他",
		"article_platform": "平台文章",
		"article_seo":      "SEO文章",
		"article_tech":     "技术文章",
		"article_product":  "产品文章",
		"article_news":     "新闻文章",
		"article_other":    "其他文章",
	}

	// 构建排名列表
	rankings := []SceneRankingItem{}
	for i, stat := range rawStats {
		sceneName := sceneNameMap[stat.SceneType]
		if sceneName == "" {
			sceneName = stat.SceneType
		}

		percentage := 0.0
		if total > 0 {
			percentage = float64(stat.UsageCount) * 100 / float64(total)
		}

		rankings = append(rankings, SceneRankingItem{
			Rank:       i + 1,
			SceneType:  stat.SceneType,
			SceneName:  sceneName,
			UsageCount: stat.UsageCount,
			UserCount:  stat.UserCount,
			Percentage: percentage,
			GrowthRate: 0, // TODO: 后续可以实现增长率计算
		})
	}

	return rankings, nil
}

// batchGetSceneRankingData 批量获取所有时期的场景数据（优化版本，减少数据库查询次数）
func (s *SceneUsageService) batchGetSceneRankingData(startDate, endDate time.Time, periods []map[string]interface{}) ([]rawSceneData, error) {
	var allData []rawSceneData

	// 一次性查询所有时期的数据（短文项目 + 长文章）
	// 注意：所有文章统一归类为 "article" 场景，不按具体的 scene_type 细分
	query := `
		SELECT 
			created_at,
			scene_type,
			user_id,
			vip_level
		FROM (
			-- 短文项目（保持原有的场景分类）
			SELECT 
				spp.created_at,
				spp.project_type as scene_type,
				spp.user_id,
				IFNULL(u.vip_level, 0) as vip_level
			FROM short_post_projects spp
			LEFT JOIN user u ON spp.user_id = u.user_id
			WHERE spp.created_at BETWEEN ? AND ?
			
			UNION ALL
			
			-- 长文章（统一归类为 'article'，不区分具体的 scene_type）
			SELECT 
				aet.created_at,
				'article' as scene_type,
				aet.user_id,
				IFNULL(u.vip_level, 0) as vip_level
			FROM article_edit_tasks aet
			LEFT JOIN user u ON aet.user_id = u.user_id
			WHERE aet.created_at BETWEEN ? AND ?
		) combined
		ORDER BY created_at
	`

	if err := s.db.Raw(query, startDate, endDate, startDate, endDate).Scan(&allData).Error; err != nil {
		return nil, err
	}

	return allData, nil
}

// buildRankingFromData 从批量数据中构建单个时期的排名
func (s *SceneUsageService) buildRankingFromData(period map[string]interface{}, allData []rawSceneData) PeriodSceneRanking {
	startDate := period["start"].(time.Time)
	endDate := period["end"].(time.Time)
	label := period["label"].(string)

	// 过滤出当前时期的数据
	periodData := []rawSceneData{}
	for _, data := range allData {
		if !data.CreatedAt.Before(startDate) && !data.CreatedAt.After(endDate) {
			periodData = append(periodData, data)
		}
	}

	// 分别统计付费、免费、所有用户的数据
	paidStats := s.aggregateSceneData(periodData, "paid")
	freeStats := s.aggregateSceneData(periodData, "free")
	allStats := s.aggregateSceneData(periodData, "all")

	return PeriodSceneRanking{
		Period:     label,
		PeriodType: period["type"].(string),
		PaidUsers:  paidStats,
		FreeUsers:  freeStats,
		AllUsers:   allStats,
	}
}

// aggregateSceneData 聚合场景数据
func (s *SceneUsageService) aggregateSceneData(data []rawSceneData, userType string) []SceneRankingItem {
	// 场景统计: scene_type -> {usage_count, unique_users}
	sceneStats := make(map[string]struct {
		usageCount int64
		users      map[string]bool
	})

	// 统计数据
	for _, item := range data {
		// 根据用户类型过滤
		isPaid := item.VipLevel > 0
		if userType == "paid" && !isPaid {
			continue
		}
		if userType == "free" && isPaid {
			continue
		}

		// 初始化场景统计
		if _, exists := sceneStats[item.SceneType]; !exists {
			sceneStats[item.SceneType] = struct {
				usageCount int64
				users      map[string]bool
			}{
				usageCount: 0,
				users:      make(map[string]bool),
			}
		}

		// 更新统计
		stat := sceneStats[item.SceneType]
		stat.usageCount++
		stat.users[item.UserID] = true
		sceneStats[item.SceneType] = stat
	}

	// 转换为排名列表
	rankings := []SceneRankingItem{}
	var totalUsage int64 = 0

	for sceneType, stat := range sceneStats {
		totalUsage += stat.usageCount
		rankings = append(rankings, SceneRankingItem{
			SceneType:  sceneType,
			SceneName:  s.getSceneName(sceneType),
			UsageCount: stat.usageCount,
			UserCount:  int64(len(stat.users)),
			Percentage: 0, // 稍后计算
		})
	}

	// 排序并计算排名和百分比
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].UsageCount > rankings[j].UsageCount
	})

	for i := range rankings {
		rankings[i].Rank = i + 1
		if totalUsage > 0 {
			rankings[i].Percentage = float64(rankings[i].UsageCount) * 100 / float64(totalUsage)
		}
	}

	return rankings
}

// getSceneName 获取场景名称
func (s *SceneUsageService) getSceneName(sceneType string) string {
	sceneNameMap := map[string]string{
		// 短文项目场景类型
		"xiaohongshu": "小红书",
		"poster":      "海报",
		"long_post":   "长图文",
		"short_post":  "短图文",
		
		// 文章场景类型（所有文章统一归为"文章"）
		"article": "文章",
		
		// 其他
		"other": "其他",
	}

	if name, exists := sceneNameMap[sceneType]; exists {
		return name
	}
	return sceneType
}

// SceneDistribution 场景分布
type SceneDistribution struct {
	SceneType  string  `json:"scene_type"`  // 场景类型: xiaohongshu/article/other
	SceneName  string  `json:"scene_name"`  // 场景名称: 小红书/文章/其他
	Count      int64   `json:"count"`       // 使用次数
	Percentage float64 `json:"percentage"`  // 占比
}

// UserUsageRankingItem 用户使用排名项
type UserUsageRankingItem struct {
	Rank              int                 `json:"rank"`               // 排名
	UserID            string              `json:"user_id"`            // 用户ID
	Username          string              `json:"username"`           // 用户名
	Nickname          *string             `json:"nickname"`           // 昵称
	Avatar            *string             `json:"avatar"`             // 头像
	Phone             *string             `json:"phone"`              // 手机号
	VipStatus         string              `json:"vip_status"`         // 会员状态：free(免费)/vip(付费)
	VipLevel          int                 `json:"vip_level"`          // 会员等级
	TotalUsage        int64               `json:"total_usage"`        // 总使用次数
	SceneDistribution []SceneDistribution `json:"scene_distribution"` // 场景分布
}

// PeriodUserRanking 时期用户排名
type PeriodUserRanking struct {
	Period     string                 `json:"period"`      // 时期
	PeriodType string                 `json:"period_type"` // 时期类型
	Rankings   []UserUsageRankingItem `json:"rankings"`    // 用户排名
}

// UserRankingResponse 用户排名响应
type UserRankingResponse struct {
	PeriodType string              `json:"period_type"` // daily/weekly/monthly
	StartDate  string              `json:"start_date"`  // 开始日期
	EndDate    string              `json:"end_date"`    // 结束日期
	Rankings   []PeriodUserRanking `json:"rankings"`    // 各时期排名列表
}

// GetUserUsageRanking 获取用户使用排名（支持每日/每周/每月）
func (s *SceneUsageService) GetUserUsageRanking(periodType string, days int, topN int) (*UserRankingResponse, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	response := &UserRankingResponse{
		PeriodType: periodType,
		StartDate:  startDate.Format("2006-01-02"),
		EndDate:    endDate.Format("2006-01-02"),
		Rankings:   []PeriodUserRanking{},
	}

	// 根据周期类型获取时间段列表
	periods := s.generatePeriods(startDate, endDate, periodType)

	// 为每个时间段生成用户排名
	for _, period := range periods {
		ranking, err := s.getUserRankingForPeriod(period, periodType, topN)
		if err != nil {
			return nil, fmt.Errorf("生成时期排名失败: %w", err)
		}
		response.Rankings = append(response.Rankings, *ranking)
	}

	return response, nil
}

// getUserRankingForPeriod 获取指定时期的用户排名
func (s *SceneUsageService) getUserRankingForPeriod(period map[string]interface{}, periodType string, topN int) (*PeriodUserRanking, error) {
	startDate := period["start"].(time.Time)
	endDate := period["end"].(time.Time)
	label := period["label"].(string)

	ranking := &PeriodUserRanking{
		Period:     label,
		PeriodType: periodType,
		Rankings:   []UserUsageRankingItem{},
	}

	// 查询用户在该时期的使用统计（合并短文项目和长文章）
	// 同时查询每个用户在不同场景的使用次数
	query := `
		SELECT 
			user_id,
			CASE 
				WHEN scene_type = 'xiaohongshu' THEN 'xiaohongshu'
				WHEN scene_type = 'article' THEN 'article'
				ELSE 'other'
			END as unified_scene_type,
			COUNT(*) as usage_count
		FROM (
			-- 短文项目统计
			SELECT 
				user_id,
				project_type as scene_type,
				created_at
			FROM short_post_projects
			WHERE created_at BETWEEN ? AND ?
			
			UNION ALL
			
			-- 长文章统计（所有文章平台统一归类为 'article'）
			SELECT 
				user_id,
				'article' as scene_type,
				created_at
			FROM article_edit_tasks
			WHERE created_at BETWEEN ? AND ?
		) combined
		GROUP BY user_id, unified_scene_type
		ORDER BY user_id
	`

	type UserSceneUsage struct {
		UserID            string `json:"user_id"`
		UnifiedSceneType  string `json:"unified_scene_type"`
		UsageCount        int64  `json:"usage_count"`
	}

	var userSceneUsages []UserSceneUsage
	if err := s.db.Raw(query, startDate, endDate, startDate, endDate).Scan(&userSceneUsages).Error; err != nil {
		return nil, err
	}

	// 按用户聚合数据
	userStatsMap := make(map[string]map[string]int64) // userID -> sceneType -> count
	userTotalMap := make(map[string]int64)            // userID -> total count

	for _, usage := range userSceneUsages {
		if userStatsMap[usage.UserID] == nil {
			userStatsMap[usage.UserID] = make(map[string]int64)
		}
		userStatsMap[usage.UserID][usage.UnifiedSceneType] = usage.UsageCount
		userTotalMap[usage.UserID] += usage.UsageCount
	}

	// 获取所有用户信息
	userIDs := make([]string, 0, len(userTotalMap))
	for userID := range userTotalMap {
		userIDs = append(userIDs, userID)
	}

	userMap := make(map[string]struct {
		Username  string
		Nickname  *string
		Avatar    *string
		Phone     *string
		VipStatus string
		VipLevel  int
	})

	if len(userIDs) > 0 {
		type UserInfo struct {
			UserID   string  `gorm:"column:user_id"`
			Username *string `gorm:"column:username"`
			Nickname *string `gorm:"column:nickname"`
			Avatar   *string `gorm:"column:avatar"`
			Phone    *string `gorm:"column:phone"`
			VipLevel int     `gorm:"column:vip_level"`
		}
		var users []UserInfo
		if err := s.db.Table("user").
			Select("user_id, username, nickname, avatar, phone, IFNULL(vip_level, 0) as vip_level").
			Where("user_id IN ?", userIDs).
			Scan(&users).Error; err != nil {
			return nil, fmt.Errorf("查询用户信息失败: %w", err)
		}

		for _, user := range users {
			username := "未知用户"
			if user.Username != nil && *user.Username != "" {
				username = *user.Username
			} else if user.Nickname != nil && *user.Nickname != "" {
				username = *user.Nickname
			}
			
			vipStatus := "free"
			if user.VipLevel > 0 {
				vipStatus = "vip"
			}
			
			userMap[user.UserID] = struct {
				Username  string
				Nickname  *string
				Avatar    *string
				Phone     *string
				VipStatus string
				VipLevel  int
			}{
				Username:  username,
				Nickname:  user.Nickname,
				Avatar:    user.Avatar,
				Phone:     user.Phone,
				VipStatus: vipStatus,
				VipLevel:  user.VipLevel,
			}
		}
	}

	// 场景名称映射
	sceneNameMap := map[string]string{
		"xiaohongshu": "小红书",
		"article":     "文章",
		"other":       "其他",
	}

	// 构建用户排名列表
	userRankings := make([]UserUsageRankingItem, 0, len(userTotalMap))
	for userID, totalUsage := range userTotalMap {
		userInfo := userMap[userID]
		
		// 构建场景分布
		sceneDistributions := []SceneDistribution{}
		for sceneType, count := range userStatsMap[userID] {
			percentage := 0.0
			if totalUsage > 0 {
				percentage = float64(count) * 100 / float64(totalUsage)
			}
			sceneDistributions = append(sceneDistributions, SceneDistribution{
				SceneType:  sceneType,
				SceneName:  sceneNameMap[sceneType],
				Count:      count,
				Percentage: percentage,
			})
		}

		// 按使用次数降序排序场景
		sort.Slice(sceneDistributions, func(i, j int) bool {
			return sceneDistributions[i].Count > sceneDistributions[j].Count
		})

		userRankings = append(userRankings, UserUsageRankingItem{
			UserID:            userID,
			Username:          userInfo.Username,
			Nickname:          userInfo.Nickname,
			Avatar:            userInfo.Avatar,
			Phone:             userInfo.Phone,
			VipStatus:         userInfo.VipStatus,
			VipLevel:          userInfo.VipLevel,
			TotalUsage:        totalUsage,
			SceneDistribution: sceneDistributions,
		})
	}

	// 按总使用次数降序排序
	sort.Slice(userRankings, func(i, j int) bool {
		return userRankings[i].TotalUsage > userRankings[j].TotalUsage
	})

	// 只取前 topN 名
	if topN > 0 && len(userRankings) > topN {
		userRankings = userRankings[:topN]
	}

	// 添加排名
	for i := range userRankings {
		userRankings[i].Rank = i + 1
	}

	ranking.Rankings = userRankings
	return ranking, nil
}

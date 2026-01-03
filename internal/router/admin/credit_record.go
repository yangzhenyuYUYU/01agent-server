package admin

import (
	"fmt"
	"time"

	"gin_web/internal/middleware"
	"gin_web/internal/models"
	"gin_web/internal/repository"

	"github.com/gin-gonic/gin"
)

// CreditRecordHandler 积分记录处理器
type CreditRecordHandler struct{}

// NewCreditRecordHandler 创建积分记录处理器
func NewCreditRecordHandler() *CreditRecordHandler {
	return &CreditRecordHandler{}
}

// parseDateRange 解析和标准化日期范围
func parseDateRangeForCredit(startDate, endDate *string, defaultDays int) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if endDate == nil || *endDate == "" {
		end = time.Now()
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, end.Location())
	} else {
		end, err = time.Parse("2006-01-02", *endDate)
		if err != nil {
			end, err = time.Parse("2006-01-02 15:04:05", *endDate)
			if err != nil {
				return start, end, fmt.Errorf("结束日期格式错误: %v", err)
			}
		}
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, end.Location())
	}

	if startDate == nil || *startDate == "" {
		start = end.AddDate(0, 0, -defaultDays)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	} else {
		start, err = time.Parse("2006-01-02", *startDate)
		if err != nil {
			start, err = time.Parse("2006-01-02 15:04:05", *startDate)
			if err != nil {
				return start, end, fmt.Errorf("开始日期格式错误: %v", err)
			}
		}
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	}

	return start, end, nil
}

// formatUserInfo 格式化用户信息为字典
func formatUserInfoForCredit(user *models.User) gin.H {
	if user == nil {
		return nil
	}

	return gin.H{
		"user_id":   user.UserID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"avatar":    user.Avatar,
		"phone":     user.Phone,
		"email":     user.Email,
		"role":      user.Role,
		"vip_level": user.VipLevel,
	}
}

// GetCreditRecordsList 获取积分记录列表（多维度查询）
func (h *CreditRecordHandler) GetCreditRecordsList(c *gin.Context) {
	// 解析查询参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	userID := c.Query("user_id")
	username := c.Query("username")
	recordTypeStr := c.Query("record_type")
	serviceCode := c.Query("service_code")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	excludeAdminStr := c.DefaultQuery("exclude_admin", "true")
	description := c.Query("description")
	orderBy := c.DefaultQuery("order_by", "created_at")
	orderDirection := c.DefaultQuery("order_direction", "desc")

	var pageNum, size int
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &size)
	if pageNum < 1 {
		pageNum = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	excludeAdmin := excludeAdminStr == "true"

	// 构建查询
	query := repository.DB.Model(&models.CreditRecord{})

	// 用户筛选
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	} else if username != "" {
		query = query.Joins("JOIN user ON credit_records.user_id = user.user_id").
			Where("user.username LIKE ?", "%"+username+"%")
	}

	// 排除管理员
	if excludeAdmin {
		// 使用子查询来排除管理员，避免重复JOIN
		if username == "" {
			query = query.Where("user_id NOT IN (SELECT user_id FROM user WHERE role = ?)", 3) // UserRoleAdmin = 3
		} else {
			query = query.Where("user.role != ?", 3) // UserRoleAdmin = 3
		}
	}

	// 记录类型筛选
	if recordTypeStr != "" {
		var recordType int
		if _, err := fmt.Sscanf(recordTypeStr, "%d", &recordType); err == nil {
			query = query.Where("record_type = ?", recordType)
		}
	}

	// 服务代码筛选
	if serviceCode != "" {
		query = query.Where("service_code = ?", serviceCode)
	}

	// 日期范围筛选
	if startDate != "" || endDate != "" {
		start, end, err := parseDateRangeForCredit(&startDate, &endDate, 30)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, err.Error()))
			return
		}
		query = query.Where("created_at >= ? AND created_at <= ?", start, end)
	}

	// 描述关键词筛选
	if description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}

	// 排序
	orderField := orderBy
	if orderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页
	offset := (pageNum - 1) * size
	var records []models.CreditRecord
	if err := query.Offset(offset).Limit(size).Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 收集所有用户ID
	userIDSet := make(map[string]bool)
	for _, record := range records {
		userIDSet[record.UserID] = true
	}

	// 批量查询用户信息
	userIDs := make([]string, 0, len(userIDSet))
	for uid := range userIDSet {
		userIDs = append(userIDs, uid)
	}

	userMap := make(map[string]*models.User)
	if len(userIDs) > 0 {
		var users []models.User
		if err := repository.DB.Where("user_id IN ?", userIDs).Find(&users).Error; err == nil {
			for i := range users {
				userMap[users[i].UserID] = &users[i]
			}
		}
	}

	// 格式化数据
	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		userInfo := formatUserInfoForCredit(userMap[record.UserID])
		items = append(items, gin.H{
			"id":               record.ID,
			"user":             userInfo,
			"record_type":      int16(record.RecordType),
			"record_type_name": getCreditRecordTypeName(record.RecordType),
			"credits":          record.Credits,
			"balance":          record.Balance,
			"description":      record.Description,
			"service_code":     record.ServiceCode,
			"created_at":       record.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	totalPages := (int(total) + size - 1) / size

	middleware.Success(c, "获取成功", gin.H{
		"items":       items,
		"total":       total,
		"page":        pageNum,
		"page_size":   size,
		"total_pages": totalPages,
	})
}

// getCreditRecordTypeName 获取积分记录类型名称
func getCreditRecordTypeName(recordType models.CreditRecordType) string {
	switch recordType {
	case models.CreditRecharge:
		return "充值"
	case models.CreditConsumption:
		return "消费"
	case models.CreditReward:
		return "奖励"
	case models.CreditExpired:
		return "过期"
	case models.CreditRefund:
		return "退款"
	default:
		return "未知"
	}
}

// GetUserSummaryStats 用户使用情况概览统计
func (h *CreditRecordHandler) GetUserSummaryStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	excludeAdminStr := c.DefaultQuery("exclude_admin", "true")
	recordTypeStr := c.DefaultQuery("record_type", "2") // 默认消费类型

	excludeAdmin := excludeAdminStr == "true"
	var recordType int
	if _, err := fmt.Sscanf(recordTypeStr, "%d", &recordType); err != nil {
		recordType = 2 // 默认消费类型
	}

	// 构建基础查询条件
	query := repository.DB.Model(&models.CreditRecord{}).Where("record_type = ?", recordType)

	// 排除管理员
	if excludeAdmin {
		query = query.Where("user_id NOT IN (SELECT user_id FROM user WHERE role = ?)", 3) // UserRoleAdmin = 3
	}

	// 日期范围
	if startDate != "" || endDate != "" {
		start, end, err := parseDateRangeForCredit(&startDate, &endDate, 30)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, err.Error()))
			return
		}
		query = query.Where("created_at >= ? AND created_at <= ?", start, end)
	}

	// 获取总用户数（去重）
	var totalUsers int64
	if err := query.Select("COUNT(DISTINCT user_id)").Scan(&totalUsers).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 获取总使用次数
	var totalUsage int64
	if err := query.Count(&totalUsage).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 获取总消耗积分
	var records []models.CreditRecord
	if err := query.Select("credits").Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	totalCredits := 0.0
	for _, record := range records {
		if record.Credits != nil {
			credits := float64(*record.Credits)
			if credits < 0 {
				credits = -credits // 取绝对值
			}
			totalCredits += credits
		}
	}

	avgUsagePerUser := 0.0
	avgCreditsPerUser := 0.0
	if totalUsers > 0 {
		avgUsagePerUser = float64(totalUsage) / float64(totalUsers)
		avgCreditsPerUser = totalCredits / float64(totalUsers)
	}

	middleware.Success(c, "统计成功", gin.H{
		"total_users":          totalUsers,
		"total_usage":          totalUsage,
		"total_credits":        totalCredits,
		"avg_usage_per_user":   fmt.Sprintf("%.2f", avgUsagePerUser),
		"avg_credits_per_user": fmt.Sprintf("%.2f", avgCreditsPerUser),
	})
}

// GetServiceStats 按服务代码统计使用情况
func (h *CreditRecordHandler) GetServiceStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	excludeAdminStr := c.DefaultQuery("exclude_admin", "true")
	recordTypeStr := c.DefaultQuery("record_type", "2") // 默认消费类型

	excludeAdmin := excludeAdminStr == "true"
	var recordType int
	if _, err := fmt.Sscanf(recordTypeStr, "%d", &recordType); err != nil {
		recordType = 2 // 默认消费类型
	}

	// 构建基础查询条件
	baseQuery := repository.DB.Model(&models.CreditRecord{})

	// 排除管理员
	if excludeAdmin {
		baseQuery = baseQuery.Where("user_id NOT IN (SELECT user_id FROM user WHERE role = ?)", 3) // UserRoleAdmin = 3
	}

	// 日期范围
	if startDate != "" || endDate != "" {
		start, end, err := parseDateRangeForCredit(&startDate, &endDate, 30)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, err.Error()))
			return
		}
		baseQuery = baseQuery.Where("credit_records.created_at >= ? AND credit_records.created_at <= ?", start, end)
	}

	// 获取消费统计（按service_code分组）
	consumeQuery := baseQuery.Where("record_type = ? AND service_code IS NOT NULL AND service_code != ''", recordType)
	var consumeRecords []struct {
		ServiceCode string
		Credits     *int
		UserID      string
	}
	if err := consumeQuery.Select("service_code, credits, user_id").Find(&consumeRecords).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 在内存中聚合
	statsMap := make(map[string]struct {
		UsageCount   int
		TotalCredits float64
		UniqueUsers  map[string]bool
	})

	for _, record := range consumeRecords {
		if record.ServiceCode == "" {
			continue
		}

		stat, exists := statsMap[record.ServiceCode]
		if !exists {
			stat = struct {
				UsageCount   int
				TotalCredits float64
				UniqueUsers  map[string]bool
			}{
				UniqueUsers: make(map[string]bool),
			}
		}

		stat.UsageCount++
		if record.Credits != nil {
			credits := float64(*record.Credits)
			if credits < 0 {
				credits = -credits // 取绝对值
			}
			stat.TotalCredits += credits
		}
		if record.UserID != "" {
			stat.UniqueUsers[record.UserID] = true
		}

		statsMap[record.ServiceCode] = stat
	}

	// 获取总消耗积分
	var allConsumeRecords []models.CreditRecord
	consumeTotalQuery := baseQuery.Where("record_type = ?", recordType)
	if err := consumeTotalQuery.Select("credits").Find(&allConsumeRecords).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	totalConsumed := 0.0
	for _, record := range allConsumeRecords {
		if record.Credits != nil {
			credits := float64(*record.Credits)
			if credits < 0 {
				credits = -credits
			}
			totalConsumed += credits
		}
	}

	// 获取总奖励积分
	var allRewardRecords []models.CreditRecord
	rewardQuery := baseQuery.Where("record_type = ?", models.CreditReward)
	if err := rewardQuery.Select("credits").Find(&allRewardRecords).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	totalReward := 0.0
	for _, record := range allRewardRecords {
		if record.Credits != nil {
			credits := float64(*record.Credits)
			if credits < 0 {
				credits = -credits
			}
			totalReward += credits
		}
	}

	// 获取总充值积分
	var allRechargeRecords []models.CreditRecord
	rechargeQuery := baseQuery.Where("record_type = ?", models.CreditRecharge)
	if err := rechargeQuery.Select("credits").Find(&allRechargeRecords).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	totalRecharge := 0.0
	for _, record := range allRechargeRecords {
		if record.Credits != nil {
			credits := float64(*record.Credits)
			if credits < 0 {
				credits = -credits
			}
			totalRecharge += credits
		}
	}

	// 获取服务名称映射
	var servicePrices []models.CreditServicePrice
	if err := repository.DB.Find(&servicePrices).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取服务名称失败: "+err.Error()))
		return
	}

	serviceNameMap := make(map[string]string)
	for _, sp := range servicePrices {
		if sp.Name != nil {
			serviceNameMap[sp.ServiceCode] = *sp.Name
		}
	}

	// 格式化结果
	result := make([]gin.H, 0, len(statsMap))
	for serviceCode, stat := range statsMap {
		uniqueUserCount := len(stat.UniqueUsers)
		avgCredits := 0.0
		if stat.UsageCount > 0 {
			avgCredits = stat.TotalCredits / float64(stat.UsageCount)
		}

		serviceName := serviceNameMap[serviceCode]
		if serviceName == "" {
			serviceName = serviceCode
		}

		result = append(result, gin.H{
			"service_code":        serviceCode,
			"service_name":        serviceName,
			"usage_count":         stat.UsageCount,
			"total_credits":       stat.TotalCredits,
			"unique_users":        uniqueUserCount,
			"avg_credits_per_use": fmt.Sprintf("%.2f", avgCredits),
		})
	}

	// 按使用次数排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i]["usage_count"].(int) < result[j]["usage_count"].(int) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// 计算总消耗积分（从统计中）
	totalConsumedFromStats := 0.0
	for _, stat := range statsMap {
		totalConsumedFromStats += stat.TotalCredits
	}

	// 使用两者中较大的值
	if totalConsumedFromStats > totalConsumed {
		totalConsumed = totalConsumedFromStats
	}

	// 计算总发放积分
	totalIssued := totalReward + totalRecharge

	// 计算使用比例
	usageRatio := 0.0
	if totalIssued > 0 {
		usageRatio = (totalConsumed / totalIssued) * 100
	}

	// 计算总使用次数
	totalUsage := 0
	for _, stat := range statsMap {
		totalUsage += stat.UsageCount
	}

	middleware.Success(c, "统计成功", gin.H{
		"stats":                  result,
		"total_services":         len(result),
		"total_usage":            totalUsage,
		"total_consumed_credits": totalConsumed,
		"total_issued_credits": gin.H{
			"total":    totalIssued,
			"reward":   totalReward,
			"recharge": totalRecharge,
		},
		"usage_ratio":       fmt.Sprintf("%.2f", usageRatio),
		"remaining_credits": totalIssued - totalConsumed,
	})
}

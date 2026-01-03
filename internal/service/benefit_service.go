package service

import (
	"fmt"
	"time"

	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"gorm.io/gorm"
)

type BenefitService struct {
	userRepo       *repository.UserRepository
	parametersRepo *repository.UserParametersRepository
}

// NewBenefitService 创建权益服务
func NewBenefitService() *BenefitService {
	return &BenefitService{
		userRepo:       repository.NewUserRepository(),
		parametersRepo: repository.NewUserParametersRepository(),
	}
}

// GetUserBenefits 获取用户订阅套餐权益信息
// 返回格式与 Python 版本一致
func (s *BenefitService) GetUserBenefits(userID string) (map[string]interface{}, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 确保积分字段不为None
	if user.Credits < 0 {
		user.Credits = 0
		s.userRepo.Update(user)
	}

	// 获取或创建用户参数
	userParam, err := s.parametersRepo.GetByUserID(userID)
	if err != nil {
		// 如果不存在，创建默认参数
		userParam = &models.UserParameters{
			UserID: userID,
		}
		s.parametersRepo.Create(userParam)
	}

	// 获取用户最新的有效订阅服务
	db := repository.GetDB()
	var userProduction models.UserProduction
	err = db.Where("user_id = ? AND status = ?", userID, "active").
		Joins("JOIN productions ON user_productions.production_id = productions.id").
		Where("productions.product_type = ?", "订阅服务").
		Order("user_productions.created_at DESC").
		Preload("Production").
		Preload("Trade").
		First(&userProduction).Error

	// 获取当日每日积分
	dailyBenefit, _ := s.getOrCreateDailyBenefit(userID)
	dailyCredits := dailyBenefit.DailyCredits
	if dailyCredits < 0 {
		dailyCredits = 0
	}

	// 初始化返回数据
	result := map[string]interface{}{
		"membership_name": "免费版",
		"vip_level":       0,
		"role":            user.Role,
		"is_active":       false,
		"expire_time":     nil,
		"credits":         user.Credits,
		"daily_credits":   dailyCredits,
		"total_credits":   user.Credits + dailyCredits,
		"production_info": nil,
		"storage_quota":   s.getStorageQuotaByVipLevel(0),
	}

	// 如果没有订阅，返回默认值
	if err == gorm.ErrRecordNotFound || userProduction.ID == 0 {
		if userParam.StorageQuota == 0 {
			userParam.StorageQuota = s.getStorageQuotaByVipLevel(0)
			s.parametersRepo.Update(userParam)
		}
		result["storage_quota"] = userParam.StorageQuota
		return result, nil
	}

	// 计算过期时间
	var expireTime *time.Time
	var isActive bool

	if userProduction.Trade != nil && userProduction.Trade.PaidAt != nil {
		paidAt := *userProduction.Trade.PaidAt
		if userProduction.Production.ValidityPeriod != nil {
			exp := paidAt.Add(time.Duration(*userProduction.Production.ValidityPeriod) * 24 * time.Hour)
			expireTime = &exp
			isActive = time.Now().Before(exp)
		} else {
			// 终身会员，永不过期
			isActive = true
		}
	} else {
		isActive = false
	}

	// 获取会员名称和等级
	membershipName := "免费版"
	vipLevel := user.VipLevel
	role := user.Role

	// 如果过期且不是管理员，降级为免费版
	if !isActive && user.Role != 0 { // 0 是管理员
		isActive = false
		vipLevel = 0
		expireTime = nil
		role = 1 // 普通用户
		membershipName = "免费版"
		user.VipLevel = vipLevel
		user.Role = role
		s.userRepo.Update(user)
		// 更新存储配额为免费版
		userParam.StorageQuota = s.getStorageQuotaByVipLevel(0)
		s.parametersRepo.Update(userParam)
	}

	if isActive && user.Role != 0 {
		if userProduction.Production != nil {
			membershipName = userProduction.Production.Name
			vipLevel = s.getVipLevelByProductName(membershipName)
			role = 2 // VIP
			user.VipLevel = vipLevel
			user.Role = role
			s.userRepo.Update(user)
			// 更新存储配额
			userParam.StorageQuota = s.getStorageQuotaByVipLevel(vipLevel)
			s.parametersRepo.Update(userParam)
		}
	}

	// 更新返回数据
	result["membership_name"] = membershipName
	result["vip_level"] = vipLevel
	result["role"] = role
	result["is_active"] = isActive

	if expireTime != nil {
		result["expire_time"] = expireTime.Format("2006-01-02 15:04:05")
	}

	result["credits"] = user.Credits
	result["daily_credits"] = dailyCredits
	result["total_credits"] = user.Credits + dailyCredits

	if userParam.StorageQuota > 0 {
		result["storage_quota"] = userParam.StorageQuota
	} else {
		result["storage_quota"] = s.getStorageQuotaByVipLevel(vipLevel)
	}

	if userProduction.Production != nil && isActive {
		result["production_info"] = map[string]interface{}{
			"id":              userProduction.Production.ID,
			"name":            userProduction.Production.Name,
			"type":            userProduction.Production.ProductType,
			"validity_period": userProduction.Production.ValidityPeriod,
			"description":     userProduction.Production.Description,
			"extra_info":      userProduction.Production.ExtraInfo,
		}
	}

	return result, nil
}

// getOrCreateDailyBenefit 获取或创建用户当日的每日权益记录
func (s *BenefitService) getOrCreateDailyBenefit(userID string) (*models.UserDailyBenefit, error) {
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())

	db := repository.GetDB()
	var dailyBenefit models.UserDailyBenefit
	err := db.Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, todayStart, todayEnd).
		First(&dailyBenefit).Error

	if err == gorm.ErrRecordNotFound {
		// 创建当天的每日权益记录
		dailyBenefit = models.UserDailyBenefit{
			UserID:       userID,
			DailyCredits: 30, // 默认每日积分
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := db.Create(&dailyBenefit).Error; err != nil {
			return nil, err
		}
	}

	return &dailyBenefit, nil
}

// getVipLevelByProductName 根据产品名称获取VIP等级
func (s *BenefitService) getVipLevelByProductName(productName string) int {
	// 根据产品名称判断VIP等级
	if productName == "种子终身会员" {
		return 4
	}
	if contains(productName, "轻量版") {
		return 2
	}
	if contains(productName, "专业版") {
		return 3
	}
	return 1
}

// getStorageQuotaByVipLevel 根据VIP等级获取存储配额（字节）
func (s *BenefitService) getStorageQuotaByVipLevel(vipLevel int) int64 {
	// 存储配额映射：VIP等级 -> 字节数
	quotaMap := map[int]int64{
		0: 300 * 1024 * 1024,       // 免费版：300MB
		1: 1 * 1024 * 1024 * 1024,  // VIP1：1GB
		2: 5 * 1024 * 1024 * 1024,  // VIP2：5GB
		3: 10 * 1024 * 1024 * 1024, // VIP3：10GB
		4: 50 * 1024 * 1024 * 1024, // VIP4：50GB
	}
	if quota, ok := quotaMap[vipLevel]; ok {
		return quota
	}
	return quotaMap[0] // 默认返回免费版配额
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

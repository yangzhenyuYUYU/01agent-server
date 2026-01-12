package service

import (
	"encoding/json"
	"fmt"
	"time"

	"01agent_server/internal/config"
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

	// 使用累加算法计算最终过期时间（考虑多次续费累加场景）
	expireTime, isActive, activeProductName := s.CalculateMembershipExpireTime(userID)

	// 获取会员名称和等级（使用累加计算后的产品名称）
	membershipName := "免费版"
	vipLevel := user.VipLevel
	role := user.Role

	if activeProductName != "" {
		membershipName = activeProductName
	} else if userProduction.Production != nil {
		membershipName = userProduction.Production.Name
	}

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
		// 使用累加计算后的产品名称（优先），或最新订阅的产品名称
		if activeProductName != "" {
			membershipName = activeProductName
		} else if userProduction.Production != nil {
			membershipName = userProduction.Production.Name
		}

		vipLevel = s.getVipLevelByProductName(membershipName)
		role = 2 // VIP
		user.VipLevel = vipLevel
		user.Role = role
		s.userRepo.Update(user)
		// 更新存储配额
		userParam.StorageQuota = s.getStorageQuotaByVipLevel(vipLevel)
		s.parametersRepo.Update(userParam)
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

// CalculateMembershipExpireTime 计算用户会员的最终过期时间（考虑多次续费累加）
//
// 计算逻辑：
// 1. 获取用户所有有效的订阅服务订单，按创建时间升序排列
// 2. 链式计算：每个订单的有效期从上一个订单的过期时间开始（如果上一个还没过期）
//
//	或者从订单创建时间开始（如果是第一个订单或上一个已过期）
//
// 3. 返回最终的过期时间和最后一个有效订单的产品名称
//
// 返回值:
// - expireTime: 最终过期时间（终身会员返回 nil）
// - isActive: 会员是否有效
// - productName: 最新产品名称（无订阅或全部过期返回空字符串）
func (s *BenefitService) CalculateMembershipExpireTime(userID string) (*time.Time, bool, string) {
	db := repository.GetDB()
	now := time.Now()

	// 获取用户所有有效的订阅服务订单，按创建时间升序排列
	var userProductions []models.UserProduction
	err := db.Where("user_id = ? AND status = ?", userID, "active").
		Joins("JOIN productions ON user_productions.production_id = productions.id").
		Where("productions.product_type = ?", "订阅服务").
		Order("user_productions.created_at ASC").
		Preload("Production").
		Preload("Trade").
		Find(&userProductions).Error

	if err != nil || len(userProductions) == 0 {
		return nil, false, ""
	}

	// 链式计算最终过期时间
	var currentExpireAt *time.Time
	hasLifetime := false
	latestProductName := ""

	for _, up := range userProductions {
		if up.Production == nil {
			continue
		}

		// 获取订单的起始时间（优先使用 trade.paid_at，其次使用 created_at）
		var orderStart time.Time
		if up.Trade != nil && up.Trade.PaidAt != nil {
			orderStart = *up.Trade.PaidAt
		} else {
			orderStart = up.CreatedAt
		}

		// 获取产品配置
		productConfig := config.GetSubscriptionProduct(up.Production.Name)

		if productConfig != nil {
			if productConfig.ValidityMonths == -1 {
				// 终身会员
				hasLifetime = true
				latestProductName = up.Production.Name
				continue
			} else if productConfig.ValidityMonths > 0 {
				// 有期限会员
				// 确定这个订单的起始时间
				var startAt time.Time
				if currentExpireAt != nil && currentExpireAt.After(orderStart) {
					// 上一个订单还没过期，从上一个订单的过期时间开始累加
					startAt = *currentExpireAt
				} else {
					// 第一个订单，或上一个订单已过期，从订单时间开始
					startAt = orderStart
				}

				// 计算这个订单的过期时间
				exp := startAt.AddDate(0, productConfig.ValidityMonths, 0)
				currentExpireAt = &exp
				latestProductName = up.Production.Name
			} else {
				// 免费版，跳过
				continue
			}
		} else if up.Production.ValidityPeriod != nil && *up.Production.ValidityPeriod > 0 {
			// 兼容旧逻辑：使用 validity_period（天数）
			var startAt time.Time
			if currentExpireAt != nil && currentExpireAt.After(orderStart) {
				startAt = *currentExpireAt
			} else {
				startAt = orderStart
			}

			exp := startAt.Add(time.Duration(*up.Production.ValidityPeriod) * 24 * time.Hour)
			currentExpireAt = &exp
			latestProductName = up.Production.Name
		}
	}

	// 终身会员优先
	if hasLifetime {
		return nil, true, latestProductName
	}

	if currentExpireAt != nil {
		isActive := currentExpireAt.After(now)
		if isActive {
			return currentExpireAt, true, latestProductName
		} else {
			return currentExpireAt, false, ""
		}
	}

	return nil, false, ""
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
		// 从配置获取每日积分
		dailyCredits := config.GetDefaultDailyCredits()

		// 创建当天的每日权益记录
		dailyBenefit = models.UserDailyBenefit{
			UserID:       userID,
			DailyCredits: dailyCredits,
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
	return config.GetVipLevelByProductName(productName)
}

// getStorageQuotaByVipLevel 根据VIP等级获取存储配额（字节）
func (s *BenefitService) getStorageQuotaByVipLevel(vipLevel int) int64 {
	return config.GetStorageQuotaByVipLevel(vipLevel)
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

// ProcessBenefitChanges 处理用户权益变更，包括积分变更和会员等级更新
// 参考 Python 版本的 BenefitManager.process_benefit_changes
func (s *BenefitService) ProcessBenefitChanges(user *models.User, production *models.Production, trade *models.Trade) (map[string]interface{}, error) {
	db := repository.GetDB()

	// 确保用户字段不为None，设置默认值
	if user.Credits < 0 {
		user.Credits = 0
	}
	if user.TotalConsumption == nil {
		zero := 0.0
		user.TotalConsumption = &zero
	}

	// 记录原始值
	oldCredits := user.Credits
	oldVipLevel := user.VipLevel
	changes := []string{}

	// 创建用户产品关联
	status := "active"
	userProduction := models.UserProduction{
		UserID:       user.UserID,
		ProductionID: production.ID,
		TradeID:      trade.ID,
		Status:       &status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&userProduction).Error; err != nil {
		return nil, fmt.Errorf("创建用户产品关联失败: %w", err)
	}

	// 如果是充值类型，更新总消费
	if trade.TradeType == "recharge" {
		if user.TotalConsumption == nil {
			zero := 0.0
			user.TotalConsumption = &zero
		}
		*user.TotalConsumption += trade.Amount
	}

	// 获取或创建用户参数
	userParam, err := s.parametersRepo.GetByUserID(user.UserID)
	if err != nil {
		userParam = &models.UserParameters{
			UserID: user.UserID,
		}
		s.parametersRepo.Create(userParam)
	}

	// 用于记录发放的每月权益积分
	monthlyCreditsIssued := 0

	// 根据产品类型处理权益
	if production.ProductType == "订阅服务" {
		// 从配置获取订阅服务产品信息
		productConfig := config.GetSubscriptionProduct(production.Name)
		if productConfig == nil {
			return nil, fmt.Errorf("未找到订阅产品配置: %s", production.Name)
		}

		user.VipLevel = productConfig.VipLevel
		userParam.StorageQuota = productConfig.StorageQuota
		changes = append(changes, productConfig.GetChanges()...)

		// 计算会员过期时间
		var membershipExpireAt *time.Time
		if trade.PaidAt != nil {
			paidAt := *trade.PaidAt
			if productConfig.ValidityMonths > 0 {
				exp := paidAt.AddDate(0, productConfig.ValidityMonths, 0)
				membershipExpireAt = &exp
			} else if productConfig.ValidityMonths == -1 {
				// 终身会员，不设过期时间
				membershipExpireAt = nil
			}
		}

		// 订阅服务的积分不直接加到用户积分，而是创建每月权益记录
		monthlyCredits := productConfig.MonthlyCredits()
		if monthlyCredits > 0 {
			today := time.Now()
			currentMonthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

			// 检查当月是否已有该订阅的权益记录
			var existingMonthly models.UserMonthlyBenefit
			err := db.Where("user_id = ? AND user_production_id = ? AND benefit_month = ?",
				user.UserID, userProduction.ID, currentMonthStart).First(&existingMonthly).Error

			if err == gorm.ErrRecordNotFound {
				// 创建首月权益记录
				monthlyBenefit := models.UserMonthlyBenefit{
					UserID:           user.UserID,
					UserProductionID: &userProduction.ID,
					MonthlyCredits:   monthlyCredits,
					BenefitMonth:     currentMonthStart,
					ExpireAt:         membershipExpireAt,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}
				if err := db.Create(&monthlyBenefit).Error; err != nil {
					return nil, fmt.Errorf("创建每月权益记录失败: %w", err)
				}
				monthlyCreditsIssued = monthlyCredits

				// 计算记录发生后的总积分
				userCredits := user.Credits
				dailyBenefit, _ := s.getOrCreateDailyBenefit(user.UserID)
				dailyCredits := dailyBenefit.DailyCredits
				if dailyCredits < 0 {
					dailyCredits = 0
				}
				allMonthlyCredits := s.getValidMonthlyCredits(user.UserID)
				totalBalance := userCredits + dailyCredits + allMonthlyCredits

				// 记录每月积分发放
				description := fmt.Sprintf("开通会员首月权益积分发放, 获得%d积分【%s, %s】",
					monthlyCredits, production.Name, currentMonthStart.Format("2006-01"))
				creditRecord := models.CreditRecord{
					UserID:      user.UserID,
					Credits:     &monthlyCredits,
					RecordType:  models.CreditReward,
					Description: &description,
					Balance:     &totalBalance,
					ServiceCode: nil,
					CreatedAt:   time.Now(),
				}
				if err := db.Create(&creditRecord).Error; err != nil {
					return nil, fmt.Errorf("创建积分记录失败: %w", err)
				}
				changes = append(changes, fmt.Sprintf("首月获得%d权益积分", monthlyCredits))
			}
		}

		// 更新用户参数
		if err := s.parametersRepo.Update(userParam); err != nil {
			return nil, fmt.Errorf("更新用户参数失败: %w", err)
		}

		// 更新用户角色
		if user.Role != 0 { // 0 是管理员
			user.Role = 2 // VIP
			changes = append(changes, "更新为VIP用户")
		}

		// 处理终身会员库存扣减
		if contains(production.Name, "终身") {
			s.decreaseProductStock(production)
		}

	} else if production.ProductType == "积分套餐" {
		// 积分套餐的积分处理，根据 production.validity_period 决定有效期
		packageConfig := config.GetCreditPackage(production.Name)
		if packageConfig == nil {
			return nil, fmt.Errorf("未找到积分套餐配置: %s", production.Name)
		}

		creditsAmount := packageConfig.Credits
		if creditsAmount > 0 {
			changes = append(changes, packageConfig.GetChanges()...)
			validityDays := production.ValidityPeriod
			if validityDays != nil && *validityDays > 0 {
				// 有期限积分
				expireAt := time.Now().Add(time.Duration(*validityDays) * 24 * time.Hour)
				timedCredit := models.UserTimedCredits{
					UserID:          user.UserID,
					Credits:         creditsAmount,
					OriginalCredits: creditsAmount,
					SourceType:      models.TimedCreditSourcePackage,
					SourceDesc:      stringPtr(fmt.Sprintf("购买%s", production.Name)),
					ExpireAt:        expireAt,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}
				if err := db.Create(&timedCredit).Error; err != nil {
					return nil, fmt.Errorf("创建有期限积分记录失败: %w", err)
				}

				// 计算记录发生后的总积分
				userCredits := user.Credits
				dailyBenefit, _ := s.getOrCreateDailyBenefit(user.UserID)
				dailyCredits := dailyBenefit.DailyCredits
				if dailyCredits < 0 {
					dailyCredits = 0
				}
				timedCredits := s.getValidTimedCredits(user.UserID)
				monthlyCredits := s.getValidMonthlyCredits(user.UserID)
				totalBalance := userCredits + dailyCredits + timedCredits + monthlyCredits

				// 记录积分获得
				description := fmt.Sprintf("购买%s, 获得%d积分（%d天有效）", production.Name, creditsAmount, *validityDays)
				creditRecord := models.CreditRecord{
					UserID:      user.UserID,
					Credits:     &creditsAmount,
					RecordType:  models.CreditRecharge,
					Description: &description,
					Balance:     &totalBalance,
					ServiceCode: nil,
					CreatedAt:   time.Now(),
				}
				if err := db.Create(&creditRecord).Error; err != nil {
					return nil, fmt.Errorf("创建积分记录失败: %w", err)
				}
				changes = append(changes, fmt.Sprintf("获得%d积分（%d天有效）", creditsAmount, *validityDays))
			} else {
				// 永久积分，直接加到用户积分
				user.Credits += creditsAmount

				// 计算记录发生后的总积分
				userCredits := user.Credits
				dailyBenefit, _ := s.getOrCreateDailyBenefit(user.UserID)
				dailyCredits := dailyBenefit.DailyCredits
				if dailyCredits < 0 {
					dailyCredits = 0
				}
				timedCredits := s.getValidTimedCredits(user.UserID)
				monthlyCredits := s.getValidMonthlyCredits(user.UserID)
				totalBalance := userCredits + dailyCredits + timedCredits + monthlyCredits

				// 记录积分获得
				description := fmt.Sprintf("购买%s, 获得%d永久积分", production.Name, creditsAmount)
				creditRecord := models.CreditRecord{
					UserID:      user.UserID,
					Credits:     &creditsAmount,
					RecordType:  models.CreditRecharge,
					Description: &description,
					Balance:     &totalBalance,
					ServiceCode: nil,
					CreatedAt:   time.Now(),
				}
				if err := db.Create(&creditRecord).Error; err != nil {
					return nil, fmt.Errorf("创建积分记录失败: %w", err)
				}
				changes = append(changes, fmt.Sprintf("获得%d永久积分", creditsAmount))
			}
		}
	}

	// 保存用户更新
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// 获取当前各类积分
	totalCredits := s.getTotalCredits(user.UserID)
	timedCredits := s.getValidTimedCredits(user.UserID)
	monthlyCredits := s.getValidMonthlyCredits(user.UserID)

	// 返回变更结果
	result := map[string]interface{}{
		"old_credits":            oldCredits,
		"new_credits":            user.Credits,
		"old_vip_level":          oldVipLevel,
		"new_vip_level":          user.VipLevel,
		"monthly_credits_issued": monthlyCreditsIssued,
		"total_timed_credits":    timedCredits,
		"total_monthly_credits":  monthlyCredits,
		"total_credits":          totalCredits,
		"changes":                changes,
		"user_production": map[string]interface{}{
			"id":              userProduction.ID,
			"product_id":      production.ID,
			"product_name":    production.Name,
			"product_type":    production.ProductType,
			"validity_period": production.ValidityPeriod,
			"extra_info":      production.ExtraInfo,
			"created_at":      userProduction.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	return result, nil
}

// 辅助方法：扣减产品库存
func (s *BenefitService) decreaseProductStock(production *models.Production) {
	if production.ExtraInfo == nil {
		return
	}

	// 解析 extra_info JSON
	var extraInfo map[string]interface{}
	if err := json.Unmarshal([]byte(*production.ExtraInfo), &extraInfo); err != nil {
		return
	}

	// 检查并扣减库存
	if stock, ok := extraInfo["stock"].(float64); ok {
		if stock > 0 {
			extraInfo["stock"] = stock - 1
			extraInfoJSON, err := json.Marshal(extraInfo)
			if err == nil {
				stockStr := string(extraInfoJSON)
				production.ExtraInfo = &stockStr
				db := repository.GetDB()
				db.Model(production).Update("extra_info", stockStr)
			}
		}
	}
}

// 辅助方法：获取有效的每月权益积分总额
func (s *BenefitService) getValidMonthlyCredits(userID string) int {
	db := repository.GetDB()
	now := time.Now()

	var monthlyBenefits []models.UserMonthlyBenefit
	db.Where("user_id = ? AND monthly_credits > 0 AND (expire_at IS NULL OR expire_at > ?)",
		userID, now).Find(&monthlyBenefits)

	total := 0
	for _, benefit := range monthlyBenefits {
		total += benefit.MonthlyCredits
	}
	return total
}

// 辅助方法：获取有效的有期限积分总额
func (s *BenefitService) getValidTimedCredits(userID string) int {
	db := repository.GetDB()
	now := time.Now()

	var timedCredits []models.UserTimedCredits
	db.Where("user_id = ? AND credits > 0 AND expire_at > ?", userID, now).Find(&timedCredits)

	total := 0
	for _, credit := range timedCredits {
		total += credit.Credits
	}
	return total
}

// 辅助方法：获取总积分
func (s *BenefitService) getTotalCredits(userID string) int {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return 0
	}

	userCredits := user.Credits
	if userCredits < 0 {
		userCredits = 0
	}

	dailyBenefit, _ := s.getOrCreateDailyBenefit(userID)
	dailyCredits := dailyBenefit.DailyCredits
	if dailyCredits < 0 {
		dailyCredits = 0
	}

	timedCredits := s.getValidTimedCredits(userID)
	monthlyCredits := s.getValidMonthlyCredits(userID)

	return userCredits + dailyCredits + timedCredits + monthlyCredits
}

// 辅助方法：字符串指针
func stringPtr(s string) *string {
	return &s
}

// BatchGetTotalCredits 批量获取用户总积分（性能优化版本）
// 返回 map[userID]totalCredits
func (s *BenefitService) BatchGetTotalCredits(userIDs []string) map[string]int {
	if len(userIDs) == 0 {
		return make(map[string]int)
	}

	db := repository.GetDB()
	now := time.Now()
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())

	result := make(map[string]int)

	// 1. 批量查询用户永久积分
	var users []models.User
	db.Where("user_id IN ?", userIDs).Find(&users)
	userCreditsMap := make(map[string]int)
	for _, user := range users {
		credits := user.Credits
		if credits < 0 {
			credits = 0
		}
		userCreditsMap[user.UserID] = credits
		result[user.UserID] = credits // 初始化为永久积分
	}

	// 2. 批量查询每日积分（当天的）
	var dailyBenefits []models.UserDailyBenefit
	db.Where("user_id IN ? AND created_at >= ? AND created_at <= ?", userIDs, todayStart, todayEnd).
		Find(&dailyBenefits)
	for _, benefit := range dailyBenefits {
		dailyCredits := benefit.DailyCredits
		if dailyCredits < 0 {
			dailyCredits = 0
		}
		if _, exists := result[benefit.UserID]; exists {
			result[benefit.UserID] += dailyCredits
		}
	}

	// 3. 批量查询有期限积分（未过期的）
	var timedCredits []models.UserTimedCredits
	db.Where("user_id IN ? AND credits > 0 AND expire_at > ?", userIDs, now).Find(&timedCredits)
	timedCreditsMap := make(map[string]int)
	for _, credit := range timedCredits {
		timedCreditsMap[credit.UserID] += credit.Credits
	}
	for userID, credits := range timedCreditsMap {
		if _, exists := result[userID]; exists {
			result[userID] += credits
		}
	}

	// 4. 批量查询每月权益积分（未过期的）
	var monthlyBenefits []models.UserMonthlyBenefit
	db.Where("user_id IN ? AND monthly_credits > 0 AND (expire_at IS NULL OR expire_at > ?)", userIDs, now).
		Find(&monthlyBenefits)
	monthlyCreditsMap := make(map[string]int)
	for _, benefit := range monthlyBenefits {
		monthlyCreditsMap[benefit.UserID] += benefit.MonthlyCredits
	}
	for userID, credits := range monthlyCreditsMap {
		if _, exists := result[userID]; exists {
			result[userID] += credits
		}
	}

	// 确保所有用户ID都有结果（即使为0）
	for _, userID := range userIDs {
		if _, exists := result[userID]; !exists {
			result[userID] = 0
		}
	}

	return result
}

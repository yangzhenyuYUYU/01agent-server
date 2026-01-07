package admin

import (
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CancelUserSubscriptionRequest 取消用户订阅请求
type CancelUserSubscriptionRequest struct {
	UserID        string `json:"user_id" binding:"required"` // 目标用户ID
	ResetVipLevel bool   `json:"reset_vip_level"`            // 是否同时重置用户VIP等级为0
}

// CancelUserSubscription 取消用户订阅权益
func (h *AdminHandler) CancelUserSubscription(c *gin.Context) {
	var req CancelUserSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 查找目标用户
	var user models.User
	if err := repository.DB.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "目标用户不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败: "+err.Error()))
		return
	}

	// 检查用户最后的会员订单是否在3天内（限制取消订阅）
	var latestTrade models.Trade
	err := repository.DB.Model(&models.Trade{}).
		Joins("JOIN user_productions ON trades.id = user_productions.trade_id").
		Joins("JOIN productions ON user_productions.production_id = productions.id").
		Where("trades.user_id = ?", req.UserID).
		Where("trades.payment_status = ?", "success").
		Where("trades.paid_at IS NOT NULL").
		Where("productions.product_type = ?", "订阅服务").
		Order("trades.paid_at DESC").
		First(&latestTrade).Error

	if err == nil && latestTrade.PaidAt != nil {
		// 检查是否超过3天
		threeDaysAgo := time.Now().AddDate(0, 0, -3)
		if latestTrade.PaidAt.Before(threeDaysAgo) {
			middleware.HandleError(c, middleware.NewBusinessError(400, "取消订阅失败：距离最后会员订单支付时间已超过3天，无法取消"))
			return
		}
	}

	// 查找用户所有的 UserProduction 记录
	var userProductions []models.UserProduction
	if err := repository.DB.Where("user_id = ?", req.UserID).
		Preload("Production").
		Preload("Trade").
		Find(&userProductions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户产品失败: "+err.Error()))
		return
	}

	// 收集要删除的信息和关联的 Trade ID
	var deletedUserProductions []gin.H
	var tradeIDs []int
	var activationCodeIDs []int

	for _, up := range userProductions {
		// 记录删除信息
		productionName := ""
		productType := ""
		if up.Production != nil {
			productionName = up.Production.Name
			productType = up.Production.ProductType
		}

		deletedUserProductions = append(deletedUserProductions, gin.H{
			"user_production_id": up.ID,
			"production_name":    productionName,
			"production_type":    productType,
			"status":             up.Status,
			"created_at":         up.CreatedAt.Format("2006-01-02 15:04:05"),
		})

		// 收集关联的 Trade ID
		if up.TradeID > 0 {
			tradeIDs = append(tradeIDs, up.TradeID)
		}
	}

	// 开始事务
	tx := repository.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除所有 UserProduction 记录
	if len(userProductions) > 0 {
		userProductionIDs := make([]int, len(userProductions))
		for i, up := range userProductions {
			userProductionIDs[i] = up.ID
		}
		if err := tx.Where("id IN ?", userProductionIDs).Delete(&models.UserProduction{}).Error; err != nil {
			tx.Rollback()
			middleware.HandleError(c, middleware.NewBusinessError(500, "删除用户产品失败: "+err.Error()))
			return
		}
	}

	// 2. 查找并处理关联的 Trade 记录和激活码
	var trades []models.Trade
	if len(tradeIDs) > 0 {
		if err := tx.Where("id IN ?", tradeIDs).Find(&trades).Error; err != nil {
			tx.Rollback()
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询交易记录失败: "+err.Error()))
			return
		}

		// 查找关联的激活码
		var activationCodes []models.ActivationCode
		if err := tx.Where("trade_id IN ?", tradeIDs).Find(&activationCodes).Error; err != nil {
			tx.Rollback()
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询激活码失败: "+err.Error()))
			return
		}

		// 重置激活码状态
		for _, ac := range activationCodes {
			activationCodeIDs = append(activationCodeIDs, ac.ID)
			ac.IsUsed = false
			ac.UsedByID = nil
			ac.TradeID = nil
			if err := tx.Save(&ac).Error; err != nil {
				tx.Rollback()
				middleware.HandleError(c, middleware.NewBusinessError(500, "重置激活码失败: "+err.Error()))
				return
			}
		}

		// 删除 Trade 记录
		if len(trades) > 0 {
			if err := tx.Where("id IN ?", tradeIDs).Delete(&models.Trade{}).Error; err != nil {
				tx.Rollback()
				middleware.HandleError(c, middleware.NewBusinessError(500, "删除交易记录失败: "+err.Error()))
				return
			}
		}
	}

	// 3. 删除用户每月权益记录（不再删除每日权益记录）
	var deletedMonthlyBenefitsCount int64
	monthlyBenefitResult := tx.Where("user_id = ?", req.UserID).Delete(&models.UserMonthlyBenefit{})
	if monthlyBenefitResult.Error != nil {
		tx.Rollback()
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除用户每月权益记录失败: "+monthlyBenefitResult.Error.Error()))
		return
	}
	deletedMonthlyBenefitsCount = monthlyBenefitResult.RowsAffected

	// 4. 删除用户积分奖励记录（CreditReward 类型，且 credits 绝对值大于 500）
	var deletedCreditRecordsCount int64
	creditRecordResult := tx.Where("user_id = ? AND record_type = ? AND (credits > 500 OR credits < -500)", req.UserID, int16(models.CreditReward)).
		Delete(&models.CreditRecord{})
	if creditRecordResult.Error != nil {
		tx.Rollback()
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除用户积分奖励记录失败: "+creditRecordResult.Error.Error()))
		return
	}
	deletedCreditRecordsCount = creditRecordResult.RowsAffected

	// 5. 如果 reset_vip_level 为 true，重置用户的 VIP 等级和角色
	oldVipLevel := user.VipLevel
	oldRole := user.Role
	if req.ResetVipLevel {
		user.VipLevel = 0
		// 恢复为普通用户（role = 1）
		if user.Role == 2 { // VIP 角色通常是 2
			user.Role = 1 // 普通用户
		}
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			middleware.HandleError(c, middleware.NewBusinessError(500, "重置用户VIP等级失败: "+err.Error()))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "提交事务失败: "+err.Error()))
		return
	}

	// 清理用户 Redis 缓存（DB3）
	tools.ClearUserCacheAsync(req.UserID)

	// 构建返回数据
	result := gin.H{
		"user_id":                        user.UserID,
		"username":                       user.Username,
		"nickname":                       user.Nickname,
		"deleted_user_productions":       deletedUserProductions,
		"deleted_trades_count":           len(trades),
		"reset_activation_codes_count":   len(activationCodeIDs),
		"deleted_monthly_benefits_count": deletedMonthlyBenefitsCount,
		"deleted_credit_records_count":   deletedCreditRecordsCount,
		"reset_vip_level":                req.ResetVipLevel,
	}

	if req.ResetVipLevel {
		result["old_vip_level"] = oldVipLevel
		result["old_role"] = oldRole
		result["new_vip_level"] = user.VipLevel
		result["new_role"] = user.Role
	}

	middleware.Success(c, "取消用户订阅权益成功", result)
}

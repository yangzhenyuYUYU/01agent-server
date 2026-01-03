package repository

import (
	"fmt"

	"01agent_server/internal/models"
)

// AutoMigrate 自动迁移数据库表
// 注意：GORM 的 AutoMigrate 只会创建新表和添加新列，不会修改或删除现有列
// 这样可以保证不会破坏现有数据库结构
func AutoMigrate() error {
	// 清理 invitation_relations 表中的重复数据（保留每个 invitee_id 的最新记录）
	// 这是为了创建唯一索引做准备
	if err := cleanupDuplicateInvitationRelations(); err != nil {
		Warnf("Failed to cleanup duplicate invitation relations: %v", err)
		// 不阻止迁移，继续执行
	}

	// 迁移所有模型
	// GORM 会自动：
	// 1. 创建不存在的表
	// 2. 添加不存在的列
	// 3. 创建不存在的索引
	// 但不会修改或删除现有列，保证数据安全
	return DB.AutoMigrate(
		// 用户相关
		&models.User{},
		&models.UserSession{},
		&models.UserParameters{},
		&models.UserPreference{},
		&models.UserAuthorization{},
		&models.UserPromiseVideo{},
		&models.UserMaterials{},
		&models.UserFeedback{},
		&models.Distributor{},
		&models.NotificationUserRecord{},
		// 邀请相关
		&models.InvitationCode{},
		&models.InvitationRelation{},
		&models.CommissionRecord{},
		// 交易相关
		&models.Trade{},
		&models.BPOrder{},
		&models.Production{},
		&models.UserProduction{},
		// 积分相关
		&models.CreditProduct{},
		&models.CreditRechargeOrder{},
		&models.CreditRecord{},
		&models.CreditServicePrice{},
		&models.UserDailyBenefit{},
		// 文章相关
		&models.ArticleEditTask{},
		&models.ArticlePublishConfig{},
		&models.ArticleTask{},
		&models.ArticleTopic{},
		&models.TaskErrorLog{},
		&models.TaskUsage{},
		&models.TotalUsageStats{},
		// AI相关
		&models.AIFormatRecord{},
		&models.AIRecommendTopic{},
		&models.AIRewriteRecord{},
		&models.AITopicPolishRecord{},
		// 系统相关
		&models.Category{},
		&models.Scene{},
		&models.SystemNotification{},
		&models.Feedback{},
		&models.ChatRecord{},
		&models.Reservation{},
		&models.MarketingActivityPlan{},
		// 其他模型（如果有的话，继续添加）
	)
}

// cleanupDuplicateInvitationRelations 清理重复的邀请关系记录
// 对于每个 invitee_id，只保留最新的记录（按 created_at 降序）
func cleanupDuplicateInvitationRelations() error {
	// 检查表是否存在
	if !DB.Migrator().HasTable(&models.InvitationRelation{}) {
		return nil // 表不存在，无需清理
	}

	// 查找所有重复的 invitee_id
	var duplicates []struct {
		InviteeID string
		Count     int64
	}

	if err := DB.Model(&models.InvitationRelation{}).
		Select("invitee_id, COUNT(*) as count").
		Group("invitee_id").
		Having("COUNT(*) > 1").
		Scan(&duplicates).Error; err != nil {
		return fmt.Errorf("failed to find duplicates: %w", err)
	}

	if len(duplicates) == 0 {
		return nil // 没有重复数据
	}

	Infof("Found %d duplicate invitee_id values, cleaning up...", len(duplicates))

	// 对每个重复的 invitee_id，删除旧记录，只保留最新的
	for _, dup := range duplicates {
		// 查找该 invitee_id 的所有记录，按 created_at 降序排列
		var relations []models.InvitationRelation
		if err := DB.Where("invitee_id = ?", dup.InviteeID).
			Order("created_at DESC").
			Find(&relations).Error; err != nil {
			Warnf("Failed to query relations for invitee_id %s: %v", dup.InviteeID, err)
			continue
		}

		if len(relations) <= 1 {
			continue // 没有重复
		}

		// 保留第一条（最新的），删除其余的
		keepID := relations[0].ID
		var idsToDelete []int
		for i := 1; i < len(relations); i++ {
			idsToDelete = append(idsToDelete, relations[i].ID)
		}

		if len(idsToDelete) > 0 {
			if err := DB.Where("id IN ?", idsToDelete).
				Delete(&models.InvitationRelation{}).Error; err != nil {
				Warnf("Failed to delete duplicate relations for invitee_id %s: %v", dup.InviteeID, err)
				continue
			}
			Infof("Deleted %d duplicate relations for invitee_id %s, kept ID %d", len(idsToDelete), dup.InviteeID, keepID)
		}
	}

	return nil
}

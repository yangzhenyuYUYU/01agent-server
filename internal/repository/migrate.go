package repository

import (
	"gin_web/internal/models"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	// 先迁移核心模型
	return DB.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.UserParameters{},
		&models.InvitationCode{},
		&models.InvitationRelation{},
		&models.CommissionRecord{},
		// 暂时注释掉其他模型，逐步添加
		// &models.Trade{},
		// &models.BPOrder{},
		// &models.ActivationCode{},
		// &models.Production{},
		// &models.CreditProduct{},
		// &models.CreditRechargeOrder{},
		// &models.CreditRecord{},
		// &models.CreditServicePrice{},
		// &models.UserProduction{},
		// &models.AIFormatRecord{},
		// &models.AIRecommendTopic{},
		// &models.AIRewriteRecord{},
		// &models.AITopicPolishRecord{},
		// &models.ArticleEditTask{},
		// &models.ArticlePublishConfig{},
		// &models.ArticleTask{},
		// &models.ArticleTopic{},
		// &models.TaskErrorLog{},
		// &models.TaskUsage{},
		// &models.TotalUsageStat{},
		// &models.Category{},
		// &models.Scene{},
		// &models.SystemNotification{},
		// &models.Feedback{},
		// &models.ChatRecord{},
		// &models.Reservation{},
		// &models.MarketingActivityPlan{},
	)
}

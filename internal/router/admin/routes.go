package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	// 可以添加其他依赖
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// SetupAdminRoutes 设置管理员路由
func SetupAdminRoutes(r *gin.Engine) {
	admin := r.Group("/api/v1/admin")
	adminHandler := NewAdminHandler()

	// 用户管理 CRUD
	userCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.User{},
		SearchFields:   []string{"username", "phone", "email", "nickname"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "user_id",
	}, repository.DB)
	userGroup := admin.Group("/user")
	userGroup.Use(middleware.AdminAuth())
	{
		userGroup.GET("/list", userCRUD.List)
		userGroup.GET("/:id", userCRUD.Detail)
		userGroup.GET("/:id/detail", adminHandler.GetUserDetail) // 用户详情（包含用户参数）
		userGroup.POST("", userCRUD.Create)
		userGroup.PUT("/:id", userCRUD.Update)
		userGroup.DELETE("/:id", userCRUD.Delete)
		userGroup.POST("/cancel_subscription", adminHandler.CancelUserSubscription) // 取消用户订阅

		// 分销商身份管理
		userGroup.POST("/:id/distributor", adminHandler.SetDistributor)      // 设置分销商身份
		userGroup.DELETE("/:id/distributor", adminHandler.RemoveDistributor) // 移除分销商身份
		userGroup.GET("/:id/distributor", adminHandler.GetDistributorInfo)   // 获取分销商信息
	}

	// 分销商管理
	distributorGroup := admin.Group("/distributor")
	distributorGroup.Use(middleware.AdminAuth())
	{
		distributorGroup.GET("/list", adminHandler.GetDistributorList) // 获取分销商列表
	}

	// 用户会话管理 CRUD
	userSessionCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.UserSession{},
		SearchFields:   []string{"user_id"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	sessionGroup := admin.Group("/user_session")
	sessionGroup.Use(middleware.AdminAuth())
	{
		sessionGroup.GET("/list", userSessionCRUD.List)
		sessionGroup.GET("/:id", userSessionCRUD.Detail)
		sessionGroup.POST("", userSessionCRUD.Create)
		sessionGroup.PUT("/:id", userSessionCRUD.Update)
		sessionGroup.DELETE("/:id", userSessionCRUD.Delete)
	}

	// 用户参数管理 CRUD
	userParameterCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.UserParameters{},
		SearchFields:   []string{"user_id"},
		DefaultOrderBy: "created_time",
		RequireAdmin:   true,
		PrimaryKey:     "param_id",
	}, repository.DB)
	parameterGroup := admin.Group("/user_parameter")
	parameterGroup.Use(middleware.AdminAuth())
	{
		parameterGroup.GET("/list", userParameterCRUD.List)
		parameterGroup.GET("/:id", userParameterCRUD.Detail)
		parameterGroup.POST("", userParameterCRUD.Create)
		parameterGroup.PUT("/:id", userParameterCRUD.Update)
		parameterGroup.DELETE("/:id", userParameterCRUD.Delete)
	}

	// 邀请码管理 CRUD
	invitationCodeCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.InvitationCode{},
		SearchFields:   []string{"code"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	invitationGroup := admin.Group("/invitation_code")
	invitationGroup.Use(middleware.AdminAuth())
	{
		invitationGroup.GET("/list", invitationCodeCRUD.List)
		invitationGroup.GET("/:id", invitationCodeCRUD.Detail)
		invitationGroup.POST("", invitationCodeCRUD.Create)
		invitationGroup.PUT("/:id", invitationCodeCRUD.Update)
		invitationGroup.DELETE("/:id", invitationCodeCRUD.Delete)
	}

	// 邀请关系管理 CRUD
	invitationRelationCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.InvitationRelation{},
		SearchFields:   []string{"code"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	invitationRelationGroup := admin.Group("/invitation_relation")
	invitationRelationGroup.Use(middleware.AdminAuth())
	{
		invitationRelationGroup.GET("/overview", adminHandler.GetInvitationRelationOverview)
		invitationRelationGroup.GET("/list", invitationRelationCRUD.List)
		invitationRelationGroup.GET("/:id", invitationRelationCRUD.Detail)
		invitationRelationGroup.POST("", invitationRelationCRUD.Create)
		invitationRelationGroup.PUT("/:id", invitationRelationCRUD.Update)
		invitationRelationGroup.DELETE("/:id", invitationRelationCRUD.Delete)
	}

	// 佣金记录管理 CRUD
	commissionRecordCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.CommissionRecord{},
		SearchFields:   []string{"user_id"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	commissionGroup := admin.Group("/commission_record")
	commissionGroup.Use(middleware.AdminAuth())
	{
		commissionGroup.GET("/list", commissionRecordCRUD.List)
		commissionGroup.GET("/:id", commissionRecordCRUD.Detail)
		commissionGroup.POST("", commissionRecordCRUD.Create)
		commissionGroup.PUT("/:id", commissionRecordCRUD.Update)
		commissionGroup.DELETE("/:id", commissionRecordCRUD.Delete)
	}

	// 场景管理 CRUD
	sceneCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.Scene{},
		SearchFields:   []string{"name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	sceneGroup := admin.Group("/scene")
	sceneGroup.Use(middleware.AdminAuth())
	{
		sceneGroup.GET("/list", sceneCRUD.List)
		sceneGroup.GET("/:id", sceneCRUD.Detail)
		sceneGroup.POST("", sceneCRUD.Create)
		sceneGroup.PUT("/:id", sceneCRUD.Update)
		sceneGroup.DELETE("/:id", sceneCRUD.Delete)
	}

	// 系统反馈管理 CRUD
	feedbackCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.Feedback{},
		SearchFields:   []string{"title", "content"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "feedback_id",
	}, repository.DB)
	feedbackGroup := admin.Group("/feedback")
	feedbackGroup.Use(middleware.AdminAuth())
	{
		feedbackGroup.GET("/list", feedbackCRUD.List)
		feedbackGroup.GET("/stats", adminHandler.GetFeedbackStats)
		feedbackGroup.POST("", feedbackCRUD.Create)
		feedbackGroup.PUT("/:id/reply", adminHandler.ReplyFeedback)
		feedbackGroup.GET("/:id", feedbackCRUD.Detail)
		feedbackGroup.PUT("/:id", feedbackCRUD.Update)
		feedbackGroup.DELETE("/:id", feedbackCRUD.Delete)
	}

	// 图片示例管理接口（需要管理员权限）
	imageExampleGroup := admin.Group("/image_example")
	imageExampleGroup.Use(middleware.AdminAuth())
	{
		imageExampleGroup.POST("/list", adminHandler.GetImageExampleList)
		imageExampleGroup.GET("/:id", adminHandler.GetImageExampleDetail)
		imageExampleGroup.POST("/save", adminHandler.SaveImageExample)
		imageExampleGroup.DELETE("/:id", adminHandler.DeleteImageExample)
	}
	// 兼容旧路径（带连字符）
	imageExampleGroupOld := admin.Group("/image-example")
	imageExampleGroupOld.Use(middleware.AdminAuth())
	{
		imageExampleGroupOld.POST("/list", adminHandler.GetImageExampleList)
		imageExampleGroupOld.GET("/:id", adminHandler.GetImageExampleDetail)
		imageExampleGroupOld.POST("/save", adminHandler.SaveImageExample)
		imageExampleGroupOld.DELETE("/:id", adminHandler.DeleteImageExample)
	}

	// 用户自定义尺寸管理接口（需要管理员权限）
	userCustomGroup := admin.Group("/user-custom")
	userCustomGroup.Use(middleware.AdminAuth())
	{
		userCustomSizeGroup := userCustomGroup.Group("/size")
		{
			userCustomSizeGroup.GET("/list", adminHandler.GetUserCustomSizeList)
		}
	}

	// 模板管理接口（需要管理员权限）
	templatesGroup := admin.Group("/templates")
	templatesGroup.Use(middleware.AdminAuth())
	{
		templatesGroup.GET("/public", adminHandler.GetPublicTemplates)
		templatesGroup.GET("/public/:template_id", adminHandler.GetPublicTemplateDetail)
		templatesGroup.PUT("/public/:template_id", adminHandler.UpdatePublicTemplate)
		templatesGroup.DELETE("/public/:template_id", adminHandler.DeletePublicTemplate)
	}

	// 样式主题管理接口（需要管理员权限）
	stylesGroup := admin.Group("/styles")
	stylesGroup.Use(middleware.AdminAuth())
	{
		themesGroup := stylesGroup.Group("/themes")
		{
			themesGroup.GET("/:theme_name/preview", adminHandler.GetThemePreview)
		}
	}

	// 用户反馈管理 CRUD
	userFeedbackCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.UserFeedback{},
		SearchFields:   []string{"user_id", "function_name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "feedback_id",
	}, repository.DB)
	userFeedbackGroup := admin.Group("/user_feedback")
	userFeedbackGroup.Use(middleware.AdminAuth())
	{
		userFeedbackGroup.GET("/list", userFeedbackCRUD.List)
		userFeedbackGroup.GET("/:id", userFeedbackCRUD.Detail)
		userFeedbackGroup.POST("", userFeedbackCRUD.Create)
		userFeedbackGroup.PUT("/:id", userFeedbackCRUD.Update)
		userFeedbackGroup.DELETE("/:id", userFeedbackCRUD.Delete)
		userFeedbackGroup.GET("/stats", adminHandler.GetUserFeedbackStats)
	}

	// 预约管理 CRUD
	reservationCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.Reservation{},
		SearchFields:   []string{"name", "phone", "email"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	reservationGroup := admin.Group("/reservation")
	reservationGroup.Use(middleware.AdminAuth())
	{
		reservationGroup.GET("/list", reservationCRUD.List)
		reservationGroup.GET("/:id", reservationCRUD.Detail)
		reservationGroup.POST("", reservationCRUD.Create)
		reservationGroup.PUT("/:id", reservationCRUD.Update)
		reservationGroup.DELETE("/:id", reservationCRUD.Delete)
	}

	// 交易管理接口（需要管理员权限）
	tradeGroup := admin.Group("/trade")
	tradeGroup.Use(middleware.AdminAuth())
	{
		tradeGroup.GET("/:id", adminHandler.GetTradeDetail)
		tradeGroup.POST("", adminHandler.CreateTrade)
		tradeGroup.PUT("/:id", adminHandler.UpdateTrade)
		tradeGroup.DELETE("/:id", adminHandler.DeleteTrade)
		tradeGroup.POST("/repair-incomplete", adminHandler.RepairIncompleteTrades)
	}
	// Trade V2 列表查询接口（需要管理员权限）
	tradeV2Group := admin.Group("/trade_v2")
	tradeV2Group.Use(middleware.AdminAuth())
	{
		tradeV2Group.POST("/list", adminHandler.GetTradeV2List)
	}

	// 用户产品管理 CRUD
	userProductionCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.UserProduction{},
		SearchFields:   []string{"user_id", "production_id", "trade_id"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	userProductionGroup := admin.Group("/user_production")
	userProductionGroup.Use(middleware.AdminAuth())
	{
		// 使用自定义List方法，返回完整的user信息而不是只有user_id
		userProductionGroup.GET("/list", adminHandler.GetUserProductionListForCRUD)
		userProductionGroup.GET("/:id", userProductionCRUD.Detail)
		userProductionGroup.POST("", userProductionCRUD.Create)
		userProductionGroup.PUT("/:id", userProductionCRUD.Update)
		userProductionGroup.DELETE("/:id", userProductionCRUD.Delete)
		// 自定义接口：获取用户产品列表（带关联信息）
		userProductionGroup.GET("/user/:user_id", adminHandler.GetUserProductionList)
		userProductionGroup.GET("/detail/:id", adminHandler.GetUserProductionDetail)
	}

	// 版本管理接口（需要管理员权限）
	versionCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.Version{},
		SearchFields:   []string{"version"},
		DefaultOrderBy: "date",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	versionGroup := admin.Group("/versions")
	versionGroup.Use(middleware.AdminAuth())
	{
		versionGroup.GET("/list", adminHandler.GetVersionList)
		versionGroup.GET("/items", adminHandler.GetVersionItems)
		versionGroup.GET("/:id", versionCRUD.Detail)
		versionGroup.POST("", versionCRUD.Create)
		versionGroup.POST("/create", adminHandler.CreateVersion)
		versionGroup.PUT("/:id", adminHandler.UpdateVersion)
		versionGroup.DELETE("/:id", versionCRUD.Delete)
	}

	// 营销活动管理接口（需要管理员权限）
	marketingGroup := admin.Group("/marketing")
	marketingGroup.Use(middleware.AdminAuth())
	{
		marketingGroup.GET("/activities", adminHandler.GetMarketingActivities)
		marketingGroup.GET("/activities/:activity_id", adminHandler.GetMarketingActivityDetail)
		marketingGroup.POST("/activities", adminHandler.CreateMarketingActivity)
		marketingGroup.PUT("/activities/:activity_id", adminHandler.UpdateMarketingActivity)
		marketingGroup.DELETE("/activities/:activity_id", adminHandler.DeleteMarketingActivity)
		marketingGroup.POST("/activities/:activity_id/purchase", adminHandler.PurchaseActivityProduct)
		marketingGroup.GET("/current-activity", adminHandler.GetCurrentActivity)
	}

	// 激活码管理接口（需要管理员权限）
	activationCodeCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.ActivationCode{},
		SearchFields:   []string{"code"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	activationCodeGroup := admin.Group("/activation_code")
	activationCodeGroup.Use(middleware.AdminAuth())
	{
		activationCodeGroup.DELETE("/:id", activationCodeCRUD.Delete)
	}
	activationCodeV2Group := admin.Group("/activation_code_v2")
	activationCodeV2Group.Use(middleware.AdminAuth())
	{
		activationCodeV2Group.GET("/list", adminHandler.GetActivationCodeList)
		activationCodeV2Group.POST("/create", adminHandler.CreateActivationCodes)
	}

	// 积分产品管理接口（需要管理员权限）
	creditProductCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.CreditProduct{},
		SearchFields:   []string{"name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	creditProductGroup := admin.Group("/credit_product")
	creditProductGroup.Use(middleware.AdminAuth())
	{
		creditProductGroup.GET("/list", creditProductCRUD.List)
		creditProductGroup.GET("/:id", creditProductCRUD.Detail)
		creditProductGroup.POST("", creditProductCRUD.Create)
		creditProductGroup.PUT("/:id", creditProductCRUD.Update)
		creditProductGroup.DELETE("/:id", creditProductCRUD.Delete)
	}

	// 积分产品管理接口 V2（需要管理员权限）
	creditProduct2Group := admin.Group("/credit_product2")
	creditProduct2Group.Use(middleware.AdminAuth())
	{
		creditProduct2Group.GET("/list", adminHandler.GetCreditProduct2List)
	}

	// 积分充值订单管理接口（需要管理员权限）
	creditRechargeOrderGroup := admin.Group("/credit_recharge_order")
	creditRechargeOrderGroup.Use(middleware.AdminAuth())
	{
		creditRechargeOrderGroup.GET("/list", adminHandler.GetCreditRechargeOrderList)
		creditRechargeOrderGroup.GET("/summary", adminHandler.GetCreditRechargeOrderSummary)
	}

	// 积分记录统计接口（需要管理员权限）
	creditRecordsGroup := admin.Group("/credit_records")
	creditRecordsGroup.Use(middleware.AdminAuth())
	{
		creditRecordsGroup.GET("/stats/overview", adminHandler.GetCreditRecordsStatsOverview)
	}

	// 积分服务价格管理接口（需要管理员权限）
	creditServicePriceCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.CreditServicePrice{},
		SearchFields:   []string{"service_code"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	creditServicePriceGroup := admin.Group("/credit_service_price")
	creditServicePriceGroup.Use(middleware.AdminAuth())
	{
		creditServicePriceGroup.GET("/list", adminHandler.GetCreditServicePriceList)
		creditServicePriceGroup.GET("/:id", creditServicePriceCRUD.Detail)
		creditServicePriceGroup.POST("", creditServicePriceCRUD.Create)
		creditServicePriceGroup.PUT("/:id", creditServicePriceCRUD.Update)
		creditServicePriceGroup.DELETE("/:id", creditServicePriceCRUD.Delete)
	}

	// 佣金管理接口（需要管理员权限）
	commissionAdminGroup := admin.Group("/commission")
	commissionAdminGroup.Use(middleware.AdminAuth())
	{
		commissionAdminGroup.GET("/overview", adminHandler.GetCommissionOverview)
		commissionAdminGroup.GET("/list", adminHandler.GetCommissionList)
	}

	// 管理员认证接口（不需要JWT认证）
	authGroup := admin.Group("/auth")
	{
		authGroup.POST("/register", adminHandler.AdminRegister)
		authGroup.POST("/login", adminHandler.AdminLogin)
	}

	// 系统通知管理 CRUD
	notificationCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.SystemNotification{},
		SearchFields:   []string{"title", "content"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "notification_id",
	}, repository.DB)
	notificationGroup := admin.Group("/notification")
	notificationGroup.Use(middleware.AdminAuth())
	{
		notificationGroup.GET("/list", adminHandler.GetNotificationList)
		notificationGroup.GET("/:id", notificationCRUD.Detail)
		notificationGroup.POST("", notificationCRUD.Create)
		notificationGroup.PUT("/:id", notificationCRUD.Update)
		notificationGroup.DELETE("/:id", notificationCRUD.Delete)
	}

	// 用户通知记录管理 CRUD
	notificationUserRecordCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.NotificationUserRecord{},
		SearchFields:   []string{"user_id", "notification_id"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "record_id",
	}, repository.DB)
	notificationUserRecordGroup := admin.Group("/notification_user_record")
	notificationUserRecordGroup.Use(middleware.AdminAuth())
	{
		notificationUserRecordGroup.GET("/list", adminHandler.GetNotificationUserRecordList)
		notificationUserRecordGroup.GET("/:id", notificationUserRecordCRUD.Detail)
		notificationUserRecordGroup.POST("", notificationUserRecordCRUD.Create)
		notificationUserRecordGroup.PUT("/:id", notificationUserRecordCRUD.Update)
		notificationUserRecordGroup.DELETE("/:id", notificationUserRecordCRUD.Delete)
	}

	// 数据分析接口（需要管理员权限）
	analyticsHandler := NewAnalyticsHandler()
	analyticsGroup := admin.Group("/analytics")
	analyticsGroup.Use(middleware.AdminAuth())
	{
		// 统一的数据分析指标接口
		analyticsGroup.GET("/metrics", analyticsHandler.GetMetrics)          // 获取指标数据
		analyticsGroup.GET("/metrics/info", analyticsHandler.GetMetricsInfo) // 获取指标信息列表

		// 用户数据统计
		analyticsGroup.GET("/user/overview", analyticsHandler.GetUserOverview)
		analyticsGroup.GET("/user/growth", analyticsHandler.GetUserGrowth)
		analyticsGroup.GET("/user/activity-trend", analyticsHandler.GetActivityTrend) // 用户活跃度趋势

		// 支付数据分析
		analyticsGroup.GET("/payment/overview", analyticsHandler.GetPaymentOverview)
		analyticsGroup.GET("/payment/trend", analyticsHandler.GetPaymentTrend)

		// 商业分析
		analyticsGroup.GET("/business/cost-analysis", analyticsHandler.GetCostAnalysis)
		analyticsGroup.GET("/business/sales-ranking", analyticsHandler.GetSalesRanking)
		analyticsGroup.GET("/business/invitation-ranking", analyticsHandler.GetInvitationRanking)

		// 流量来源分析
		analyticsGroup.GET("/traffic/source-distribution", analyticsHandler.GetRegistrationSourceDistribution)
	}

	// 会员统计接口（需要管理员权限）
	membershipHandler := NewMembershipHandler()
	membershipGroup := admin.Group("/membership")
	membershipGroup.Use(middleware.AdminAuth())
	{
		membershipGroup.GET("/overview", membershipHandler.GetMembershipOverview)     // 获取会员购买概览
		membershipGroup.GET("/trend", membershipHandler.GetMembershipTrend)           // 获取会员购买趋势
		membershipGroup.GET("/product-trend", membershipHandler.GetProductSalesTrend) // 获取产品销售趋势（折线图）
	}

	// 用户列表V2接口（需要管理员权限）
	adminAuth := admin.Group("")
	adminAuth.Use(middleware.AdminAuth())
	{
		adminAuth.POST("/user_v2/list", adminHandler.UserV2List)
		adminAuth.GET("/user_v2/:id/token", adminHandler.GetUserToken)
	}

	// 付费用户公众号信息接口（需要管理员权限）
	admin.GET("/paid_users_wechat_accounts", middleware.AdminAuth(), adminHandler.GetPaidUsersWechatAccounts)

	// 文章主题管理 CRUD
	articleTopicCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.ArticleTopic{},
		SearchFields:   []string{"title"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	articleTopicGroup := admin.Group("/article_topic")
	articleTopicGroup.Use(middleware.AdminAuth())
	{
		articleTopicGroup.GET("/list", articleTopicCRUD.List)
		articleTopicGroup.GET("/:id", articleTopicCRUD.Detail)
		articleTopicGroup.POST("", articleTopicCRUD.Create)
		articleTopicGroup.PUT("/:id", articleTopicCRUD.Update)
		articleTopicGroup.DELETE("/:id", articleTopicCRUD.Delete)
	}

	// AI推荐主题管理 CRUD
	aiRecommendTopicCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.AIRecommendTopic{},
		SearchFields:   []string{"title"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	aiRecommendTopicGroup := admin.Group("/ai_recommend_topic")
	aiRecommendTopicGroup.Use(middleware.AdminAuth())
	{
		aiRecommendTopicGroup.GET("/list", aiRecommendTopicCRUD.List)
		aiRecommendTopicGroup.GET("/:id", aiRecommendTopicCRUD.Detail)
		aiRecommendTopicGroup.POST("", aiRecommendTopicCRUD.Create)
		aiRecommendTopicGroup.PUT("/:id", aiRecommendTopicCRUD.Update)
		aiRecommendTopicGroup.DELETE("/:id", aiRecommendTopicCRUD.Delete)
	}

	// 文章任务管理 CRUD
	articleTaskCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.ArticleTask{},
		SearchFields:   []string{"topic", "title"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	articleTaskGroup := admin.Group("/article_task")
	articleTaskGroup.Use(middleware.AdminAuth())
	{
		articleTaskGroup.GET("/list", articleTaskCRUD.List)
		articleTaskGroup.GET("/:id", articleTaskCRUD.Detail)
		articleTaskGroup.POST("", articleTaskCRUD.Create)
		articleTaskGroup.PUT("/:id", articleTaskCRUD.Update)
		articleTaskGroup.DELETE("/:id", articleTaskCRUD.Delete)
	}

	// 产品管理接口（需要管理员权限）
	productGroup := admin.Group("/product")
	productGroup.Use(middleware.AdminAuth())
	{
		productGroup.GET("/list", adminHandler.GetProductList)
	}

	// 短图文工程管理接口（需要管理员权限）
	shortPostGroup := admin.Group("/short-post/project")
	shortPostGroup.Use(middleware.AdminAuth())
	{
		shortPostGroup.GET("/all/content", adminHandler.GetAllContentList)
	}

	// 积分记录管理接口（需要管理员权限）
	creditRecordHandler := NewCreditRecordHandler()
	creditRecordGroup := admin.Group("/credit_records")
	creditRecordGroup.Use(middleware.AdminAuth())
	{
		creditRecordGroup.GET("/list", creditRecordHandler.GetCreditRecordsList)
		creditRecordGroup.GET("/stats/user/summary", creditRecordHandler.GetUserSummaryStats)
		creditRecordGroup.GET("/stats/service", creditRecordHandler.GetServiceStats)
	}

	// 缓存管理接口（需要管理员权限）
	cacheHandler := NewCacheHandler()
	cacheGroup := admin.Group("/cache")
	cacheGroup.Use(middleware.AdminAuth())
	{
		cacheGroup.GET("/list", cacheHandler.ListCache)
		cacheGroup.GET("/detail", cacheHandler.GetCacheDetail)
		cacheGroup.POST("/create", cacheHandler.CreateCache)
		cacheGroup.POST("/clear", cacheHandler.ClearCache)
		cacheGroup.PUT("/update", cacheHandler.UpdateCache)
		cacheGroup.DELETE("/delete", cacheHandler.DeleteCache)
	}

	// 健康检查（不需要管理员权限，用于测试）
	admin.GET("/health", func(c *gin.Context) {
		middleware.Success(c, "Admin API is running", gin.H{
			"status": "ok",
		})
	})
}

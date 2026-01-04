package admin

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/models/short_post"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		feedbackGroup.GET("/:id", feedbackCRUD.Detail)
		feedbackGroup.POST("", feedbackCRUD.Create)
		feedbackGroup.PUT("/:id", feedbackCRUD.Update)
		feedbackGroup.DELETE("/:id", feedbackCRUD.Delete)
		feedbackGroup.GET("/stats", adminHandler.GetFeedbackStats)
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
		versionGroup.PUT("/:id", versionCRUD.Update)
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
		// 用户数据统计
		analyticsGroup.GET("/user/overview", analyticsHandler.GetUserOverview)
		analyticsGroup.GET("/user/growth", analyticsHandler.GetUserGrowth)

		// 支付数据分析
		analyticsGroup.GET("/payment/overview", analyticsHandler.GetPaymentOverview)
		analyticsGroup.GET("/payment/trend", analyticsHandler.GetPaymentTrend)

		// 商业分析
		analyticsGroup.GET("/business/cost-analysis", analyticsHandler.GetCostAnalysis)
		analyticsGroup.GET("/business/sales-ranking", analyticsHandler.GetSalesRanking)
		analyticsGroup.GET("/business/invitation-ranking", analyticsHandler.GetInvitationRanking)
	}

	// 用户列表V2接口（需要管理员权限）
	adminAuth := admin.Group("")
	adminAuth.Use(middleware.AdminAuth())
	{
		adminAuth.POST("/user_v2/list", adminHandler.UserV2List)
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

	// 健康检查（不需要管理员权限，用于测试）
	admin.GET("/health", func(c *gin.Context) {
		middleware.Success(c, "Admin API is running", gin.H{
			"status": "ok",
		})
	})
}

// AdminRegisterRequest 管理员注册请求
type AdminRegisterRequest struct {
	Username string `json:"username" binding:"required"` // 可以是username、phone或email
	Password string `json:"password" binding:"required"`
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"` // 可以是username、phone或email
	Password string `json:"password" binding:"required"`
}

// findUserByIdentifier 通过username、phone或email查找用户
func findUserByIdentifier(identifier string) (*models.User, error) {
	userRepo := repository.NewUserRepository()

	// 先尝试通过username查找
	if user, err := userRepo.GetByUsername(identifier); err == nil {
		return user, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 如果username找不到，尝试通过phone查找
	if user, err := userRepo.GetByPhone(identifier); err == nil {
		return user, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 如果phone也找不到，尝试通过email查找
	if user, err := userRepo.GetByEmail(identifier); err == nil {
		return user, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return nil, gorm.ErrRecordNotFound
}

// AdminRegister 管理员注册
func (h *AdminHandler) AdminRegister(c *gin.Context) {
	var req AdminRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	username := strings.TrimSpace(req.Username)
	password := req.Password

	if username == "" || password == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户名和密码不能为空"))
		return
	}

	// 查找匹配的用户
	user, err := findUserByIdentifier(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(400, "没找到01agent用户，无注册资格"))
			return
		}
		repository.Errorf("Failed to find user: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查找用户失败"))
		return
	}

	// 检查用户是否是管理员（role = 3）
	if user.Role != 3 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "该用户不是管理员，无注册资格"))
		return
	}

	// 检查用户是否已经有密码
	if user.PasswordHash != nil && *user.PasswordHash != "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "该用户已设置密码，请直接登录"))
		return
	}

	// 设置密码（会自动转成hash）
	if err := user.HashPassword(password); err != nil {
		repository.Errorf("Failed to hash password: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "设置密码失败"))
		return
	}

	// 保存用户
	userRepo := repository.NewUserRepository()
	if err := userRepo.Update(user); err != nil {
		repository.Errorf("Failed to update user: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "保存用户失败"))
		return
	}

	// 构建响应
	response := gin.H{
		"user_id":  user.UserID,
		"username": tools.GetStringValue(user.Username),
		"phone":    tools.GetStringValue(user.Phone),
		"email":    tools.GetStringValue(user.Email),
		"nickname": tools.GetStringValue(user.Nickname),
	}

	middleware.Success(c, "注册成功", response)
}

// AdminLogin 管理员登录
func (h *AdminHandler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	username := strings.TrimSpace(req.Username)
	password := req.Password

	if username == "" || password == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户名和密码不能为空"))
		return
	}

	// 查找匹配的用户
	user, err := findUserByIdentifier(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(400, "用户不存在"))
			return
		}
		repository.Errorf("Failed to find user: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查找用户失败"))
		return
	}

	// 检查用户是否是管理员（role = 3）
	if user.Role != 3 {
		middleware.HandleError(c, middleware.NewBusinessError(403, "非管理员用户，无权登录"))
		return
	}

	// 检查用户是否有密码
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "该用户未设置密码，请先设置密码"))
		return
	}

	// 验证密码
	if !user.CheckPassword(password) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "密码错误"))
		return
	}

	// 构建用户信息（用于返回和生成token）
	userInfo := gin.H{
		"sub":               user.UserID,
		"id":                user.UserID,
		"nickname":          tools.GetStringValue(user.Nickname),
		"avatar":            tools.GetStringValue(user.Avatar),
		"username":          tools.GetStringValue(user.Username),
		"appid":             tools.GetStringValue(user.AppID),
		"openid":            tools.GetStringValue(user.OpenID),
		"phone":             tools.GetStringValue(user.Phone),
		"email":             tools.GetStringValue(user.Email),
		"credits":           user.Credits,
		"is_active":         user.IsActive,
		"vip_level":         user.VipLevel,
		"role":              user.Role,
		"status":            user.Status,
		"registration_date": user.RegistrationDate.Format(time.RFC3339),
		"last_login_time":   user.LastLoginTime.Format(time.RFC3339),
		"usage_count":       user.UsageCount,
		"total_consumption": func() string {
			if user.TotalConsumption != nil {
				return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", *user.TotalConsumption), "0"), ".")
			}
			return ""
		}(),
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	}

	// 获取客户端IP
	ipAddress := c.ClientIP()

	// 查找是否有已登录可用的session
	sessionRepo := repository.NewUserSessionRepository()
	activeSessions, err := sessionRepo.GetActiveSessionsByUserID(user.UserID)
	if err != nil {
		repository.Errorf("Failed to get active sessions: %v", err)
	}

	var token string
	var existingSession *models.UserSession

	// 查找最新的活跃会话（按创建时间降序，取第一个）
	if len(activeSessions) > 0 {
		existingSession = &activeSessions[0]
		if existingSession.Token != "" {
			// 如果存在可用的session，直接复用，更新活跃时间
			existingSession.LastActiveTime = time.Now()
			if err := sessionRepo.UpdateLastActiveTime(existingSession.ID); err != nil {
				repository.Errorf("Failed to update session: %v", err)
			} else {
				token = existingSession.Token
			}
		}
	}

	// 如果没有可用的session，创建新的session
	if token == "" {
		// 生成token
		usernameStr := tools.GetStringValue(user.Username)
		if usernameStr == "" {
			usernameStr = user.UserID
		}
		token, err = tools.GenerateToken(user.UserID, usernameStr)
		if err != nil {
			repository.Errorf("Failed to generate token: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "生成token失败"))
			return
		}

		// 创建新的session
		session := &models.UserSession{
			UserID:         user.UserID,
			Token:          token,
			LoginType:      "web",
			IPAddress:      ipAddress,
			Status:         1,
			LoginTime:      time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour), // 24小时过期
			IsActive:       true,
			LastActiveTime: time.Now(),
			CreatedAt:      time.Now(),
		}

		if err := sessionRepo.Create(session); err != nil {
			repository.Errorf("Failed to create session: %v", err)
			// 即使创建session失败，也返回token（不影响登录）
		}
	}

	// 更新最后登录时间
	userRepo := repository.NewUserRepository()
	if err := userRepo.UpdateLastLoginTime(user.UserID); err != nil {
		repository.Errorf("Failed to update last login time: %v", err)
	}

	// 返回响应
	response := gin.H{
		"token": token,
		"user":  userInfo,
	}

	middleware.Success(c, "登录成功", response)
}

// UserV2ListRequest 用户列表查询请求模型V2
type UserV2ListRequest struct {
	Page                int      `json:"page" binding:"min=1"`
	PageSize            int      `json:"page_size" binding:"min=1,max=100"`
	UserID              *string  `json:"user_id"`
	Username            *string  `json:"username"`
	Phone               *string  `json:"phone"`
	Email               *string  `json:"email"`
	Nickname            *string  `json:"nickname"`
	Roles               []string `json:"roles"`
	Statuses            []string `json:"statuses"`
	VipLevels           []int    `json:"vip_levels"`
	Channels            []string `json:"channels"`
	MinTotalConsumption *float64 `json:"min_total_consumption"`
	MaxTotalConsumption *float64 `json:"max_total_consumption"`
	MinCredits          *float64 `json:"min_credits"`
	MaxCredits          *float64 `json:"max_credits"`
	StartDate           *string  `json:"start_date"`
	EndDate             *string  `json:"end_date"`
	OrderBy             *string  `json:"order_by"`
	OrderDirection      string   `json:"order_direction"`
}

// UserV2List 用户列表查询接口V2
func (h *AdminHandler) UserV2List(c *gin.Context) {
	var req UserV2ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}
	if req.OrderBy == nil || *req.OrderBy == "" {
		orderBy := "created_at"
		req.OrderBy = &orderBy
	}

	// 如果提供了用户ID，直接根据用户ID查询
	if req.UserID != nil && *req.UserID != "" {
		userRepo := repository.NewUserRepository()
		user, err := userRepo.GetByID(*req.UserID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.Success(c, "success", gin.H{
					"total":     0,
					"items":     []gin.H{},
					"page":      req.Page,
					"page_size": req.PageSize,
				})
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败: "+err.Error()))
			return
		}

		// 计算用户统计信息
		var paymentCount, articleCount, copilotCount, aiEditCount, invitationCount int64

		repository.DB.Model(&models.Trade{}).
			Where("user_id = ? AND payment_status = ?", user.UserID, 1). // PaymentStatus.SUCCESS = 1
			Count(&paymentCount)

		repository.DB.Model(&models.ArticleTask{}).
			Where("user_id = ?", user.UserID).
			Count(&articleCount)

		// 注意：CopilotChatSession和AIRewriteRecords可能不存在，先注释
		// repository.DB.Model(&models.CopilotChatSession{}).
		// 	Where("user_id = ?", user.UserID).
		// 	Count(&copilotCount)

		// repository.DB.Model(&models.AIRewriteRecord{}).
		// 	Where("user_id = ?", user.UserID).
		// 	Count(&aiEditCount)

		repository.DB.Model(&models.InvitationRelation{}).
			Where("inviter_id = ?", user.UserID).
			Count(&invitationCount)

		// 获取用户最近登录认证凭证token
		var session models.UserSession
		var token *string
		if err := repository.DB.Where("user_id = ? AND is_active = ?", user.UserID, true).
			Order("created_at DESC").First(&session).Error; err == nil {
			tokenStr := session.Token
			token = &tokenStr
		}

		totalConsumption := 0.0
		if user.TotalConsumption != nil {
			totalConsumption = *user.TotalConsumption
		}

		// 构建单个用户记录返回数据
		result := []gin.H{{
			"user_id":           user.UserID,
			"username":          user.Username,
			"phone":             user.Phone,
			"email":             user.Email,
			"nickname":          user.Nickname,
			"avatar":            user.Avatar,
			"role":              user.Role,
			"status":            user.Status,
			"vip_level":         user.VipLevel,
			"credits":           float64(user.Credits),
			"total_consumption": totalConsumption,
			"usage_count":       user.UsageCount,
			"payment_count":     paymentCount,
			"article_count":     articleCount,
			"copilot_count":     copilotCount,
			"ai_edit_count":     aiEditCount,
			"utm_source":        user.UtmSource,
			"invitation_count":  invitationCount,
			"token":             token,
			"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}}

		middleware.Success(c, "success", gin.H{
			"total":     1,
			"items":     result,
			"page":      1,
			"page_size": 1,
		})
		return
	}

	// 构建复杂查询
	query := repository.DB.Model(&models.User{})

	// 时间范围筛选
	if req.StartDate != nil && *req.StartDate != "" {
		start, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		query = query.Where("created_at >= ?", start)
	}

	if req.EndDate != nil && *req.EndDate != "" {
		end, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
		query = query.Where("created_at <= ?", end)
	}

	// 用户基本信息筛选
	if req.Username != nil && *req.Username != "" {
		query = query.Where("username LIKE ?", "%"+*req.Username+"%")
	}
	if req.Phone != nil && *req.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+*req.Phone+"%")
	}
	if req.Email != nil && *req.Email != "" {
		query = query.Where("email LIKE ?", "%"+*req.Email+"%")
	}
	if req.Nickname != nil && *req.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+*req.Nickname+"%")
	}

	// 角色筛选
	if len(req.Roles) > 0 {
		roleValues := make([]int16, 0)
		roleMapping := map[string]int16{
			"user":  1, // UserRoleNormal
			"vip":   2, // UserRoleVIP
			"admin": 3, // UserRoleAdmin
		}
		for _, roleStr := range req.Roles {
			if role, ok := roleMapping[strings.ToLower(roleStr)]; ok {
				roleValues = append(roleValues, role)
			} else {
				// 尝试解析为数字
				var role int16
				if _, err := fmt.Sscanf(roleStr, "%d", &role); err == nil {
					roleValues = append(roleValues, role)
				}
			}
		}
		if len(roleValues) > 0 {
			query = query.Where("role IN ?", roleValues)
		}
	}

	// 状态筛选
	if len(req.Statuses) > 0 {
		statusValues := make([]int16, 0)
		statusMapping := map[string]int16{
			"inactive": 0,
			"active":   1,
		}
		for _, statusStr := range req.Statuses {
			if status, ok := statusMapping[strings.ToLower(statusStr)]; ok {
				statusValues = append(statusValues, status)
			} else {
				// 尝试解析为数字
				var status int16
				if _, err := fmt.Sscanf(statusStr, "%d", &status); err == nil {
					statusValues = append(statusValues, status)
				}
			}
		}
		if len(statusValues) > 0 {
			query = query.Where("status IN ?", statusValues)
		}
	}

	// VIP等级筛选
	if len(req.VipLevels) > 0 {
		query = query.Where("vip_level IN ?", req.VipLevels)
	}

	// 渠道筛选
	if len(req.Channels) > 0 {
		query = query.Where("utm_source IN ?", req.Channels)
	}

	// 消费金额区间筛选
	if req.MinTotalConsumption != nil {
		query = query.Where("total_consumption >= ?", *req.MinTotalConsumption)
	}
	if req.MaxTotalConsumption != nil {
		query = query.Where("total_consumption <= ?", *req.MaxTotalConsumption)
	}

	// 积分区间筛选
	if req.MinCredits != nil {
		query = query.Where("credits >= ?", int(*req.MinCredits))
	}
	if req.MaxCredits != nil {
		query = query.Where("credits <= ?", int(*req.MaxCredits))
	}

	// 排序
	orderField := *req.OrderBy
	if req.OrderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var users []models.User
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(users))
	for _, user := range users {
		// 获取用户最近登录认证凭证token
		var session models.UserSession
		var token *string
		if err := repository.DB.Where("user_id = ? AND is_active = ?", user.UserID, true).
			Order("created_at DESC").First(&session).Error; err == nil {
			tokenStr := session.Token
			token = &tokenStr
		}

		totalConsumption := 0.0
		if user.TotalConsumption != nil {
			totalConsumption = *user.TotalConsumption
		}

		result = append(result, gin.H{
			"user_id":           user.UserID,
			"username":          user.Username,
			"phone":             user.Phone,
			"email":             user.Email,
			"nickname":          user.Nickname,
			"avatar":            user.Avatar,
			"role":              user.Role,
			"status":            user.Status,
			"vip_level":         user.VipLevel,
			"token":             token,
			"credits":           float64(user.Credits),
			"total_consumption": totalConsumption,
			"usage_count":       user.UsageCount,
			"utm_source":        user.UtmSource,
			"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetUserFeedbackStats 获取用户反馈统计
func (h *AdminHandler) GetUserFeedbackStats(c *gin.Context) {
	// 统计各类型反馈的数量
	var noneCount, satisfiedCount, dissatisfiedCount, totalCount int64

	// 统计未操作类型（0）
	if err := repository.DB.Model(&models.UserFeedback{}).
		Where("feedback_type = ?", models.FeedbackTypeNone).
		Count(&noneCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计满意类型（1）
	if err := repository.DB.Model(&models.UserFeedback{}).
		Where("feedback_type = ?", models.FeedbackTypeSatisfied).
		Count(&satisfiedCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计不满意类型（2）
	if err := repository.DB.Model(&models.UserFeedback{}).
		Where("feedback_type = ?", models.FeedbackTypeUnsatisfied).
		Count(&dissatisfiedCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计总数
	if err := repository.DB.Model(&models.UserFeedback{}).
		Count(&totalCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取反馈统计成功", gin.H{
		"none_count":         noneCount,
		"satisfied_count":    satisfiedCount,
		"dissatisfied_count": dissatisfiedCount,
		"total_count":        totalCount,
	})
}

// GetPaidUsersWechatAccounts 获取付费用户绑定的公众号信息
func (h *AdminHandler) GetPaidUsersWechatAccounts(c *gin.Context) {
	var req struct {
		Page            int    `form:"page" binding:"min=1"`
		PageSize        int    `form:"page_size" binding:"min=1,max=100"`
		Keyword         string `form:"keyword"`
		HasAppid        *bool  `form:"has_appid"`
		FetchWechatInfo bool   `form:"fetch_wechat_info"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 1. 获取所有VIP和管理员用户ID（role = 2 或 3）
	var vipAdminUserIDs []string
	if err := repository.DB.Model(&models.User{}).
		Where("role IN ?", []int16{2, 3}). // UserRoleVIP = 2, UserRoleAdmin = 3
		Pluck("user_id", &vipAdminUserIDs).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询VIP和管理员用户失败: "+err.Error()))
		return
	}

	// 2. 获取所有有成功充值订单的用户ID（trade_type = "recharge", payment_status = "success"）
	var paidUserIDs []string
	if err := repository.DB.Model(&models.Trade{}).
		Where("payment_status = ? AND trade_type = ?", "success", "recharge").
		Distinct("user_id").
		Pluck("user_id", &paidUserIDs).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询付费用户失败: "+err.Error()))
		return
	}

	// 3. 合并用户ID（去重）
	userIDMap := make(map[string]bool)
	for _, id := range vipAdminUserIDs {
		userIDMap[id] = true
	}
	for _, id := range paidUserIDs {
		userIDMap[id] = true
	}
	var allPaidUserIDs []string
	for id := range userIDMap {
		allPaidUserIDs = append(allPaidUserIDs, id)
	}

	if len(allPaidUserIDs) == 0 {
		middleware.Success(c, "获取成功", gin.H{
			"total":     0,
			"page":      req.Page,
			"page_size": req.PageSize,
			"list":      []gin.H{},
		})
		return
	}

	// 4. 构建查询条件
	query := repository.DB.Model(&models.User{}).Where("user_id IN ?", allPaidUserIDs)

	// 5. 默认只返回已绑定appid的用户（has_appid默认为true）
	hasAppidFilter := req.HasAppid == nil || (req.HasAppid != nil && *req.HasAppid)
	if hasAppidFilter {
		query = query.Where("appid IS NOT NULL AND appid != ''")
	} else {
		// 如果明确要求查看未绑定的用户，重新构建查询
		query = repository.DB.Model(&models.User{}).Where("user_id IN ?", allPaidUserIDs)
		if req.Keyword != "" {
			keyword := "%" + req.Keyword + "%"
			query = query.Where("username LIKE ? OR phone LIKE ? OR nickname LIKE ?", keyword, keyword, keyword)
		}
		query = query.Where("appid IS NULL OR appid = ''")
	}

	// 6. 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("username LIKE ? OR phone LIKE ? OR nickname LIKE ?", keyword, keyword, keyword)
	}

	// 7. 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计总数失败: "+err.Error()))
		return
	}

	// 8. 分页查询
	var users []models.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&users).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败: "+err.Error()))
		return
	}

	// 9. 构建返回数据
	resultList := []gin.H{}
	for _, user := range users {
		// 获取用户的支付统计（只统计充值订单）
		var paidCount int64
		var totalAmount float64
		if err := repository.DB.Model(&models.Trade{}).
			Where("user_id = ? AND payment_status = ? AND trade_type = ?", user.UserID, "success", "recharge").
			Count(&paidCount).Error; err == nil {
			var amounts []float64
			if err := repository.DB.Model(&models.Trade{}).
				Where("user_id = ? AND payment_status = ? AND trade_type = ?", user.UserID, "success", "recharge").
				Pluck("amount", &amounts).Error; err == nil {
				for _, amount := range amounts {
					totalAmount += amount
				}
			}
		}

		// 获取用户参数
		var userParam models.UserParameters
		repository.DB.Where("user_id = ?", user.UserID).First(&userParam)

		// 确定角色名称
		roleName := "普通用户"
		if user.Role == 3 {
			roleName = "管理员"
		} else if user.Role == 2 {
			roleName = "VIP"
		}

		// 构建用户数据
		userData := gin.H{
			"user_id":       user.UserID,
			"username":      user.Username,
			"nickname":      user.Nickname,
			"phone":         user.Phone,
			"email":         user.Email,
			"avatar":        user.Avatar,
			"role":          user.Role,
			"role_name":     roleName,
			"appid":         user.AppID,
			"has_appid":     user.AppID != nil && *user.AppID != "",
			"is_gzh_bind":   userParam.IsGzhBind,
			"paid_count":    paidCount,
			"total_amount":  totalAmount,
			"wechat_info":   nil,
			"wechat_status": "未绑定",
		}

		// 格式化注册日期
		if !user.RegistrationDate.IsZero() {
			userData["registration_date"] = user.RegistrationDate.Format("2006-01-02 15:04:05")
		} else {
			userData["registration_date"] = nil
		}

		// 如果用户有appid，处理公众号信息
		if user.AppID != nil && *user.AppID != "" {
			if req.FetchWechatInfo {
				// TODO: 调用微信API获取公众号详细信息
				// 这里暂时返回占位符，实际需要调用微信API
				userData["wechat_status"] = "已配置(未获取详情)"
				userData["wechat_info"] = gin.H{
					"success": nil,
					"message": "已配置appid，但未获取详细信息。设置 fetch_wechat_info=true 可获取详情",
				}
			} else {
				userData["wechat_status"] = "已配置(未获取详情)"
				userData["wechat_info"] = gin.H{
					"success": nil,
					"message": "已配置appid，但未获取详细信息。设置 fetch_wechat_info=true 可获取详情",
				}
			}
		}

		resultList = append(resultList, userData)
	}

	middleware.Success(c, "获取成功", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      resultList,
	})
}

// GetProductList 获取产品列表
func (h *AdminHandler) GetProductList(c *gin.Context) {
	var req struct {
		Page           int     `form:"page" binding:"min=1"`
		PageSize       int     `form:"page_size" binding:"min=1,max=9999"`
		Search         string  `form:"search"`
		OrderBy        string  `form:"order_by"`
		Status         *int    `form:"status"`
		ProductType    *string `form:"product_type"`
		IsCustom       bool    `form:"is_custom"`
		OrderDirection string  `form:"order_direction"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}

	// 构建查询
	query := repository.DB.Model(&models.Production{})

	// 产品类型筛选
	if req.ProductType != nil && *req.ProductType != "" {
		query = query.Where("product_type = ?", *req.ProductType)
	}

	// 状态筛选
	if req.Status != nil && (*req.Status == 0 || *req.Status == 1) {
		query = query.Where("status = ?", *req.Status)
	}

	// 搜索
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 排序
	orderField := req.OrderBy
	if req.OrderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var products []models.Production
	if err := query.Offset(offset).Limit(req.PageSize).Find(&products).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(products))
	for _, item := range products {
		result = append(result, gin.H{
			"id":              item.ID,
			"name":            item.Name,
			"description":     item.Description,
			"price":           item.Price,
			"original_price":  item.OriginalPrice,
			"extra_info":      item.ExtraInfo,
			"status":          item.Status,
			"validity_period": item.ValidityPeriod,
			"product_type":    item.ProductType,
			"created_at":      item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// 如果是管理员且is_custom为true，添加"专业版开通测试"产品
	if req.IsCustom {
		var customProduct models.Production
		if err := repository.DB.Where("name = ?", "专业版开通测试").First(&customProduct).Error; err == nil {
			result = append(result, gin.H{
				"id":              customProduct.ID,
				"name":            customProduct.Name,
				"description":     customProduct.Description,
				"price":           customProduct.Price,
				"original_price":  customProduct.OriginalPrice,
				"extra_info":      customProduct.ExtraInfo,
				"status":          customProduct.Status,
				"validity_period": customProduct.ValidityPeriod,
				"product_type":    customProduct.ProductType,
				"created_at":      customProduct.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"updated_at":      customProduct.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetAllContentList 获取全部用户的内容列表
func (h *AdminHandler) GetAllContentList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"min=1"`
		PageSize int    `form:"page_size" binding:"min=1,max=100"`
		Category string `form:"category"` // long_post/xiaohongshu/short_post/poster/other
		Keyword  string `form:"keyword"`
		UserID   string `form:"user_id"`
		OrderBy  string `form:"order_by"`
		Order    string `form:"order"` // asc/desc
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.OrderBy == "" {
		req.OrderBy = "updated_at"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	items := []gin.H{}
	var total int64

	// 查询长图文（当category为空或为long_post时）
	if req.Category == "" || req.Category == "long_post" {
		query := repository.DB.Model(&models.ArticleEditTask{})

		if req.UserID != "" {
			query = query.Where("user_id = ?", req.UserID)
		}

		if req.Keyword != "" {
			keywordPattern := "%" + req.Keyword + "%"
			query = query.Where("title LIKE ?", keywordPattern)
		}

		if req.Category == "long_post" {
			// 只查询长图文
			orderField := req.OrderBy
			if req.Order == "desc" {
				orderField = orderField + " DESC"
			} else {
				orderField = orderField + " ASC"
			}
			query = query.Order(orderField)

			if err := query.Count(&total).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
				return
			}

			offset := (req.Page - 1) * req.PageSize
			var longPosts []models.ArticleEditTask
			if err := query.Offset(offset).Limit(req.PageSize).Find(&longPosts).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
				return
			}

			for _, p := range longPosts {
				// 获取用户信息
				var user models.User
				repository.DB.Where("user_id = ?", p.UserID).First(&user)

				content := ""
				if p.Content != "" {
					if len(p.Content) > 100 {
						content = p.Content[:100] + "..."
					} else {
						content = p.Content
					}
				}

				items = append(items, gin.H{
					"id":              p.ID,
					"article_task_id": p.ArticleTaskID,
					"user_id":         p.UserID,
					"user_name":       tools.GetStringValue(user.Nickname),
					"title":           p.Title,
					"theme":           p.Theme,
					"content":         content,
					"section_html":    tools.GetStringValue(p.SectionHTML),
					"status":          p.Status,
					"is_public":       p.IsPublic,
					"tags":            p.Tags,
					"category":        "long_post",
					"created_at":      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"published_at": func() *string {
						if p.PublishedAt != nil {
							formatted := p.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
							return &formatted
						}
						return nil
					}(),
				})
			}

			middleware.Success(c, "success", gin.H{
				"items":     items,
				"total":     total,
				"page":      req.Page,
				"page_size": req.PageSize,
			})
			return
		}
	}

	// 查询ShortPostProject（当category为xiaohongshu/short_post/poster/other时）
	categoryToProjectType := map[string]string{
		"xiaohongshu": "xiaohongshu",
		"short_post":  "short_post",
		"poster":      "poster",
		"other":       "other",
	}

	if projectType, ok := categoryToProjectType[req.Category]; ok {
		query := repository.DB.Model(&short_post.ShortPostProject{})

		if req.UserID != "" {
			query = query.Where("user_id = ?", req.UserID)
		}

		if req.Keyword != "" {
			keywordPattern := "%" + req.Keyword + "%"
			query = query.Where("name LIKE ?", keywordPattern)
		}

		query = query.Where("project_type = ?", projectType)

		// 排序
		orderField := req.OrderBy
		if req.Order == "desc" {
			orderField = orderField + " DESC"
		} else {
			orderField = orderField + " ASC"
		}
		query = query.Order(orderField)

		if err := query.Count(&total).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}

		offset := (req.Page - 1) * req.PageSize
		var shortPosts []short_post.ShortPostProject
		if err := query.Offset(offset).Limit(req.PageSize).Find(&shortPosts).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}

		for _, p := range shortPosts {
			// 获取用户信息
			var user models.User
			repository.DB.Where("user_id = ?", p.UserID).First(&user)

			items = append(items, gin.H{
				"id":           p.ID,
				"user_id":      p.UserID,
				"user_name":    tools.GetStringValue(user.Nickname),
				"name":         p.Name,
				"description":  tools.GetStringValue(p.Description),
				"cover_image":  tools.GetStringValue(p.CoverImage),
				"thumbnail":    tools.GetStringValue(p.Thumbnail),
				"project_type": string(p.ProjectType),
				"status":       string(p.Status),
				"frame_count":  p.FrameCount,
				"category":     req.Category,
				"created_at":   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"updated_at":   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"saved_at": func() *string {
					if p.SavedAt != nil {
						formatted := p.SavedAt.Format("2006-01-02T15:04:05Z07:00")
						return &formatted
					}
					return nil
				}(),
			})
		}

		middleware.Success(c, "success", gin.H{
			"items":     items,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		})
		return
	}

	// 如果category不匹配任何已知类型，返回空结果
	middleware.Success(c, "success", gin.H{
		"items":     []gin.H{},
		"total":     0,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetFeedbackStats 获取系统反馈统计
func (h *AdminHandler) GetFeedbackStats(c *gin.Context) {
	// 统计各状态的反馈数量
	var pendingCount, processingCount, resolvedCount, totalCount int64

	// 统计待处理状态
	if err := repository.DB.Model(&models.Feedback{}).
		Where("status = ?", "pending").
		Count(&pendingCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计处理中状态
	if err := repository.DB.Model(&models.Feedback{}).
		Where("status = ?", "processing").
		Count(&processingCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计已解决状态
	if err := repository.DB.Model(&models.Feedback{}).
		Where("status = ?", "resolved").
		Count(&resolvedCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计总数
	if err := repository.DB.Model(&models.Feedback{}).
		Count(&totalCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取反馈统计成功", gin.H{
		"pending_count":    pendingCount,
		"processing_count": processingCount,
		"resolved_count":   resolvedCount,
		"total_count":      totalCount,
	})
}

// GetImageExampleList 获取图文生成示例列表
func (h *AdminHandler) GetImageExampleList(c *gin.Context) {
	var req struct {
		Page      int                    `json:"page" binding:"min=1"`
		PageSize  int                    `json:"page_size" binding:"min=1"`
		Name      string                 `json:"name"`
		Tags      []string               `json:"tags"`
		ExtraData map[string]interface{} `json:"extra_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建查询
	query := repository.DB.Model(&models.ImageExample{})

	// 名称搜索
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// Tags搜索（如果tags是JSON数组，需要特殊处理）
	// 这里简化处理，实际可能需要根据数据库类型调整

	// ExtraData搜索
	if req.ExtraData != nil {
		// 特殊处理：如果size是"1080×自适应"且width != height
		if size, ok := req.ExtraData["size"].(string); ok && size == "1080×自适应" {
			if width, wOk := req.ExtraData["width"]; wOk {
				if height, hOk := req.ExtraData["height"]; hOk && width != height {
					// 查询width=1080的记录
					query = query.Where("extra_data LIKE ?", "%\"width\":1080%")
				}
			}
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var examples []models.ImageExample
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&examples).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(examples))
	for _, item := range examples {
		images := []string{}
		// 如果是小红书类型，获取详情中的图片
		if item.ProjectType == "xiaohongshu" {
			var detail models.ImageExampleDetail
			if err := repository.DB.Where("example_id = ?", item.ID).First(&detail).Error; err == nil {
				if detail.Images != nil {
					// 解析JSON数组
					// 这里简化处理，实际需要解析JSON
				}
			}
		}

		result = append(result, gin.H{
			"id":           item.ID,
			"name":         item.Name,
			"prompt":       item.Prompt,
			"cover_url":    tools.GetStringValue(item.CoverURL),
			"tags":         item.Tags,
			"sort_order":   item.SortOrder,
			"is_visible":   item.IsVisible,
			"extra_data":   item.ExtraData,
			"images":       images,
			"project_type": item.ProjectType,
			"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items":     result,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetUserCustomSizeList 获取用户自定义尺寸列表
func (h *AdminHandler) GetUserCustomSizeList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"min=1"`
		PageSize int    `form:"page_size" binding:"min=1"`
		Status   *int   `form:"status"`
		Name     string `form:"name"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	// 构建查询
	query := repository.DB.Model(&models.UserCustomSize{})

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 名称搜索
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var sizes []models.UserCustomSize
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&sizes).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(sizes))
	for _, size := range sizes {
		result = append(result, gin.H{
			"id":         size.ID,
			"user_id":    size.UserID,
			"name":       size.Name,
			"data":       size.Data,
			"status":     size.Status,
			"is_default": size.IsDefault,
			"sort_order": size.SortOrder,
			"created_at": size.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": size.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items":     result,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetPublicTemplates 获取公开模板列表（管理员接口，包含所有状态）
func (h *AdminHandler) GetPublicTemplates(c *gin.Context) {
	var req struct {
		Page         int    `form:"page" binding:"min=1"`
		PageSize     int    `form:"page_size" binding:"min=1"`
		Search       string `form:"search"`
		Category     string `form:"category"`
		TemplateType *int   `form:"template_type"`
		Status       *int   `form:"status"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	// 构建查询 - 管理员接口，不限制状态
	query := repository.DB.Model(&models.PublicTemplate{})

	// 搜索
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ?", searchPattern)
	}

	// 分类筛选
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	// 类型筛选
	if req.TemplateType != nil {
		query = query.Where("template_type = ?", *req.TemplateType)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询，按创建时间降序排序
	offset := (req.Page - 1) * req.PageSize
	var templates []models.PublicTemplate
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&templates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 获取缩略图内容
	thumbnailContent := tools.GetThumbnailContent()
	processor := tools.NewUnifiedMarkdownProcessor()

	// 构建返回数据
	result := make([]gin.H, 0, len(templates))
	for _, template := range templates {
		// 生成缩略图HTML（数据库未存储时动态生成）
		sectionHTML := tools.GetStringValue(template.SectionHTML)
		if sectionHTML == "" {
			// 使用UnifiedMarkdownProcessor处理markdown
			if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
				sectionHTML = processedHTML
			}
		}

		price := 0.0
		if template.Price > 0 {
			price = template.Price
		}

		var originalPrice *float64
		if template.OriginalPrice != nil && *template.OriginalPrice > 0 {
			originalPrice = template.OriginalPrice
		}

		templateType := "unified"
		if template.TemplateType == models.TemplateTypeWechat {
			templateType = "wechat"
		}

		nameEn := tools.GetStringValue(template.NameEn)
		if nameEn == "" {
			nameEn = template.Name
		}

		description := tools.GetStringValue(template.Description)
		if description == "" {
			description = fmt.Sprintf("%s 主题", template.Name)
		}

		result = append(result, gin.H{
			"id":             template.TemplateID,
			"template_id":    template.TemplateID,
			"label":          template.Name,
			"labelEn":        nameEn,
			"value":          template.TemplateID,
			"name":           template.Name,
			"name_en":        nameEn,
			"description":    description,
			"author":         template.Author,
			"template_type":  template.TemplateType,
			"type":           templateType,
			"status":         template.Status,
			"price_type":     template.PriceType,
			"price":          price,
			"original_price": originalPrice,
			"is_public":      template.IsPublic,
			"is_featured":    template.IsFeatured,
			"is_official":    template.IsOfficial,
			"preview_url":    tools.GetStringValue(template.PreviewURL),
			"previewUrl":     tools.GetStringValue(template.PreviewURL),
			"thumbnail_url":  tools.GetStringValue(template.ThumbnailURL),
			"primary_color":  template.PrimaryColor,
			"tags":           template.Tags,
			"category":       tools.GetStringValue(template.Category),
			"download_count": template.DownloadCount,
			"use_count":      template.UseCount,
			"like_count":     template.LikeCount,
			"view_count":     template.ViewCount,
			"sort_order":     template.SortOrder,
			"section_html":   sectionHTML,
			"created_at":     template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})

		if template.PublishedAt != nil {
			result[len(result)-1]["published_at"] = template.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		} else {
			result[len(result)-1]["published_at"] = nil
		}
	}

	middleware.Success(c, "获取成功", gin.H{
		"templates": result,
		"pagination": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"total":     total,
			"pages":     (int(total) + req.PageSize - 1) / req.PageSize,
		},
	})
}

// GetPublicTemplateDetail 获取公开模板详情（管理员接口，包含完整配置数据）
func (h *AdminHandler) GetPublicTemplateDetail(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("模板 '%s' 不存在", templateID)))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 生成缩略图HTML
	sectionHTML := tools.GetStringValue(template.SectionHTML)
	if sectionHTML == "" {
		thumbnailContent := tools.GetThumbnailContent()
		processor := tools.NewUnifiedMarkdownProcessor()
		if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	price := 0.0
	if template.Price > 0 {
		price = template.Price
	}

	var originalPrice *float64
	if template.OriginalPrice != nil && *template.OriginalPrice > 0 {
		originalPrice = template.OriginalPrice
	}

	templateType := "unified"
	if template.TemplateType == models.TemplateTypeWechat {
		templateType = "wechat"
	}

	nameEn := tools.GetStringValue(template.NameEn)
	if nameEn == "" {
		nameEn = template.Name
	}

	description := tools.GetStringValue(template.Description)
	if description == "" {
		description = fmt.Sprintf("%s 主题", template.Name)
	}

	result := gin.H{
		"id":             template.TemplateID,
		"template_id":    template.TemplateID,
		"label":          template.Name,
		"labelEn":        nameEn,
		"value":          template.TemplateID,
		"name":           template.Name,
		"name_en":        nameEn,
		"description":    description,
		"author":         template.Author,
		"template_type":  template.TemplateType,
		"type":           templateType,
		"status":         template.Status,
		"price_type":     template.PriceType,
		"price":          price,
		"original_price": originalPrice,
		"is_public":      template.IsPublic,
		"is_featured":    template.IsFeatured,
		"is_official":    template.IsOfficial,
		"preview_url":    tools.GetStringValue(template.PreviewURL),
		"previewUrl":     tools.GetStringValue(template.PreviewURL),
		"thumbnail_url":  tools.GetStringValue(template.ThumbnailURL),
		"primary_color":  template.PrimaryColor,
		"tags":           template.Tags,
		"category":       tools.GetStringValue(template.Category),
		"download_count": template.DownloadCount,
		"use_count":      template.UseCount,
		"like_count":     template.LikeCount,
		"view_count":     template.ViewCount,
		"sort_order":     template.SortOrder,
		"section_html":   sectionHTML,
		"config":         template.TemplateData,
		"template_data":  template.TemplateData,
		"created_at":     template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":     template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if template.PublishedAt != nil {
		result["published_at"] = template.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		result["published_at"] = nil
	}

	middleware.Success(c, "获取成功", result)
}

// GetThemePreview 获取主题预览HTML
func (h *AdminHandler) GetThemePreview(c *gin.Context) {
	themeName := c.Param("theme_name")
	if themeName == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "主题名称不能为空"))
		return
	}

	// 查找模板（theme_name可能是template_id或name_en）
	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ? OR name_en = ?", themeName, themeName).
		Where("status = ? AND is_public = ?", models.TemplateStatusPublished, true).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "主题不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 生成缩略图HTML
	sectionHTML := tools.GetStringValue(template.SectionHTML)
	if sectionHTML == "" {
		thumbnailContent := tools.GetThumbnailContent()
		processor := tools.NewUnifiedMarkdownProcessor()
		if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	// 构建完整预览HTML
	fullPreview := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 预览</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background-color: #f5f5f5;
        }
        .preview-container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border-radius: 8px;
        }
    </style>
</head>
<body>
    <div class="preview-container">
        %s
    </div>
</body>
</html>`, template.Name, sectionHTML)

	templateType := "unified"
	if template.TemplateType == models.TemplateTypeWechat {
		templateType = "wechat"
	}

	result := gin.H{
		"theme_name":   template.TemplateID,
		"chinese_name": template.Name,
		"type":         templateType,
		"section_html": sectionHTML,
		"full_preview": fullPreview,
	}

	// 如果有template_data，也返回配置信息
	if template.TemplateData != nil {
		result["meta"] = gin.H{} // 可以从template_data中提取meta信息
		result["config"] = template.TemplateData
	}

	middleware.Success(c, "获取成功", result)
}

// UpdatePublicTemplate 更新官方模板
func (h *AdminHandler) UpdatePublicTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("模板 '%s' 不存在", templateID)))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var req struct {
		Name          *string                `json:"name"`
		NameEn        *string                `json:"name_en"`
		Description   *string                `json:"description"`
		Author        *string                `json:"author"`
		TemplateType  *int                   `json:"template_type"`
		Status        *int                   `json:"status"`
		PriceType     *int                   `json:"price_type"`
		Price         *float64               `json:"price"`
		OriginalPrice *float64               `json:"original_price"`
		IsPublic      *bool                  `json:"is_public"`
		IsFeatured    *bool                  `json:"is_featured"`
		IsOfficial    *bool                  `json:"is_official"`
		PreviewURL    *string                `json:"preview_url"`
		ThumbnailURL  *string                `json:"thumbnail_url"`
		PrimaryColor  *string                `json:"primary_color"`
		Tags          interface{}            `json:"tags"`
		Category      *string                `json:"category"`
		SortOrder     *int                   `json:"sort_order"`
		TemplateData  map[string]interface{} `json:"template_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	hasUpdate := false

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "模板名称不能为空"))
			return
		}
		updates["name"] = name
		hasUpdate = true
	}

	if req.NameEn != nil {
		nameEn := strings.TrimSpace(*req.NameEn)
		if nameEn == "" {
			updates["name_en"] = nil
		} else {
			updates["name_en"] = nameEn
		}
		hasUpdate = true
	}

	if req.Description != nil {
		updates["description"] = *req.Description
		hasUpdate = true
	}

	if req.Author != nil {
		updates["author"] = *req.Author
		hasUpdate = true
	}

	if req.TemplateType != nil {
		updates["template_type"] = models.TemplateType(*req.TemplateType)
		hasUpdate = true
	}

	if req.Status != nil {
		newStatus := models.TemplateStatus(*req.Status)
		updates["status"] = newStatus
		// 如果状态改为已发布且之前没有发布时间，设置发布时间
		if newStatus == models.TemplateStatusPublished && template.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = &now
		}
		hasUpdate = true
	}

	if req.PriceType != nil {
		updates["price_type"] = models.PriceType(*req.PriceType)
		hasUpdate = true
	}

	if req.Price != nil {
		updates["price"] = *req.Price
		hasUpdate = true
	}

	if req.OriginalPrice != nil {
		if *req.OriginalPrice > 0 {
			updates["original_price"] = *req.OriginalPrice
		} else {
			updates["original_price"] = nil
		}
		hasUpdate = true
	}

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
		hasUpdate = true
	}

	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
		hasUpdate = true
	}

	if req.IsOfficial != nil {
		updates["is_official"] = *req.IsOfficial
		hasUpdate = true
	}

	if req.PreviewURL != nil {
		updates["preview_url"] = *req.PreviewURL
		hasUpdate = true
	}

	if req.ThumbnailURL != nil {
		updates["thumbnail_url"] = *req.ThumbnailURL
		hasUpdate = true
	}

	if req.PrimaryColor != nil {
		updates["primary_color"] = *req.PrimaryColor
		hasUpdate = true
	}

	if req.Tags != nil {
		// 将tags转换为JSON字符串
		if tagsJSON, err := json.Marshal(req.Tags); err == nil {
			tagsStr := string(tagsJSON)
			updates["tags"] = &tagsStr
		}
		hasUpdate = true
	}

	if req.Category != nil {
		updates["category"] = *req.Category
		hasUpdate = true
	}

	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
		hasUpdate = true
	}

	if req.TemplateData != nil {
		// 将template_data转换为JSON字符串
		if templateDataJSON, err := json.Marshal(req.TemplateData); err == nil {
			templateDataStr := string(templateDataJSON)
			updates["template_data"] = &templateDataStr
			// 如果更新了模板数据，清除section_html以便重新生成
			updates["section_html"] = nil
		}
		hasUpdate = true
	}

	if !hasUpdate {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	// 执行更新 - 使用 Select 确保所有字段都能更新
	updateFields := make([]string, 0, len(updates))
	for k := range updates {
		updateFields = append(updateFields, k)
	}

	if err := repository.DB.Model(&template).Select(updateFields).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 重新查询以获取最新数据
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询更新后的数据失败: "+err.Error()))
		return
	}

	// 生成新的section_html
	sectionHTML := tools.GetStringValue(template.SectionHTML)
	if sectionHTML == "" {
		thumbnailContent := tools.GetThumbnailContent()
		processor := tools.NewUnifiedMarkdownProcessor()
		if processedHTML, err := processor.ProcessMarkdown(thumbnailContent, template.TemplateID); err == nil {
			sectionHTML = processedHTML
		}
	}

	middleware.Success(c, "模板更新成功", gin.H{
		"id":           template.TemplateID,
		"template_id":  template.TemplateID,
		"name":         template.Name,
		"name_en":      tools.GetStringValue(template.NameEn),
		"description":  tools.GetStringValue(template.Description),
		"status":       template.Status,
		"is_public":    template.IsPublic,
		"is_featured":  template.IsFeatured,
		"sort_order":   template.SortOrder,
		"section_html": sectionHTML,
		"updated_at":   template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeletePublicTemplate 删除官方模板
func (h *AdminHandler) DeletePublicTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	hardDelete := c.DefaultQuery("hard_delete", "false") == "true"

	var template models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, fmt.Sprintf("模板 '%s' 不存在", templateID)))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	templateName := template.Name

	if hardDelete {
		// 物理删除
		if err := repository.DB.Delete(&template).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
			return
		}
		middleware.Success(c, "模板已物理删除", gin.H{
			"id":          templateID,
			"template_id": templateID,
			"name":        templateName,
			"hard_delete": true,
			"deleted_at":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	} else {
		// 软删除
		if err := repository.DB.Model(&template).Update("status", models.TemplateStatusDeleted).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
			return
		}
		// 重新查询以获取更新后的状态
		repository.DB.Where("template_id = ?", templateID).First(&template)
		middleware.Success(c, "模板已删除（软删除）", gin.H{
			"id":          templateID,
			"template_id": templateID,
			"name":        templateName,
			"hard_delete": false,
			"deleted_at":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// SaveImageExample 创建或更新图文生成示例
func (h *AdminHandler) SaveImageExample(c *gin.Context) {
	var req struct {
		ID          *int                   `json:"id"`
		Name        *string                `json:"name"`
		Prompt      *string                `json:"prompt"`
		CoverURL    *string                `json:"cover_url"`
		JsxCode     *string                `json:"jsx_code"`
		Tags        interface{}            `json:"tags"` // 可能是数组或JSON字符串
		ExtraData   map[string]interface{} `json:"extra_data"`
		NodeData    map[string]interface{} `json:"node_data"`
		Images      interface{}            `json:"images"` // 可能是数组或JSON字符串
		ProjectType *string                `json:"project_type"`
		SortOrder   *int                   `json:"sort_order"`
		IsVisible   *bool                  `json:"is_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	var example models.ImageExample
	var isNew bool

	if req.ID != nil && *req.ID > 0 {
		// 更新现有记录
		if err := repository.DB.Where("id = ?", *req.ID).First(&example).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(404, "示例不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}
		isNew = false
	} else {
		// 创建新记录
		if req.Name == nil || *req.Name == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "名称不能为空"))
			return
		}
		if req.Prompt == nil || *req.Prompt == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "提示词不能为空"))
			return
		}
		example = models.ImageExample{
			Name:        *req.Name,
			Prompt:      *req.Prompt,
			ProjectType: "other",
			IsVisible:   true,
			SortOrder:   0,
		}
		isNew = true
	}

	// 更新字段
	if req.Name != nil {
		example.Name = *req.Name
	}
	if req.Prompt != nil {
		example.Prompt = *req.Prompt
	}
	if req.CoverURL != nil {
		example.CoverURL = req.CoverURL
	}
	if req.SortOrder != nil {
		example.SortOrder = *req.SortOrder
	}
	if req.IsVisible != nil {
		example.IsVisible = *req.IsVisible
	}
	if req.ProjectType != nil {
		// 只有明确传入 "xiaohongshu" 时才使用，否则使用 "other"
		if *req.ProjectType == "xiaohongshu" {
			example.ProjectType = "xiaohongshu"
		} else {
			example.ProjectType = "other"
		}
	}

	// 处理 Tags（转换为JSON字符串）
	if req.Tags != nil {
		tagsJSON, err := json.Marshal(req.Tags)
		if err == nil {
			tagsStr := string(tagsJSON)
			example.Tags = &tagsStr
		}
	}

	// 处理 ExtraData（转换为JSON字符串）
	if req.ExtraData != nil {
		extraDataJSON, err := json.Marshal(req.ExtraData)
		if err == nil {
			extraDataStr := string(extraDataJSON)
			example.ExtraData = &extraDataStr
		}
	}

	// 保存主表
	if isNew {
		if err := repository.DB.Create(&example).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
			return
		}
	} else {
		if err := repository.DB.Save(&example).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
			return
		}
	}

	// 处理详情表（存储大字段）
	var detail models.ImageExampleDetail
	if err := repository.DB.Where("example_id = ?", example.ID).First(&detail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新详情记录
			detail = models.ImageExampleDetail{
				ExampleID: example.ID,
			}
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询详情失败: "+err.Error()))
			return
		}
	}

	// 更新详情字段
	if req.JsxCode != nil {
		detail.JsxCode = req.JsxCode
	}
	if req.NodeData != nil {
		nodeDataJSON, err := json.Marshal(req.NodeData)
		if err == nil {
			nodeDataStr := string(nodeDataJSON)
			detail.NodeData = &nodeDataStr
		}
	}
	if req.Images != nil {
		imagesJSON, err := json.Marshal(req.Images)
		if err == nil {
			imagesStr := string(imagesJSON)
			detail.Images = &imagesStr
		}
	}

	// 保存详情表
	if detail.ID == 0 {
		if err := repository.DB.Create(&detail).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建详情失败: "+err.Error()))
			return
		}
	} else {
		if err := repository.DB.Save(&detail).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新详情失败: "+err.Error()))
			return
		}
	}

	// 构建返回数据
	result := gin.H{
		"id":           example.ID,
		"name":         example.Name,
		"prompt":       example.Prompt,
		"cover_url":    tools.GetStringValue(example.CoverURL),
		"jsx_code":     tools.GetStringValue(detail.JsxCode),
		"tags":         example.Tags,
		"node_data":    detail.NodeData,
		"images":       detail.Images,
		"project_type": example.ProjectType,
		"sort_order":   example.SortOrder,
		"is_visible":   example.IsVisible,
		"extra_data":   example.ExtraData,
		"created_at":   example.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   example.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	middleware.Success(c, "success", result)
}

// GetImageExampleDetail 获取图文生成示例详情
func (h *AdminHandler) GetImageExampleDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	// 查询主表记录
	var example models.ImageExample
	if err := repository.DB.Where("id = ?", id).First(&example).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "未找到示例"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 关联查询详情表获取大字段数据
	var detail models.ImageExampleDetail
	repository.DB.Where("example_id = ?", id).First(&detail)

	// 构建返回数据
	result := gin.H{
		"id":           example.ID,
		"name":         example.Name,
		"prompt":       example.Prompt,
		"cover_url":    example.CoverURL,
		"jsx_code":     nil,
		"tags":         example.Tags,
		"node_data":    nil,
		"images":       nil,
		"project_type": example.ProjectType,
		"sort_order":   example.SortOrder,
		"is_visible":   example.IsVisible,
		"extra_data":   example.ExtraData,
		"created_at":   example.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   example.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 如果有详情记录，添加大字段数据
	if detail.ID > 0 {
		result["jsx_code"] = tools.GetStringValue(detail.JsxCode)
		result["node_data"] = detail.NodeData
		result["images"] = detail.Images
	}

	middleware.Success(c, "success", result)
}

// DeleteImageExample 删除图文生成示例
func (h *AdminHandler) DeleteImageExample(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	// 检查记录是否存在
	var example models.ImageExample
	if err := repository.DB.Where("id = ?", id).First(&example).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "示例不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 删除详情记录
	if err := repository.DB.Where("example_id = ?", id).Delete(&models.ImageExampleDetail{}).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除详情失败: "+err.Error()))
		return
	}

	// 删除主表记录
	if err := repository.DB.Delete(&example).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "success", gin.H{})
}

// GetTradeV2List 交易列表查询V2 - 支持复杂查询条件
func (h *AdminHandler) GetTradeV2List(c *gin.Context) {
	var req struct {
		Page           int      `json:"page" binding:"min=1"`
		PageSize       int      `json:"page_size" binding:"min=1"`
		StartDate      *string  `json:"start_date"` // YYYY-MM-DD
		EndDate        *string  `json:"end_date"`   // YYYY-MM-DD
		TradeNo        *string  `json:"trade_no"`
		PaymentStatus  []string `json:"payment_status"`
		TradeType      []string `json:"trade_type"`
		PaymentChannel []string `json:"payment_channel"`
		UserID         *string  `json:"user_id"`
		Username       *string  `json:"username"`
		Phone          *string  `json:"phone"`
		MinAmount      *float64 `json:"min_amount"`
		MaxAmount      *float64 `json:"max_amount"`
		OrderBy        *string  `json:"order_by"`
		OrderDirection *string  `json:"order_direction"` // asc | desc
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.OrderBy == nil || *req.OrderBy == "" {
		orderBy := "created_at"
		req.OrderBy = &orderBy
	}
	if req.OrderDirection == nil || *req.OrderDirection == "" {
		orderDir := "desc"
		req.OrderDirection = &orderDir
	}

	// 如果提供了交易号，直接根据交易号查询
	if req.TradeNo != nil && *req.TradeNo != "" {
		var trade models.Trade
		if err := repository.DB.Where("trade_no = ?", *req.TradeNo).First(&trade).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.Success(c, "success", gin.H{
					"total":     0,
					"items":     []gin.H{},
					"page":      req.Page,
					"page_size": req.PageSize,
				})
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}

		// 加载用户信息
		loadUserInfo(&trade)

		// 构建单个交易记录返回数据
		result := buildTradeResponse(&trade)
		middleware.Success(c, "success", gin.H{
			"total":     1,
			"items":     []gin.H{result},
			"page":      1,
			"page_size": 1,
		})
		return
	}

	// 构建复杂查询
	query := repository.DB.Model(&models.Trade{})

	// 时间范围筛选
	if req.StartDate != nil && *req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		query = query.Where("created_at >= ?", startTime)
	}

	if req.EndDate != nil && *req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		// 设置为当天23:59:59
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
		query = query.Where("created_at <= ?", endTime)
	}

	// 支付状态筛选（数组）
	if len(req.PaymentStatus) > 0 {
		query = query.Where("payment_status IN ?", req.PaymentStatus)
	}

	// 交易类型筛选（数组）
	if len(req.TradeType) > 0 {
		query = query.Where("trade_type IN ?", req.TradeType)
	}

	// 支付渠道筛选（数组）
	if len(req.PaymentChannel) > 0 {
		query = query.Where("payment_channel IN ?", req.PaymentChannel)
	}

	// 用户相关筛选
	if req.UserID != nil && *req.UserID != "" {
		query = query.Where("user_id = ?", *req.UserID)
	}

	// 用户名和手机号筛选需要 JOIN users 表
	needJoinUser := (req.Username != nil && *req.Username != "") || (req.Phone != nil && *req.Phone != "")
	if needJoinUser {
		query = query.Joins("LEFT JOIN users ON trades.user_id = users.user_id")
		if req.Username != nil && *req.Username != "" {
			query = query.Where("users.username LIKE ?", "%"+*req.Username+"%")
		}
		if req.Phone != nil && *req.Phone != "" {
			query = query.Where("users.phone LIKE ?", "%"+*req.Phone+"%")
		}
	}

	// 价格区间筛选
	if req.MinAmount != nil {
		query = query.Where("amount >= ?", *req.MinAmount)
	}

	if req.MaxAmount != nil {
		query = query.Where("amount <= ?", *req.MaxAmount)
	}

	// 价格区间验证
	if req.MinAmount != nil && req.MaxAmount != nil {
		if *req.MinAmount > *req.MaxAmount {
			middleware.HandleError(c, middleware.NewBusinessError(400, "最小金额不能大于最大金额"))
			return
		}
	}

	// 排序 - 如果使用了 JOIN，需要指定表名
	orderByField := *req.OrderBy
	if needJoinUser && orderByField != "" {
		// 检查是否是 trades 表的字段
		if orderByField == "created_at" || orderByField == "updated_at" || orderByField == "amount" {
			orderByField = "trades." + orderByField
		}
	}
	if *req.OrderDirection == "asc" {
		query = query.Order(orderByField + " ASC")
	} else {
		query = query.Order(orderByField + " DESC")
	}

	// 获取总数 - 如果使用了 JOIN，需要去重
	var total int64
	if needJoinUser {
		// 使用子查询来避免 JOIN 导致的重复计数
		countQuery := repository.DB.Model(&models.Trade{}).Where("user_id IN (SELECT user_id FROM users WHERE 1=1")
		if req.Username != nil && *req.Username != "" {
			countQuery = countQuery.Where("username LIKE ?", "%"+*req.Username+"%")
		}
		if req.Phone != nil && *req.Phone != "" {
			countQuery = countQuery.Where("phone LIKE ?", "%"+*req.Phone+"%")
		}
		countQuery = countQuery.Select("user_id)")

		// 重新构建基础查询条件（不包括用户筛选）
		baseQuery := repository.DB.Model(&models.Trade{})
		if req.StartDate != nil && *req.StartDate != "" {
			startTime, _ := time.Parse("2006-01-02", *req.StartDate)
			baseQuery = baseQuery.Where("created_at >= ?", startTime)
		}
		if req.EndDate != nil && *req.EndDate != "" {
			endTime, _ := time.Parse("2006-01-02", *req.EndDate)
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
			baseQuery = baseQuery.Where("created_at <= ?", endTime)
		}
		if len(req.PaymentStatus) > 0 {
			baseQuery = baseQuery.Where("payment_status IN ?", req.PaymentStatus)
		}
		if len(req.TradeType) > 0 {
			baseQuery = baseQuery.Where("trade_type IN ?", req.TradeType)
		}
		if len(req.PaymentChannel) > 0 {
			baseQuery = baseQuery.Where("payment_channel IN ?", req.PaymentChannel)
		}
		if req.UserID != nil && *req.UserID != "" {
			baseQuery = baseQuery.Where("user_id = ?", *req.UserID)
		}
		if req.MinAmount != nil {
			baseQuery = baseQuery.Where("amount >= ?", *req.MinAmount)
		}
		if req.MaxAmount != nil {
			baseQuery = baseQuery.Where("amount <= ?", *req.MaxAmount)
		}

		// 使用子查询获取符合条件的用户ID
		var userIDs []string
		userQuery := repository.DB.Model(&models.User{})
		if req.Username != nil && *req.Username != "" {
			userQuery = userQuery.Where("username LIKE ?", "%"+*req.Username+"%")
		}
		if req.Phone != nil && *req.Phone != "" {
			userQuery = userQuery.Where("phone LIKE ?", "%"+*req.Phone+"%")
		}
		userQuery.Select("user_id").Find(&userIDs)

		if len(userIDs) > 0 {
			baseQuery = baseQuery.Where("user_id IN ?", userIDs)
		} else {
			// 如果没有匹配的用户，返回空结果
			middleware.Success(c, "success", gin.H{
				"total":     0,
				"items":     []gin.H{},
				"page":      req.Page,
				"page_size": req.PageSize,
			})
			return
		}

		if err := baseQuery.Count(&total).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}
	} else {
		if err := query.Count(&total).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}
	}

	// 分页查询 - 如果使用了 JOIN，需要指定表名
	offset := (req.Page - 1) * req.PageSize
	var trades []models.Trade
	queryFind := query
	if needJoinUser {
		queryFind = queryFind.Select("trades.*")
	}
	if err := queryFind.Offset(offset).Limit(req.PageSize).Find(&trades).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 批量加载用户信息
	userIDs := make([]string, 0, len(trades))
	for _, trade := range trades {
		if trade.UserID != "" {
			userIDs = append(userIDs, trade.UserID)
		}
	}

	// 查询所有相关用户
	usersMap := make(map[string]models.User)
	if len(userIDs) > 0 {
		var users []models.User
		repository.DB.Where("user_id IN ?", userIDs).Find(&users)
		for _, user := range users {
			usersMap[user.UserID] = user
		}
	}

	// 构建返回数据
	items := make([]gin.H, 0, len(trades))
	for _, trade := range trades {
		// 设置用户信息
		if user, ok := usersMap[trade.UserID]; ok {
			trade.User = user
		}
		items = append(items, buildTradeResponse(&trade))
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     items,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// loadUserInfo 加载用户信息
func loadUserInfo(trade *models.Trade) {
	if trade.UserID != "" {
		var user models.User
		if err := repository.DB.Where("user_id = ?", trade.UserID).First(&user).Error; err == nil {
			trade.User = user
		}
	}
}

// buildTradeResponse 构建交易响应数据
func buildTradeResponse(trade *models.Trade) gin.H {
	result := gin.H{
		"id":              trade.ID,
		"trade_no":        trade.TradeNo,
		"user_id":         trade.UserID,
		"amount":          trade.Amount,
		"trade_type":      trade.TradeType,
		"payment_channel": trade.PaymentChannel,
		"payment_status":  trade.PaymentStatus,
		"payment_id":      trade.PaymentID,
		"title":           trade.Title,
		"created_at":      trade.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 如果有用户信息
	if trade.User.UserID != "" {
		result["username"] = trade.User.Username
		result["phone"] = trade.User.Phone
		result["nickname"] = trade.User.Nickname
		result["avatar"] = trade.User.Avatar
	}

	// 解析 metadata
	if trade.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*trade.Metadata), &metadata); err == nil {
			result["metadata"] = metadata
		}
	}

	if trade.PaidAt != nil {
		result["paid_at"] = trade.PaidAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		result["paid_at"] = nil
	}

	return result
}

// GetTradeDetail 获取交易详情
func (h *AdminHandler) GetTradeDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	var trade models.Trade
	if err := repository.DB.Where("id = ?", id).First(&trade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "交易不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 加载用户信息
	loadUserInfo(&trade)

	result := buildTradeResponse(&trade)
	middleware.Success(c, "success", result)
}

// CreateTrade 创建交易
func (h *AdminHandler) CreateTrade(c *gin.Context) {
	var req struct {
		TradeNo        string                 `json:"trade_no" binding:"required"`
		UserID         string                 `json:"user_id" binding:"required"`
		Amount         float64                `json:"amount" binding:"required,min=0"`
		TradeType      string                 `json:"trade_type" binding:"required"`
		PaymentChannel string                 `json:"payment_channel" binding:"required"`
		PaymentStatus  string                 `json:"payment_status"`
		Title          string                 `json:"title" binding:"required"`
		Metadata       map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认支付状态
	if req.PaymentStatus == "" {
		req.PaymentStatus = "pending"
	}

	// 转换 metadata 为 JSON 字符串
	var metadataStr *string
	if req.Metadata != nil {
		if metadataJSON, err := json.Marshal(req.Metadata); err == nil {
			metaStr := string(metadataJSON)
			metadataStr = &metaStr
		}
	}

	trade := models.Trade{
		TradeNo:        req.TradeNo,
		UserID:         req.UserID,
		Amount:         req.Amount,
		TradeType:      req.TradeType,
		PaymentChannel: req.PaymentChannel,
		PaymentStatus:  req.PaymentStatus,
		Title:          req.Title,
		Metadata:       metadataStr,
	}

	if err := repository.DB.Create(&trade).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	// 加载用户信息
	loadUserInfo(&trade)

	result := buildTradeResponse(&trade)
	middleware.Success(c, "success", result)
}

// UpdateTrade 更新交易
func (h *AdminHandler) UpdateTrade(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	var trade models.Trade
	if err := repository.DB.Where("id = ?", id).First(&trade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "交易不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var req struct {
		Amount         *float64               `json:"amount"`
		TradeType      *string                `json:"trade_type"`
		PaymentChannel *string                `json:"payment_channel"`
		PaymentStatus  *string                `json:"payment_status"`
		Title          *string                `json:"title"`
		Metadata       map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.TradeType != nil {
		updates["trade_type"] = *req.TradeType
	}
	if req.PaymentChannel != nil {
		updates["payment_channel"] = *req.PaymentChannel
	}
	if req.PaymentStatus != nil {
		updates["payment_status"] = *req.PaymentStatus
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Metadata != nil {
		if metadataJSON, err := json.Marshal(req.Metadata); err == nil {
			metaStr := string(metadataJSON)
			updates["metadata"] = &metaStr
		}
	}

	if len(updates) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	if err := repository.DB.Model(&trade).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 重新加载
	repository.DB.First(&trade, id)
	// 加载用户信息
	loadUserInfo(&trade)

	result := buildTradeResponse(&trade)
	middleware.Success(c, "success", result)
}

// DeleteTrade 删除交易
func (h *AdminHandler) DeleteTrade(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	var trade models.Trade
	if err := repository.DB.Where("id = ?", id).First(&trade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "交易不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	if err := repository.DB.Delete(&trade).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "success", gin.H{})
}

// RepairIncompleteTrades 修复未完成订单
func (h *AdminHandler) RepairIncompleteTrades(c *gin.Context) {
	var req struct {
		TradeNos []string `json:"trade_nos" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	fixList := make([]gin.H, 0)

	// 处理每个 trade_no
	for _, tradeNo := range req.TradeNos {
		var trade models.Trade
		if err := repository.DB.Where("trade_no = ?", tradeNo).First(&trade).Error; err != nil {
			continue
		}

		// 加载用户信息
		loadUserInfo(&trade)

		// 如果已经是成功状态，跳过
		if trade.PaymentStatus == "success" {
			continue
		}

		// 更新为成功状态
		now := time.Now()
		if err := repository.DB.Model(&trade).Updates(map[string]interface{}{
			"payment_status": "success",
			"paid_at":        &now,
		}).Error; err != nil {
			continue
		}

		// 解析 metadata 获取产品信息
		var metadata map[string]interface{}
		if trade.Metadata != nil {
			json.Unmarshal([]byte(*trade.Metadata), &metadata)
		}

		productID, _ := metadata["product_id"].(float64)
		if productID > 0 {
			var product models.Production
			if err := repository.DB.Where("id = ?", int(productID)).First(&product).Error; err == nil {
				fixList = append(fixList, gin.H{
					"user_id":         trade.UserID,
					"user_phone":      trade.User.Phone,
					"user_name":       trade.User.Username,
					"product_name":    product.Name,
					"product_type":    product.ProductType,
					"benefit_changes": gin.H{}, // 这里可以添加权益变更逻辑
				})
			}
		}
	}

	middleware.Success(c, "修复完成", fixList)
}

// GetUserDetail 获取用户详情（包含用户参数）
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	// 查询用户
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 查询用户参数，如果不存在则创建
	var userParam models.UserParameters
	if err := repository.DB.Where("user_id = ?", userID).First(&userParam).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认用户参数
			userParam = models.UserParameters{
				ParamID:             userID, // 使用 user_id 作为 param_id
				UserID:              userID,
				EnableHeadInfo:      false,
				EnableKnowledgeBase: false,
				DefaultTheme:        "countryside",
				IsGzhBind:           false,
				IsWechatAuthorized:  false,
				HasAuthReminded:     false,
				PublishTarget:       0,
				StorageQuota:        314572800, // 300MB
			}
			if err := repository.DB.Create(&userParam).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "创建用户参数失败: "+err.Error()))
				return
			}
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户参数失败: "+err.Error()))
			return
		}
	}

	// 构建返回数据
	result := gin.H{
		"user_id":           user.UserID,
		"nickname":          user.Nickname,
		"avatar":            user.Avatar,
		"username":          user.Username,
		"phone":             user.Phone,
		"email":             user.Email,
		"openid":            user.OpenID,
		"credits":           user.Credits,
		"is_active":         user.IsActive,
		"total_consumption": user.TotalConsumption,
		"vip_level":         user.VipLevel,
		"role":              user.Role,
		"status":            user.Status,
		"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"user_parameter": gin.H{
			"param_id":              userParam.ParamID,
			"enable_head_info":      userParam.EnableHeadInfo,
			"enable_knowledge_base": userParam.EnableKnowledgeBase,
			"default_theme":         userParam.DefaultTheme,
			"is_wechat_authorized":  userParam.IsWechatAuthorized,
			"is_gzh_bind":           userParam.IsGzhBind,
			"has_auth_reminded":     userParam.HasAuthReminded,
			"publish_target":        userParam.PublishTarget,
			"qrcode_data":           userParam.QrcodeData,
			"created_time":          userParam.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_time":          userParam.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	middleware.Success(c, "success", result)
}

// GetVersionList 获取版本列表（CRUD 列表接口）
func (h *AdminHandler) GetVersionList(c *gin.Context) {
	// 使用 CRUD handler 的 List 方法
	versionCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.Version{},
		SearchFields:   []string{"version"},
		DefaultOrderBy: "date",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	versionCRUD.List(c)
}

// GetVersionItems 获取版本列表（用于前端渲染版本更新记录页）
func (h *AdminHandler) GetVersionItems(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"min=1"`
		PageSize int `form:"page_size" binding:"min=1"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 获取总数
	var total int64
	if err := repository.DB.Model(&models.Version{}).Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询，按日期和ID降序排序
	offset := (req.Page - 1) * req.PageSize
	var versions []models.Version
	if err := repository.DB.Order("date DESC, id DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&versions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(versions))
	for _, item := range versions {
		// 解析 highlights（可能是 JSON 字符串或已经是数组）
		var highlights interface{}
		if err := json.Unmarshal([]byte(item.Highlights), &highlights); err != nil {
			// 如果解析失败，尝试作为字符串处理
			highlights = item.Highlights
		}

		versionType := ""
		if item.Type != nil {
			versionType = string(*item.Type)
		}

		result = append(result, gin.H{
			"id":         item.ID,
			"version":    item.Version,
			"date":       item.Date.Format("2006-01-02"),
			"title":      item.Title,
			"type":       versionType,
			"highlights": highlights,
			"created_at": item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      result,
	})
}

// CreateVersion 创建版本
func (h *AdminHandler) CreateVersion(c *gin.Context) {
	var req struct {
		Version    string   `json:"version" binding:"required"`
		Date       string   `json:"date" binding:"required"` // YYYY-MM-DD
		Title      string   `json:"title" binding:"required"`
		Type       *string  `json:"type"` // major / minor / patch
		Highlights []string `json:"highlights" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，应为 YYYY-MM-DD"))
		return
	}

	// 转换 highlights 为 JSON 字符串
	highlightsJSON, err := json.Marshal(req.Highlights)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "高亮信息格式错误"))
		return
	}

	// 转换版本类型
	var versionType *models.VersionType
	if req.Type != nil && *req.Type != "" {
		vt := models.VersionType(*req.Type)
		versionType = &vt
	}

	version := models.Version{
		Version:    req.Version,
		Date:       date,
		Title:      req.Title,
		Highlights: string(highlightsJSON),
		Type:       versionType,
	}

	if err := repository.DB.Create(&version).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	middleware.Success(c, "success", gin.H{})
}

// GetMarketingActivities 获取营销活动列表
func (h *AdminHandler) GetMarketingActivities(c *gin.Context) {
	var req struct {
		Status *int `form:"status"` // 0-待开始, 1-进行中, 2-已结束
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	query := repository.DB.Model(&models.MarketingActivityPlan{})
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var activities []models.MarketingActivityPlan
	if err := query.Order("created_at DESC").Find(&activities).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(activities))
	for _, activity := range activities {
		var config map[string]interface{}
		if activity.Config != "" {
			json.Unmarshal([]byte(activity.Config), &config)
		}

		result = append(result, gin.H{
			"activity_id": activity.ActivityID,
			"name":        activity.Name,
			"description": activity.Description,
			"start_time":  activity.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			"end_time":    activity.EndTime.Format("2006-01-02T15:04:05Z07:00"),
			"status":      activity.Status,
			"is_visible":  activity.IsVisible,
			"config":      config,
			"created_at":  activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":  activity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "获取活动列表成功", result)
}

// GetMarketingActivityDetail 获取营销活动详情
func (h *AdminHandler) GetMarketingActivityDetail(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var config map[string]interface{}
	if activity.Config != "" {
		json.Unmarshal([]byte(activity.Config), &config)
	}

	result := gin.H{
		"activity_id": activity.ActivityID,
		"name":        activity.Name,
		"description": activity.Description,
		"start_time":  activity.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		"end_time":    activity.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		"config":      config,
		"status":      activity.Status,
		"is_visible":  activity.IsVisible,
		"created_at":  activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":  activity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	middleware.Success(c, "获取活动详情成功", result)
}

// parseTimeString 解析时间字符串，支持多种格式
func parseTimeString(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,                  // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,              // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05",         // 不带时区
		"2006-01-02T15:04:05.000Z",    // 带毫秒和Z
		"2006-01-02T15:04:05.000000Z", // 带微秒和Z
		"2006-01-02T15:04:05Z",        // 简单Z格式
		"2006-01-02 15:04:05",         // 空格分隔
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// CreateMarketingActivity 创建营销活动
func (h *AdminHandler) CreateMarketingActivity(c *gin.Context) {
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description" binding:"required"`
		StartTime   string                 `json:"start_time" binding:"required"` // ISO 8601 格式
		EndTime     string                 `json:"end_time" binding:"required"`   // ISO 8601 格式
		Config      map[string]interface{} `json:"config" binding:"required"`
		IsVisible   *bool                  `json:"is_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 解析时间

	startTime, err := parseTimeString(req.StartTime)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误: "+err.Error()))
		return
	}

	endTime, err := parseTimeString(req.EndTime)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误: "+err.Error()))
		return
	}

	if startTime.After(endTime) || startTime.Equal(endTime) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间必须大于开始时间"))
		return
	}

	// 检查时间段内是否已有其他活动
	var existingActivity models.MarketingActivityPlan
	if err := repository.DB.Where("start_time < ? AND end_time > ? AND status IN ?", endTime, startTime, []int16{0, 1}).
		First(&existingActivity).Error; err == nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "该时间段内已存在其他活动"))
		return
	}

	// 转换 config 为 JSON 字符串
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "配置格式错误"))
		return
	}

	// 根据时间自动设置状态
	now := time.Now()
	status := int16(models.ActivityStatusPending)
	if startTime.Before(now) && endTime.After(now) {
		status = int16(models.ActivityStatusOngoing)
	}

	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	// 生成活动ID
	activityID := fmt.Sprintf("ACT%d", time.Now().Unix())

	activity := models.MarketingActivityPlan{
		ActivityID:  activityID,
		Name:        req.Name,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Config:      string(configJSON),
		Status:      status,
		IsVisible:   isVisible,
	}

	if err := repository.DB.Create(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动创建成功", gin.H{
		"activity_id": activity.ActivityID,
		"status":      activity.Status,
	})
}

// UpdateMarketingActivity 更新营销活动
func (h *AdminHandler) UpdateMarketingActivity(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var req struct {
		Name        *string                `json:"name"`
		Description *string                `json:"description"`
		StartTime   *string                `json:"start_time"`
		EndTime     *string                `json:"end_time"`
		Config      map[string]interface{} `json:"config"`
		IsVisible   *bool                  `json:"is_visible"`
		Status      *int16                 `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	updates := make(map[string]interface{})
	hasUpdate := false

	if req.Name != nil {
		updates["name"] = *req.Name
		hasUpdate = true
	}

	if req.Description != nil {
		updates["description"] = *req.Description
		hasUpdate = true
	}

	startTime := activity.StartTime
	endTime := activity.EndTime

	if req.StartTime != nil {
		st, err := parseTimeString(*req.StartTime)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误: "+err.Error()))
			return
		}
		startTime = st
		updates["start_time"] = startTime
		hasUpdate = true
	}

	if req.EndTime != nil {
		et, err := parseTimeString(*req.EndTime)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误: "+err.Error()))
			return
		}
		endTime = et
		updates["end_time"] = endTime
		hasUpdate = true
	}

	if req.StartTime != nil || req.EndTime != nil {
		if startTime.After(endTime) || startTime.Equal(endTime) {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间必须大于开始时间"))
			return
		}

		// 检查时间冲突
		var existingActivity models.MarketingActivityPlan
		if err := repository.DB.Where("activity_id != ? AND start_time < ? AND end_time > ? AND status IN ?", activityID, endTime, startTime, []int16{0, 1}).
			First(&existingActivity).Error; err == nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "该时间段内已存在其他活动"))
			return
		}

		// 根据时间自动更新状态
		if req.Status == nil {
			now := time.Now()
			if now.Before(startTime) {
				updates["status"] = int16(models.ActivityStatusPending)
			} else if now.After(endTime) {
				updates["status"] = int16(models.ActivityStatusEnded)
			} else {
				updates["status"] = int16(models.ActivityStatusOngoing)
			}
		}
	}

	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "配置格式错误"))
			return
		}
		updates["config"] = string(configJSON)
		hasUpdate = true
	}

	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
		hasUpdate = true
	}

	if req.Status != nil {
		updates["status"] = *req.Status
		hasUpdate = true
	}

	if !hasUpdate {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	if err := repository.DB.Model(&activity).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动更新成功", gin.H{})
}

// DeleteMarketingActivity 删除营销活动
func (h *AdminHandler) DeleteMarketingActivity(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	if err := repository.DB.Delete(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动删除成功", gin.H{})
}

// PurchaseActivityProduct 购买活动产品
func (h *AdminHandler) PurchaseActivityProduct(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var req struct {
		ProductID int `json:"product_id" binding:"required"`
		Quantity  int `json:"quantity" binding:"min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认数量
	if req.Quantity == 0 {
		req.Quantity = 1
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 检查活动状态
	if activity.Status != int16(models.ActivityStatusOngoing) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动未开始或已结束"))
		return
	}

	// 解析配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(activity.Config), &config); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "活动配置解析失败"))
		return
	}

	products, ok := config["products"].([]interface{})
	if !ok || len(products) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动没有产品"))
		return
	}

	// 查找匹配的产品
	var targetProduct map[string]interface{}
	for _, p := range products {
		product, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		pid, ok := product["product_id"].(float64)
		if ok && int(pid) == req.ProductID {
			targetProduct = product
			break
		}
	}

	if targetProduct == nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "产品不匹配"))
		return
	}

	// 检查库存
	stock, ok := targetProduct["stock"].(float64)
	if !ok || int(stock) < req.Quantity {
		middleware.HandleError(c, middleware.NewBusinessError(400, "库存不足"))
		return
	}

	// 检查每人限购
	if limitPerUser, ok := targetProduct["limit_per_user"].(float64); ok {
		if req.Quantity > int(limitPerUser) {
			middleware.HandleError(c, middleware.NewBusinessError(400, fmt.Sprintf("每人限购%d件", int(limitPerUser))))
			return
		}
	}

	// 更新库存
	targetProduct["stock"] = stock - float64(req.Quantity)

	// 保存更新后的配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新配置失败"))
		return
	}

	if err := repository.DB.Model(&activity).Update("config", string(configJSON)).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新库存失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动产品购买成功, 库存已更新", gin.H{
		"product": targetProduct,
	})
}

// GetCurrentActivity 获取当前正在进行的活动
func (h *AdminHandler) GetCurrentActivity(c *gin.Context) {
	now := time.Now()
	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("start_time <= ? AND end_time > ? AND status = ? AND is_visible = ?", now, now, int16(models.ActivityStatusOngoing), true).
		First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "当前没有进行中的活动"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 解析配置
	var config map[string]interface{}
	if activity.Config != "" {
		json.Unmarshal([]byte(activity.Config), &config)
	}

	result := gin.H{
		"activity_id": activity.ActivityID,
		"name":        activity.Name,
		"description": activity.Description,
		"start_time":  activity.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		"end_time":    activity.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		"config":      config,
	}

	middleware.Success(c, "获取当前活动成功", result)
}

// GetActivationCodeList 获取激活码列表
func (h *AdminHandler) GetActivationCodeList(c *gin.Context) {
	var req struct {
		Page           int    `form:"page" binding:"min=1"`
		PageSize       int    `form:"page_size" binding:"min=1"`
		Search         string `form:"search"`
		IsUsed         *bool  `form:"is_used"`
		CardType       string `form:"card_type"`
		OrderBy        string `form:"order_by"`
		OrderDirection string `form:"order_direction"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}

	query := repository.DB.Model(&models.ActivationCode{})

	// 筛选条件
	if req.IsUsed != nil {
		query = query.Where("is_used = ?", *req.IsUsed)
	}
	if req.CardType != "" {
		query = query.Where("card_type = ?", req.CardType)
	}
	if req.Search != "" {
		query = query.Where("code LIKE ?", "%"+req.Search+"%")
	}

	// 排序
	orderBy := "created_at"
	if req.OrderBy != "" {
		orderBy = req.OrderBy
	}
	if req.OrderDirection == "desc" {
		query = query.Order(orderBy + " DESC")
	} else {
		query = query.Order(orderBy + " ASC")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var codes []models.ActivationCode
	if err := query.Offset(offset).Limit(req.PageSize).Find(&codes).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	items := make([]gin.H, 0, len(codes))
	for _, code := range codes {
		// 获取产品名称
		var productName *string
		if code.CardType == string(models.CardTypeMembership) {
			var product models.Production
			if err := repository.DB.Where("id = ?", code.ProductID).First(&product).Error; err == nil {
				name := product.Name
				productName = &name
			}
		} else if code.CardType == string(models.CardTypeCredits) {
			var product models.CreditProduct
			if err := repository.DB.Where("id = ?", code.ProductID).First(&product).Error; err == nil {
				productName = product.Name
			}
		}

		// 获取使用用户手机号
		var userPhone *string
		if code.UsedByID != nil {
			var user models.User
			if err := repository.DB.Where("user_id = ?", *code.UsedByID).First(&user).Error; err == nil {
				userPhone = user.Phone
			}
		}

		items = append(items, gin.H{
			"id":           code.ID,
			"code":         code.Code,
			"card_type":    code.CardType,
			"product_id":   code.ProductID,
			"product_name": productName,
			"is_used":      code.IsUsed,
			"remark":       code.Remark,
			"used_by":      userPhone,
			"created_at":   code.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "获取兑换码列表成功", gin.H{
		"total":           total,
		"items":           items,
		"page":            req.Page,
		"page_size":       req.PageSize,
		"order_direction": req.OrderDirection,
		"search":          req.Search,
	})
}

// GetCreditProduct2List 获取积分产品列表
func (h *AdminHandler) GetCreditProduct2List(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"min=1"`
		PageSize int `form:"page_size" binding:"min=1"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 获取总数
	var total int64
	if err := repository.DB.Model(&models.CreditProduct{}).Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var products []models.CreditProduct
	if err := repository.DB.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&products).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(products))
	for _, item := range products {
		result = append(result, gin.H{
			"id":         item.ID,
			"name":       item.Name,
			"credits":    item.Credits,
			"price":      item.Price,
			"status":     item.Status,
			"created_at": item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetCreditRechargeOrderList 获取积分充值订单列表
func (h *AdminHandler) GetCreditRechargeOrderList(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"min=1"`
		PageSize int `form:"page_size" binding:"min=1"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 获取总数
	var total int64
	if err := repository.DB.Model(&models.CreditRechargeOrder{}).Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var orders []models.CreditRechargeOrder
	if err := repository.DB.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&orders).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(orders))
	for _, order := range orders {
		// 加载关联数据
		var product models.CreditProduct
		repository.DB.Where("id = ?", order.ProductID).First(&product)

		var trade models.Trade
		repository.DB.Where("id = ?", order.TradeID).First(&trade)

		var user models.User
		repository.DB.Where("user_id = ?", order.UserID).First(&user)

		result = append(result, gin.H{
			"id":           order.ID,
			"user_id":      order.UserID,
			"user_phone":   user.Phone,
			"product_id":   order.ProductID,
			"product_name": product.Name,
			"trade_id":     order.TradeID,
			"created_at":   order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "获取成功", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetCreditRechargeOrderSummary 获取积分充值订单汇总
func (h *AdminHandler) GetCreditRechargeOrderSummary(c *gin.Context) {
	// 总订单数
	var totalOrders int64
	repository.DB.Model(&models.CreditRechargeOrder{}).Count(&totalOrders)

	// 总充值金额（通过关联的 Trade 表）
	var totalAmount float64
	repository.DB.Table("credit_recharge_orders").
		Select("COALESCE(SUM(trades.amount), 0)").
		Joins("LEFT JOIN trades ON credit_recharge_orders.trade_id = trades.id").
		Scan(&totalAmount)

	// 总充值积分（通过关联的 CreditProduct 表）
	var totalCredits int64
	repository.DB.Table("credit_recharge_orders").
		Select("COALESCE(SUM(credit_products.credits), 0)").
		Joins("LEFT JOIN credit_products ON credit_recharge_orders.product_id = credit_products.id").
		Scan(&totalCredits)

	// 今日订单数
	today := time.Now().Format("2006-01-02")
	var todayOrders int64
	repository.DB.Model(&models.CreditRechargeOrder{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayOrders)

	// 今日充值金额
	var todayAmount float64
	repository.DB.Table("credit_recharge_orders").
		Select("COALESCE(SUM(trades.amount), 0)").
		Joins("LEFT JOIN trades ON credit_recharge_orders.trade_id = trades.id").
		Where("DATE(credit_recharge_orders.created_at) = ?", today).
		Scan(&todayAmount)

	// 今日充值积分
	var todayCredits int64
	repository.DB.Table("credit_recharge_orders").
		Select("COALESCE(SUM(credit_products.credits), 0)").
		Joins("LEFT JOIN credit_products ON credit_recharge_orders.product_id = credit_products.id").
		Where("DATE(credit_recharge_orders.created_at) = ?", today).
		Scan(&todayCredits)

	middleware.Success(c, "获取成功", gin.H{
		"total_orders":  totalOrders,
		"total_amount":  totalAmount,
		"total_credits": totalCredits,
		"today_orders":  todayOrders,
		"today_amount":  todayAmount,
		"today_credits": todayCredits,
	})
}

// GetCreditRecordsStatsOverview 获取积分记录概览统计
func (h *AdminHandler) GetCreditRecordsStatsOverview(c *gin.Context) {
	var req struct {
		StartDate    string `form:"start_date"`
		EndDate      string `form:"end_date"`
		ExcludeAdmin bool   `form:"exclude_admin"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if !req.ExcludeAdmin {
		req.ExcludeAdmin = true
	}

	// 构建基础查询
	query := repository.DB.Model(&models.CreditRecord{})

	// 排除管理员
	if req.ExcludeAdmin {
		query = query.Joins("LEFT JOIN users ON credit_records.user_id = users.user_id").
			Where("users.role != ? OR users.role IS NULL", 1) // 假设 1 是管理员角色
	}

	// 日期范围
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query = query.Where("credit_records.created_at >= ?", startTime)
		}
	}
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// 设置为当天的最后一刻
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("credit_records.created_at <= ?", endTime)
		}
	}

	// 总记录数
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 按类型统计
	typeStats := make(map[string]gin.H)
	recordTypes := []models.CreditRecordType{
		models.CreditRecharge,
		models.CreditConsumption,
		models.CreditReward,
		models.CreditExpired,
		models.CreditRefund,
	}

	for _, recordType := range recordTypes {
		typeQuery := query.Where("record_type = ?", int16(recordType))

		var count int64
		typeQuery.Count(&count)

		var totalCredits int64
		typeQuery.Select("COALESCE(SUM(credits), 0)").
			Scan(&totalCredits)

		typeName := ""
		switch recordType {
		case models.CreditRecharge:
			typeName = "RECHARGE"
		case models.CreditConsumption:
			typeName = "CONSUME"
		case models.CreditReward:
			typeName = "REWARD"
		case models.CreditExpired:
			typeName = "EXPIRED"
		case models.CreditRefund:
			typeName = "REFUND"
		}

		typeStats[typeName] = gin.H{
			"count":         count,
			"total_credits": totalCredits,
		}
	}

	// 有service_code的记录数
	var recordsWithService int64
	query.Where("service_code IS NOT NULL AND record_type = ?", int16(models.CreditConsumption)).
		Count(&recordsWithService)

	// 唯一用户数
	var uniqueUserCount int64
	repository.DB.Table("(?) as subquery", query).
		Select("COUNT(DISTINCT user_id)").
		Scan(&uniqueUserCount)

	// 唯一服务数
	var uniqueServiceCount int64
	query.Where("service_code IS NOT NULL").
		Select("COUNT(DISTINCT service_code)").
		Scan(&uniqueServiceCount)

	middleware.Success(c, "获取成功", gin.H{
		"total_records":             totalRecords,
		"type_stats":                typeStats,
		"records_with_service_code": recordsWithService,
		"unique_users":              uniqueUserCount,
		"unique_services":           uniqueServiceCount,
		"date_range": gin.H{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
	})
}

// GetCreditServicePriceList 获取积分服务价格列表
func (h *AdminHandler) GetCreditServicePriceList(c *gin.Context) {
	// 使用 CRUD handler 的 List 方法
	creditServicePriceCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &models.CreditServicePrice{},
		SearchFields:   []string{"service_code"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	creditServicePriceCRUD.List(c)
}

// generateActivationCode 生成兑换码
func generateActivationCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除容易混淆的字符：0, O, 1, I
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// CreateActivationCodes 创建兑换码
func (h *AdminHandler) CreateActivationCodes(c *gin.Context) {
	var req struct {
		CardType  string `json:"card_type" binding:"required"` // "membership" 或 "credits"
		ProductID int    `json:"product_id" binding:"required"`
		Count     int    `json:"count" binding:"min=1,max=1000"`
		Remark    string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 1
	}

	// 验证卡片类型
	if req.CardType != string(models.CardTypeMembership) && req.CardType != string(models.CardTypeCredits) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "无效的卡片类型"))
		return
	}

	// 验证产品是否存在
	var productName *string
	if req.CardType == string(models.CardTypeMembership) {
		var product models.Production
		if err := repository.DB.Where("id = ?", req.ProductID).First(&product).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(400, "会员产品不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询产品失败: "+err.Error()))
			return
		}
		productName = &product.Name
	} else if req.CardType == string(models.CardTypeCredits) {
		var product models.CreditProduct
		if err := repository.DB.Where("id = ?", req.ProductID).First(&product).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(400, "积分产品不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询产品失败: "+err.Error()))
			return
		}
		productName = product.Name
	}

	// 批量创建兑换码
	createdCodes := make([]gin.H, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		// 生成唯一的兑换码
		var code string
		attempts := 0
		for attempts < 10 {
			code = generateActivationCode(10)
			// 检查是否已存在
			var existing models.ActivationCode
			if err := repository.DB.Where("code = ?", code).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					break // 代码不存在，可以使用
				}
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询兑换码失败: "+err.Error()))
				return
			}
			attempts++
		}
		if attempts >= 10 {
			middleware.HandleError(c, middleware.NewBusinessError(500, "生成兑换码失败，请重试"))
			return
		}

		// 创建兑换码记录
		remark := req.Remark
		activationCode := models.ActivationCode{
			Code:      code,
			CardType:  req.CardType,
			ProductID: req.ProductID,
			IsUsed:    false,
		}
		if remark != "" {
			activationCode.Remark = &remark
		}

		if err := repository.DB.Create(&activationCode).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建兑换码失败: "+err.Error()))
			return
		}

		createdCodes = append(createdCodes, gin.H{
			"id":         activationCode.ID,
			"code":       activationCode.Code,
			"card_type":  activationCode.CardType,
			"product_id": activationCode.ProductID,
			"remark":     activationCode.Remark,
		})
	}

	middleware.Success(c, fmt.Sprintf("成功创建 %d 个兑换码", len(createdCodes)), gin.H{
		"count":        len(createdCodes),
		"codes":        createdCodes,
		"product_name": productName,
	})
}

// GetCommissionOverview 获取佣金概览统计
func (h *AdminHandler) GetCommissionOverview(c *gin.Context) {
	// 总佣金记录数
	var totalRecords int64
	repository.DB.Model(&models.CommissionRecord{}).Count(&totalRecords)

	// 按状态统计
	statusStats := make(map[string]gin.H)
	statuses := []models.CommissionStatus{
		models.CommissionPending,
		models.CommissionIssued,
		models.CommissionWithdrawn,
		models.CommissionRejected,
		models.CommissionApplying,
	}

	for _, status := range statuses {
		var count int64
		var totalAmount float64

		repository.DB.Model(&models.CommissionRecord{}).
			Where("status = ?", int(status)).
			Count(&count)

		repository.DB.Model(&models.CommissionRecord{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("status = ?", int(status)).
			Scan(&totalAmount)

		statusName := ""
		switch status {
		case models.CommissionPending:
			statusName = "PENDING"
		case models.CommissionIssued:
			statusName = "ISSUED"
		case models.CommissionWithdrawn:
			statusName = "WITHDRAWN"
		case models.CommissionRejected:
			statusName = "REJECTED"
		case models.CommissionApplying:
			statusName = "APPLYING"
		}

		statusStats[statusName] = gin.H{
			"count":  count,
			"amount": totalAmount,
		}
	}

	// 总佣金金额（所有状态）
	var totalAmount float64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)

	// 已发放和已提现的佣金总额
	var issuedAmount float64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status IN ?", []int{int(models.CommissionIssued), int(models.CommissionWithdrawn)}).
		Scan(&issuedAmount)

	// 待发放佣金总额
	var pendingAmount float64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status = ?", int(models.CommissionPending)).
		Scan(&pendingAmount)

	// 唯一用户数
	var uniqueUsers int64
	repository.DB.Model(&models.CommissionRecord{}).
		Select("COUNT(DISTINCT user_id)").
		Scan(&uniqueUsers)

	middleware.Success(c, "获取成功", gin.H{
		"total_records":  totalRecords,
		"total_amount":   totalAmount,
		"issued_amount":  issuedAmount,
		"pending_amount": pendingAmount,
		"unique_users":   uniqueUsers,
		"status_stats":   statusStats,
	})
}

// GetCommissionList 获取佣金列表
func (h *AdminHandler) GetCommissionList(c *gin.Context) {
	var req struct {
		Page      int    `form:"page" binding:"min=1"`
		PageSize  int    `form:"page_size" binding:"min=1"`
		Status    *int   `form:"status"`
		UserID    string `form:"user_id"`
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建查询
	query := repository.DB.Model(&models.CommissionRecord{})

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 用户筛选
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 日期范围筛选
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// 设置为当天的最后一刻
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", endTime)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var records []models.CommissionRecord
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(records))
	for _, record := range records {
		// 加载用户信息
		var user models.User
		repository.DB.Where("user_id = ?", record.UserID).First(&user)

		// 加载邀请关系信息
		var relation models.InvitationRelation
		var invitee models.User
		if err := repository.DB.Where("id = ?", record.RelationID).First(&relation).Error; err == nil {
			repository.DB.Where("user_id = ?", relation.InviteeID).First(&invitee)
		}

		// 加载订单信息（如果有）
		var order models.Trade
		if record.OrderID != nil {
			repository.DB.Where("id = ?", *record.OrderID).First(&order)
		}

		item := gin.H{
			"id":      record.ID,
			"user_id": record.UserID,
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"username": user.Username,
				"phone":    user.Phone,
			},
			"amount":      record.Amount,
			"status":      int(record.Status),
			"status_text": record.Status.String(),
			"description": record.Description,
			"created_at":  record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if record.IssueTime != nil {
			item["issue_time"] = record.IssueTime.Format("2006-01-02T15:04:05Z07:00")
		}
		if record.WithdrawalTime != nil {
			item["withdrawal_time"] = record.WithdrawalTime.Format("2006-01-02T15:04:05Z07:00")
		}

		if invitee.UserID != "" {
			item["invitee"] = gin.H{
				"user_id":  invitee.UserID,
				"nickname": invitee.Nickname,
				"username": invitee.Username,
			}
		}

		if order.ID != 0 {
			item["order"] = gin.H{
				"trade_no": order.TradeNo,
				"amount":   order.Amount,
				"title":    order.Title,
			}
		}

		result = append(result, item)
	}

	middleware.Success(c, "获取成功", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      result,
	})
}

// GetInvitationRelationOverview 获取邀请关系概览统计
func (h *AdminHandler) GetInvitationRelationOverview(c *gin.Context) {
	// 总邀请关系数
	var totalRelations int64
	repository.DB.Model(&models.InvitationRelation{}).Count(&totalRelations)

	// 唯一邀请人数
	var uniqueInviters int64
	repository.DB.Model(&models.InvitationRelation{}).
		Select("COUNT(DISTINCT inviter_id)").
		Scan(&uniqueInviters)

	// 唯一被邀请人数
	var uniqueInvitees int64
	repository.DB.Model(&models.InvitationRelation{}).
		Select("COUNT(DISTINCT invitee_id)").
		Scan(&uniqueInvitees)

	// 今日新增邀请关系
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var todayCount int64
	repository.DB.Model(&models.InvitationRelation{}).
		Where("created_at >= ?", todayStart).
		Count(&todayCount)

	// 本周新增邀请关系
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	if now.Weekday() == time.Sunday {
		weekStart = todayStart.AddDate(0, 0, -6)
	}
	var weekCount int64
	repository.DB.Model(&models.InvitationRelation{}).
		Where("created_at >= ?", weekStart).
		Count(&weekCount)

	// 本月新增邀请关系
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthCount int64
	repository.DB.Model(&models.InvitationRelation{}).
		Where("created_at >= ?", monthStart).
		Count(&monthCount)

	// 平均每个邀请人邀请的人数
	var avgInvitationsPerInviter float64
	if uniqueInviters > 0 {
		avgInvitationsPerInviter = float64(totalRelations) / float64(uniqueInviters)
	}

	// 邀请人数最多的前10个邀请人
	type TopInviter struct {
		InviterID string `gorm:"column:inviter_id"`
		Count     int64  `gorm:"column:count"`
	}
	var topInviters []TopInviter
	repository.DB.Model(&models.InvitationRelation{}).
		Select("inviter_id, COUNT(*) as count").
		Group("inviter_id").
		Order("count DESC").
		Limit(10).
		Scan(&topInviters)

	// 构建返回数据
	topInvitersData := make([]gin.H, 0, len(topInviters))
	for _, inviter := range topInviters {
		// 获取邀请人信息
		var user models.User
		repository.DB.Where("user_id = ?", inviter.InviterID).First(&user)

		topInvitersData = append(topInvitersData, gin.H{
			"inviter_id": inviter.InviterID,
			"count":      inviter.Count,
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"username": user.Username,
			},
		})
	}

	middleware.Success(c, "获取成功", gin.H{
		"total_relations":             totalRelations,
		"unique_inviters":             uniqueInviters,
		"unique_invitees":             uniqueInvitees,
		"today_count":                 todayCount,
		"week_count":                  weekCount,
		"month_count":                 monthCount,
		"avg_invitations_per_inviter": avgInvitationsPerInviter,
		"top_inviters":                topInvitersData,
	})
}

// GetNotificationList 获取通知列表
func (h *AdminHandler) GetNotificationList(c *gin.Context) {
	var req struct {
		Page   int    `form:"page"`
		Limit  int    `form:"limit"`
		Status string `form:"status"`
		Type   string `form:"type"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	// 构建查询
	query := repository.DB.Model(&models.SystemNotification{})

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 类型筛选
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.Limit
	var notifications []models.SystemNotification
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&notifications).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(notifications))
	for _, notification := range notifications {
		item := gin.H{
			"notification_id": notification.NotificationID,
			"user_id":         notification.UserID,
			"title":           notification.Title,
			"content":         notification.Content,
			"type":            notification.Type,
			"link":            notification.Link,
			"is_important":    notification.IsImportant,
			"status":          notification.Status,
			"created_at":      notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      notification.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if notification.ReadTime != nil {
			item["read_time"] = notification.ReadTime.Format("2006-01-02T15:04:05Z07:00")
		}
		if notification.ExpireTime != nil {
			item["expire_time"] = notification.ExpireTime.Format("2006-01-02T15:04:05Z07:00")
		}

		result = append(result, item)
	}

	middleware.Success(c, "获取通知列表成功", gin.H{
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
		"list":  result,
	})
}

// GetNotificationUserRecordList 获取用户通知记录列表
func (h *AdminHandler) GetNotificationUserRecordList(c *gin.Context) {
	var req struct {
		Page           int    `form:"page"`
		PageSize       int    `form:"page_size"`
		OrderBy        string `form:"order_by"`
		OrderDirection string `form:"order_direction"`
		Relations      string `form:"relations"`
		RelationDepth  int    `form:"relation_depth"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}

	// 构建查询
	query := repository.DB.Model(&models.NotificationUserRecord{})

	// 排序
	orderField := req.OrderBy
	if req.OrderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var records []models.NotificationUserRecord
	if err := query.Offset(offset).Limit(req.PageSize).Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 加载关联数据
	if req.Relations != "" {
		relations := strings.Split(req.Relations, ",")
		loadUser := false
		loadNotification := false
		for _, relation := range relations {
			relation = strings.TrimSpace(relation)
			if relation == "user" {
				loadUser = true
			}
			if relation == "notification" {
				loadNotification = true
			}
		}

		// 批量加载用户信息
		if loadUser {
			userIDs := make([]string, 0)
			userIDMap := make(map[string]bool)
			for _, record := range records {
				if !userIDMap[record.UserID] {
					userIDs = append(userIDs, record.UserID)
					userIDMap[record.UserID] = true
				}
			}
			if len(userIDs) > 0 {
				var users []models.User
				repository.DB.Where("user_id IN ?", userIDs).Find(&users)
				userMap := make(map[string]*models.User)
				for i := range users {
					userMap[users[i].UserID] = &users[i]
				}
				for i := range records {
					if user, ok := userMap[records[i].UserID]; ok {
						records[i].User = user
					}
				}
			}
		}

		// 批量加载通知信息
		if loadNotification {
			notificationIDs := make([]string, 0)
			notificationIDMap := make(map[string]bool)
			for _, record := range records {
				if !notificationIDMap[record.NotificationID] {
					notificationIDs = append(notificationIDs, record.NotificationID)
					notificationIDMap[record.NotificationID] = true
				}
			}
			if len(notificationIDs) > 0 {
				var notifications []models.SystemNotification
				repository.DB.Where("notification_id IN ?", notificationIDs).Find(&notifications)
				notificationMap := make(map[string]*models.SystemNotification)
				for i := range notifications {
					notificationMap[notifications[i].NotificationID] = &notifications[i]
				}
				for i := range records {
					if notification, ok := notificationMap[records[i].NotificationID]; ok {
						records[i].Notification = notification
					}
				}
			}
		}
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(records))
	for _, record := range records {
		item := gin.H{
			"record_id":       record.RecordID,
			"notification_id": record.NotificationID,
			"user_id":         record.UserID,
			"status":          string(record.Status),
			"created_at":      record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if record.ReadTime != nil {
			item["read_time"] = record.ReadTime.Format("2006-01-02T15:04:05Z07:00")
		}
		if record.DeletedTime != nil {
			item["deleted_time"] = record.DeletedTime.Format("2006-01-02T15:04:05Z07:00")
		}

		// 添加关联数据
		if record.User != nil {
			item["user"] = gin.H{
				"user_id":  record.User.UserID,
				"username": record.User.Username,
				"nickname": record.User.Nickname,
			}
		}
		if record.Notification != nil {
			item["notification"] = gin.H{
				"notification_id": record.Notification.NotificationID,
				"title":           record.Notification.Title,
				"content":         record.Notification.Content,
				"type":            record.Notification.Type,
			}
		}

		result = append(result, item)
	}

	middleware.Success(c, "获取成功", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      result,
	})
}

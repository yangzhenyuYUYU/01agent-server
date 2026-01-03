package router

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService       *service.UserService
	invitationService *service.InvitationService
}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService:       service.NewUserService(),
		invitationService: service.NewInvitationService(),
	}
}

// Hello 简单的hello接口
func (h *UserHandler) Hello(c *gin.Context) {
	// 使用新的成功响应格式 (code = 0)
	middleware.Success(c, "hello", gin.H{
		"message": "Hello from Gin Web Server!",
		"version": "1.0.0",
	})
}

// GetUserInfo 获取当前用户信息 - 对应Python的/user/info接口
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		repository.Errorf("GetUserInfo: userID not found in context")
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	// 获取详情模式参数
	detail := c.Query("detail")

	repository.Infof("GetUserInfo: attempting to find user with ID: %s, detail: %s", userID, detail)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		repository.Errorf("GetUserInfo: failed to get user by ID %s: %v", userID, err)
		middleware.HandleError(c, middleware.NewBusinessError(404, "未找到用户记录"))
		return
	}

	// 获取用户参数
	userParams, err := h.userService.GetUserParameters(userID)
	if err != nil {
		repository.Warnf("GetUserInfo: failed to get user parameters for %s: %v", userID, err)
		userParams = nil
	}

	// 构建基础用户信息
	info := h.buildUserInfo(user, userParams)

	// 如果需要完整信息
	if detail == "all" {
		h.addDetailedInfo(info, user)
	}

	repository.Infof("GetUserInfo: successfully found user: %s", userID)
	// 成功响应 (code = 0)
	middleware.Success(c, "获取用户信息成功", info)
}

// buildUserInfo 构建基础用户信息
func (h *UserHandler) buildUserInfo(user *models.User, userParams *models.UserParameters) gin.H {
	// 处理用户参数
	var userParamDict gin.H
	if userParams != nil {
		// 计算综合发布状态：0=不同步，1=同步到草稿箱，2=直接发布上线
		publishStatus := 0
		if userParams.IsWechatAuthorized {
			if userParams.PublishTarget == 1 { // 假设1为草稿箱
				publishStatus = 1
			} else { // 直接发布
				publishStatus = 2
			}
		}

		userParamDict = gin.H{
			"enable_head_info":      userParams.EnableHeadInfo,
			"enable_knowledge_base": userParams.EnableKnowledgeBase,
			"default_theme":         userParams.DefaultTheme,
			"is_wechat_authorized":  userParams.IsWechatAuthorized,
			"has_auth_reminded":     userParams.HasAuthReminded,
			"is_gzh_bind":           userParams.IsGzhBind,
			"publish_target":        userParams.PublishTarget,
			"publish_status":        publishStatus,
		}
	}

	// 安全地处理指针字段
	var nickname, avatar, openid, username, phone, email, appid interface{}
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	if user.Avatar != nil {
		avatar = *user.Avatar
	}
	if user.OpenID != nil {
		openid = *user.OpenID
	}
	if user.Username != nil {
		username = *user.Username
	}
	if user.Phone != nil {
		phone = *user.Phone
	}
	if user.Email != nil {
		email = *user.Email
	}
	if user.AppID != nil {
		appid = *user.AppID
	}

	// 构建基础返回信息
	info := gin.H{
		"id":                user.UserID,
		"nickname":          nickname,
		"avatar":            avatar,
		"openid":            openid,
		"username":          username,
		"phone":             phone,
		"email":             email,
		"appid":             appid,
		"credits":           user.Credits,
		"is_active":         user.IsActive,
		"vip_level":         user.VipLevel,
		"role":              user.Role,
		"status":            user.Status,
		"registration_date": user.RegistrationDate.Format("2006-01-02T15:04:05Z07:00"),
		"last_login_time":   user.LastLoginTime.Format("2006-01-02T15:04:05Z07:00"),
		"user_param":        userParamDict,
		"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return info
}

// addDetailedInfo 添加详细信息 (对应Python中的detail="all"逻辑)
func (h *UserHandler) addDetailedInfo(info gin.H, user *models.User) {
	// 这里可以添加更多详细信息，如：
	// - 邀请码信息
	// - 邀请人信息
	// - 使用次数统计
	// - 权益信息等

	// 暂时添加一些基础的详细信息
	info["invitation_code"] = nil         // 邀请码
	info["invitation_count"] = 0          // 邀请数量
	info["inviter"] = nil                 // 邀请人信息
	info["usage_count"] = user.UsageCount // 使用次数
	info["benefit_info"] = gin.H{         // 权益信息
		"membership_name": "",
		"is_active":       false,
		"expire_time":     nil,
		"production_info": nil,
	}
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 调用服务层更新
	user, err := h.userService.Update(userID, &req)
	if err != nil {
		repository.Errorf("UpdateUserInfo failed: %v", err)
		if err.Error() == "user not found" {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		if err.Error() == "email already exists" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "邮箱已被使用"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 构建响应
	response := gin.H{
		"user_id":  user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"phone":    user.Phone,
		"message":  "更新成功",
	}

	middleware.Success(c, "更新成功", response)
}

// GetUserParameters 获取用户参数配置
func (h *UserHandler) GetUserParameters(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	params, err := h.userService.GetUserParameters(userID)
	if err != nil {
		repository.Errorf("GetUserParameters failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取用户参数失败: "+err.Error()))
		return
	}

	// 构建响应
	response := gin.H{
		"enable_head_info":      params.EnableHeadInfo,
		"enable_knowledge_base": params.EnableKnowledgeBase,
		"default_theme":         params.DefaultTheme,
		"is_wechat_authorized":  params.IsWechatAuthorized,
		"has_auth_reminded":     params.HasAuthReminded,
		"is_gzh_bind":           params.IsGzhBind,
		"publish_target":        params.PublishTarget,
	}

	middleware.Success(c, "获取用户参数成功", response)
}

// UpdateUserParameters 更新用户参数配置
func (h *UserHandler) UpdateUserParameters(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req models.UserParameters
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 调用服务层更新
	if err := h.userService.UpdateUserParameters(userID, &req); err != nil {
		repository.Errorf("UpdateUserParameters failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新用户参数失败: "+err.Error()))
		return
	}

	middleware.Success(c, "更新用户参数成功", gin.H{
		"message": "更新成功",
	})
}

// GetUserSessions 获取用户活跃会话
func (h *UserHandler) GetUserSessions(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	sessions, err := h.userService.GetActiveSessions(userID)
	if err != nil {
		repository.Errorf("GetUserSessions failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取会话列表失败: "+err.Error()))
		return
	}

	// 构建响应
	var sessionList []gin.H
	for _, session := range sessions {
		sessionList = append(sessionList, gin.H{
			"id":         session.ID,
			"ip_address": session.IPAddress,
			"login_time": session.LoginTime.Format("2006-01-02T15:04:05Z07:00"),
			"expires_at": session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			"is_active":  session.IsActive,
			"login_type": session.LoginType,
		})
	}

	middleware.Success(c, "获取会话列表成功", gin.H{
		"sessions": sessionList,
		"count":    len(sessionList),
	})
}

// SetupUserRoutes 设置用户路由
func SetupUserRoutes(r *gin.Engine, userHandler *UserHandler) {
	// 公开路由
	public := r.Group("/api/v1")
	{
		// 基础路由
		public.GET("/hello", userHandler.Hello)
		public.GET("/", userHandler.Hello)
	}

	// 需要认证的用户路由 - 放在 /api/v1/user 下
	userGroup := r.Group("/api/v1/user")
	userGroup.Use(middleware.JWTAuth())
	{
		// 用户信息接口 - /api/v1/user/info
		userGroup.GET("/info", userHandler.GetUserInfo)
		// 更新用户信息 - /api/v1/user/info
		userGroup.PUT("/info", userHandler.UpdateUserInfo)
		// 获取用户参数 - /api/v1/user/parameters
		userGroup.GET("/parameters", userHandler.GetUserParameters)
		// 更新用户参数 - /api/v1/user/parameters
		userGroup.PUT("/parameters", userHandler.UpdateUserParameters)
		// 获取用户会话列表 - /api/v1/user/sessions
		userGroup.GET("/sessions", userHandler.GetUserSessions)
	}
}

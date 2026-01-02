package router

import (
	"fmt"
	"time"

	"gin_web/internal/middleware"
	"gin_web/internal/models"
	"gin_web/internal/repository"
	"gin_web/internal/service"
	tools "gin_web/internal/tools"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userService: service.NewUserService(),
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 调用服务层注册
	user, err := h.userService.Register(&req)
	if err != nil {
		repository.Errorf("Register failed: %v", err)
		if err.Error() == "email already exists" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "邮箱已被注册"))
			return
		}
		if err.Error() == "username already exists" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "用户名已被使用"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "注册失败: "+err.Error()))
		return
	}

	// 构建响应
	response := gin.H{
		"user_id":  user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"message":  "注册成功",
	}

	middleware.Success(c, "注册成功", response)
}

// LoginRequest 登录请求 - 对应Python的LoginData
type LoginRequest struct {
	LoginType  string `json:"login_type" binding:"required"` // phone, email, username, wxgzh
	Identifier string `json:"identifier" binding:"required"` // 标识符（手机号、邮箱、用户名或openid）
	InviteCode string `json:"invite_code"`                   // 邀请码（可选）
	UtmSource  string `json:"utm_source"`                    // 用户来源渠道（可选）
}

// Login 用户登录 - 对应Python的/auth/login接口
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	// 获取客户端IP
	ipAddress := getClientIP(c)
	// 获取设备ID
	deviceID := c.GetHeader("X-Device-ID")

	// 获取token（如果存在，需要使旧会话失效）
	authHeader := c.GetHeader("Authorization")
	var oldToken string
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		oldToken = authHeader[7:]
	}

	// 转换为服务层请求
	serviceReq := &service.LoginRequest{
		LoginType:  req.LoginType,
		Identifier: req.Identifier,
		InviteCode: req.InviteCode,
		UtmSource:  req.UtmSource,
	}

	// 调用服务层登录（支持自动创建用户）
	result, err := h.userService.LoginWithTypeV2(serviceReq, ipAddress, deviceID, oldToken)
	if err != nil {
		repository.Errorf("Login failed: %v", err)
		c.JSON(400, models.ErrorResponse(400, err.Error()))
		return
	}

	user := result.User
	token := result.Token
	session := result.Session

	// 构建用户信息（与Python版本格式一致）
	userInfo := gin.H{
		"sub":               user.UserID,
		"id":                user.UserID,
		"nickname":          tools.GetStringValue(user.Nickname),
		"avatar":            tools.GetStringValue(user.Avatar),
		"username":          tools.GetStringValue(user.Username),
		"password_hash":     tools.GetStringValue(user.PasswordHash),
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
				return fmt.Sprintf("%.2f", *user.TotalConsumption)
			}
			return "0.00"
		}(),
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	}

	// 构建会话信息
	sessionInfo := gin.H{
		"token":      session.Token,
		"login_type": session.LoginType,
		"ip_address": session.IPAddress,
		"device_id":  tools.GetStringValue(session.DeviceID),
		"created_at": session.CreatedAt.Format(time.RFC3339),
	}

	// 返回响应（与Python版本格式一致：code: 0, msg, data）
	c.JSON(200, models.SuccessResponse(result.LoginMsg, gin.H{
		"token":        token,
		"user":         userInfo,
		"session":      sessionInfo,
		"max_sessions": result.MaxSessions,
	}))
}

// Logout 用户登出 - 对应Python的/auth/logout (PUT)
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	// 获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "缺少token"))
		return
	}

	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		token = authHeader
	}

	// 调用登出服务
	if err := h.userService.Logout(userID, token); err != nil {
		repository.Errorf("Logout failed: %v", err)
	}

	repository.Infof("User %s logged out", userID)
	// 与Python版本格式一致
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "退出成功",
	})
}

// DeleteAccount 注销用户账号 - 对应Python的/auth/account (DELETE)
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	// 调用服务层注销账号
	if err := h.userService.DeleteAccount(userID); err != nil {
		repository.Errorf("DeleteAccount failed: %v", err)
		c.JSON(400, models.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "账号已注销",
	})
}

// BindPhoneRequest 绑定手机号请求
type BindPhoneRequest struct {
	Phone      string `json:"phone" binding:"required"`
	Code       string `json:"code" binding:"required"`
	Identifier string `json:"identifier" binding:"required"` // openid
}

// BindPhone 绑定手机号 - 对应Python的/auth/bind-phone
func (h *AuthHandler) BindPhone(c *gin.Context) {
	var req BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	// TODO: 验证验证码
	// 这里需要集成验证码验证逻辑

	// 调用服务层绑定手机号
	user, err := h.userService.BindPhone(req.Phone, req.Identifier)
	if err != nil {
		repository.Errorf("BindPhone failed: %v", err)
		c.JSON(400, models.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "绑定成功",
		"data": gin.H{
			"id": user.UserID,
		},
	})
}

// SendCodeRequest 发送验证码请求（可选，POST方式）
type SendCodeRequest struct {
	Phone    string `json:"phone" binding:"required"`
	AreaCode string `json:"area_code"`
}

// SendCode 发送验证码 - 对应Python的/auth/send-code (GET)
func (h *AuthHandler) SendCode(c *gin.Context) {
	phone := c.Query("phone")
	areaCode := c.Query("area_code")

	if phone == "" {
		c.JSON(400, models.ErrorResponse(400, "手机号不能为空"))
		return
	}

	if areaCode == "" {
		areaCode = "86"
	}

	// 验证手机号格式
	if len(phone) != 11 || !isDigit(phone) {
		c.JSON(400, models.ErrorResponse(400, "无效的手机号码"))
		return
	}

	// TODO: 检查是否频繁发送（使用Redis）
	// TODO: 发送验证码（集成SMS服务）

	// 模拟成功响应
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "验证码发送成功",
	})
}

// VerifyCodeRequest 验证验证码请求
type VerifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// VerifyCode 验证验证码 - 对应Python的/auth/verify-code
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	// 验证手机号格式
	if len(req.Phone) != 11 || !isDigit(req.Phone) {
		c.JSON(400, models.ErrorResponse(400, "无效的手机号码"))
		return
	}

	// TODO: 从Redis获取验证码并验证
	// 这里需要集成Redis验证逻辑

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "验证成功",
	})
}

// CheckPhone 检查手机号是否绑定 - 对应Python的/auth/check-phone
func (h *AuthHandler) CheckPhone(c *gin.Context) {
	phone := c.Query("phone")
	identifier := c.Query("identifier")

	if phone == "" {
		c.JSON(400, models.ErrorResponse(400, "手机号不能为空"))
		return
	}

	// 调用服务层检查
	if err := h.userService.CheckPhone(phone, identifier); err != nil {
		c.JSON(400, models.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "当前号码可绑定",
	})
}

// CheckEmail 检查邮箱是否绑定 - 对应Python的/auth/check-email
func (h *AuthHandler) CheckEmail(c *gin.Context) {
	email := c.Query("email")
	identifier := c.Query("identifier")

	if email == "" {
		c.JSON(400, models.ErrorResponse(400, "邮箱不能为空"))
		return
	}

	// 调用服务层检查
	if err := h.userService.CheckEmail(email, identifier); err != nil {
		c.JSON(400, models.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "当前邮箱可绑定",
	})
}

// SendEmailCode 发送邮箱验证码 - 对应Python的/auth/send-email-code
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	email := c.Query("email")

	if email == "" {
		c.JSON(400, models.ErrorResponse(400, "邮箱不能为空"))
		return
	}

	// TODO: 实现邮箱验证码发送逻辑

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "邮箱验证码发送成功",
	})
}

// VerifyEmailCodeRequest 验证邮箱验证码请求
type VerifyEmailCodeRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// VerifyEmailCode 验证邮箱验证码 - 对应Python的/auth/verify-email-code
func (h *AuthHandler) VerifyEmailCode(c *gin.Context) {
	var req VerifyEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	// TODO: 实现邮箱验证码验证逻辑

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "验证成功",
	})
}

// getClientIP 获取客户端真实IP地址
func getClientIP(c *gin.Context) string {
	// 按优先级依次尝试获取真实IP
	if xRealIP := c.GetHeader("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP，第一个是客户端真实IP
		for i := 0; i < len(xForwardedFor); i++ {
			if xForwardedFor[i] == ',' {
				return xForwardedFor[:i]
			}
		}
		return xForwardedFor
	}
	return c.ClientIP()
}

// isDigit 检查字符串是否全为数字
func isDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// SetupAuthRoutes 设置认证路由
func SetupAuthRoutes(r *gin.Engine, authHandler *AuthHandler) {
	// 公开路由 - 不需要认证（与Python的/auth前缀对应）
	public := r.Group("/api/v1")
	{
		// 用户注册
		public.POST("/register", authHandler.Register)
		// 用户登录
		public.POST("/login", authHandler.Login)
	}

	// 认证相关公开路由 - /api/v1/auth
	authPublic := r.Group("/api/v1/auth")
	{
		// 用户登录（兼容/auth/login路径）
		authPublic.POST("/login", authHandler.Login)
		// 发送验证码
		authPublic.GET("/send-code", authHandler.SendCode)
		// 验证验证码
		authPublic.POST("/verify-code", authHandler.VerifyCode)
		// 检查手机号
		authPublic.GET("/check-phone", authHandler.CheckPhone)
		// 检查邮箱
		authPublic.GET("/check-email", authHandler.CheckEmail)
		// 发送邮箱验证码
		authPublic.GET("/send-email-code", authHandler.SendEmailCode)
		// 验证邮箱验证码
		authPublic.POST("/verify-email-code", authHandler.VerifyEmailCode)
		// 绑定手机号
		authPublic.POST("/bind-phone", authHandler.BindPhone)
	}

	// 需要认证的路由
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.JWTAuth())
	{
		// 用户登出（使用PUT方法与Python一致）
		authGroup.PUT("/logout", authHandler.Logout)
		// 兼容POST方法
		authGroup.POST("/logout", authHandler.Logout)
	}

	// 需要认证的auth路由
	authPrivate := r.Group("/api/v1/auth")
	authPrivate.Use(middleware.JWTAuth())
	{
		// 用户登出
		authPrivate.PUT("/logout", authHandler.Logout)
		// 注销账号
		authPrivate.DELETE("/account", authHandler.DeleteAccount)
	}
}

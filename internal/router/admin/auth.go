package admin

import (
	"fmt"
	"strings"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
		if existingSession.Token != nil {
			// 如果存在可用的session，直接复用，更新活跃时间
			existingSession.LastActiveTime = time.Now()
			if err := sessionRepo.UpdateLastActiveTime(existingSession.ID); err != nil {
				repository.Errorf("Failed to update session: %v", err)
			} else {
				token = *existingSession.Token
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
			Token:          tools.StringPtr(token),
			LoginType:      "web",
			IPAddress:      ipAddress,
			Status:         1,
			LastActiveTime: time.Now(),
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


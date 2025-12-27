package service

import (
	"fmt"
	"time"

	"gin_web/internal/models"
	"gin_web/internal/repository"
	"gin_web/internal/tools"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo       *repository.UserRepository
	sessionRepo    *repository.UserSessionRepository
	parametersRepo *repository.UserParametersRepository
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo:       repository.NewUserRepository(),
		sessionRepo:    repository.NewUserSessionRepository(),
		parametersRepo: repository.NewUserParametersRepository(),
	}
}

// generateUserID 生成用户ID
func (s *UserService) generateUserID() string {
	return uuid.New().String()
}

// Register 用户注册
func (s *UserService) Register(req *models.UserRegisterRequest) (*models.User, error) {
	// 检查邮箱是否已存在
	if exists, err := s.userRepo.IsEmailExists(req.Email); err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	} else if exists {
		return nil, fmt.Errorf("email already exists")
	}

	// 检查用户名是否已存在
	if req.Username != "" {
		if exists, err := s.userRepo.IsUsernameExists(req.Username); err != nil {
			return nil, fmt.Errorf("failed to check username: %w", err)
		} else if exists {
			return nil, fmt.Errorf("username already exists")
		}
	}

	// 创建用户
	user := &models.User{
		UserID:           s.generateUserID(),
		Username:         tools.StringPtr(req.Username),
		Email:            tools.StringPtr(req.Email),
		Nickname:         tools.StringPtr(req.Nickname),
		RegistrationDate: time.Now(),
		LastLoginTime:    time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 设置密码
	if err := user.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 开始事务
	tx := repository.GetDB().Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 创建用户
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 创建用户参数
	userParams := &models.UserParameters{
		UserID: user.UserID,
	}
	if err := tx.Create(userParams).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user parameters: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *models.UserLoginRequest, ipAddress string) (*models.User, string, error) {
	var user *models.User
	var err error

	// 根据用户名或邮箱查找用户
	if req.Username != "" {
		user, err = s.userRepo.GetByUsername(req.Username)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, "", fmt.Errorf("failed to get user by username: %w", err)
		}
	}

	if user == nil && req.Email != "" {
		user, err = s.userRepo.GetByEmail(req.Email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, "", fmt.Errorf("failed to get user by email: %w", err)
		}
	}

	if user == nil {
		return nil, "", fmt.Errorf("user not found")
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return nil, "", fmt.Errorf("invalid password")
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateLastLoginTime(user.UserID); err != nil {
		repository.Errorf("Failed to update last login time: %v", err)
	}

	// 生成JWT token
	token, err := tools.GenerateToken(user.UserID, tools.GetStringValue(user.Username))
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// 创建会话记录
	session := &models.UserSession{
		UserID:    user.UserID,
		Token:     token,
		IPAddress: ipAddress,
		LoginTime: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时过期
		IsActive:  true,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		repository.Errorf("Failed to create session: %v", err)
	}

	return user, token, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(userID string) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}

// Update 更新用户信息
func (s *UserService) Update(userID string, req *models.UserUpdateRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查邮箱是否被其他用户使用
	if req.Email != "" && req.Email != tools.GetStringValue(user.Email) {
		if exists, err := s.userRepo.IsEmailExists(req.Email); err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		} else if exists {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = tools.StringPtr(req.Email)
	}

	// 更新其他字段
	if req.Nickname != "" {
		user.Nickname = tools.StringPtr(req.Nickname)
	}
	if req.Avatar != "" {
		user.Avatar = tools.StringPtr(req.Avatar)
	}
	if req.Phone != "" {
		user.Phone = tools.StringPtr(req.Phone)
	}
	if req.TotalConsumption != nil {
		user.TotalConsumption = tools.Float64Ptr(*req.TotalConsumption)
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// List 获取用户列表
func (s *UserService) List(page, size int) ([]models.User, int64, error) {
	return s.userRepo.List(page, size)
}

// GetUserParameters 获取用户参数
func (s *UserService) GetUserParameters(userID string) (*models.UserParameters, error) {
	return s.parametersRepo.GetByUserID(userID)
}

// UpdateUserParameters 更新用户参数
func (s *UserService) UpdateUserParameters(userID string, params *models.UserParameters) error {
	params.UserID = userID
	params.UpdatedAt = time.Now()
	return s.parametersRepo.Update(params)
}

// Logout 用户登出
func (s *UserService) Logout(userID, token string) error {
	return s.sessionRepo.DeactivateByToken(token)
}

// GetActiveSessions 获取用户活跃会话
func (s *UserService) GetActiveSessions(userID string) ([]models.UserSession, error) {
	return s.sessionRepo.GetByUserID(userID)
}

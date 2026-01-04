package service

import (
	"fmt"
	"strings"
	"time"

	"01agent_server/internal/config"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole 用户角色枚举
type UserRole int16

const (
	UserRoleNormal      UserRole = 1 // 普通用户
	UserRoleVIP         UserRole = 2 // VIP用户
	UserRoleAdmin       UserRole = 3 // 管理员
	UserRoleDistributor UserRole = 4 // 分销商/合作方
)

// SessionStatus 会话状态枚举
type SessionStatus int16

const (
	SessionStatusInactive SessionStatus = 0 // 不活跃
	SessionStatusActive   SessionStatus = 1 // 活跃
)

// 用户角色和VIP等级对应的最大同时在线设备数配置
var MaxSessionsConfig = map[UserRole]interface{}{
	UserRoleAdmin:  9999, // 管理员：无限制
	UserRoleNormal: 1,    // 普通用户：1个设备
	UserRoleVIP: map[int]int{ // VIP用户：根据等级
		1: 3,
		2: 3,
		3: 5,
		4: 6,
	},
}

// GetMaxSessions 获取用户允许的最大同时在线设备数
func GetMaxSessions(user *models.User) int {
	role := UserRole(user.Role)
	switch role {
	case UserRoleAdmin:
		return MaxSessionsConfig[UserRoleAdmin].(int)
	case UserRoleVIP:
		vipConfig := MaxSessionsConfig[UserRoleVIP].(map[int]int)
		if maxSessions, ok := vipConfig[user.VipLevel]; ok {
			return maxSessions
		}
		return 1
	default:
		return MaxSessionsConfig[UserRoleNormal].(int)
	}
}

type UserService struct {
	userRepo       *repository.UserRepository
	sessionRepo    *repository.UserSessionRepository
	parametersRepo *repository.UserParametersRepository
	invitationRepo *repository.InvitationRepository
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo:       repository.NewUserRepository(),
		sessionRepo:    repository.NewUserSessionRepository(),
		parametersRepo: repository.NewUserParametersRepository(),
		invitationRepo: repository.NewInvitationRepository(),
	}
}

// generateUserID 生成用户ID
func (s *UserService) generateUserID() string {
	return uuid.New().String()
}

// generateNotificationID 生成通知ID
func (s *UserService) generateNotificationID() string {
	return uuid.New().String()
}

// GetInitialCredits 获取初始积分配置
func (s *UserService) GetInitialCredits() int {
	if config.AppConfig != nil && config.AppConfig.Credits.Initial > 0 {
		return config.AppConfig.Credits.Initial
	}
	return 100 // 默认初始积分
}

// GetInvitationReward 获取邀请奖励积分
func (s *UserService) GetInvitationReward() int {
	// 可以从配置中读取，这里默认100
	return 100
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

	initialCredits := s.GetInitialCredits()

	// 创建用户
	user := &models.User{
		UserID:           s.generateUserID(),
		Username:         tools.StringPtr(req.Username),
		Email:            tools.StringPtr(req.Email),
		Nickname:         tools.StringPtr(req.Nickname),
		Credits:          initialCredits,
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
		ParamID: uuid.New().String(),
		UserID:  user.UserID,
	}
	if err := tx.Create(userParams).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user parameters: %w", err)
	}

	// 创建新人注册积分奖励记录
	creditRecord := &models.CreditRecord{
		UserID:      user.UserID,
		RecordType:  models.CreditReward,
		Credits:     tools.IntPtr(initialCredits),
		Balance:     tools.IntPtr(user.Credits),
		Description: tools.StringPtr("新人注册积分奖励"),
		CreatedAt:   time.Now(),
	}
	if err := tx.Create(creditRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create credit record: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

// Login 用户登录（传统登录方式）
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
		Token:     tools.StringPtr(token),
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
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
	params.UpdatedTime = time.Now()
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

// LoginRequest 登录请求（用于多种登录类型）- 对应Python的LoginData
type LoginRequest struct {
	LoginType  string `json:"login_type"`  // phone, email, username, wxgzh
	Identifier string `json:"identifier"`  // 标识符
	InviteCode string `json:"invite_code"` // 邀请码
	UtmSource  string `json:"utm_source"`  // 用户来源
}

// LoginResult 登录结果
type LoginResult struct {
	User        *models.User
	Token       string
	Session     *models.UserSession
	MaxSessions int
	LoginMsg    string
	IsNewUser   bool
}

// LoginWithType 支持多种登录类型的登录方法（对应Python的/auth/login）
func (s *UserService) LoginWithType(req *LoginRequest, ipAddress, deviceID, oldToken string) (*models.User, string, *models.UserSession, error) {
	result, err := s.LoginWithTypeV2(req, ipAddress, deviceID, oldToken)
	if err != nil {
		return nil, "", nil, err
	}
	return result.User, result.Token, result.Session, nil
}

// LoginWithTypeV2 支持多种登录类型的登录方法（返回更详细的结果）
func (s *UserService) LoginWithTypeV2(req *LoginRequest, ipAddress, deviceID, oldToken string) (*LoginResult, error) {
	db := repository.GetDB()
	var user *models.User
	var err error
	isNewUser := false
	initialCredits := s.GetInitialCredits()

	// 设置默认utm_source
	utmSource := req.UtmSource
	if utmSource == "" {
		utmSource = "direct"
	}

	// 根据登录类型查找或创建用户
	switch req.LoginType {
	case "phone":
		user, err = s.userRepo.GetByPhone(req.Identifier)
		if err == gorm.ErrRecordNotFound {
			isNewUser = true
			user = &models.User{
				UserID:           s.generateUserID(),
				Phone:            tools.StringPtr(req.Identifier),
				Credits:          initialCredits,
				UtmSource:        tools.StringPtr(utmSource),
				RegistrationDate: time.Now(),
				LastLoginTime:    time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := db.Create(user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			// 创建新人注册积分奖励记录
			s.createCreditRecord(db, user.UserID, initialCredits, user.Credits, "新人注册积分奖励")
		} else if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

	case "email":
		user, err = s.userRepo.GetByEmail(req.Identifier)
		if err == gorm.ErrRecordNotFound {
			isNewUser = true
			user = &models.User{
				UserID:           s.generateUserID(),
				Email:            tools.StringPtr(req.Identifier),
				Credits:          initialCredits,
				UtmSource:        tools.StringPtr(utmSource),
				RegistrationDate: time.Now(),
				LastLoginTime:    time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := db.Create(user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			s.createCreditRecord(db, user.UserID, initialCredits, user.Credits, "新人注册积分奖励")
		} else if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

	case "username":
		user, err = s.userRepo.GetByUsername(req.Identifier)
		if err == gorm.ErrRecordNotFound {
			isNewUser = true
			user = &models.User{
				UserID:           s.generateUserID(),
				Username:         tools.StringPtr(req.Identifier),
				Credits:          initialCredits,
				UtmSource:        tools.StringPtr(utmSource),
				RegistrationDate: time.Now(),
				LastLoginTime:    time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := db.Create(user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			s.createCreditRecord(db, user.UserID, initialCredits, user.Credits, "新人注册积分奖励")
		} else if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

	case "wxgzh":
		user, err = s.userRepo.GetByOpenID(req.Identifier)
		if err == gorm.ErrRecordNotFound {
			isNewUser = true
			user = &models.User{
				UserID:           s.generateUserID(),
				OpenID:           tools.StringPtr(req.Identifier),
				Credits:          initialCredits,
				UtmSource:        tools.StringPtr(utmSource),
				RegistrationDate: time.Now(),
				LastLoginTime:    time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := db.Create(user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			s.createCreditRecord(db, user.UserID, initialCredits, user.Credits, "新人注册积分奖励")
		} else if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported login type: %s", req.LoginType)
	}

	// 确保用户参数存在
	_, err = s.parametersRepo.GetByUserID(user.UserID)
	if err == gorm.ErrRecordNotFound {
		userParams := &models.UserParameters{
			ParamID: uuid.New().String(),
			UserID:  user.UserID,
		}
		s.parametersRepo.Create(userParams)
	}

	// 如果提供了旧token，使旧会话失效
	if oldToken != "" {
		s.sessionRepo.DeactivateByToken(oldToken)
	}

	// 检查用户状态
	if user.Status != 1 { // 1 表示活跃
		return nil, fmt.Errorf("用户已失效")
	}

	// 生成默认昵称（如果为空）
	if user.Nickname == nil || *user.Nickname == "" {
		nickname := s.generateDefaultNickname(user)
		user.Nickname = tools.StringPtr(nickname)
		s.userRepo.Update(user)
	}

	// 处理邀请码逻辑
	if req.InviteCode != "" && isNewUser {
		s.processInviteCode(db, user, req.InviteCode)
	}

	// 更新最后登录时间
	user.LastLoginTime = time.Now()
	s.userRepo.Update(user)

	// 构建JWT token信息
	username := tools.GetStringValue(user.Username)
	if username == "" {
		username = user.UserID
	}

	// 生成JWT token
	token, err := tools.GenerateToken(user.UserID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 获取用户允许的最大设备数
	maxSessions := GetMaxSessions(user)

	// 根据用户角色处理会话
	loginMsg := s.handleSessionByRole(user, token, maxSessions)

	// 创建会话记录
	session := &models.UserSession{
		UserID:         user.UserID,
		Token:          tools.StringPtr(token),
		LoginType:      "web",
		IPAddress:      ipAddress,
		DeviceID:       tools.StringPtr(deviceID),
		Status:         1, // 活跃
		LastActiveTime: time.Now(),
		CreatedAt:      time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 清理会话：根据用户等级保留对应数量的在线session
	deletedCount, _ := s.sessionRepo.CleanupSessionsKeepRecent(user.UserID, maxSessions)

	// 发送登录成功系统通知
	s.sendLoginNotification(db, user, maxSessions, deletedCount)

	return &LoginResult{
		User:        user,
		Token:       token,
		Session:     session,
		MaxSessions: maxSessions,
		LoginMsg:    loginMsg,
		IsNewUser:   isNewUser,
	}, nil
}

// createCreditRecord 创建积分记录
func (s *UserService) createCreditRecord(db *gorm.DB, userID string, credits int, balance int, description string) {
	creditRecord := &models.CreditRecord{
		UserID:      userID,
		RecordType:  models.CreditReward,
		Credits:     tools.IntPtr(credits),
		Balance:     tools.IntPtr(balance),
		Description: tools.StringPtr(description),
		CreatedAt:   time.Now(),
	}
	db.Create(creditRecord)
}

// generateDefaultNickname 生成默认昵称
func (s *UserService) generateDefaultNickname(user *models.User) string {
	if user.Phone != nil && *user.Phone != "" {
		phone := *user.Phone
		if len(phone) >= 4 {
			return "用户" + phone[len(phone)-4:]
		}
		return "用户" + phone
	}
	if user.Username != nil && *user.Username != "" {
		username := *user.Username
		if len(username) > 8 {
			return "用户" + username[:8]
		}
		return "用户" + username
	}
	if user.Email != nil && *user.Email != "" {
		email := *user.Email
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			username := parts[0]
			if len(username) > 8 {
				return "用户" + username[:8]
			}
			return "用户" + username
		}
	}
	if len(user.UserID) > 8 {
		return "用户" + user.UserID[:8]
	}
	return "用户" + user.UserID
}

// processInviteCode 处理邀请码
func (s *UserService) processInviteCode(db *gorm.DB, user *models.User, inviteCode string) {
	// 检查是否已有邀请关系
	existingRelation, _ := s.invitationRepo.GetInvitationRelationByInvitee(user.UserID)
	if existingRelation != nil {
		return // 已有邀请关系，不再处理
	}

	// 查找邀请码
	inviterCode, err := s.invitationRepo.GetInvitationCodeByCode(inviteCode)
	if err != nil || inviterCode == nil {
		return // 邀请码无效
	}

	// 不能邀请自己
	if inviterCode.UserID == user.UserID {
		return
	}

	// 获取邀请人
	inviter, err := s.userRepo.GetByID(inviterCode.UserID)
	if err != nil {
		return
	}

	creditsReward := s.GetInvitationReward()

	// 创建邀请关系
	relation := &models.InvitationRelation{
		InviterID: inviterCode.UserID,
		InviteeID: user.UserID,
		CodeID:    inviterCode.ID,
		CreatedAt: time.Now(),
	}
	if err := s.invitationRepo.CreateInvitationRelation(relation); err != nil {
		return
	}

	// 邀请人奖励积分
	inviter.Credits += creditsReward
	s.userRepo.Update(inviter)

	// 用户奖励积分
	user.Credits += creditsReward
	s.userRepo.Update(user)

	// 创建双方积分记录
	s.createCreditRecord(db, inviter.UserID, creditsReward, inviter.Credits, "邀请新用户注册成功奖励")
	s.createCreditRecord(db, user.UserID, creditsReward, user.Credits, "绑定邀请人成功奖励")
}

// handleSessionByRole 根据用户角色处理会话
func (s *UserService) handleSessionByRole(user *models.User, token string, maxSessions int) string {
	role := UserRole(user.Role)
	switch role {
	case UserRoleAdmin:
		return "管理员登录成功"
	case UserRoleVIP:
		sessionCount, _ := s.sessionRepo.CountActiveSessionsByUserID(user.UserID)
		if sessionCount > int64(maxSessions) {
			s.sessionRepo.DeactivateOtherSessions(user.UserID, token)
			sessionCount = 1
		}
		remaining := maxSessions - int(sessionCount)
		if remaining < 0 {
			remaining = 0
		}
		return fmt.Sprintf("VIP登录成功，还剩%d设备可登录", remaining)
	default:
		// 普通用户只允许1个设备，使其他会话失效
		s.sessionRepo.DeactivateOtherSessions(user.UserID, token)
		return "登录成功"
	}
}

// sendLoginNotification 发送登录成功通知
func (s *UserService) sendLoginNotification(db *gorm.DB, user *models.User, maxSessions int, deletedCount int64) {
	var welcomeMessages []string

	// 欢迎消息
	if !user.LastLoginTime.IsZero() {
		daysSinceLastLogin := int(time.Since(user.LastLoginTime).Hours() / 24)
		if daysSinceLastLogin > 0 {
			welcomeMessages = append(welcomeMessages, fmt.Sprintf("欢迎回来！距离您上次登录已过去 %d 天。", daysSinceLastLogin))
		} else {
			welcomeMessages = append(welcomeMessages, "欢迎回来！")
		}
	} else {
		welcomeMessages = append(welcomeMessages, "欢迎使用！")
	}

	// 积分信息
	welcomeMessages = append(welcomeMessages, fmt.Sprintf("当前积分余额：%d", user.Credits))

	// VIP等级信息
	role := UserRole(user.Role)
	switch role {
	case UserRoleVIP:
		if user.VipLevel > 0 {
			welcomeMessages = append(welcomeMessages, fmt.Sprintf("VIP等级：%d级，最多支持 %d 个设备同时在线", user.VipLevel, maxSessions))
		}
	case UserRoleAdmin:
		welcomeMessages = append(welcomeMessages, "管理员账号，无设备数量限制")
	default:
		welcomeMessages = append(welcomeMessages, fmt.Sprintf("普通用户，最多支持 %d 个设备同时在线", maxSessions))
	}

	// 使用统计
	if user.UsageCount > 0 {
		welcomeMessages = append(welcomeMessages, fmt.Sprintf("累计使用次数：%d 次", user.UsageCount))
	}

	// 如果清理了旧会话，添加提示
	if deletedCount > 0 {
		welcomeMessages = append(welcomeMessages, fmt.Sprintf("已自动清理 %d 个旧登录会话，确保账户安全。", deletedCount))
	}

	// 使用建议
	if user.Credits < 100 {
		welcomeMessages = append(welcomeMessages, "💡 提示：积分不足时可通过充值或邀请好友获得更多积分。")
	}

	content := strings.Join(welcomeMessages, "\n")

	// 创建系统通知
	notification := &models.SystemNotification{
		NotificationID: s.generateNotificationID(),
		UserID:         tools.StringPtr(user.UserID),
		Type:           "system",
		Title:          "登录成功",
		Content:        content,
		IsImportant:    false,
		Status:         "unread",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.Create(notification).Error; err != nil {
		repository.Errorf("发送登录成功通知失败: %v", err)
	}
}

// DeleteAccount 注销用户账号
func (s *UserService) DeleteAccount(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 管理员不能注销
	if UserRole(user.Role) == UserRoleAdmin {
		return fmt.Errorf("管理员账号不能注销")
	}

	// 删除用户（软删除）
	return s.userRepo.Delete(userID)
}

// BindPhone 绑定手机号
func (s *UserService) BindPhone(phone, identifier string) (*models.User, error) {
	// 查找是否已有该手机号的用户
	existingUser, err := s.userRepo.GetByPhone(phone)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check phone: %w", err)
	}

	if existingUser != nil {
		// 手机号已存在
		if existingUser.OpenID != nil && *existingUser.OpenID == identifier {
			return nil, fmt.Errorf("当前微信已绑定过该手机号")
		}
		if existingUser.OpenID != nil && *existingUser.OpenID != identifier {
			return nil, fmt.Errorf("手机号已绑定其他账号")
		}
		// 绑定openid到已有用户
		existingUser.OpenID = tools.StringPtr(identifier)
		if err := s.userRepo.Update(existingUser); err != nil {
			return nil, fmt.Errorf("failed to bind phone: %w", err)
		}
		return existingUser, nil
	}

	// 创建新用户
	initialCredits := s.GetInitialCredits()
	user := &models.User{
		UserID:           s.generateUserID(),
		Phone:            tools.StringPtr(phone),
		OpenID:           tools.StringPtr(identifier),
		Credits:          initialCredits,
		RegistrationDate: time.Now(),
		LastLoginTime:    time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	db := repository.GetDB()
	if err := db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// CheckPhone 检查手机号是否可绑定
func (s *UserService) CheckPhone(phone, identifier string) error {
	user, err := s.userRepo.GetByPhone(phone)
	if err == gorm.ErrRecordNotFound {
		return nil // 手机号未被使用，可以绑定
	}
	if err != nil {
		return fmt.Errorf("failed to check phone: %w", err)
	}

	// 手机号已存在
	if user.OpenID != nil && *user.OpenID == identifier {
		return fmt.Errorf("当前账号已绑定手机号")
	}
	if user.OpenID != nil && *user.OpenID != identifier {
		return fmt.Errorf("手机号已绑定其他账号")
	}

	return nil
}

// CheckEmail 检查邮箱是否可绑定
func (s *UserService) CheckEmail(email, identifier string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err == gorm.ErrRecordNotFound {
		return nil // 邮箱未被使用，可以绑定
	}
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}

	// 邮箱已存在
	if user.OpenID != nil && *user.OpenID == identifier {
		return fmt.Errorf("当前账号已绑定邮箱")
	}

	return fmt.Errorf("邮箱已绑定其他账号")
}

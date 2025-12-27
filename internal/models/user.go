package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型 - 匹配现有数据库结构
type User struct {
	UserID           string          `json:"user_id" gorm:"primaryKey;column:user_id;type:varchar(50)" binding:"required"`
	Nickname         *string         `json:"nickname" gorm:"column:nickname;type:varchar(50)"`
	Avatar           *string         `json:"avatar" gorm:"column:avatar;type:varchar(255)"`
	Username         *string         `json:"username" gorm:"column:username;type:varchar(50);uniqueIndex"`
	PasswordHash     *string         `json:"-" gorm:"column:password_hash;type:longtext"`
	AppID            *string         `json:"appid" gorm:"column:appid;type:varchar(50)"`
	OpenID           *string         `json:"openid" gorm:"column:openid;type:varchar(100)"`
	Phone            *string         `json:"phone" gorm:"column:phone;type:varchar(20)"`
	Email            *string         `json:"email" gorm:"column:email;type:varchar(100);uniqueIndex"`
	UtmSource        *string         `json:"utm_source" gorm:"column:utm_source;type:varchar(20);default:'direct'" description:"用户来源渠道"`
	Credits          int             `json:"credits" gorm:"column:credits;default:0"`
	IsActive         bool            `json:"is_active" gorm:"column:is_active;default:true"`
	VipLevel         int             `json:"vip_level" gorm:"column:vip_level;default:0"`
	Role             int16           `json:"role" gorm:"column:role;default:1"`
	Status           int16           `json:"status" gorm:"column:status;default:1"`
	RegistrationDate time.Time       `json:"registration_date" gorm:"column:registration_date"`
	TotalConsumption *float64        `json:"total_consumption" gorm:"column:total_consumption;type:decimal(10,2)"`
	LastLoginTime    time.Time       `json:"last_login_time" gorm:"column:last_login_time"`
	UsageCount       int             `json:"usage_count" gorm:"column:usage_count;default:0"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt        *gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserSession 用户会话模型
type UserSession struct {
	ID             int       `json:"id" gorm:"primaryKey;column:id"`
	UserID         string    `json:"user_id" gorm:"column:user_id;type:varchar(50)"`
	Token          string    `json:"token" gorm:"column:token;type:longtext"`
	LoginType      string    `json:"login_type" gorm:"column:login_type;type:varchar(20);default:'web'"`
	IPAddress      string    `json:"ip_address" gorm:"column:ip_address;type:varchar(45)"`
	DeviceID       *string   `json:"device_id" gorm:"column:device_id;type:varchar(100)"`
	Status         int16     `json:"status" gorm:"column:status;default:1"`
	LoginTime      time.Time `json:"login_time" gorm:"column:login_time"`
	ExpiresAt      time.Time `json:"expires_at" gorm:"column:expires_at"`
	IsActive       bool      `json:"is_active" gorm:"column:is_active;default:true"`
	LastActiveTime time.Time `json:"last_active_time" gorm:"column:last_active_time"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
}

// UserParameters 用户参数模型
type UserParameters struct {
	ParamID             string    `json:"param_id" gorm:"primaryKey;column:param_id;type:varchar(50)"`
	UserID              string    `json:"user_id" gorm:"column:user_id;type:varchar(50)"`
	EnableHeadInfo      bool      `json:"enable_head_info" gorm:"column:enable_head_info;default:false"`
	EnableKnowledgeBase bool      `json:"enable_knowledge_base" gorm:"column:enable_knowledge_base;default:false"`
	DefaultTheme        string    `json:"default_theme" gorm:"column:default_theme;type:varchar(50);default:'countryside'"`
	IsGzhBind           bool      `json:"is_gzh_bind" gorm:"column:is_gzh_bind;default:false"`
	IsWechatAuthorized  bool      `json:"is_wechat_authorized" gorm:"column:is_wechat_authorized;default:false"`
	PublishTarget       int       `json:"publish_target" gorm:"column:publish_target"`
	HasAuthReminded     bool      `json:"has_auth_reminded" gorm:"column:has_auth_reminded;default:false"`
	QrcodeData          *string   `json:"qrcode_data" gorm:"column:qrcode_data;type:json"`
	StorageQuota        int64     `json:"storage_quota" gorm:"column:storage_quota;default:314572800" description:"资源存储空间配额（字节），默认300MB"`
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// 请求和响应结构
type UserLoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
}

type UserUpdateRequest struct {
	Nickname         string   `json:"nickname" binding:"omitempty,max=50"`
	Email            string   `json:"email" binding:"omitempty,email"`
	Avatar           string   `json:"avatar"`
	Phone            string   `json:"phone" binding:"omitempty,max=20"`
	TotalConsumption *float64 `json:"total_consumption"`
}

type UserResponse struct {
	UserID           string   `json:"user_id"`
	Username         *string  `json:"username"`
	Nickname         *string  `json:"nickname"`
	Email            *string  `json:"email"`
	Avatar           *string  `json:"avatar"`
	Phone            *string  `json:"phone"`
	Credits          int      `json:"credits"`
	IsActive         bool     `json:"is_active"`
	VipLevel         int      `json:"vip_level"`
	Role             int16    `json:"role"`
	Status           int16    `json:"status"`
	RegistrationDate string   `json:"registration_date"`
	LastLoginTime    string   `json:"last_login_time"`
	UsageCount       int      `json:"usage_count"`
	TotalConsumption *float64 `json:"total_consumption"`
}

// 表名设置
func (User) TableName() string {
	return "user"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (UserParameters) TableName() string {
	return "user_parameters"
}

// HashPassword 密码加密
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hashedPassword)
	u.PasswordHash = &hashStr
	return nil
}

// CheckPassword 检查密码
func (u *User) CheckPassword(password string) bool {
	if u.PasswordHash == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password))
	return err == nil
}

// UserPreference 用户偏好表模型 - 存储reflection节点返回的信息
type UserPreference struct {
	PreferenceID     string    `json:"preference_id" gorm:"primaryKey;column:preference_id;type:varchar(50)" description:"偏好ID"`
	UserID           string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"用户ID"`
	StyleRules       *string   `json:"style_rules" gorm:"column:style_rules;type:longtext" description:"风格规则和指导方针"`
	UserProfile      *string   `json:"user_profile" gorm:"column:user_profile;type:longtext" description:"用户画像信息"`
	SystemPrompt     *string   `json:"system_prompt" gorm:"column:system_prompt;type:longtext" description:"系统提示词"`
	IsActive         bool      `json:"is_active" gorm:"column:is_active;default:true" description:"是否激活"`
	IsUserPreference bool      `json:"is_user_preference" gorm:"column:is_user_preference;default:true" description:"是否开启用户偏好"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// UserAuthorization 用户第三方平台授权表模型
type UserAuthorization struct {
	AuthID           string    `json:"auth_id" gorm:"primaryKey;column:auth_id;type:varchar(50)" description:"授权ID"`
	UserID           string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Platform         string    `json:"platform" gorm:"column:platform;type:varchar(20);not null;index" description:"平台类型(wechat/xiaohongshu/douyin/toutiao/zhihu/weibo/bilibili/kuaishou/csdn)"`
	AppID            *string   `json:"appid" gorm:"column:appid;type:varchar(100)" description:"应用ID（如微信appid）"`
	OpenID           string    `json:"openid" gorm:"column:openid;type:varchar(200);not null" description:"用户在平台的唯一标识"`
	PlatformNickname *string   `json:"platform_nickname" gorm:"column:platform_nickname;type:varchar(100)" description:"平台昵称"`
	Status           string    `json:"status" gorm:"column:status;type:varchar(20);not null;default:'enabled';index" description:"授权状态(enabled/disabled)"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// UserPromiseVideoStatus 用户承诺视频状态枚举
type UserPromiseVideoStatus int16

const (
	PromiseVideoNormal  UserPromiseVideoStatus = 1 // 正常
	PromiseVideoExpired UserPromiseVideoStatus = 2 // 过期
)

// UserPromiseVideo 用户承诺视频表
type UserPromiseVideo struct {
	ID        int                    `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID    string                 `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	VideoPath *string                `json:"video_path" gorm:"column:video_path;type:varchar(256)" description:"视频存储路径"`
	Status    UserPromiseVideoStatus `json:"status" gorm:"column:status;type:smallint;default:1" description:"状态"`
	CreatedAt time.Time              `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt time.Time              `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// MaterialTypes 素材类型枚举
type MaterialTypes string

const (
	MaterialTypeText     MaterialTypes = "text"
	MaterialTypeImage    MaterialTypes = "image"
	MaterialTypeAudio    MaterialTypes = "audio"
	MaterialTypeVideo    MaterialTypes = "video"
	MaterialTypeGroup    MaterialTypes = "group"
	MaterialTypeDocument MaterialTypes = "document"
)

// UserMaterials 用户创作素材表
type UserMaterials struct {
	ID           int           `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID       string        `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Name         string        `json:"name" gorm:"column:name;type:varchar(50);not null" description:"素材自定义名称"`
	MaterialType MaterialTypes `json:"material_type" gorm:"column:material_type;type:varchar(20);not null;default:'image';index" description:"素材类型"`
	Data         *string       `json:"data" gorm:"column:data;type:json" description:"素材数据，包括如资源路径，或者其他信息"`
	Tags         *string       `json:"tags" gorm:"column:tags;type:json" description:"素材相关标签，分类属性"`
	IsPublic     int           `json:"is_public" gorm:"column:is_public;default:0" description:"0/1表示:否/是"`
	Size         int64         `json:"size" gorm:"column:size;default:0" description:"素材大小（字节）"`
	SortOrder    int           `json:"sort_order" gorm:"column:sort_order;default:0;index" description:"排序顺序，数字越小越靠前"`
	CreatedAt    time.Time     `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// FeedbackType 反馈类型枚举
type FeedbackType int16

const (
	FeedbackTypeNone        FeedbackType = 0 // 未操作
	FeedbackTypeSatisfied   FeedbackType = 1 // 满意
	FeedbackTypeUnsatisfied FeedbackType = 2 // 不满意
)

// UserFeedback 用户快速反馈日志表
type UserFeedback struct {
	FeedbackID   string       `json:"feedback_id" gorm:"primaryKey;column:feedback_id;type:varchar(50)" description:"反馈ID"`
	UserID       string       `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	FeedbackType FeedbackType `json:"feedback_type" gorm:"column:feedback_type;type:smallint;not null;index" description:"反馈类型：0-未操作，1-满意，2-不满意"`
	FunctionName *string      `json:"function_name" gorm:"column:function_name;type:varchar(100)" description:"功能描述标识"`
	Suggestion   *string      `json:"suggestion" gorm:"column:suggestion;type:longtext" description:"建议说明"`
	CreatedAt    time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// Distributor 分销商/合作方信息表 - 仅当用户role为DISTRIBUTOR时使用
type Distributor struct {
	DistributorID  string    `json:"distributor_id" gorm:"primaryKey;column:distributor_id;type:varchar(50)" description:"分销商ID"`
	UserID         string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;uniqueIndex" description:"关联用户ID"`
	CommissionRate float64   `json:"commission_rate" gorm:"column:commission_rate;type:decimal(5,4);default:0.2" description:"佣金比例，默认0.2(20%)"`
	ExtraParams    *string   `json:"extra_params" gorm:"column:extra_params;type:json" description:"额外参数，JSON格式存储，用于扩展业务需求"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// NotificationUserStatus 通知用户状态枚举
type NotificationUserStatus string

const (
	NotificationUserStatusRead    NotificationUserStatus = "read"    // 已读
	NotificationUserStatusDeleted NotificationUserStatus = "deleted" // 已删除
)

// NotificationUserRecord 系统通知用户状态记录表
type NotificationUserRecord struct {
	RecordID       string                 `json:"record_id" gorm:"primaryKey;column:record_id;type:varchar(50)" description:"记录ID"`
	NotificationID string                 `json:"notification_id" gorm:"column:notification_id;type:varchar(50);not null;index" description:"通知ID"`
	UserID         string                 `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"用户ID"`
	Status         NotificationUserStatus `json:"status" gorm:"column:status;type:varchar(20);not null;default:'read'" description:"用户操作状态"`
	ReadTime       *time.Time             `json:"read_time" gorm:"column:read_time" description:"阅读时间"`
	DeletedTime    *time.Time             `json:"deleted_time" gorm:"column:deleted_time" description:"删除时间"`
	CreatedAt      time.Time              `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt      time.Time              `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	Notification *SystemNotification `json:"notification,omitempty" gorm:"-"`
	User         *User               `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (UserPreference) TableName() string {
	return "user_preferences"
}

func (UserAuthorization) TableName() string {
	return "user_authorizations"
}

func (UserPromiseVideo) TableName() string {
	return "user_promise_videos"
}

func (UserMaterials) TableName() string {
	return "user_materials"
}

func (UserFeedback) TableName() string {
	return "user_feedbacks"
}

func (Distributor) TableName() string {
	return "distributors"
}

func (NotificationUserRecord) TableName() string {
	return "notification_user_record"
}

// 响应结构
type UserPreferenceResponse struct {
	PreferenceID     string  `json:"preference_id"`
	StyleRules       *string `json:"style_rules"`
	UserProfile      *string `json:"user_profile"`
	SystemPrompt     *string `json:"system_prompt"`
	IsActive         bool    `json:"is_active"`
	IsUserPreference bool    `json:"is_user_preference"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type UserAuthorizationResponse struct {
	AuthID           string  `json:"auth_id"`
	Platform         string  `json:"platform"`
	AppID            *string `json:"appid"`
	OpenID           string  `json:"openid"`
	PlatformNickname *string `json:"platform_nickname"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type UserMaterialsResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	MaterialType string `json:"material_type"`
	IsPublic     int    `json:"is_public"`
	Size         int64  `json:"size"`
	SortOrder    int    `json:"sort_order"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type UserFeedbackResponse struct {
	FeedbackID   string  `json:"feedback_id"`
	FeedbackType int16   `json:"feedback_type"`
	FunctionName *string `json:"function_name"`
	Suggestion   *string `json:"suggestion"`
	CreatedAt    string  `json:"created_at"`
}

type DistributorResponse struct {
	DistributorID  string  `json:"distributor_id"`
	CommissionRate float64 `json:"commission_rate"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ToResponse 转换为响应结构
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		UserID:           u.UserID,
		Username:         u.Username,
		Nickname:         u.Nickname,
		Email:            u.Email,
		Avatar:           u.Avatar,
		Phone:            u.Phone,
		Credits:          u.Credits,
		IsActive:         u.IsActive,
		VipLevel:         u.VipLevel,
		Role:             u.Role,
		Status:           u.Status,
		RegistrationDate: u.RegistrationDate.Format("2006-01-02 15:04:05"),
		LastLoginTime:    u.LastLoginTime.Format("2006-01-02 15:04:05"),
		UsageCount:       u.UsageCount,
		TotalConsumption: u.TotalConsumption,
	}
}

func (up *UserPreference) ToResponse() UserPreferenceResponse {
	return UserPreferenceResponse{
		PreferenceID:     up.PreferenceID,
		StyleRules:       up.StyleRules,
		UserProfile:      up.UserProfile,
		SystemPrompt:     up.SystemPrompt,
		IsActive:         up.IsActive,
		IsUserPreference: up.IsUserPreference,
		CreatedAt:        up.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        up.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (ua *UserAuthorization) ToResponse() UserAuthorizationResponse {
	return UserAuthorizationResponse{
		AuthID:           ua.AuthID,
		Platform:         ua.Platform,
		AppID:            ua.AppID,
		OpenID:           ua.OpenID,
		PlatformNickname: ua.PlatformNickname,
		Status:           ua.Status,
		CreatedAt:        ua.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        ua.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (um *UserMaterials) ToResponse() UserMaterialsResponse {
	return UserMaterialsResponse{
		ID:           um.ID,
		Name:         um.Name,
		MaterialType: string(um.MaterialType),
		IsPublic:     um.IsPublic,
		Size:         um.Size,
		SortOrder:    um.SortOrder,
		CreatedAt:    um.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    um.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (uf *UserFeedback) ToResponse() UserFeedbackResponse {
	return UserFeedbackResponse{
		FeedbackID:   uf.FeedbackID,
		FeedbackType: int16(uf.FeedbackType),
		FunctionName: uf.FunctionName,
		Suggestion:   uf.Suggestion,
		CreatedAt:    uf.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (d *Distributor) ToResponse() DistributorResponse {
	return DistributorResponse{
		DistributorID:  d.DistributorID,
		CommissionRate: d.CommissionRate,
		CreatedAt:      d.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      d.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

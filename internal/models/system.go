package models

import (
	"time"
)

// SystemNotification 系统通知模型
type SystemNotification struct {
	NotificationID string     `json:"notification_id" gorm:"primaryKey;column:notification_id;type:varchar(50)" description:"通知ID"`
	UserID         *string    `json:"user_id" gorm:"column:user_id;type:varchar(50)" description:"接收用户ID，为空表示全体用户"`
	Type           string     `json:"type" gorm:"column:type;type:varchar(20);not null;default:'system'" description:"通知类型"`
	Title          string     `json:"title" gorm:"column:title;type:varchar(100);not null" description:"通知标题"`
	Content        string     `json:"content" gorm:"column:content;type:longtext;not null" description:"通知内容"`
	Link           *string    `json:"link" gorm:"column:link;type:varchar(255)" description:"相关链接"`
	IsImportant    bool       `json:"is_important" gorm:"column:is_important;default:false" description:"是否重要"`
	Status         string     `json:"status" gorm:"column:status;type:varchar(20);not null;default:'unread'" description:"通知状态"`
	ReadTime       *time.Time `json:"read_time" gorm:"column:read_time" description:"阅读时间"`
	ExpireTime     *time.Time `json:"expire_time" gorm:"column:expire_time" description:"过期时间"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// Feedback 用户反馈模型
type Feedback struct {
	FeedbackID  string    `json:"feedback_id" gorm:"primaryKey;column:feedback_id;type:varchar(50)" description:"反馈ID"`
	UserID      *string   `json:"user_id" gorm:"column:user_id;type:varchar(50)" description:"用户ID"`
	Type        string    `json:"type" gorm:"column:type;type:varchar(20);not null;default:'other'" description:"反馈类型"`
	Title       *string   `json:"title" gorm:"column:title;type:varchar(100)" description:"反馈标题"`
	Content     string    `json:"content" gorm:"column:content;type:longtext;not null" description:"反馈内容"`
	ContactInfo *string   `json:"contact_info" gorm:"column:contact_info;type:varchar(100)" description:"联系方式"`
	Images      *string   `json:"images" gorm:"column:images;type:json" description:"相关图片URL列表"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(20);not null;default:'pending'" description:"处理状态"`
	AdminReply  *string   `json:"admin_reply" gorm:"column:admin_reply;type:longtext" description:"管理员回复"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// 注意：Reservation 已在 books.go 中定义
// 注意：Category, Scene, ChatRecord, MarketingActivityPlan 已分别移到独立文件中

// 表名设置
func (SystemNotification) TableName() string {
	return "system_notification"
}

func (Feedback) TableName() string {
	return "feedback"
}

// 响应结构
type SystemNotificationResponse struct {
	NotificationID string  `json:"notification_id"`
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	Content        string  `json:"content"`
	Link           *string `json:"link"`
	IsImportant    bool    `json:"is_important"`
	Status         string  `json:"status"`
	ReadTime       *string `json:"read_time"`
	ExpireTime     *string `json:"expire_time"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ToResponse 转换方法
func (sn *SystemNotification) ToResponse() SystemNotificationResponse {
	resp := SystemNotificationResponse{
		NotificationID: sn.NotificationID,
		Type:           sn.Type,
		Title:          sn.Title,
		Content:        sn.Content,
		Link:           sn.Link,
		IsImportant:    sn.IsImportant,
		Status:         sn.Status,
		CreatedAt:      sn.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      sn.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if sn.ReadTime != nil {
		readTimeStr := sn.ReadTime.Format("2006-01-02 15:04:05")
		resp.ReadTime = &readTimeStr
	}
	if sn.ExpireTime != nil {
		expireTimeStr := sn.ExpireTime.Format("2006-01-02 15:04:05")
		resp.ExpireTime = &expireTimeStr
	}

	return resp
}

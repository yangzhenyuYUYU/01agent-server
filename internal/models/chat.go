package models

import (
	"time"
)

// MessageType 消息类型枚举
type MessageType string

const (
	MessageTypeUser      MessageType = "user"      // 用户消息
	MessageTypeAssistant MessageType = "assistant" // AI助手消息
	MessageTypeSystem    MessageType = "system"    // 系统消息
)

// MessageStatus 消息状态枚举
type MessageStatus string

const (
	MessageStatusPending    MessageStatus = "pending"    // 待处理
	MessageStatusProcessing MessageStatus = "processing" // 处理中
	MessageStatusCompleted  MessageStatus = "completed"  // 已完成
	MessageStatusFailed     MessageStatus = "failed"     // 失败
)

// ChatRecord 聊天记录模型
type ChatRecord struct {
	ID           string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"聊天记录ID"`
	SessionID    *string   `json:"session_id" gorm:"column:session_id;type:varchar(100);index" description:"会话ID"`
	MessageType  string    `json:"message_type" gorm:"column:message_type;type:varchar(20);not null" description:"消息类型"`
	Content      string    `json:"content" gorm:"column:content;type:longtext;not null" description:"消息内容"`
	Status       string    `json:"status" gorm:"column:status;type:varchar(20);not null;default:'completed'" description:"消息状态"`
	Metadata     *string   `json:"metadata" gorm:"column:metadata;type:json" description:"消息元数据(如工具调用信息等)"`
	Tokens       *string   `json:"tokens" gorm:"column:tokens;type:varchar(64)" description:"消耗tokens"`
	ModelVersion *string   `json:"model_version" gorm:"column:model_version;type:varchar(50);default:''" description:"AI模型版本"`
	ErrorLog     *string   `json:"error_log" gorm:"column:error_log;type:json" description:"错误日志详情"`
	UserID       string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"最后更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (ChatRecord) TableName() string {
	return "chat_records"
}


package models

import (
	"time"
)

// VoiceToneStatus 音色状态
type VoiceToneStatus string

const (
	VoiceToneStatusActive   VoiceToneStatus = "active"   // 启用
	VoiceToneStatusDisabled VoiceToneStatus = "disabled" // 禁用
)

// SourceStatus 合成状态
type SourceStatus string

const (
	SourceStatusDraft     SourceStatus = "draft"     // 草稿
	SourceStatusPending   SourceStatus = "pending"   // 待处理
	SourceStatusProcessing SourceStatus = "processing" // 处理中
	SourceStatusCompleted SourceStatus = "completed" // 完成
	SourceStatusFailed    SourceStatus = "failed"    // 失败
	SourceStatusDeleted   SourceStatus = "deleted"   // 已删除
)

// VoiceTone 音色模型
type VoiceTone struct {
	ID            int            `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	ToneName      string         `json:"tone_name" gorm:"column:tone_name;type:varchar(50);not null" description:"音色名称"`
	RecommendScene string        `json:"recommend_scene" gorm:"column:recommend_scene;type:varchar(50);not null" description:"推荐场景"`
	Language      string         `json:"language" gorm:"column:language;type:varchar(20);not null" description:"语言"`
	VoiceType     string         `json:"voice_type" gorm:"column:voice_type;type:varchar(50);uniqueIndex" description:"音色类型ID"`
	Status        VoiceToneStatus `json:"status" gorm:"column:status;type:varchar(20);default:'active'" description:"状态：active-启用，disabled-禁用"`
	SortOrder     int            `json:"sort_order" gorm:"column:sort_order;default:0" description:"排序序号"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// SourceSynthesisRecord 语音合成记录模型
type SourceSynthesisRecord struct {
	ID          int          `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID       string      `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	VoiceToneID int          `json:"voice_tone_id" gorm:"column:voice_tone_id;not null;index" description:"关联音色ID"`
	TextContent string       `json:"text_content" gorm:"column:text_content;type:longtext;not null" description:"合成文本内容"`
	Duration    *float64     `json:"duration" gorm:"column:duration;default:0" description:"合成时长"`
	AudioURL    *string      `json:"audio_url" gorm:"column:audio_url;type:varchar(255)" description:"合成音频URL"`
	Status      SourceStatus `json:"status" gorm:"column:status;type:varchar(20);default:'draft'" description:"状态：draft-草稿，pending-待处理，processing-处理中，completed-完成，failed-失败，deleted-已删除"`
	ErrorMsg    *string      `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	CreatedAt   time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User      *User      `json:"user,omitempty" gorm:"-"`
	VoiceTone *VoiceTone `json:"voice_tone,omitempty" gorm:"-"`
}

// 表名设置
func (VoiceTone) TableName() string {
	return "voice_tone"
}

func (SourceSynthesisRecord) TableName() string {
	return "source_synthesis_records"
}


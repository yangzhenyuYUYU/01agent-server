package digital

import (
	"gin_web/internal/models"
	"time"
)

// TranslationStatus 翻译状态枚举
type TranslationStatus int16

const (
	TranslationStatusPending    TranslationStatus = 0  // 等待中
	TranslationStatusProcessing TranslationStatus = 1  // 处理中
	TranslationStatusSuccess    TranslationStatus = 2  // 成功
	TranslationStatusFailed     TranslationStatus = -1 // 失败
)

// TranslationRecord AI翻译记录模型
type TranslationRecord struct {
	ID             int               `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID         string            `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"用户ID"`
	SourceText     string            `json:"source_text" gorm:"column:source_text;type:longtext;not null" description:"原文内容"`
	TargetText     *string           `json:"target_text" gorm:"column:target_text;type:longtext" description:"翻译结果"`
	SourceLanguage string            `json:"source_language" gorm:"column:source_language;type:varchar(10);not null" description:"源语言"`
	TargetLanguage string            `json:"target_language" gorm:"column:target_language;type:varchar(10);not null" description:"目标语言"`
	Model          string            `json:"model" gorm:"column:model;type:varchar(50);not null" description:"使用的AI模型"`
	TokensUsed     *int              `json:"tokens_used" gorm:"column:tokens_used" description:"消耗的token数量"`
	ResponseTime   *float64          `json:"response_time" gorm:"column:response_time" description:"响应时间(秒)"`
	Status         TranslationStatus `json:"status" gorm:"column:status;type:smallint;default:0" description:"翻译状态"`
	ErrorMessage   *string           `json:"error_message" gorm:"column:error_message;type:longtext" description:"错误信息"`
	CreatedAt      time.Time         `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt      time.Time         `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *models.User `json:"user,omitempty" gorm:"-"`
}

// BroadcastLengthType 文案长度类型
type BroadcastLengthType string

const (
	BroadcastLengthTypeShort  BroadcastLengthType = "short"  // 短篇，约50字
	BroadcastLengthTypeMedium BroadcastLengthType = "medium" // 中篇，约100字
	BroadcastLengthTypeLong   BroadcastLengthType = "long"   // 长篇，约500字
)

// BroadcastType 生成类型
type BroadcastType int16

const (
	BroadcastTypeSync   BroadcastType = 0 // 同步生成
	BroadcastTypeStream BroadcastType = 1 // 流式生成
)

// BroadcastRecord 口播文案生成记录模型
type BroadcastRecord struct {
	ID            int                 `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID        string              `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"用户ID"`
	PromptID      *int                `json:"prompt_id" gorm:"column:prompt_id;index" description:"使用的提示词ID"`
	LengthType    BroadcastLengthType `json:"length_type" gorm:"column:length_type;type:varchar(20);not null" description:"文案长度类型"`
	BroadcastType BroadcastType       `json:"broadcast_type" gorm:"column:broadcast_type;type:smallint;not null" description:"生成类型：0-同步，1-流式"`
	InputContent  string              `json:"input_content" gorm:"column:input_content;type:longtext;not null" description:"用户输入内容"`
	OutputContent *string             `json:"output_content" gorm:"column:output_content;type:longtext" description:"生成的文案内容"`
	Model         string              `json:"model" gorm:"column:model;type:varchar(50);not null" description:"使用的AI模型"`
	Tokens        int                 `json:"tokens" gorm:"column:tokens;default:0" description:"消耗的tokens"`
	Status        int16               `json:"status" gorm:"column:status;default:0" description:"状态：0-生成中，1-成功，2-失败"`
	ErrorMsg      *string             `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	CreatedAt     time.Time           `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time           `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User   *models.User   `json:"user,omitempty" gorm:"-"`
	Prompt *DigitalPrompt `json:"prompt,omitempty" gorm:"-"`
}

// 表名设置
func (TranslationRecord) TableName() string {
	return "translation_records"
}

func (BroadcastRecord) TableName() string {
	return "broadcast_records"
}

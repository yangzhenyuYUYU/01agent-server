package digital

import (
	"01agent_server/internal/models"
	"time"
)

// TemplateStatus 数字人模板状态
type DigitalTemplateStatus string

const (
	DigitalTemplateStatusPending    DigitalTemplateStatus = "pending"    // 等待中
	DigitalTemplateStatusProcessing DigitalTemplateStatus = "processing" // 处理中
	DigitalTemplateStatusCompleted  DigitalTemplateStatus = "completed"  // 完成
	DigitalTemplateStatusFailed     DigitalTemplateStatus = "failed"     // 失败
)

// DigitalTemplate 用户数字人模版模型
type DigitalTemplate struct {
	ID          int                   `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID      string                `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	MediaID     *string               `json:"media_id" gorm:"column:media_id;type:varchar(100)" description:"媒体ID"`
	DigitalID   *string               `json:"digital_id" gorm:"column:digital_id;type:varchar(100)" description:"数字人模版ID"`
	Name        *string               `json:"name" gorm:"column:name;type:varchar(100)" description:"模版名称"`
	Description *string               `json:"description" gorm:"column:description;type:longtext" description:"场景说明"`
	IsOpen      bool                  `json:"is_open" gorm:"column:is_open;default:false" description:"是否公开"`
	ModelPath   *string               `json:"model_path" gorm:"column:model_path;type:varchar(255)" description:"模型文件路径"`
	CategoryID  *string               `json:"category_id" gorm:"column:category_id;type:varchar(20)" description:"分类ID"`
	Thumbnail   *string               `json:"thumbnail" gorm:"column:thumbnail;type:varchar(255)" description:"缩略图"`
	Status      DigitalTemplateStatus `json:"status" gorm:"column:status;type:varchar(20);default:'pending'" description:"状态：pending-等待中，processing-处理中，completed-完成，failed-失败"`
	TaskID      *string               `json:"task_id" gorm:"column:task_id;type:varchar(100)" description:"任务ID"`
	Price       *float64              `json:"price" gorm:"column:price;type:decimal(10,2)" description:"模板价格"`
	ErrorMsg    *string               `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	Properties  *string               `json:"properties" gorm:"column:properties;type:json" description:"合成视频属性"`
	CreatedAt   time.Time             `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *models.User `json:"user,omitempty" gorm:"-"`
}

// SynthesisRecord 用户合成数字人记录模型
type SynthesisRecord struct {
	ID                int                   `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID            string                `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	DigitalTemplateID int                   `json:"digital_template_id" gorm:"column:digital_template_id;not null;index" description:"关联数字人模板ID"`
	MediaID           *string               `json:"media_id" gorm:"column:media_id;type:varchar(100)" description:"媒体ID"`
	AudioPath         string                `json:"audio_path" gorm:"column:audio_path;type:varchar(255);not null" description:"音频文件路径"`
	Duration          *int                  `json:"duration" gorm:"column:duration" description:"音频时长(秒)"`
	Name              string                `json:"name" gorm:"column:name;type:varchar(100);not null" description:"合成名称"`
	Description       *string               `json:"description" gorm:"column:description;type:longtext" description:"场景说明"`
	TextContent       *string               `json:"text_content" gorm:"column:text_content;type:longtext" description:"合成文本内容"`
	ResultPath        *string               `json:"result_path" gorm:"column:result_path;type:varchar(255)" description:"结果视频路径"`
	Status            DigitalTemplateStatus `json:"status" gorm:"column:status;type:varchar(20);default:'pending'" description:"状态：pending-等待中，processing-处理中，completed-完成，failed-失败"`
	Properties        *string               `json:"properties" gorm:"column:properties;type:json" description:"合成视频属性"`
	TaskID            *string               `json:"task_id" gorm:"column:task_id;type:varchar(100)" description:"任务ID"`
	ErrorMsg          *string               `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	CreatedAt         time.Time             `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt         time.Time             `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User            *models.User     `json:"user,omitempty" gorm:"-"`
	DigitalTemplate *DigitalTemplate `json:"digital_template,omitempty" gorm:"-"`
}

// 表名设置
func (DigitalTemplate) TableName() string {
	return "digital_templates"
}

func (SynthesisRecord) TableName() string {
	return "synthesis_records"
}

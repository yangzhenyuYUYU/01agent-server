package digital

import (
	"01agent_server/internal/models"
	"time"
)

// VoiceTrainStatus 声音训练状态
type VoiceTrainStatus string

const (
	VoiceTrainStatusDraft     VoiceTrainStatus = "draft"     // 草稿
	VoiceTrainStatusRecording VoiceTrainStatus = "recording" // 录制中
	VoiceTrainStatusSubmitted VoiceTrainStatus = "submitted" // 已提交
	VoiceTrainStatusTraining  VoiceTrainStatus = "training"  // 训练中
	VoiceTrainStatusCompleted VoiceTrainStatus = "completed" // 完成
	VoiceTrainStatusFailed    VoiceTrainStatus = "failed"    // 失败
)

// VoiceTrainProvider 声音训练提供商
type VoiceTrainProvider string

const (
	VoiceTrainProviderVolc VoiceTrainProvider = "volc" // 火山引擎
	VoiceTrainProviderXF   VoiceTrainProvider = "xf"   // 讯飞
)

// VoiceTrainTask 声音训练任务模型
type VoiceTrainTask struct {
	ID           int              `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Provider     *string          `json:"provider" gorm:"column:provider;type:varchar(20)" description:"提供商: volc-火山引擎  xf-讯飞"`
	UserID       string           `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	IsOpen       bool             `json:"is_open" gorm:"column:is_open;default:false" description:"是否公开"`
	CategoryID   *string          `json:"category_id" gorm:"column:category_id;type:varchar(20)" description:"分类ID"`
	Price        *float64         `json:"price" gorm:"column:price;type:decimal(10,2)" description:"模板价格"`
	TaskName     string           `json:"task_name" gorm:"column:task_name;type:varchar(100);not null" description:"任务名称"`
	TaskID       *string          `json:"task_id" gorm:"column:task_id;type:varchar(100)" description:"讯飞平台任务ID"`
	ResourceName string           `json:"resource_name" gorm:"column:resource_name;type:varchar(100);not null" description:"音库名称"`
	Sex          int              `json:"sex" gorm:"column:sex;not null" description:"性别: 1-男 2-女 0-未知"`
	AgeGroup     int              `json:"age_group" gorm:"column:age_group;not null" description:"年龄组: 1-儿童 2-青年 3-中年 4-中老年 0-未知"`
	Language     string           `json:"language" gorm:"column:language;type:varchar(10);not null;default:''" description:"训练语种"`
	Version      *int             `json:"version" gorm:"column:version" description:"火山引擎训练版本，如V1、V2等"`
	Status       VoiceTrainStatus `json:"status" gorm:"column:status;type:varchar(20);default:'draft'" description:"状态：draft-草稿，recording-录制中，submitted-已提交，training-训练中，completed-完成，failed-失败"`
	TrainVID     *string          `json:"train_vid" gorm:"column:train_vid;type:varchar(100)" description:"训练后音库ID"`
	AssetID      *string          `json:"asset_id" gorm:"column:asset_id;type:varchar(100)" description:"训练后音色ID"`
	ErrorMsg     *string          `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	CreatedAt    time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt    time.Time        `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *models.User `json:"user,omitempty" gorm:"-"`
}

// VoiceTrainAudio 声音训练音频模型
type VoiceTrainAudio struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	TrainTaskID int       `json:"train_task_id" gorm:"column:train_task_id;not null;index" description:"关联训练任务ID"`
	TextID      *int      `json:"text_id" gorm:"column:text_id" description:"训练文本ID"`
	TextSegID   *int      `json:"text_seg_id" gorm:"column:text_seg_id" description:"文本段落ID"`
	TextContent *string   `json:"text_content" gorm:"column:text_content;type:longtext" description:"文本内容"`
	AudioURL    string    `json:"audio_url" gorm:"column:audio_url;type:varchar(255);not null" description:"音频文件URL"`
	Duration    *float64  `json:"duration" gorm:"column:duration" description:"音频时长(秒)"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(20);default:'pending'" description:"状态：draft-草稿，recording-录制中，submitted-已提交，training-训练中，completed-完成，failed-失败"`
	ErrorMsg    *string   `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	TrainTask *VoiceTrainTask `json:"train_task,omitempty" gorm:"-"`
}

// VoiceModel 声音模型模型
type VoiceModel struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID      string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	TrainTaskID int       `json:"train_task_id" gorm:"column:train_task_id;not null;index" description:"关联训练任务ID"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(100);not null" description:"模型名称"`
	Description *string   `json:"description" gorm:"column:description;type:longtext" description:"模型描述"`
	TrainVID    *string   `json:"train_vid" gorm:"column:train_vid;type:varchar(100)" description:"音库ID"`
	AssetID     *string   `json:"asset_id" gorm:"column:asset_id;type:varchar(100)" description:"音色ID"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(20);default:'active'" description:"状态：active-启用，inactive-停用"`
	IsDefault   bool      `json:"is_default" gorm:"column:is_default;default:false" description:"是否默认音色"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User      *models.User    `json:"user,omitempty" gorm:"-"`
	TrainTask *VoiceTrainTask `json:"train_task,omitempty" gorm:"-"`
}

// VoiceSynthesisRecord 声音合成记录模型
type VoiceSynthesisRecord struct {
	ID           int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID       string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	VoiceModelID int       `json:"voice_model_id" gorm:"column:voice_model_id;not null;index" description:"关联声音模型ID"`
	TextContent  string    `json:"text_content" gorm:"column:text_content;type:longtext;not null" description:"合成文本内容"`
	AudioURL     *string   `json:"audio_url" gorm:"column:audio_url;type:varchar(255)" description:"合成音频URL"`
	Duration     *float64  `json:"duration" gorm:"column:duration" description:"音频时长(秒)"`
	Volume       int       `json:"volume" gorm:"column:volume;default:50" description:"音量(0-100)"`
	Speed        int       `json:"speed" gorm:"column:speed;default:50" description:"语速(0-100)"`
	Pitch        int       `json:"pitch" gorm:"column:pitch;default:50" description:"语调(0-100)"`
	IsOpen       bool      `json:"is_open" gorm:"column:is_open;default:false" description:"是否公开"`
	CategoryID   *string   `json:"category_id" gorm:"column:category_id;type:varchar(20)" description:"分类ID"`
	Status       string    `json:"status" gorm:"column:status;type:varchar(20);default:'pending'" description:"状态：pending-处理中，completed-完成，failed-失败"`
	ErrorMsg     *string   `json:"error_msg" gorm:"column:error_msg;type:longtext" description:"错误信息"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User       *models.User `json:"user,omitempty" gorm:"-"`
	VoiceModel *VoiceModel  `json:"voice_model,omitempty" gorm:"-"`
}

// 表名设置
func (VoiceTrainTask) TableName() string {
	return "voice_train_tasks"
}

func (VoiceTrainAudio) TableName() string {
	return "voice_train_audios"
}

func (VoiceModel) TableName() string {
	return "voice_models"
}

func (VoiceSynthesisRecord) TableName() string {
	return "voice_synthesis_records"
}

package conversation

import (
	"gin_web/internal/models"
	"time"
)

// ConversationMessage 单条对话消息记录模型
type ConversationMessage struct {
	ID               int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	SceneID          string    `json:"scene_id" gorm:"column:scene_id;type:varchar(64);not null;index" description:"场景ID"`
	UserID           string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Mode             string    `json:"mode" gorm:"column:mode;type:varchar(32);not null" description:"模式：normal_reply | product_json"`
	InputText        string    `json:"input_text" gorm:"column:input_text;type:longtext;not null" description:"输入文本"`
	OutputText       *string   `json:"output_text" gorm:"column:output_text;type:longtext" description:"输出文本"`
	JsonData         *string   `json:"json_data" gorm:"column:json_data;type:json" description:"JSON数据"`
	Status           string    `json:"status" gorm:"column:status;type:varchar(16);default:'doing'" description:"状态：ready | doing | done | failed"`
	TokensPrompt     int       `json:"tokens_prompt" gorm:"column:tokens_prompt;default:0" description:"提示词tokens"`
	TokensCompletion int       `json:"tokens_completion" gorm:"column:tokens_completion;default:0" description:"完成tokens"`
	TokensTotal      int       `json:"tokens_total" gorm:"column:tokens_total;default:0" description:"总tokens"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`

	// 关联关系
	Scene *ConversationScene `json:"scene,omitempty" gorm:"-"`
	User  *models.User       `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (ConversationMessage) TableName() string {
	return "conv_messages"
}

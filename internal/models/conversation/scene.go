package conversation

import (
	"01agent_server/internal/models"
	"time"
)

// ConversationScene 对话场景容器模型
type ConversationScene struct {
	SceneID   string    `json:"scene_id" gorm:"primaryKey;column:scene_id;type:varchar(64)" description:"场景ID"`
	UserID    string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Title     *string   `json:"title" gorm:"column:title;type:varchar(255)" description:"标题"`
	Summary   *string   `json:"summary" gorm:"column:summary;type:longtext" description:"摘要"`
	IsActive  bool      `json:"is_active" gorm:"column:is_active;default:true" description:"是否活跃"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User     *models.User          `json:"user,omitempty" gorm:"-"`
	Messages []ConversationMessage `json:"messages,omitempty" gorm:"-"`
}

// 表名设置
func (ConversationScene) TableName() string {
	return "conv_scenes"
}

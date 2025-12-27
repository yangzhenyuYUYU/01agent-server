package models

import (
	"time"
)

// Scene 场景模型
type Scene struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"场景ID"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(128);not null" description:"场景名称"`
	Prompt    *string   `json:"prompt" gorm:"column:prompt;type:longtext" description:"提示词"`
	IsActive  bool      `json:"is_active" gorm:"column:is_active;default:true" description:"是否启用"`
	Order     int       `json:"order" gorm:"column:order;not null;default:0" description:"排序"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// 表名设置
func (Scene) TableName() string {
	return "scenes"
}

// SceneResponse 响应结构
type SceneResponse struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Prompt    *string `json:"prompt"`
	IsActive  bool    `json:"is_active"`
	Order     int     `json:"order"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// ToResponse 转换方法
func (s *Scene) ToResponse() SceneResponse {
	return SceneResponse{
		ID:        s.ID,
		Name:      s.Name,
		Prompt:    s.Prompt,
		IsActive:  s.IsActive,
		Order:     s.Order,
		CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}


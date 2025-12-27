package models

import (
	"time"
)

// PromptTemplateStatus 提示词模板状态枚举
type PromptTemplateStatus int16

const (
	PromptTemplateStatusActive   PromptTemplateStatus = 1 // 启用
	PromptTemplateStatusInactive PromptTemplateStatus = 0 // 禁用
)

// UserPromptTemplate 用户自定义提示词模板模型
type UserPromptTemplate struct {
	ID          int                   `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID      string                `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Name        string                `json:"name" gorm:"column:name;type:varchar(100);not null" description:"模板名称"`
	Description *string               `json:"description" gorm:"column:description;type:varchar(500)" description:"模板描述"`
	Data        string                `json:"data" gorm:"column:data;type:json;not null" description:"提示词数据(JSON格式)"`
	Status      PromptTemplateStatus  `json:"status" gorm:"column:status;type:smallint;default:1" description:"状态"`
	IsDefault   bool                  `json:"is_default" gorm:"column:is_default;default:false" description:"是否为默认模板"`
	SortOrder   int                   `json:"sort_order" gorm:"column:sort_order;default:0" description:"排序顺序"`
	CreatedAt   time.Time             `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (UserPromptTemplate) TableName() string {
	return "user_prompt_templates"
}


package models

import (
	"time"
)

// CustomConfigStatus 自定义配置状态枚举
type CustomConfigStatus int16

const (
	CustomConfigStatusActive   CustomConfigStatus = 1 // 启用
	CustomConfigStatusInactive CustomConfigStatus = 0 // 禁用
)

// UserCustomSize 用户自定义尺寸模型
type UserCustomSize struct {
	ID        int                `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID    string             `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Name      string             `json:"name" gorm:"column:name;type:varchar(100);not null" description:"尺寸名称"`
	Data      string             `json:"data" gorm:"column:data;type:json;not null" description:"尺寸数据(JSON格式)"`
	Status    CustomConfigStatus `json:"status" gorm:"column:status;type:smallint;default:1" description:"状态"`
	IsDefault bool               `json:"is_default" gorm:"column:is_default;default:false" description:"是否为默认尺寸"`
	SortOrder int                `json:"sort_order" gorm:"column:sort_order;default:0" description:"排序顺序"`
	CreatedAt time.Time          `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time          `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// UserCustomTheme 用户自定义主题配色模型
type UserCustomTheme struct {
	ID        int                `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID    string             `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Name      string             `json:"name" gorm:"column:name;type:varchar(100);not null" description:"主题名称"`
	Data      string             `json:"data" gorm:"column:data;type:json;not null" description:"主题配色数据(JSON格式)"`
	Status    CustomConfigStatus `json:"status" gorm:"column:status;type:smallint;default:1" description:"状态"`
	IsDefault bool               `json:"is_default" gorm:"column:is_default;default:false" description:"是否为默认主题"`
	SortOrder int                `json:"sort_order" gorm:"column:sort_order;default:0" description:"排序顺序"`
	CreatedAt time.Time          `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time          `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (UserCustomSize) TableName() string {
	return "user_custom_sizes"
}

func (UserCustomTheme) TableName() string {
	return "user_custom_themes"
}


package models

import (
	"time"
)

// Status 状态枚举
type DigitalStatus string

const (
	DigitalStatusActive   DigitalStatus = "active"   // 启用
	DigitalStatusInactive DigitalStatus = "inactive" // 禁用
)

// DigitalCategory 数字人分类模型
type DigitalCategory struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(50);not null" description:"分类名称"`
	Key       string    `json:"key" gorm:"column:key;type:varchar(50);not null" description:"分类键值，例如：1-1, 1-1-2"`
	Position  int       `json:"position" gorm:"column:position;default:0" description:"位置"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// DigitalCountry 数字人国家模型
type DigitalCountry struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Thumbnail   *string   `json:"thumbnail" gorm:"column:thumbnail;type:varchar(255)" description:"国家缩略图"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(100);not null" description:"国家名称"`
	EnglishName string    `json:"english_name" gorm:"column:english_name;type:varchar(100);not null" description:"英文名称"`
	Language    string    `json:"language" gorm:"column:language;type:varchar(50);not null" description:"主要使用语言"`
	LanguageCode string   `json:"language_code" gorm:"column:language_code;type:varchar(10);not null" description:"语言代码，例如：zh-CN, en-US"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(20);default:'active'" description:"可用状态，例如：active, inactive"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// DigitalPrompt 数字人提示词模型
type DigitalPrompt struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Category  string    `json:"category" gorm:"column:category;type:varchar(50);not null" description:"分类"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(100);not null" description:"提示词名称"`
	Content   *string   `json:"content" gorm:"column:content;type:longtext" description:"提示词描述"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// 表名设置
func (DigitalCategory) TableName() string {
	return "digital_categories"
}

func (DigitalCountry) TableName() string {
	return "digital_countries"
}

func (DigitalPrompt) TableName() string {
	return "digital_prompts"
}


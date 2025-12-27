package models

import (
	"time"
)

// VersionType 版本类型枚举
type VersionType string

const (
	VersionTypeMajor VersionType = "major" // 主版本
	VersionTypeMinor VersionType = "minor" // 次版本
	VersionTypePatch VersionType = "patch" // 补丁版本
)

// Version 版本迭代模型
// 用于记录 01Editor Web 的版本历史，方便在前端展示版本更新记录
type Version struct {
	ID         int         `json:"id" gorm:"primaryKey;column:id" description:"主键ID"`
	Version    string      `json:"version" gorm:"column:version;type:varchar(32);not null" description:"版本号，如 v0.3.0"`
	Date       time.Time   `json:"date" gorm:"column:date;type:date;not null" description:"发布日期"`
	Title      string      `json:"title" gorm:"column:title;type:varchar(255);not null" description:"版本标题"`
	Highlights string      `json:"highlights" gorm:"column:highlights;type:json;not null" description:"高亮信息列表（字符串数组）"`
	Type       *VersionType `json:"type" gorm:"column:type;type:varchar(20)" description:"版本类型：major / minor / patch"`
	CreatedAt  time.Time   `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt  time.Time   `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// 表名设置
func (Version) TableName() string {
	return "versions"
}


package models

import (
	"time"
)

// Category 分类模型
type Category struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"分类ID"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(128);not null" description:"分类名称"`
	Key         string    `json:"key" gorm:"column:key;type:varchar(128);not null" description:"分类标识"`
	SceneID     int       `json:"scene_id" gorm:"column:scene_id;not null;index" description:"所属场景ID"`
	Order       int       `json:"order" gorm:"column:order;not null;default:0" description:"排序"`
	Description string    `json:"description" gorm:"column:description;type:longtext;not null" description:"分类描述"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	Scene *Scene `json:"scene,omitempty" gorm:"-"`
}

// 表名设置
func (Category) TableName() string {
	return "categories"
}

// CategoryResponse 响应结构
type CategoryResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	SceneID     int    `json:"scene_id"`
	Order       int    `json:"order"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToResponse 转换方法
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Key:         c.Key,
		SceneID:     c.SceneID,
		Order:       c.Order,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}


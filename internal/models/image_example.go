package models

import (
	"time"
)

// ImageExample 图文生成示例模型 - 轻量级主表，用于列表快速检索
type ImageExample struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(100);not null" description:"示例名称"`
	Prompt      string    `json:"prompt" gorm:"column:prompt;type:longtext;not null" description:"图文生成提示词"`
	CoverURL    *string   `json:"cover_url" gorm:"column:cover_url;type:varchar(500)" description:"封面图片URL"`
	Tags        *string   `json:"tags" gorm:"column:tags;type:json" description:"标签列表"`
	SortOrder   int       `json:"sort_order" gorm:"column:sort_order;default:0" description:"排序顺序"`
	IsVisible   bool      `json:"is_visible" gorm:"column:is_visible;default:true" description:"是否显示"`
	ExtraData   *string   `json:"extra_data" gorm:"column:extra_data;type:json" description:"额外数据，示例: {\"model\": \"doubao-1-5-vision-pro-32k-250115\", \"size\": \"1024x1024\"}"`
	ProjectType string    `json:"project_type" gorm:"column:project_type;type:varchar(50);default:'other'" description:"工程类型：xiaohongshu(小红书) 或 other(其他)"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// ImageExampleDetail 图文生成示例详情模型 - 存储大字段数据（node_data, jsx_code）
type ImageExampleDetail struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	ExampleID int       `json:"example_id" gorm:"column:example_id;not null;uniqueIndex" description:"关联的示例ID"`
	JsxCode   *string   `json:"jsx_code" gorm:"column:jsx_code;type:longtext" description:"JSX代码"`
	NodeData  *string   `json:"node_data" gorm:"column:node_data;type:json" description:"节点数据"`
	Images    *string   `json:"images" gorm:"column:images;type:json" description:"图组列表"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	Example *ImageExample `json:"example,omitempty" gorm:"-"`
}

// 表名设置
func (ImageExample) TableName() string {
	return "image_examples"
}

func (ImageExampleDetail) TableName() string {
	return "image_example_details"
}


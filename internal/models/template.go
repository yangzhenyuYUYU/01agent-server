package models

import (
	"time"
)

// TemplateStatus 模板状态枚举
type TemplateStatus int16

const (
	TemplateStatusDraft     TemplateStatus = 0 // 草稿
	TemplateStatusPublished TemplateStatus = 1 // 已发布
	TemplateStatusDisabled  TemplateStatus = 2 // 已禁用
	TemplateStatusDeleted   TemplateStatus = 3 // 已删除
)

// TemplateType 模板类型枚举
type TemplateType int16

const (
	TemplateTypeWechat TemplateType = 1 // 微信公众号
	TemplateTypeUnified TemplateType = 2 // 统一格式
	TemplateTypeCustom  TemplateType = 3 // 自定义
)

// PriceType 价格类型枚举
type PriceType int16

const (
	PriceTypeFree   PriceType = 0 // 免费
	PriceTypePaid   PriceType = 1 // 付费
	PriceTypeVIPOnly PriceType = 2 // 仅VIP
)

// VisibilityType 可见性类型枚举
type VisibilityType int16

const (
	VisibilityTypePrivate VisibilityType = 0 // 私有
	VisibilityTypePublic  VisibilityType = 1 // 公开
	VisibilityTypeShared  VisibilityType = 2 // 分享链接可见
)

// PublicTemplate 官方模板模型
type PublicTemplate struct {
	TemplateID    string         `json:"template_id" gorm:"primaryKey;column:template_id;type:varchar(50)" description:"模板ID"`
	Name          string         `json:"name" gorm:"column:name;type:varchar(100);not null" description:"模板名称"`
	NameEn        *string        `json:"name_en" gorm:"column:name_en;type:varchar(100)" description:"英文名称"`
	Description   *string        `json:"description" gorm:"column:description;type:longtext" description:"模板描述"`
	Author        string         `json:"author" gorm:"column:author;type:varchar(50);default:'system'" description:"作者"`
	TemplateType  TemplateType   `json:"template_type" gorm:"column:template_type;type:smallint;default:2" description:"模板类型"`
	Status        TemplateStatus `json:"status" gorm:"column:status;type:smallint;default:0" description:"模板状态"`
	PriceType     PriceType      `json:"price_type" gorm:"column:price_type;type:smallint;default:0" description:"价格类型"`
	Price         float64        `json:"price" gorm:"column:price;type:decimal(10,2);default:0" description:"价格"`
	OriginalPrice *float64       `json:"original_price" gorm:"column:original_price;type:decimal(10,2)" description:"原价"`
	IsPublic      bool           `json:"is_public" gorm:"column:is_public;default:true" description:"是否公开"`
	IsFeatured    bool           `json:"is_featured" gorm:"column:is_featured;default:false" description:"是否推荐"`
	IsOfficial    bool           `json:"is_official" gorm:"column:is_official;default:true" description:"是否官方模板"`
	TemplateData  *string        `json:"template_data" gorm:"column:template_data;type:json" description:"完整的模板JSON数据"`
	PreviewURL    *string        `json:"preview_url" gorm:"column:preview_url;type:varchar(500)" description:"预览URL"`
	ThumbnailURL  *string        `json:"thumbnail_url" gorm:"column:thumbnail_url;type:varchar(500)" description:"缩略图URL"`
	SectionHTML   *string        `json:"section_html" gorm:"column:section_html;type:longtext" description:"预览HTML片段"`
	PrimaryColor  string         `json:"primary_color" gorm:"column:primary_color;type:varchar(20);default:'#000000'" description:"主色调"`
	Tags          *string        `json:"tags" gorm:"column:tags;type:json" description:"标签列表"`
	Category      *string        `json:"category" gorm:"column:category;type:varchar(50)" description:"分类"`
	DownloadCount int            `json:"download_count" gorm:"column:download_count;default:0" description:"下载次数"`
	UseCount      int            `json:"use_count" gorm:"column:use_count;default:0" description:"使用次数"`
	LikeCount     int            `json:"like_count" gorm:"column:like_count;default:0" description:"点赞数"`
	ViewCount     int            `json:"view_count" gorm:"column:view_count;default:0" description:"查看次数"`
	SortOrder     int            `json:"sort_order" gorm:"column:sort_order;default:0" description:"排序权重"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
	PublishedAt   *time.Time     `json:"published_at" gorm:"column:published_at" description:"发布时间"`
}

// UserTemplate 用户模板模型
type UserTemplate struct {
	TemplateID    string         `json:"template_id" gorm:"primaryKey;column:template_id;type:varchar(50)" description:"模板ID"`
	UserID        string         `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"创建用户ID"`
	Name          string         `json:"name" gorm:"column:name;type:varchar(100);not null" description:"模板名称"`
	Description   *string        `json:"description" gorm:"column:description;type:longtext" description:"模板描述"`
	TemplateType  TemplateType   `json:"template_type" gorm:"column:template_type;type:smallint;default:3" description:"模板类型"`
	Status        TemplateStatus `json:"status" gorm:"column:status;type:smallint;default:0" description:"模板状态"`
	Visibility    VisibilityType `json:"visibility" gorm:"column:visibility;type:smallint;default:0" description:"可见性"`
	PriceType     PriceType      `json:"price_type" gorm:"column:price_type;type:smallint;default:0" description:"价格类型"`
	Price         float64        `json:"price" gorm:"column:price;type:decimal(10,2);default:0" description:"价格"`
	TemplateData  *string        `json:"template_data" gorm:"column:template_data;type:json" description:"完整的模板JSON数据"`
	BaseTemplateID *string       `json:"base_template_id" gorm:"column:base_template_id;type:varchar(50);index" description:"基于的官方模板ID"`
	PreviewURL    *string        `json:"preview_url" gorm:"column:preview_url;type:varchar(500)" description:"预览URL"`
	ThumbnailURL  *string        `json:"thumbnail_url" gorm:"column:thumbnail_url;type:varchar(500)" description:"缩略图URL"`
	SectionHTML   *string        `json:"section_html" gorm:"column:section_html;type:longtext" description:"预览HTML片段"`
	PrimaryColor  string         `json:"primary_color" gorm:"column:primary_color;type:varchar(20);default:'#000000'" description:"主色调"`
	Tags          *string        `json:"tags" gorm:"column:tags;type:json" description:"标签列表"`
	Category      *string        `json:"category" gorm:"column:category;type:varchar(50)" description:"分类"`
	ShareCode     *string        `json:"share_code" gorm:"column:share_code;type:varchar(20);uniqueIndex" description:"分享码"`
	ShareCount    int            `json:"share_count" gorm:"column:share_count;default:0" description:"分享次数"`
	DownloadCount int            `json:"download_count" gorm:"column:download_count;default:0" description:"下载次数"`
	UseCount      int            `json:"use_count" gorm:"column:use_count;default:0" description:"使用次数"`
	LikeCount     int            `json:"like_count" gorm:"column:like_count;default:0" description:"点赞数"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
	PublishedAt   *time.Time     `json:"published_at" gorm:"column:published_at" description:"发布时间"`
	LastUsedAt    *time.Time     `json:"last_used_at" gorm:"column:last_used_at" description:"最后使用时间"`

	// 关联关系
	User        *User           `json:"user,omitempty" gorm:"-"`
	BaseTemplate *PublicTemplate `json:"base_template,omitempty" gorm:"-"`
}

// 表名设置
func (PublicTemplate) TableName() string {
	return "public_templates"
}

func (UserTemplate) TableName() string {
	return "user_templates"
}


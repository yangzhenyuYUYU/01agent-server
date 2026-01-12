package models

import (
	"time"
)

// BlogPost 博客文章模型
type BlogPost struct {
	ID             string     `json:"id" gorm:"primaryKey;column:id;type:varchar(36)" description:"文章ID"`
	Slug           string     `json:"slug" gorm:"column:slug;type:varchar(255);uniqueIndex;not null" description:"URL友好的标识符"`
	Title          string     `json:"title" gorm:"column:title;type:varchar(500);not null" description:"文章标题"`
	Summary        string     `json:"summary" gorm:"column:summary;type:text;not null" description:"文章摘要"`
	Content        string     `json:"content" gorm:"column:content;type:longtext;not null" description:"Markdown格式的正文"`
	Category       string     `json:"category" gorm:"column:category;type:varchar(50);not null;index" description:"分类"`
	CoverImage     *string    `json:"cover_image" gorm:"column:cover_image;type:varchar(500)" description:"封面图URL"`
	Author         string     `json:"author" gorm:"column:author;type:varchar(100);default:'01Agent Team'" description:"作者"`
	AuthorAvatar   *string    `json:"author_avatar" gorm:"column:author_avatar;type:varchar(500)" description:"作者头像URL"`
	PublishDate    time.Time  `json:"publish_date" gorm:"column:publish_date;not null;index" description:"发布时间"`
	UpdatedDate    *time.Time `json:"updated_date" gorm:"column:updated_date" description:"更新时间"`
	ReadTime       *int       `json:"read_time" gorm:"column:read_time" description:"阅读时间（分钟）"`
	Views          int        `json:"views" gorm:"column:views;default:0" description:"浏览量"`
	Likes          int        `json:"likes" gorm:"column:likes;default:0" description:"点赞数"`
	IsFeatured     bool       `json:"is_featured" gorm:"column:is_featured;default:false" description:"是否精选"`
	SEODescription *string    `json:"seo_description" gorm:"column:seo_description;type:varchar(500)" description:"SEO描述"`
	Status         string     `json:"status" gorm:"column:status;type:varchar(20);default:'published';index" description:"状态: draft/published/archived"`
	ThemeName      *string    `json:"theme_name" gorm:"column:theme_name;type:varchar(100)" description:"主题样式名称"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`

	// 关联字段
	Tags        []BlogTag `json:"tags,omitempty" gorm:"many2many:blog_post_tags;foreignKey:ID;joinForeignKey:PostID;References:ID;joinReferences:TagID"`
	SEOKeywords []string  `json:"seo_keywords,omitempty" gorm:"-"`
}

// BlogTag 博客标签模型
type BlogTag struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement;column:id" description:"标签ID"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(50);uniqueIndex;not null" description:"标签名"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
}

// BlogPostTag 文章标签关联模型
type BlogPostTag struct {
	PostID string `gorm:"primaryKey;column:post_id;type:varchar(36)"`
	TagID  int    `gorm:"primaryKey;column:tag_id"`
}

// BlogSEOKeyword SEO关键词模型
type BlogSEOKeyword struct {
	ID      int    `json:"id" gorm:"primaryKey;autoIncrement;column:id" description:"关键词ID"`
	PostID  string `json:"post_id" gorm:"column:post_id;type:varchar(36);not null;index" description:"文章ID"`
	Keyword string `json:"keyword" gorm:"column:keyword;type:varchar(100);not null" description:"关键词"`
}

// 表名设置
func (BlogPost) TableName() string {
	return "blog_posts"
}

func (BlogTag) TableName() string {
	return "blog_tags"
}

func (BlogPostTag) TableName() string {
	return "blog_post_tags"
}

func (BlogSEOKeyword) TableName() string {
	return "blog_seo_keywords"
}

// 常量定义
const (
	BlogStatusDraft     = "draft"
	BlogStatusPublished = "published"
	BlogStatusArchived  = "archived"
)

// 分类映射
var CategoryNames = map[string]string{
	"product-updates":   "产品动态",
	"tutorials":         "使用教程",
	"tips-and-tricks":   "运营技巧",
	"industry-insights": "行业洞察",
	"case-studies":      "案例故事",
}

// BlogCreateRequest 创建博客文章请求
type BlogCreateRequest struct {
	Slug           string   `json:"slug" binding:"required" description:"URL友好的标识符"`
	Title          string   `json:"title" binding:"required" description:"文章标题"`
	Summary        string   `json:"summary" binding:"required" description:"文章摘要"`
	Content        string   `json:"content" binding:"required" description:"Markdown格式的正文"`
	Category       string   `json:"category" binding:"required" description:"分类"`
	CoverImage     *string  `json:"cover_image" description:"封面图URL"`
	Author         string   `json:"author" description:"作者"`
	AuthorAvatar   *string  `json:"author_avatar" description:"作者头像URL"`
	ReadTime       *int     `json:"read_time" description:"阅读时间（分钟）"`
	IsFeatured     bool     `json:"is_featured" description:"是否精选"`
	SEODescription *string  `json:"seo_description" description:"SEO描述"`
	Status         string   `json:"status" description:"状态: draft/published/archived"`
	ThemeName      *string  `json:"theme_name" description:"主题样式名称"`
	Tags           []string `json:"tags" description:"标签列表"`
	SEOKeywords    []string `json:"seo_keywords" description:"SEO关键词列表"`
}

// BlogUpdateRequest 更新博客文章请求
type BlogUpdateRequest struct {
	Title          *string  `json:"title" description:"文章标题"`
	Summary        *string  `json:"summary" description:"文章摘要"`
	Content        *string  `json:"content" description:"Markdown格式的正文"`
	Category       *string  `json:"category" description:"分类"`
	CoverImage     *string  `json:"cover_image" description:"封面图URL"`
	Author         *string  `json:"author" description:"作者"`
	AuthorAvatar   *string  `json:"author_avatar" description:"作者头像URL"`
	ReadTime       *int     `json:"read_time" description:"阅读时间（分钟）"`
	IsFeatured     *bool    `json:"is_featured" description:"是否精选"`
	SEODescription *string  `json:"seo_description" description:"SEO描述"`
	Status         *string  `json:"status" description:"状态"`
	ThemeName      *string  `json:"theme_name" description:"主题样式名称"`
	Tags           []string `json:"tags" description:"标签列表"`
	SEOKeywords    []string `json:"seo_keywords" description:"SEO关键词列表"`
}

// BlogPostResponse 博客文章响应结构
type BlogPostResponse struct {
	ID             string     `json:"id"`
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Summary        string     `json:"summary"`
	Content        *string    `json:"content,omitempty"` // 列表接口不返回content
	Category       string     `json:"category"`
	CategoryName   string     `json:"category_name"`
	CoverImage     *string    `json:"cover_image"`
	Author         string     `json:"author"`
	AuthorAvatar   *string    `json:"author_avatar"`
	PublishDate    time.Time  `json:"publish_date"`
	UpdatedDate    *time.Time `json:"updated_date"`
	ReadTime       *int       `json:"read_time"`
	Views          int        `json:"views"`
	Likes          int        `json:"likes"`
	IsFeatured     bool       `json:"is_featured"`
	ThemeName      *string    `json:"theme_name"`
	Tags           []string   `json:"tags"`
	SEOKeywords    []string   `json:"seo_keywords,omitempty"`
	SEODescription *string    `json:"seo_description,omitempty"`
	Status         string     `json:"status"`
}

// ToResponse 转换为响应结构
func (bp *BlogPost) ToResponse(includeContent bool) *BlogPostResponse {
	categoryName, ok := CategoryNames[bp.Category]
	if !ok {
		categoryName = bp.Category
	}

	resp := &BlogPostResponse{
		ID:           bp.ID,
		Slug:         bp.Slug,
		Title:        bp.Title,
		Summary:      bp.Summary,
		Category:     bp.Category,
		CategoryName: categoryName,
		CoverImage:   bp.CoverImage,
		Author:       bp.Author,
		AuthorAvatar: bp.AuthorAvatar,
		PublishDate:  bp.PublishDate,
		UpdatedDate:  bp.UpdatedDate,
		ReadTime:     bp.ReadTime,
		Views:        bp.Views,
		Likes:        bp.Likes,
		IsFeatured:   bp.IsFeatured,
		ThemeName:    bp.ThemeName,
		Tags:         []string{},
		SEOKeywords:  []string{},
		Status:       bp.Status,
	}

	if includeContent {
		resp.Content = &bp.Content
		resp.SEODescription = bp.SEODescription
	}

	// 转换标签
	if len(bp.Tags) > 0 {
		for _, tag := range bp.Tags {
			resp.Tags = append(resp.Tags, tag.Name)
		}
	}

	// SEO关键词
	if len(bp.SEOKeywords) > 0 {
		resp.SEOKeywords = bp.SEOKeywords
	}

	return resp
}

// BlogListResponse 博客列表响应
type BlogListResponse struct {
	Items      []BlogPostResponse `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

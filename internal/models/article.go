package models

import (
	"time"
)

// ArticleSceneType 文章场景类型枚举
type ArticleSceneType string

const (
	ArticleSceneWeixin   ArticleSceneType = "weixin"   // 微信公众号
	ArticleSceneZhihu    ArticleSceneType = "zhihu"    // 知乎
	ArticleSceneToutiao  ArticleSceneType = "toutiao"  // 今日头条
	ArticleSceneJianshu  ArticleSceneType = "jianshu"  // 简书
	ArticleSceneCSDN     ArticleSceneType = "csdn"     // CSDN
	ArticleSceneJuejin   ArticleSceneType = "juejin"   // 掘金
	ArticleSceneBlog     ArticleSceneType = "blog"     // 个人博客
	ArticleSceneOther    ArticleSceneType = "other"    // 其他
)

// ArticleEditTask 文章编辑任务模型
type ArticleEditTask struct {
	ID            string           `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"编辑任务ID"`
	ArticleTaskID *string          `json:"article_task_id" gorm:"column:article_task_id;type:char(36);index" description:"关联文章任务ID"`
	UserID        string           `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Title         string           `json:"title" gorm:"column:title;type:varchar(255);not null" description:"文章标题"`
	Theme         string           `json:"theme" gorm:"column:theme;type:varchar(100);default:'default'" description:"文章主题"`
	SceneType     ArticleSceneType `json:"scene_type" gorm:"column:scene_type;type:varchar(20);default:'other';index" description:"场景类型"`
	Params        *string          `json:"params" gorm:"column:params;type:json" description:"编辑参数"`
	Content       string           `json:"content" gorm:"column:content;type:longtext;not null" description:"文章内容"`
	SectionHTML   *string          `json:"section_html" gorm:"column:section_html;type:longtext" description:"文章内容的HTML格式（用于预览）"`
	Status        string           `json:"status" gorm:"column:status;type:varchar(20);not null;default:'editing'" description:"编辑状态(editing编辑中/pending待发布/published已发布)"`
	IsPublic      bool             `json:"is_public" gorm:"column:is_public;default:false" description:"是否公开"`
	Tags          *string          `json:"tags" gorm:"column:tags;type:json" description:"分类标签"`
	PublishedAt   *time.Time       `json:"published_at" gorm:"column:published_at" description:"发布时间"`
	CreatedAt     time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time        `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User        *User        `json:"user,omitempty" gorm:"-"`
	ArticleTask *ArticleTask `json:"article_task,omitempty" gorm:"-"`
}

// ArticlePublishConfig 文章发布配置模型
type ArticlePublishConfig struct {
	ID                   string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"发布配置ID"`
	EditTaskID           string    `json:"edit_task_id" gorm:"column:edit_task_id;type:char(36);not null;index" description:"关联编辑任务ID"`
	PublishTitle         string    `json:"publish_title" gorm:"column:publish_title;type:varchar(255);not null" description:"发布标题"`
	AuthorName           string    `json:"author_name" gorm:"column:author_name;type:varchar(100);not null" description:"作者名称"`
	Summary              *string   `json:"summary" gorm:"column:summary;type:longtext" description:"文章摘要"`
	CoverImage           *string   `json:"cover_image" gorm:"column:cover_image;type:varchar(500)" description:"封面图片URL"`
	EnableComments       bool      `json:"enable_comments" gorm:"column:enable_comments;default:true" description:"是否开放评论区"`
	FollowersOnlyComment bool      `json:"followers_only_comment" gorm:"column:followers_only_comment;default:false" description:"是否仅粉丝可评论"`
	CreatedAt            time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	EditTask *ArticleEditTask `json:"edit_task,omitempty" gorm:"-"`
}

// 注意：ArticleTask, ArticleTopic, TaskUsage, TaskErrorLog, TotalUsageStats 已在 agent_search.go 中定义

// ArticleEditStatus 文章编辑状态常量
const (
	ArticleEditStatusEditing   = "editing"   // 编辑中
	ArticleEditStatusPending   = "pending"   // 待发布(已保存)
	ArticleEditStatusDraft     = "draft"     // 已同步到草稿箱
	ArticleEditStatusPublished = "published" // 已发布
)

// 表名设置
func (ArticleEditTask) TableName() string {
	return "article_edit_tasks"
}

func (ArticlePublishConfig) TableName() string {
	return "article_publish_configs"
}

// 注意：ArticleTask, ArticleTopic, TaskUsage, TaskErrorLog, TotalUsageStats 的 TableName 方法已在 agent_search.go 中定义

// 注意：ArticleTask 和 ArticleTopic 的响应结构和 ToResponse 方法应在 agent_search.go 中定义

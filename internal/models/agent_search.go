package models

import (
	"time"
)

// ArticleTask 文章生成任务模型
type ArticleTask struct {
	ID                string     `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"任务ID"`
	ClientID          string     `json:"client_id" gorm:"column:client_id;type:varchar(100);uniqueIndex" description:"客户端ID"`
	UserID            string     `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Theme             *string    `json:"theme" gorm:"column:theme;type:varchar(30)" description:"文章排版主题"`
	Topic             string     `json:"topic" gorm:"column:topic;type:longtext;not null" description:"文章主题"`
	AuthorName        string     `json:"author_name" gorm:"column:author_name;type:varchar(100);not null" description:"作者名称"`
	IsPublic          bool       `json:"is_public" gorm:"column:is_public;default:false" description:"是否公开"`
	Status            string     `json:"status" gorm:"column:status;type:varchar(20);default:'pending'" description:"任务状态"`
	CurrentStep       *string    `json:"current_step" gorm:"column:current_step;type:varchar(20)" description:"当前步骤"`
	Steps             *string    `json:"steps" gorm:"column:steps;type:json" description:"所有步骤的详细状态"`
	Title             *string    `json:"title" gorm:"column:title;type:varchar(255)" description:"文章标题"`
	Snippet           *string    `json:"snippet" gorm:"column:snippet;type:longtext" description:"文章摘要"`
	Content           *string    `json:"content" gorm:"column:content;type:longtext" description:"文章内容"`
	WordCount         *int       `json:"word_count" gorm:"column:word_count" description:"文章字数"`
	KbContent         *string    `json:"kb_content" gorm:"column:kb_content;type:longtext" description:"知识库搜索内容"`
	Images            *string    `json:"images" gorm:"column:images;type:json" description:"文章相关图片"`
	IsWebSearch       *bool      `json:"is_web_search" gorm:"column:is_web_search" description:"是否进行互联网搜索"`
	UserLinks         *string    `json:"user_links" gorm:"column:user_links;type:json" description:"用户上传链接"`
	IsPublished       bool       `json:"is_published" gorm:"column:is_published;default:false" description:"是否已发布"`
	PublishURL        *string    `json:"publish_url" gorm:"column:publish_url;type:varchar(255)" description:"发布URL"`
	StartTime         time.Time  `json:"start_time" gorm:"column:start_time;autoCreateTime" description:"开始时间"`
	EndTime           *time.Time `json:"end_time" gorm:"column:end_time" description:"结束时间"`
	TotalDuration     *int       `json:"total_duration" gorm:"column:total_duration" description:"总执行时间(秒)"`
	SearchDuration    *int       `json:"search_duration" gorm:"column:search_duration" description:"搜索步骤执行时间(秒)"`
	ParseDuration     *int       `json:"parse_duration" gorm:"column:parse_duration" description:"解析步骤执行时间(秒)"`
	GenerateDuration  *int       `json:"generate_duration" gorm:"column:generate_duration" description:"生成步骤执行时间(秒)"`
	CompleteDuration  *int       `json:"complete_duration" gorm:"column:complete_duration" description:"完成步骤执行时间(秒)"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// ArticleTopic 文章主题清单模型
type ArticleTopic struct {
	ID           string     `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"任务ID"`
	UserID       string     `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	RelatedTask  *string    `json:"related_task" gorm:"column:related_task;type:varchar(255)" description:"关联任务ID"`
	Title        *string    `json:"title" gorm:"column:title;type:varchar(255)" description:"文章主题标题"`
	Description  *string    `json:"description" gorm:"column:description;type:longtext" description:"主题描述"`
	AuthorName   *string    `json:"author_name" gorm:"column:author_name;type:varchar(100)" description:"作者名称"`
	Category     *string    `json:"category" gorm:"column:category;type:varchar(100)" description:"主题分类"`
	Tags         *string    `json:"tags" gorm:"column:tags;type:json" description:"主题标签"`
	Status       *string    `json:"status" gorm:"column:status;type:varchar(20);default:'draft'" description:"主题状态(draft草稿/published已发布/scheduled计划发布)"`
	PublishDate  *time.Time `json:"publish_date" gorm:"column:publish_date;type:date" description:"计划发布日期"`
	PublishTime  *time.Time `json:"publish_time" gorm:"column:publish_time" description:"实际发布时间"`
	ViewCount    int        `json:"view_count" gorm:"column:view_count;default:0" description:"浏览次数"`
	LikeCount    int        `json:"like_count" gorm:"column:like_count;default:0" description:"点赞次数"`
	CommentCount int        `json:"comment_count" gorm:"column:comment_count;default:0" description:"评论次数"`
	TokenUsage   *string    `json:"token_usage" gorm:"column:token_usage;type:json" description:"Token使用情况"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// TaskUsage 任务具体使用统计模型
type TaskUsage struct {
	ID               string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"统计ID"`
	TaskID           string    `json:"task_id" gorm:"column:task_id;type:char(36);not null;index" description:"关联任务ID"`
	TaskName         string    `json:"task_name" gorm:"column:task_name;type:varchar(100);not null" description:"任务名称"`
	Provider         string    `json:"provider" gorm:"column:provider;type:varchar(100);not null" description:"服务提供商"`
	Model            string    `json:"model" gorm:"column:model;type:varchar(100);not null" description:"模型名称"`
	ActModel         string    `json:"act_model" gorm:"column:act_model;type:varchar(100);not null" description:"实际使用的模型"`
	PromptTokens     int       `json:"prompt_tokens" gorm:"column:prompt_tokens;not null" description:"输入token数量"`
	CompletionTokens int       `json:"completion_tokens" gorm:"column:completion_tokens;not null" description:"输出token数量"`
	TotalTokens      int       `json:"total_tokens" gorm:"column:total_tokens;not null" description:"总token数量"`
	TotalCost        float64   `json:"total_cost" gorm:"column:total_cost;not null" description:"总成本(元)"`
	ExecutionTime    float64   `json:"execution_time" gorm:"column:execution_time;not null" description:"执行时间(秒)"`
	Prompt           *string   `json:"prompt" gorm:"column:prompt;type:longtext" description:"提示词"`
	Response         *string   `json:"response" gorm:"column:response;type:longtext" description:"响应内容"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	Task *ArticleTask `json:"task,omitempty" gorm:"-"`
}

// TotalUsageStats 总体使用统计模型
type TotalUsageStats struct {
	ID                    string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"统计ID"`
	TaskID                string    `json:"task_id" gorm:"column:task_id;type:char(36);not null;index" description:"关联任务ID"`
	UserID                string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	TaskType              string    `json:"task_type" gorm:"column:task_type;type:varchar(100);not null" description:"任务类型"`
	TotalPromptTokens     int       `json:"total_prompt_tokens" gorm:"column:total_prompt_tokens;default:0" description:"总输入token数量"`
	TotalCompletionTokens int       `json:"total_completion_tokens" gorm:"column:total_completion_tokens;default:0" description:"总输出token数量"`
	TotalTokens           int       `json:"total_tokens" gorm:"column:total_tokens;default:0" description:"总token数量"`
	TotalCost             float64   `json:"total_cost" gorm:"column:total_cost;default:0" description:"总成本(元)"`
	TotalExecutionTime    float64   `json:"total_execution_time" gorm:"column:total_execution_time;default:0" description:"总执行时间(秒)"`
	CreatedAt             time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	Task *ArticleTask `json:"task,omitempty" gorm:"-"`
	User *User        `json:"user,omitempty" gorm:"-"`
}

// TaskErrorLog 任务错误日志模型
type TaskErrorLog struct {
	ID             string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"日志ID"`
	TaskID         *string   `json:"task_id" gorm:"column:task_id;type:char(36);index" description:"关联任务ID"`
	UserID         *string   `json:"user_id" gorm:"column:user_id;type:varchar(50);index" description:"关联用户ID"`
	ErrorMessage   string    `json:"error_message" gorm:"column:error_message;type:longtext;not null" description:"错误信息"`
	ErrorType      string    `json:"error_type" gorm:"column:error_type;type:varchar(100);not null" description:"错误类型"`
	ErrorTraceback *string   `json:"error_traceback" gorm:"column:error_traceback;type:longtext" description:"错误堆栈跟踪"`
	Step           string    `json:"step" gorm:"column:step;type:varchar(50);not null" description:"发生错误的步骤"`
	SubStep        *string   `json:"sub_step" gorm:"column:sub_step;type:varchar(50)" description:"发生错误的子步骤"`
	ClientID       *string   `json:"client_id" gorm:"column:client_id;type:varchar(100)" description:"客户端ID"`
	TaskType       *string   `json:"task_type" gorm:"column:task_type;type:varchar(20)" description:"任务类型(wx/xhs)"`
	AdditionalInfo *string   `json:"additional_info" gorm:"column:additional_info;type:json" description:"额外信息"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`

	// 关联关系
	Task *ArticleTask `json:"task,omitempty" gorm:"-"`
	User *User        `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (ArticleTask) TableName() string {
	return "article_tasks"
}

func (ArticleTopic) TableName() string {
	return "article_topics"
}

func (TaskUsage) TableName() string {
	return "task_usages"
}

func (TotalUsageStats) TableName() string {
	return "total_usage_stats"
}

func (TaskErrorLog) TableName() string {
	return "task_error_logs"
}


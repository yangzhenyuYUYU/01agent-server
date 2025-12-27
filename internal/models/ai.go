package models

import (
	"time"
)

// AIFormatRecord AI内容排版记录模型
type AIFormatRecord struct {
	ID               int       `json:"id" gorm:"primaryKey;column:id" description:"唯一记录ID"`
	OriginalContent  string    `json:"original_content" gorm:"column:original_content;type:longtext;not null" description:"原始内容"`
	FormattedContent *string   `json:"formatted_content" gorm:"column:formatted_content;type:longtext" description:"排版后内容"`
	FormatType       string    `json:"format_type" gorm:"column:format_type;type:varchar(10);not null" description:"目标格式类型"`
	Status           string    `json:"status" gorm:"column:status;type:varchar(15);not null;default:'pending'" description:"处理状态"`
	Tokens           *string   `json:"tokens" gorm:"column:tokens;type:varchar(64)" description:"消耗tokens"`
	ErrorLog         *string   `json:"error_log" gorm:"column:error_log;type:json" description:"错误日志详情"`
	ModelVersion     *string   `json:"model_version" gorm:"column:model_version;type:varchar(50);default:''" description:"AI模型版本"`
	UserID           string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"最后更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// AIRecommendTopic AI推荐主题模型
type AIRecommendTopic struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"唯一主题ID"`
	Title       string    `json:"title" gorm:"column:title;type:varchar(255);not null" description:"主题标题"`
	Description *string   `json:"description" gorm:"column:description;type:longtext" description:"主题描述"`
	Category    *string   `json:"category" gorm:"column:category;type:varchar(50)" description:"主题分类"`
	Tags        *string   `json:"tags" gorm:"column:tags;type:json" description:"主题标签"`
	Status      int       `json:"status" gorm:"column:status;not null;default:1" description:"状态：0-禁用，1-启用"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"最后更新时间"`
}

// AIRewriteRecord AI智能文案改写记录模型
type AIRewriteRecord struct {
	ID            int       `json:"id" gorm:"primaryKey;column:id" description:"唯一记录ID"`
	OriginalText  *string   `json:"original_text" gorm:"column:original_text;type:longtext" description:"原始文本内容"`
	RewrittenText *string   `json:"rewritten_text" gorm:"column:rewritten_text;type:longtext" description:"改写后文本内容"`
	Status        string    `json:"status" gorm:"column:status;type:varchar(15);not null;default:'pending'" description:"处理状态"`
	Tokens        *string   `json:"tokens" gorm:"column:tokens;type:varchar(64)" description:"消耗tokens"`
	ErrorLog      *string   `json:"error_log" gorm:"column:error_log;type:json" description:"错误日志详情"`
	ModelVersion  *string   `json:"model_version" gorm:"column:model_version;type:varchar(50);default:''" description:"AI模型版本"`
	UserID        string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"最后更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// AITopicPolishRecord AI主题润色记录模型
type AITopicPolishRecord struct {
	ID            int       `json:"id" gorm:"primaryKey;column:id" description:"唯一记录ID"`
	OriginalTopic string    `json:"original_topic" gorm:"column:original_topic;type:longtext;not null" description:"原始主题"`
	PolishedTopic *string   `json:"polished_topic" gorm:"column:polished_topic;type:longtext" description:"润色后主题"`
	Status        string    `json:"status" gorm:"column:status;type:varchar(15);not null;default:'pending'" description:"处理状态"`
	Tokens        *string   `json:"tokens" gorm:"column:tokens;type:varchar(64)" description:"消耗tokens"`
	ErrorLog      *string   `json:"error_log" gorm:"column:error_log;type:json" description:"错误日志详情"`
	ModelVersion  *string   `json:"model_version" gorm:"column:model_version;type:varchar(50);default:''" description:"AI模型版本"`
	UserID        string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"最后更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (AIFormatRecord) TableName() string {
	return "ai_format_records"
}

func (AIRecommendTopic) TableName() string {
	return "ai_recommend_topics"
}

func (AIRewriteRecord) TableName() string {
	return "ai_rewrite_records"
}

func (AITopicPolishRecord) TableName() string {
	return "ai_topic_polish_records"
}

// 响应结构
type AIFormatRecordResponse struct {
	ID               int     `json:"id"`
	OriginalContent  string  `json:"original_content"`
	FormattedContent *string `json:"formatted_content"`
	FormatType       string  `json:"format_type"`
	Status           string  `json:"status"`
	Tokens           *string `json:"tokens"`
	ModelVersion     *string `json:"model_version"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type AIRecommendTopicResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Tags        *string `json:"tags"`
	Status      int     `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ToResponse 转换方法
func (a *AIFormatRecord) ToResponse() AIFormatRecordResponse {
	return AIFormatRecordResponse{
		ID:               a.ID,
		OriginalContent:  a.OriginalContent,
		FormattedContent: a.FormattedContent,
		FormatType:       a.FormatType,
		Status:           a.Status,
		Tokens:           a.Tokens,
		ModelVersion:     a.ModelVersion,
		CreatedAt:        a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (a *AIRecommendTopic) ToResponse() AIRecommendTopicResponse {
	return AIRecommendTopicResponse{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		Category:    a.Category,
		Tags:        a.Tags,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

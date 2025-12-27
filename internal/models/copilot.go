package models

import (
	"time"
)

// WorkflowStatus 工作流状态枚举
type WorkflowStatus string

const (
	WorkflowStatusPending     WorkflowStatus = "pending"     // 待处理
	WorkflowStatusRunning     WorkflowStatus = "running"     // 运行中
	WorkflowStatusCompleted   WorkflowStatus = "completed"   // 已完成
	WorkflowStatusFailed      WorkflowStatus = "failed"      // 失败
	WorkflowStatusInterrupted WorkflowStatus = "interrupted" // 中断
	WorkflowStatusCancelled   WorkflowStatus = "cancelled"   // 用户取消
)

// CopilotScene Copilot场景类型枚举
type CopilotScene string

const (
	CopilotSceneContext CopilotScene = "context" // 文本内容（公众号文章等）
	CopilotSceneCanvas  CopilotScene = "canvas"  // 画布内容（Canvas画板编辑器）
	CopilotSceneVideo   CopilotScene = "video"   // 视频内容
	CopilotSceneDigital CopilotScene = "digital" // 数字人内容
)

// CopilotChatThread 对话线程模型 - 存储对话线程的基本信息（父级表）
type CopilotChatThread struct {
	ID        string       `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"线程ID"`
	UserID    string       `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	ThreadID  string       `json:"thread_id" gorm:"column:thread_id;type:varchar(100);uniqueIndex" description:"线程标识ID"`
	Label     *string      `json:"label" gorm:"column:label;type:varchar(200)" description:"线程标签/名称"`
	Scene     *CopilotScene `json:"scene" gorm:"column:scene;type:varchar(20);default:'context'" description:"场景类型：CONTEXT-文本内容，CANVAS-画布内容，VIDEO-视频内容，DIGITAL-数字人内容"`
	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// CopilotChatSession 对话聊天模型 - 存储用户输入和AI回答
type CopilotChatSession struct {
	ID             string         `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"会话记录ID"`
	UserID         string         `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	ThreadID       string         `json:"thread_id" gorm:"column:thread_id;type:varchar(100);index" description:"线程ID"`
	WorkflowID     string         `json:"workflow_id" gorm:"column:workflow_id;type:varchar(100);index" description:"工作流ID"`
	ClientID       *string        `json:"client_id" gorm:"column:client_id;type:varchar(100)" description:"客户端ID"`
	UserQuery      string         `json:"user_query" gorm:"column:user_query;type:longtext;not null" description:"用户查询内容"`
	AiResponse     *string        `json:"ai_response" gorm:"column:ai_response;type:json" description:"AI会话数据"`
	Feedback       int            `json:"feedback" gorm:"column:feedback;default:0" description:"反馈状态，0-无，1-赞，-1-踩"`
	FeedbackContent *string       `json:"feedback_content" gorm:"column:feedback_content;type:varchar(200)" description:"反馈内容"`
	Status         WorkflowStatus `json:"status" gorm:"column:status;type:varchar(20);default:'pending'" description:"会话状态"`
	CreatedAt      time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
	CompletedAt    *time.Time     `json:"completed_at" gorm:"column:completed_at" description:"完成时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// CopilotWorkflowRecord Copilot工作流记录模型
type CopilotWorkflowRecord struct {
	ID             string     `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"记录ID"`
	ThreadID       string     `json:"thread_id" gorm:"column:thread_id;type:varchar(100);index" description:"线程ID"`
	WorkflowID     string     `json:"workflow_id" gorm:"column:workflow_id;type:varchar(100);index" description:"工作流ID"`
	UserID         string     `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Config         *string    `json:"config" gorm:"column:config;type:json" description:"会话配置参数"`
	WorkflowData   *string    `json:"workflow_data" gorm:"column:workflow_data;type:json" description:"公众数据流"`
	TopicContent   *string    `json:"topic_content" gorm:"column:topic_content;type:json" description:"选题计划"`
	ArticleContent *string    `json:"article_content" gorm:"column:article_content;type:longtext" description:"文章内容"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// TokenUsageRecord Token耗费计算模型 - 记录token_usage_summary
type TokenUsageRecord struct {
	ID                string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"Token使用记录ID"`
	WorkflowID        string    `json:"workflow_id" gorm:"column:workflow_id;type:varchar(100);index" description:"工作流ID"`
	UserID            string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	TotalInputTokens  int       `json:"total_input_tokens" gorm:"column:total_input_tokens;default:0" description:"总输入Token数"`
	TotalOutputTokens int       `json:"total_output_tokens" gorm:"column:total_output_tokens;default:0" description:"总输出Token数"`
	TotalTokens       int       `json:"total_tokens" gorm:"column:total_tokens;default:0" description:"总Token数"`
	TotalCost         float64   `json:"total_cost" gorm:"column:total_cost;type:decimal(10,6);default:0" description:"总成本"`
	ModelCount        int       `json:"model_count" gorm:"column:model_count;default:0" description:"使用的模型数量"`
	SessionCount      int       `json:"session_count" gorm:"column:session_count;default:0" description:"会话调用次数"`
	ModelBreakdown    *string   `json:"model_breakdown" gorm:"column:model_breakdown;type:json" description:"按模型分组的使用统计"`
	SessionDetails    *string   `json:"session_details" gorm:"column:session_details;type:json" description:"每次调用的详细信息"`
	PrimaryModel      *string   `json:"primary_model" gorm:"column:primary_model;type:varchar(100);index" description:"主要使用的模型"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (CopilotChatThread) TableName() string {
	return "copilot_chat_threads"
}

func (CopilotChatSession) TableName() string {
	return "copilot_chat_sessions"
}

func (CopilotWorkflowRecord) TableName() string {
	return "copilot_workflow_records"
}

func (TokenUsageRecord) TableName() string {
	return "token_usage_records"
}


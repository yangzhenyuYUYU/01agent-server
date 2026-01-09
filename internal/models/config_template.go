package models

import (
	"time"
)

// ConfigTemplateStatus 配置模板状态枚举
type ConfigTemplateStatus int

const (
	ConfigTemplateStatusInactive ConfigTemplateStatus = 0 // 禁用
	ConfigTemplateStatusActive   ConfigTemplateStatus = 1 // 启用
)

// ConfigTemplateType 配置模板类型枚举
type ConfigTemplateType int

const (
	ConfigTemplateTypeArticle     ConfigTemplateType = 1  // 公众号文章
	ConfigTemplateTypePoster      ConfigTemplateType = 2  // 图文海报
	ConfigTemplateTypeXiaohongshu ConfigTemplateType = 3  // 小红书
	ConfigTemplateTypeVideo       ConfigTemplateType = 4  // 视频
	ConfigTemplateTypeGeo         ConfigTemplateType = 5  // GEO模式
	ConfigTemplateTypeCustom      ConfigTemplateType = 99 // 自定义
)

// UserConfigTemplate 用户配置模板模型
//
// 一套完整的配置模板，用户选中即可绑定整套配置。
// 支持文章、图文、视频等多种场景使用。
//
// config_data 结构示例:
//
//	{
//	    // ========== 示例提示词配置 ==========
//	    "example_prompts": [
//	        {"id": "1", "text": "写一篇...", "type": "text"},
//	        {"id": "2", "text": "基于链接...", "type": "link", "url": "https://...", "urlTitle": "..."},
//	        {"id": "3", "text": "分析文档...", "type": "document"}
//	    ],
//	    "placeholders": ["描述你想要创作的内容...", "上传参考图，AI帮你创作..."],
//
//	    // ========== 模型与生成配置 ==========
//	    "llm": "deepseek-v3",                    // 使用的语言模型: deepseek-v3, gpt-4o, claude-3-5-sonnet 等
//	    "mode": "auto",                          // 处理模式: auto, research_topic_agent, created_agent
//	    "is_web_search": true,                   // 是否进行网络搜索
//	    "is_insert_imgs": true,                  // 是否需要配图
//	    "is_info_mining": false,                 // 是否进行信息挖掘
//	    "is_auto_accepted_plan": true,           // 是否自动接受创作计划
//
//	    // ========== 系统提示词配置 ==========
//	    "is_system_prompt": true,                // 是否使用系统提示词
//	    "system_prompt_name": "default",         // 系统提示词名称
//	    "system_prompt": null,                   // 自定义系统提示词内容
//	    "is_user_preference": false,             // 是否使用用户偏好
//
//	    // ========== 风格与作者配置 ==========
//	    "author_name": "01Editor",               // 作者名称
//	    "theme": "default",                      // 主题
//	    "style_rules": null,                    // 用户风格规则
//	    "user_profile": null,                    // 用户画像信息
//
//	    // ========== 海报尺寸配置 ==========
//	    "poster_size": {
//	        "value": "poster_long_adaptive",
//	        "label": "长图",
//	        "width": 1080,
//	        "height": 0,
//	        "unit": "px",
//	        "desc": "1080x自适应",
//	        "platform": "图文海报"
//	    },
//
//	    // ========== 配色方案配置 ==========
//	    "poster_color_scheme": {
//	        "value": "ocean",
//	        "label": "科技蓝调",
//	        "colors": ["#1E40AF", "#60A5FA", "#DBEAFE"],
//	        "desc": "稳重专业"
//	    },
//
//	    // ========== 参考文件配置 ==========
//	    "files": [
//	        {"id": "f1", "type": "image", "path": "/uploads/xxx.jpg", "url": "https://...", "name": "参考图1.jpg"},
//	        {"id": "f2", "type": "document", "path": "/uploads/xxx.pdf", "url": "https://...", "name": "素材.pdf"}
//	    ],
//	    "imgs": [
//	        {"path": "/uploads/xxx.jpg", "url": "https://..."}
//	    ],
//	    "urls": [
//	        {"path": "article-link", "url": "https://..."}
//	    ],
//
//	    // ========== 项目类型 ==========
//	    "project_type": null                     // 项目类型标识
//
//	    // ... 可随意扩展更多配置
//	}
//
// metadata 结构示例:
//
//	{
//	    "description": "适合科技类文章的配置模板",
//	    "tags": ["科技", "AI", "专业"],
//	    "thumbnail": "https://...",
//	    "sort_order": 1,
//	    "is_default": false,
//	    "version": "1.0"
//	}
type UserConfigTemplate struct {
	ID           int                  `json:"id" gorm:"primaryKey;column:id" description:"模板ID"`
	UserID       string               `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"用户ID"`
	Name         string               `json:"name" gorm:"column:name;type:varchar(128);not null" description:"模板名称"`
	TemplateType ConfigTemplateType   `json:"template_type" gorm:"column:template_type;type:int;default:99;index" description:"模板类型"`
	ConfigData   string               `json:"config_data" gorm:"column:config_data;type:json" description:"配置数据(JSON)"`
	Metadata     *string              `json:"metadata" gorm:"column:metadata;type:json" description:"扩展元数据(JSON)"`
	Status       ConfigTemplateStatus `json:"status" gorm:"column:status;type:int;default:1;index" description:"状态"`
	CreatedAt    time.Time            `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt    time.Time            `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// TableName 指定表名
func (UserConfigTemplate) TableName() string {
	return "user_config_templates"
}

// SaveConfigTemplateRequest 保存配置模板请求
type SaveConfigTemplateRequest struct {
	ID           *int                  `json:"id" description:"模板ID，更新时必传"`
	UserID       string                `json:"user_id" description:"用户ID，创建时必传"`
	Name         *string               `json:"name" binding:"omitempty,max=128" description:"模板名称"`
	TemplateType *ConfigTemplateType   `json:"template_type" description:"模板类型"`
	ConfigData   interface{}           `json:"config_data" description:"配置数据(JSON对象)"`
	Metadata     interface{}           `json:"metadata" description:"扩展元数据(JSON对象)"`
	Status       *ConfigTemplateStatus `json:"status" description:"状态"`
}

// ConfigTemplateListRequest 配置模板列表请求
type ConfigTemplateListRequest struct {
	Page         int                   `form:"page" binding:"min=1" description:"页码"`
	PageSize     int                   `form:"page_size" binding:"min=1,max=100" description:"每页数量"`
	TemplateType *ConfigTemplateType   `form:"template_type" description:"模板类型"`
	Status       *ConfigTemplateStatus `form:"status" description:"状态"`
	Name         string                `form:"name" description:"名称搜索"`
	UserID       string                `form:"user_id" description:"用户ID"`
}

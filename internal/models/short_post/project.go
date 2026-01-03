package short_post

import (
	"time"

	"01agent_server/internal/models"
)

// ProjectStatus 工程状态枚举
type ProjectStatus string

const (
	ProjectStatusDraft    ProjectStatus = "draft"    // 草稿中
	ProjectStatusSaved    ProjectStatus = "saved"    // 已保存
	ProjectStatusArchived ProjectStatus = "archived" // 已归档
)

// ExportFormat 导出格式枚举
type ExportFormat string

const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatImage    ExportFormat = "image"
	ExportFormatPDF      ExportFormat = "pdf"
	ExportFormatHTML     ExportFormat = "html"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatZIP      ExportFormat = "zip"
	ExportFormatAudio    ExportFormat = "audio"
	ExportFormatDOC      ExportFormat = "doc"
	ExportFormatDOCX     ExportFormat = "docx"
	ExportFormatPPT      ExportFormat = "ppt"
	ExportFormatPPTX     ExportFormat = "pptx"
	ExportFormatXLS      ExportFormat = "xls"
	ExportFormatXLSX     ExportFormat = "xlsx"
	ExportFormatCSV      ExportFormat = "csv"
	ExportFormatTSV      ExportFormat = "tsv"
	ExportFormatVideo    ExportFormat = "video"
)

// ProjectType 工程类型枚举
type ProjectType string

const (
	ProjectTypeXiaohongshu ProjectType = "xiaohongshu" // 小红书
	ProjectTypeOther       ProjectType = "other"       // 其他
	ProjectTypeLongPost    ProjectType = "long_post"   // 长图文
	ProjectTypeShortPost   ProjectType = "short_post"  // 短图文
	ProjectTypePoster      ProjectType = "poster"      // 海报
)

// ShortPostProject 短图文工程主模型（父表 - 轻量级）
type ShortPostProject struct {
	ID          string        `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"工程ID"`
	UserID      string        `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	ThreadID    *string       `json:"thread_id" gorm:"column:thread_id;type:varchar(100);index" description:"关联的对话线程ID"`
	Name        string        `json:"name" gorm:"column:name;type:varchar(200);not null" description:"工程名称"`
	Description *string       `json:"description" gorm:"column:description;type:longtext" description:"工程描述"`
	CoverImage  *string       `json:"cover_image" gorm:"column:cover_image;type:varchar(500)" description:"封面图片URL"`
	Thumbnail   *string       `json:"thumbnail" gorm:"column:thumbnail;type:varchar(500)" description:"缩略图URL"`
	ProjectType ProjectType   `json:"project_type" gorm:"column:project_type;type:varchar(20);default:'xiaohongshu'" description:"工程类型"`
	Metadata    *string       `json:"metadata" gorm:"column:metadata;type:json" description:"工程元数据"`
	Status      ProjectStatus `json:"status" gorm:"column:status;type:varchar(20);default:'draft'" description:"工程状态"`
	FrameCount  int           `json:"frame_count" gorm:"column:frame_count;default:0" description:"Frame节点数量"`
	CreatedAt   time.Time     `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt   time.Time     `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;index" description:"更新时间"`
	SavedAt     *time.Time    `json:"saved_at" gorm:"column:saved_at" description:"最后保存时间"`

	// 关联关系
	User *models.User `json:"user,omitempty" gorm:"-"`
}

// ShortPostProjectContent 短图文工程内容模型（子表 - 存储JSON数据）
type ShortPostProjectContent struct {
	ID           string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"内容ID"`
	ProjectID    string    `json:"project_id" gorm:"column:project_id;type:char(36);not null;index" description:"关联工程ID"`
	CanvasConfig *string   `json:"canvas_config" gorm:"column:canvas_config;type:json" description:"画板配置"`
	FramesData   *string   `json:"frames_data" gorm:"column:frames_data;type:json" description:"Frame节点数据列表"`
	ElementsData *string   `json:"elements_data" gorm:"column:elements_data;type:json" description:"画板元素数据"`
	Metadata     *string   `json:"metadata" gorm:"column:metadata;type:json" description:"工程元数据"`
	Version      int       `json:"version" gorm:"column:version;default:1" description:"内容版本号"`
	IsLatest     bool      `json:"is_latest" gorm:"column:is_latest;default:true;index" description:"是否是最新版本"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	Project *ShortPostProject `json:"project,omitempty" gorm:"-"`
}

// ShortPostProjectCopywriting 短图文工程文案模型
type ShortPostProjectCopywriting struct {
	ID        string    `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"文案ID"`
	ProjectID string    `json:"project_id" gorm:"column:project_id;type:char(36);not null;uniqueIndex" description:"关联工程ID"`
	Title     *string   `json:"title" gorm:"column:title;type:varchar(200)" description:"标题"`
	Content   *string   `json:"content" gorm:"column:content;type:longtext" description:"内容"`
	Topics    *string   `json:"topics" gorm:"column:topics;type:json" description:"话题标签列表（数组）"`
	Images    *string   `json:"images" gorm:"column:images;type:json" description:"图片URL列表（数组）"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`

	// 关联关系
	Project *ShortPostProject `json:"project,omitempty" gorm:"-"`
}

// ShortPostExportRecord 短图文导出记录模型
type ShortPostExportRecord struct {
	ID           string       `json:"id" gorm:"primaryKey;column:id;type:char(36)" description:"导出记录ID"`
	UserID       string       `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	ExportName   string       `json:"export_name" gorm:"column:export_name;type:varchar(200);not null" description:"导出名称"`
	ExportFormat ExportFormat `json:"export_format" gorm:"column:export_format;type:varchar(20);not null" description:"导出格式"`
	FileURLs     *string      `json:"file_urls" gorm:"column:file_urls;type:json" description:"导出文件URL列表"`
	FileSize     *int64       `json:"file_size" gorm:"column:file_size" description:"文件大小（字节）"`
	ExportConfig *string      `json:"export_config" gorm:"column:export_config;type:json" description:"导出配置参数"`
	ExportedData *string      `json:"exported_data" gorm:"column:exported_data;type:json" description:"导出的数据快照"`
	CreatedAt    time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime;index" description:"创建时间"`

	// 关联关系
	User *models.User `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (ShortPostProject) TableName() string {
	return "short_post_projects"
}

func (ShortPostProjectContent) TableName() string {
	return "short_post_project_contents"
}

func (ShortPostProjectCopywriting) TableName() string {
	return "short_post_project_copywriting"
}

func (ShortPostExportRecord) TableName() string {
	return "short_post_export_records"
}

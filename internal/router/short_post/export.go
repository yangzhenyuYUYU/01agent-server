package short_post

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models/short_post"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExportHandler short post export handler
type ExportHandler struct {
	db *gorm.DB
}

// NewExportHandler create export handler
func NewExportHandler() *ExportHandler {
	return &ExportHandler{
		db: repository.DB,
	}
}

// ========================= Request/Response Models =========================

// CreateExportRequest create export request
type CreateExportRequest struct {
	ExportName  string                  `json:"export_name" binding:"required,max=200"`
	ExportFormat short_post.ExportFormat `json:"export_format" binding:"required"`
	FileURLs    []map[string]interface{} `json:"file_urls"`
	FileSize    *int64                  `json:"file_size"`
	ExportConfig map[string]interface{}  `json:"export_config"`
	ExportedData map[string]interface{}  `json:"exported_data"`
}

// ========================= Export Handlers =========================

// CreateExportRecord create export record
func (h *ExportHandler) CreateExportRecord(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req CreateExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	// 序列化 JSON 字段
	var fileURLsJSON, exportConfigJSON, exportedDataJSON *string
	if req.FileURLs != nil {
		bytes, _ := json.Marshal(req.FileURLs)
		str := string(bytes)
		fileURLsJSON = &str
	}
	if req.ExportConfig != nil {
		bytes, _ := json.Marshal(req.ExportConfig)
		str := string(bytes)
		exportConfigJSON = &str
	}
	if req.ExportedData != nil {
		bytes, _ := json.Marshal(req.ExportedData)
		str := string(bytes)
		exportedDataJSON = &str
	}

	exportRecord := &short_post.ShortPostExportRecord{
		ID:           uuid.New().String(),
		UserID:       userID,
		ExportName:   req.ExportName,
		ExportFormat: req.ExportFormat,
		FileURLs:     fileURLsJSON,
		FileSize:     req.FileSize,
		ExportConfig: exportConfigJSON,
		ExportedData: exportedDataJSON,
	}

	if err := h.db.Create(exportRecord).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建导出记录失败: %v", err)))
		return
	}

	// 解析 file_urls 返回
	var fileURLs interface{}
	if fileURLsJSON != nil {
		json.Unmarshal([]byte(*fileURLsJSON), &fileURLs)
	}

	middleware.Success(c, "创建成功", gin.H{
		"id":           exportRecord.ID,
		"export_name":  exportRecord.ExportName,
		"export_format": string(exportRecord.ExportFormat),
		"file_urls":    fileURLs,
		"created_at":   exportRecord.CreatedAt.Format(time.RFC3339),
	})
}

// GetExportList get export list
func (h *ExportHandler) GetExportList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	exportFormatStr := c.Query("export_format")

	query := h.db.Model(&short_post.ShortPostExportRecord{}).Where("user_id = ?", userID)

	if exportFormatStr != "" {
		exportFormat := short_post.ExportFormat(exportFormatStr)
		query = query.Where("export_format = ?", exportFormat)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	var records []short_post.ShortPostExportRecord
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(records))
	for _, r := range records {
		// 解析 file_urls
		var fileURLs interface{}
		if r.FileURLs != nil {
			json.Unmarshal([]byte(*r.FileURLs), &fileURLs)
		}

		item := map[string]interface{}{
			"id":           r.ID,
			"export_name":  r.ExportName,
			"export_format": string(r.ExportFormat),
			"file_urls":    fileURLs,
			"file_size":    r.FileSize,
			"created_at":   r.CreatedAt.Format(time.RFC3339),
		}

		items = append(items, item)
	}

	middleware.Success(c, "success", gin.H{
		"items":      items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// GetExportDetail get export detail
func (h *ExportHandler) GetExportDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	exportID := c.Param("export_id")

	var record short_post.ShortPostExportRecord
	if err := h.db.Where("id = ? AND user_id = ?", exportID, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "导出记录不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 解析 JSON 字段
	var fileURLs, exportConfig, exportedData interface{}
	if record.FileURLs != nil {
		json.Unmarshal([]byte(*record.FileURLs), &fileURLs)
	}
	if record.ExportConfig != nil {
		json.Unmarshal([]byte(*record.ExportConfig), &exportConfig)
	}
	if record.ExportedData != nil {
		json.Unmarshal([]byte(*record.ExportedData), &exportedData)
	}

	middleware.Success(c, "success", gin.H{
		"id":            record.ID,
		"export_name":   record.ExportName,
		"export_format": string(record.ExportFormat),
		"file_urls":     fileURLs,
		"file_size":     record.FileSize,
		"export_config": exportConfig,
		"exported_data": exportedData,
		"created_at":    record.CreatedAt.Format(time.RFC3339),
	})
}

// DeleteExportRecord delete export record
func (h *ExportHandler) DeleteExportRecord(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	exportID := c.Param("export_id")

	var record short_post.ShortPostExportRecord
	if err := h.db.Where("id = ? AND user_id = ?", exportID, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "导出记录不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&record).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "删除成功", nil)
}

// BatchDeleteExportRequest batch delete request
type BatchDeleteExportRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// BatchDeleteExportRecords batch delete export records
func (h *ExportHandler) BatchDeleteExportRecords(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req BatchDeleteExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "参数错误：需要提供ids列表"))
		return
	}

	if len(req.IDs) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ids列表不能为空"))
		return
	}

	result := h.db.Where("id IN ? AND user_id = ?", req.IDs, userID).Delete(&short_post.ShortPostExportRecord{})
	if result.Error != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("批量删除失败: %v", result.Error)))
		return
	}

	middleware.Success(c, fmt.Sprintf("成功删除 %d 条记录", result.RowsAffected), gin.H{
		"deleted_count": result.RowsAffected,
	})
}


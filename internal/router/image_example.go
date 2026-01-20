package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImageExampleHandler image example handler
type ImageExampleHandler struct {
	db *gorm.DB
}

// NewImageExampleHandler create image example handler
func NewImageExampleHandler() *ImageExampleHandler {
	return &ImageExampleHandler{
		db: repository.DB,
	}
}

// ========================= Request/Response Models =========================

// ListImageExampleParams list image example params
type ListImageExampleParams struct {
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
	Name        *string                `json:"name"`
	Tags        []string               `json:"tags"`
	ExtraData   map[string]interface{} `json:"extra_data"`
	IsVisible   *bool                  `json:"is_visible"`
	ProjectType *string                `json:"project_type"`
}

// SaveImageExampleParams save image example params
type SaveImageExampleParams struct {
	ID          *int                   `json:"id"`
	Name        *string                `json:"name"`
	Prompt      *string                `json:"prompt"`
	CoverURL    *string                `json:"cover_url"`
	JsxCode     *string                `json:"jsx_code"`
	Tags        []string               `json:"tags"`
	ExtraData   map[string]interface{} `json:"extra_data"`
	NodeData    map[string]interface{} `json:"node_data"`
	Images      []interface{}          `json:"images"`
	ProjectType *string                `json:"project_type"`
	SortOrder   *int                   `json:"sort_order"`
	IsVisible   *bool                  `json:"is_visible"`
}

// ========================= Image Example Handlers =========================

// GetImageExampleList get image example list
func (h *ImageExampleHandler) GetImageExampleList(c *gin.Context) {
	var req ListImageExampleParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	// 设置默认值
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.ImageExample{}).Where("is_visible = ?", true)

	if req.Name != nil && *req.Name != "" {
		query = query.Where("name LIKE ?", "%"+*req.Name+"%")
	}

	// Tags 过滤：GORM 对 JSON 字段的查询支持有限，这里使用 LIKE 查询
	if len(req.Tags) > 0 {
		// 简化处理：使用第一个标签进行 LIKE 查询
		query = query.Where("tags LIKE ?", "%"+req.Tags[0]+"%")
	}

	// ExtraData 过滤：逐个键值对进行 LIKE 查询，模拟 Tortoise ORM 的 __contains 行为
	for key, value := range req.ExtraData {
		// 特殊处理：如果 size 是 "1080×自适应"，转换为 width:1080 查询
		if key == "size" {
			if size, ok := value.(string); ok && size == "1080×自适应" {
				query = query.Where("extra_data LIKE ?", "%\"width\":1080%")
				continue
			}
		}
		// 将单个键值对序列化为 JSON 片段进行匹配
		valueBytes, _ := json.Marshal(value)
		// 构造匹配模式，如 "size":"2560×1080"
		pattern := fmt.Sprintf(`"%s":%s`, key, string(valueBytes))
		query = query.Where("extra_data LIKE ?", "%"+pattern+"%")
	}

	if req.ProjectType != nil && *req.ProjectType != "" {
		query = query.Where("project_type = ?", *req.ProjectType)
	}

	var total int64
	query.Count(&total)

	var items []models.ImageExample
	if err := query.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 如果 project_type 是 xiaohongshu，需要获取 images 信息
	needImages := req.ProjectType != nil && *req.ProjectType == "xiaohongshu"

	itemsData := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		// 解析 JSON 字段
		var tags interface{}
		if item.Tags != nil {
			json.Unmarshal([]byte(*item.Tags), &tags)
		}

		var extraData interface{}
		if item.ExtraData != nil {
			json.Unmarshal([]byte(*item.ExtraData), &extraData)
		}

		itemData := map[string]interface{}{
			"id":           item.ID,
			"name":         item.Name,
			"prompt":       item.Prompt,
			"cover_url":    item.CoverURL,
			"tags":         tags,
			"sort_order":   item.SortOrder,
			"is_visible":   item.IsVisible,
			"extra_data":   extraData,
			"project_type": item.ProjectType,
			"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 如果是 xiaohongshu 类型，获取 images 信息
		if needImages || item.ProjectType == "xiaohongshu" {
			var detail models.ImageExampleDetail
			if err := h.db.Where("example_id = ?", item.ID).First(&detail).Error; err == nil {
				var images interface{}
				if detail.Images != nil {
					json.Unmarshal([]byte(*detail.Images), &images)
				}
				itemData["images"] = images
			} else {
				itemData["images"] = nil
			}
		}

		itemsData = append(itemsData, itemData)
	}

	middleware.Success(c, "success", gin.H{
		"items":     itemsData,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// SaveImageExample save image example
func (h *ImageExampleHandler) SaveImageExample(c *gin.Context) {
	var req SaveImageExampleParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	var item models.ImageExample
	var isNew bool

	if req.ID != nil && *req.ID > 0 {
		// 更新模式
		if err := h.db.Where("id = ?", *req.ID).First(&item).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到示例"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
			return
		}
		isNew = false
	} else {
		// 创建模式
		if req.Name == nil || *req.Name == "" {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "名称不能为空"))
			return
		}
		if req.Prompt == nil || *req.Prompt == "" {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "提示词不能为空"))
			return
		}

		projectTypeValue := "other"
		if req.ProjectType != nil && *req.ProjectType == "xiaohongshu" {
			projectTypeValue = "xiaohongshu"
		}

		item = models.ImageExample{
			Name:        *req.Name,
			Prompt:      *req.Prompt,
			ProjectType: projectTypeValue,
			IsVisible:   true,
			SortOrder:   0,
		}
		if req.IsVisible != nil {
			item.IsVisible = *req.IsVisible
		}
		if req.SortOrder != nil {
			item.SortOrder = *req.SortOrder
		}
		isNew = true
	}

	// 更新主表字段
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Prompt != nil {
		item.Prompt = *req.Prompt
	}
	if req.CoverURL != nil {
		item.CoverURL = req.CoverURL
	}
	if req.Tags != nil {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsStr := string(tagsBytes)
		item.Tags = &tagsStr
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.IsVisible != nil {
		item.IsVisible = *req.IsVisible
	}
	if req.ExtraData != nil {
		extraDataBytes, _ := json.Marshal(req.ExtraData)
		extraDataStr := string(extraDataBytes)
		item.ExtraData = &extraDataStr
	}
	if req.ProjectType != nil {
		// 只有明确传入 "xiaohongshu" 时才使用，否则使用 "other"
		if *req.ProjectType == "xiaohongshu" {
			item.ProjectType = "xiaohongshu"
		} else {
			item.ProjectType = "other"
		}
	}

	if isNew {
		if err := h.db.Create(&item).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
			return
		}
	} else {
		if err := h.db.Save(&item).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
	}

	// 处理详情表（存储大字段）
	var detail models.ImageExampleDetail
	if err := h.db.Where("example_id = ?", item.ID).First(&detail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新详情记录
			detail = models.ImageExampleDetail{
				ExampleID: item.ID,
			}
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询详情失败: %v", err)))
			return
		}
	}

	// 更新详情字段
	if req.JsxCode != nil {
		detail.JsxCode = req.JsxCode
	}
	if req.NodeData != nil {
		nodeDataBytes, _ := json.Marshal(req.NodeData)
		nodeDataStr := string(nodeDataBytes)
		detail.NodeData = &nodeDataStr
	}
	if req.Images != nil {
		imagesBytes, _ := json.Marshal(req.Images)
		imagesStr := string(imagesBytes)
		detail.Images = &imagesStr
	}

	if detail.ID == 0 {
		// 创建新记录
		if err := h.db.Create(&detail).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建详情失败: %v", err)))
			return
		}
	} else {
		// 更新现有记录
		if err := h.db.Save(&detail).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新详情失败: %v", err)))
			return
		}
	}

	// 重新获取详情数据用于返回
	h.db.Where("example_id = ?", item.ID).First(&detail)

	// 解析 JSON 字段返回
	var tags interface{}
	if item.Tags != nil {
		json.Unmarshal([]byte(*item.Tags), &tags)
	}

	var extraData interface{}
	if item.ExtraData != nil {
		json.Unmarshal([]byte(*item.ExtraData), &extraData)
	}

	var nodeData interface{}
	if detail.NodeData != nil {
		json.Unmarshal([]byte(*detail.NodeData), &nodeData)
	}

	var images interface{}
	if detail.Images != nil {
		json.Unmarshal([]byte(*detail.Images), &images)
	}

	middleware.Success(c, "success", gin.H{
		"id":           item.ID,
		"name":         item.Name,
		"prompt":       item.Prompt,
		"cover_url":    item.CoverURL,
		"jsx_code":     detail.JsxCode,
		"tags":         tags,
		"node_data":    nodeData,
		"images":       images,
		"project_type": item.ProjectType,
		"sort_order":   item.SortOrder,
		"is_visible":   item.IsVisible,
		"extra_data":   extraData,
		"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetImageExampleDetail get image example detail
func (h *ImageExampleHandler) GetImageExampleDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var item models.ImageExample
	if err := h.db.Where("id = ?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到示例"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 关联查询详情表获取大字段数据
	var detail models.ImageExampleDetail
	h.db.Where("example_id = ?", id).First(&detail)

	// 解析 JSON 字段
	var tags interface{}
	if item.Tags != nil {
		json.Unmarshal([]byte(*item.Tags), &tags)
	}

	var extraData interface{}
	if item.ExtraData != nil {
		json.Unmarshal([]byte(*item.ExtraData), &extraData)
	}

	var nodeData interface{}
	if detail.NodeData != nil {
		json.Unmarshal([]byte(*detail.NodeData), &nodeData)
	}

	var images interface{}
	if detail.Images != nil {
		json.Unmarshal([]byte(*detail.Images), &images)
	}

	middleware.Success(c, "success", gin.H{
		"id":           item.ID,
		"name":         item.Name,
		"prompt":       item.Prompt,
		"cover_url":    item.CoverURL,
		"jsx_code":     detail.JsxCode,
		"tags":         tags,
		"node_data":    nodeData,
		"images":       images,
		"project_type": item.ProjectType,
		"sort_order":   item.SortOrder,
		"is_visible":   item.IsVisible,
		"extra_data":   extraData,
		"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteImageExample delete image example
func (h *ImageExampleHandler) DeleteImageExample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "ID格式错误"))
		return
	}

	var item models.ImageExample
	if err := h.db.Where("id = ?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到示例"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 删除详情记录（如果有）
	h.db.Where("example_id = ?", id).Delete(&models.ImageExampleDetail{})

	// 删除主表记录
	if err := h.db.Delete(&item).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// ReorderImageExample reorder image example sort_order
func (h *ImageExampleHandler) ReorderImageExample(c *gin.Context) {
	// 获取所有记录，按创建时间排序
	var items []models.ImageExample
	if err := h.db.Order("created_at ASC").Find(&items).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if len(items) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "没有需要排序的记录"))
		return
	}

	// 使用事务确保数据一致性
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updatedCount := 0
	for index, item := range items {
		// 从0开始，依次设置sort_order
		if item.SortOrder != index {
			if err := tx.Model(&item).Updates(map[string]interface{}{
				"sort_order": index,
			}).Error; err != nil {
				tx.Rollback()
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新排序失败: %v", err)))
				return
			}
			updatedCount++
		}
	}

	if err := tx.Commit().Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("提交事务失败: %v", err)))
		return
	}

	middleware.Success(c, "success", gin.H{
		"updated_count": updatedCount,
		"total_count":   len(items),
		"message":       fmt.Sprintf("成功更新 %d 条记录的排序", updatedCount),
	})
}

// SetupImageExampleRoutes setup image example routes
func SetupImageExampleRoutes(r *gin.Engine) {
	handler := NewImageExampleHandler()

	imageExampleGroup := r.Group("/api/v1/image-example")
	imageExampleGroup.Use(middleware.JWTAuth())
	{
		imageExampleGroup.POST("/list", handler.GetImageExampleList)
		imageExampleGroup.POST("/save", handler.SaveImageExample)
		imageExampleGroup.GET("/:id", handler.GetImageExampleDetail)
		imageExampleGroup.DELETE("/:id", handler.DeleteImageExample)
		imageExampleGroup.POST("/reorder", handler.ReorderImageExample)
	}
}

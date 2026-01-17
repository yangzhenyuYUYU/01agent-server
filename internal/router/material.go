package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MaterialHandler material handler
type MaterialHandler struct {
	db             *gorm.DB
	benefitService *service.BenefitService
}

// NewMaterialHandler create material handler
func NewMaterialHandler() *MaterialHandler {
	return &MaterialHandler{
		db:             repository.DB,
		benefitService: service.NewBenefitService(),
	}
}

// ========================= Request/Response Models =========================

// SaveMaterialParams save material request
type SaveMaterialParams struct {
	ID           *int                `json:"id"`
	Name         *string             `json:"name"`
	Tags         []string            `json:"tags"`
	MaterialType *models.MaterialTypes `json:"material_type"`
	IsPublic     *int                `json:"is_public"`
	Data         map[string]interface{} `json:"data"`
	Size         *int64              `json:"size"`
}

// UpdateMaterialsOrderParams update materials order request
type UpdateMaterialsOrderParams struct {
	MaterialIDs []interface{} `json:"material_ids" binding:"required"` // 素材ID数组，按新顺序排列
}

// ========================= Helper Functions =========================

// getFileSizeFromURL 从 URL 获取文件大小（字节）
// 优先使用 HEAD 请求获取 Content-Length，如果失败则流式下载计算大小
func getFileSizeFromURL(url string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // 允许重定向
		},
	}

	// 先尝试 HEAD 请求获取 Content-Length（更高效）
	headReq, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %v", err)
	}

	headResp, err := client.Do(headReq)
	if err == nil {
		headResp.Body.Close()
		if headResp.StatusCode == http.StatusOK {
			contentLength := headResp.Header.Get("Content-Length")
			if contentLength != "" {
				if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
					return size, nil
				}
			}
		}
	}

	// 如果 HEAD 请求无法获取大小，使用流式下载计算
	getReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := client.Do(getReq)
	if err != nil {
		return 0, fmt.Errorf("下载文件失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 流式读取并计算大小
	var totalSize int64
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			totalSize += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("读取文件失败: %v", err)
		}
	}

	return totalSize, nil
}

// ========================= Material Handlers =========================

// GetMaterialList get material list
func (h *MaterialHandler) GetMaterialList(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	materialTypeStr := c.Query("material_type")
	isPublicStr := c.Query("is_public")
	name := c.Query("name")
	tagsStr := c.Query("tags")

	query := h.db.Model(&models.UserMaterials{}).Where("user_id = ?", userID)

	if materialTypeStr != "" {
		materialType := models.MaterialTypes(materialTypeStr)
		query = query.Where("material_type = ?", materialType)
	}

	if isPublicStr != "" {
		if isPublic, err := strconv.Atoi(isPublicStr); err == nil {
			query = query.Where("is_public = ?", isPublic)
		}
	}

	if name != "" {
		query = query.Where("name = ?", name)
	}

	// tags 过滤：如果 tags 参数是 JSON 数组字符串，需要解析后查询
	if tagsStr != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsStr), &tags); err == nil && len(tags) > 0 {
			// 这里简化处理，实际可能需要更复杂的 JSON 查询
			// 由于 GORM 对 JSON 字段的查询支持有限，这里先使用 LIKE 查询
			query = query.Where("tags LIKE ?", "%"+tags[0]+"%")
		}
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	var materials []models.UserMaterials
	// 先按 sort_order 升序排序，再按创建时间降序排序
	if err := query.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&materials).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(materials))
	for _, material := range materials {
		// 解析 JSON 字段
		var data interface{}
		if material.Data != nil {
			json.Unmarshal([]byte(*material.Data), &data)
		}

		var tags []string
		if material.Tags != nil {
			json.Unmarshal([]byte(*material.Tags), &tags)
		}

		item := map[string]interface{}{
			"id":            material.ID,
			"user_id":      material.UserID,
			"name":         material.Name,
			"material_type": string(material.MaterialType),
			"data":         data,
			"tags":         tags,
			"is_public":    material.IsPublic,
			"size":         material.Size,
			"sort_order":   material.SortOrder,
			"created_at":   material.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   material.UpdatedAt.Format("2006-01-02 15:04:05"),
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

// SaveMaterial create or update material
func (h *MaterialHandler) SaveMaterial(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req SaveMaterialParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	// 如果 size 为 0 或 None，且 data 中有 url，则自动获取文件大小
	finalSize := int64(0)
	if req.Size != nil {
		finalSize = *req.Size
	}

	// 确定要检查的 data
	dataToCheck := req.Data
	if dataToCheck == nil {
		dataToCheck = make(map[string]interface{})
	}

	// 如果 size 为 0 且 data 中有 url，尝试获取文件大小
	if finalSize == 0 && len(dataToCheck) > 0 {
		if url, ok := dataToCheck["url"].(string); ok && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
			if size, err := getFileSizeFromURL(url); err == nil {
				finalSize = size
			}
		}
	}

	var material models.UserMaterials
	var isUpdate bool

	if req.ID != nil {
		// 更新模式
		if err := h.db.Where("id = ? AND user_id = ?", *req.ID, userID).First(&material).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "素材不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
			return
		}
		isUpdate = true
	} else {
		// 创建模式
		material.UserID = userID
		material.MaterialType = models.MaterialTypeImage // 默认值
		if req.MaterialType != nil {
			material.MaterialType = *req.MaterialType
		}
		material.Name = ""
		if req.Name != nil {
			material.Name = *req.Name
		}
		material.IsPublic = 0
		if req.IsPublic != nil {
			material.IsPublic = *req.IsPublic
		}
		material.Size = finalSize
		isUpdate = false
	}

	// 更新字段
	if req.Name != nil {
		if len(*req.Name) > 50 {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "素材名称不能超过50个字符"))
			return
		}
		material.Name = *req.Name
	}

	if req.Tags != nil {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsStr := string(tagsBytes)
		material.Tags = &tagsStr
	}

	if req.IsPublic != nil {
		material.IsPublic = *req.IsPublic
	}

	if req.MaterialType != nil {
		material.MaterialType = *req.MaterialType
	}

	if req.Data != nil {
		// 如果更新了 data 且包含 url，且 size 为 0 或 None，重新获取大小
		if finalSize == 0 || (req.Size != nil && *req.Size == 0) {
			if url, ok := req.Data["url"].(string); ok && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
				if size, err := getFileSizeFromURL(url); err == nil {
					finalSize = size
				}
			}
		}

		dataBytes, _ := json.Marshal(req.Data)
		dataStr := string(dataBytes)
		material.Data = &dataStr
	} else if isUpdate && (material.Size == 0 || finalSize == 0) {
		// 如果 params.data 没有提供，且当前素材的 size 是 0，检查现有素材的 data 中是否有 URL
		if material.Data != nil {
			var data map[string]interface{}
			if json.Unmarshal([]byte(*material.Data), &data) == nil {
				if url, ok := data["url"].(string); ok && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
					if size, err := getFileSizeFromURL(url); err == nil {
						finalSize = size
					}
				}
			}
		}
	}

	// 更新 size：如果明确提供了大于 0 的 size 则使用，否则使用通过 URL 获取的大小
	if req.Size != nil && *req.Size > 0 {
		material.Size = *req.Size
	} else if finalSize > 0 {
		material.Size = finalSize
	} else if req.Size != nil {
		material.Size = 0
	}

	if isUpdate {
		if err := h.db.Save(&material).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
			return
		}
	} else {
		if err := h.db.Create(&material).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
			return
		}
	}

	// 解析 JSON 字段返回
	var data interface{}
	if material.Data != nil {
		json.Unmarshal([]byte(*material.Data), &data)
	}

	var tags []string
	if material.Tags != nil {
		json.Unmarshal([]byte(*material.Tags), &tags)
	}

	middleware.Success(c, "success", gin.H{
		"id":            material.ID,
		"user_id":      material.UserID,
		"name":         material.Name,
		"material_type": string(material.MaterialType),
		"data":         data,
		"tags":         tags,
		"is_public":    material.IsPublic,
		"size":         material.Size,
		"sort_order":   material.SortOrder,
		"created_at":   material.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   material.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// SetMaterialPublic set material public
func (h *MaterialHandler) SetMaterialPublic(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	// 注意：Python 代码中没有指定 material ID，这可能是设计问题
	// 这里我们需要从查询参数或请求体中获取 material ID
	materialIDStr := c.Query("id")
	if materialIDStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "缺少素材ID"))
		return
	}

	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "素材ID格式错误"))
		return
	}

	var material models.UserMaterials
	if err := h.db.Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到素材"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	material.IsPublic = 1
	if err := h.db.Save(&material).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// SetMaterialPrivate set material private
func (h *MaterialHandler) SetMaterialPrivate(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	materialIDStr := c.Query("id")
	if materialIDStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "缺少素材ID"))
		return
	}

	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "素材ID格式错误"))
		return
	}

	var material models.UserMaterials
	if err := h.db.Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到素材"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	material.IsPublic = 0
	if err := h.db.Save(&material).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// GetMaterialDetail get material detail
func (h *MaterialHandler) GetMaterialDetail(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	materialID := c.Param("id")

	var material models.UserMaterials
	if err := h.db.Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到对应素材"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 解析 JSON 字段
	var data interface{}
	if material.Data != nil {
		json.Unmarshal([]byte(*material.Data), &data)
	}

	var tags []string
	if material.Tags != nil {
		json.Unmarshal([]byte(*material.Tags), &tags)
	}

	middleware.Success(c, "success", gin.H{
		"id":            material.ID,
		"user_id":      material.UserID,
		"name":         material.Name,
		"material_type": string(material.MaterialType),
		"data":         data,
		"tags":         tags,
		"is_public":    material.IsPublic,
		"size":         material.Size,
		"sort_order":   material.SortOrder,
		"created_at":   material.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   material.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteMaterial delete material
func (h *MaterialHandler) DeleteMaterial(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	materialIDStr := c.Param("id")

	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "素材ID格式错误"))
		return
	}

	var material models.UserMaterials
	if err := h.db.Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "未找到素材，无法删除"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 如果素材data中有url字段，尝试删除OSS中的文件
	// TODO: 实现 OSS 文件删除功能
	// 这里先记录日志，不阻塞删除操作
	if material.Data != nil {
		var data map[string]interface{}
		if json.Unmarshal([]byte(*material.Data), &data) == nil {
			if url, ok := data["url"].(string); ok && url != "" {
				repository.Infof("尝试删除 OSS 文件: %s (TODO: 实现 OSS 删除功能)", url)
				// oss_client.delete_file(url) // TODO: 实现 OSS 删除
			}
		}
	}

	// 删除数据库中的素材记录
	if err := h.db.Delete(&material).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	middleware.Success(c, "success", nil)
}

// UpdateMaterialsOrder update materials order
func (h *MaterialHandler) UpdateMaterialsOrder(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req UpdateMaterialsOrderParams
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "参数错误：需要提供material_ids列表"))
		return
	}

	if len(req.MaterialIDs) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "素材ID数组不能为空"))
		return
	}

	// 转换 material_ids 为整数数组
	materialIDs := make([]int, 0, len(req.MaterialIDs))
	for _, mid := range req.MaterialIDs {
		switch v := mid.(type) {
		case float64:
			materialIDs = append(materialIDs, int(v))
		case int:
			materialIDs = append(materialIDs, v)
		case string:
			if id, err := strconv.Atoi(v); err == nil {
				materialIDs = append(materialIDs, id)
			}
		}
	}

	if len(materialIDs) != len(req.MaterialIDs) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "素材ID格式错误"))
		return
	}

	// 验证所有素材是否属于当前用户
	var materials []models.UserMaterials
	if err := h.db.Where("id IN ? AND user_id = ?", materialIDs, userID).Find(&materials).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if len(materials) != len(materialIDs) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "部分素材不存在或不属于当前用户"))
		return
	}

	// 创建 ID 到素材对象的映射
	materialDict := make(map[int]*models.UserMaterials)
	for i := range materials {
		materialDict[materials[i].ID] = &materials[i]
	}

	// 按数组顺序更新 sort_order
	for index, materialID := range materialIDs {
		if material, ok := materialDict[materialID]; ok {
			material.SortOrder = index
			if err := h.db.Save(material).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新排序失败: %v", err)))
				return
			}
		}
	}

	middleware.Success(c, "success", nil)
}

// GetStorageInfo get storage info
func (h *MaterialHandler) GetStorageInfo(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	// 获取用户权益信息（包含存储配额）
	benefits, err := h.benefitService.GetUserBenefits(userID)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("获取用户权益失败: %v", err)))
		return
	}

	storageQuota := int64(314572800) // 默认300MB
	if quota, ok := benefits["storage_quota"].(int64); ok {
		storageQuota = quota
	}

	// 使用数据库聚合函数直接计算已使用的存储空间总和
	var result struct {
		TotalSize int64
	}
	h.db.Model(&models.UserMaterials{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(size), 0) as total_size").
		Scan(&result)

	usedStorage := result.TotalSize

	// 计算剩余空间
	remainingStorage := storageQuota - usedStorage
	if remainingStorage < 0 {
		remainingStorage = 0
	}

	// 判断是否已满
	isFull := usedStorage >= storageQuota

	usagePercentage := 0.0
	if storageQuota > 0 {
		usagePercentage = float64(usedStorage) / float64(storageQuota) * 100
		// 四舍五入到2位小数
		usagePercentage = float64(int(usagePercentage*100+0.5)) / 100.0
	}

	middleware.Success(c, "success", gin.H{
		"used_storage":      usedStorage,
		"total_storage":     storageQuota,
		"remaining_storage": remainingStorage,
		"is_full":           isFull,
		"usage_percentage":  usagePercentage,
	})
}

// SetupMaterialRoutes setup material routes
func SetupMaterialRoutes(r *gin.Engine) {
	materialHandler := NewMaterialHandler()

	materialGroup := r.Group("/api/v1/material")
	materialGroup.Use(middleware.JWTAuth())
	{
		materialGroup.GET("/list", materialHandler.GetMaterialList)
		materialGroup.POST("/save", materialHandler.SaveMaterial)
		materialGroup.PUT("/public", materialHandler.SetMaterialPublic)
		materialGroup.PUT("/private", materialHandler.SetMaterialPrivate)
		materialGroup.GET("/:id", materialHandler.GetMaterialDetail)
		materialGroup.DELETE("/:id", materialHandler.DeleteMaterial)
		materialGroup.PUT("/order", materialHandler.UpdateMaterialsOrder)
		materialGroup.GET("/storage/info", materialHandler.GetStorageInfo)
	}
}


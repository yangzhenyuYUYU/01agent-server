package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarketingHandler marketing handler
type MarketingHandler struct {
	db    *gorm.DB
	redis *tools.Redis
}

// NewMarketingHandler create marketing handler
func NewMarketingHandler() *MarketingHandler {
	return &MarketingHandler{
		db:    repository.DB,
		redis: tools.GetRedisInstance(),
	}
}

// ========================= Request/Response Models =========================

// ActivityConfigModel activity config model
type ActivityConfigModel struct {
	Products  []map[string]interface{} `json:"products" binding:"required"`
	Benefits  map[string]interface{}   `json:"benefits" binding:"required"`
	Conditions map[string]interface{}  `json:"conditions" binding:"required"`
}

// ActivityCreateModel activity create model
type ActivityCreateModel struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description" binding:"required"`
	StartTime   time.Time           `json:"start_time" binding:"required"`
	EndTime     time.Time           `json:"end_time" binding:"required"`
	Config      ActivityConfigModel `json:"config" binding:"required"`
	IsVisible   bool                `json:"is_visible"`
}

// ActivityUpdateModel activity update model
type ActivityUpdateModel struct {
	Name        *string             `json:"name"`
	Description *string             `json:"description"`
	StartTime   *time.Time          `json:"start_time"`
	EndTime     *time.Time          `json:"end_time"`
	Config      *ActivityConfigModel `json:"config"`
	IsVisible   *bool               `json:"is_visible"`
	Status      *int16              `json:"status"`
}

// ActivityPurchaseModel activity purchase model
type ActivityPurchaseModel struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity"`
}

// ========================= Helper Functions =========================

// clearCurrentActivityCache 清理当前活动缓存
func (h *MarketingHandler) clearCurrentActivityCache() {
	cacheKey := "marketing:current_activity"
	if err := h.redis.Delete(cacheKey, 0); err != nil {
		repository.Warnf("清理缓存失败: %v", err)
	}
}

// ========================= Marketing Handlers =========================

// GetCurrentActivity get current activity
func (h *MarketingHandler) GetCurrentActivity(c *gin.Context) {
	cacheKey := "marketing:current_activity"

	// 先尝试从缓存获取
	cachedData, err := h.redis.Get(cacheKey, 0)
	if err == nil && cachedData != "" {
		var cachedActivity map[string]interface{}
		if json.Unmarshal([]byte(cachedData), &cachedActivity) == nil {
			middleware.Success(c, "获取当前活动成功", cachedActivity)
			return
		}
		// 缓存数据格式错误，删除缓存继续查询数据库
		h.redis.Delete(cacheKey, 0)
	}

	// 缓存未命中，查询数据库
	now := time.Now()
	var activity models.MarketingActivityPlan
	if err := h.db.Where("start_time <= ? AND end_time > ? AND status = ? AND is_visible = ?",
		now, now, int16(models.ActivityStatusOngoing), true).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "当前没有进行中的活动"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 解析 config JSON
	var configData map[string]interface{}
	if err := json.Unmarshal([]byte(activity.Config), &configData); err != nil {
		repository.Warnf("解析活动配置失败: %v", err)
		configData = make(map[string]interface{})
	}

	activityData := map[string]interface{}{
		"activity_id": activity.ActivityID,
		"name":        activity.Name,
		"description": activity.Description,
		"start_time":  activity.StartTime.Format(time.RFC3339),
		"end_time":    activity.EndTime.Format(time.RFC3339),
		"config":      configData,
	}

	// 存入缓存，过期时间一天（86400秒）
	activityDataBytes, _ := json.Marshal(activityData)
	if err := h.redis.Set(cacheKey, string(activityDataBytes), 86400, 0); err != nil {
		repository.Warnf("缓存写入失败: %v", err)
	}

	middleware.Success(c, "获取当前活动成功", activityData)
}

// CreateActivity create activity
func (h *MarketingHandler) CreateActivity(c *gin.Context) {
	var req ActivityCreateModel
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	if req.StartTime.After(req.EndTime) || req.StartTime.Equal(req.EndTime) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "结束时间必须大于开始时间"))
		return
	}

	now := time.Now()
	initialStatus := int16(models.ActivityStatusOngoing)
	if now.Before(req.StartTime) {
		initialStatus = int16(models.ActivityStatusPending)
	} else if now.After(req.EndTime) || now.Equal(req.EndTime) {
		initialStatus = int16(models.ActivityStatusEnded)
	}

	// 检查时间段内是否已有其他活动
	var existingActivity models.MarketingActivityPlan
	if err := h.db.Where("start_time < ? AND end_time > ? AND status IN ?",
		req.EndTime, req.StartTime, []int16{int16(models.ActivityStatusPending), int16(models.ActivityStatusOngoing)}).First(&existingActivity).Error; err == nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "该时间段内已存在其他活动"))
		return
	}

	// 序列化 config
	configBytes, _ := json.Marshal(req.Config)
	activityID := uuid.New().String()

	activity := models.MarketingActivityPlan{
		ActivityID:  activityID,
		Name:        req.Name,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Config:      string(configBytes),
		IsVisible:   req.IsVisible,
		Status:      initialStatus,
	}

	if err := h.db.Create(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("创建失败: %v", err)))
		return
	}

	// 清理当前活动缓存
	h.clearCurrentActivityCache()

	middleware.Success(c, "活动创建成功", gin.H{
		"activity_id": activity.ActivityID,
		"status":      initialStatus,
	})
}

// ListActivities list activities
func (h *MarketingHandler) ListActivities(c *gin.Context) {
	statusStr := c.Query("status")

	query := h.db.Model(&models.MarketingActivityPlan{})

	if statusStr != "" {
		var status int16
		if _, err := fmt.Sscanf(statusStr, "%d", &status); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	var activities []models.MarketingActivityPlan
	if err := query.Order("created_at DESC").Find(&activities).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	items := make([]map[string]interface{}, 0, len(activities))
	for _, activity := range activities {
		items = append(items, map[string]interface{}{
			"activity_id": activity.ActivityID,
			"name":        activity.Name,
			"description": activity.Description,
			"start_time":  activity.StartTime.Format(time.RFC3339),
			"end_time":    activity.EndTime.Format(time.RFC3339),
			"status":      activity.Status,
			"is_visible":  activity.IsVisible,
			"created_at":  activity.CreatedAt.Format(time.RFC3339),
		})
	}

	middleware.Success(c, "获取活动列表成功", items)
}

// GetActivity get activity detail
func (h *MarketingHandler) GetActivity(c *gin.Context) {
	activityID := c.Param("activity_id")

	var activity models.MarketingActivityPlan
	if err := h.db.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	// 解析 config JSON
	var configData map[string]interface{}
	json.Unmarshal([]byte(activity.Config), &configData)

	middleware.Success(c, "获取活动详情成功", map[string]interface{}{
		"activity_id": activity.ActivityID,
		"name":        activity.Name,
		"description": activity.Description,
		"start_time":  activity.StartTime.Format(time.RFC3339),
		"end_time":    activity.EndTime.Format(time.RFC3339),
		"config":      configData,
		"status":      activity.Status,
		"is_visible":  activity.IsVisible,
		"created_at":  activity.CreatedAt.Format(time.RFC3339),
		"updated_at":  activity.UpdatedAt.Format(time.RFC3339),
	})
}

// UpdateActivity update activity
func (h *MarketingHandler) UpdateActivity(c *gin.Context) {
	activityID := c.Param("activity_id")

	var req ActivityUpdateModel
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	var activity models.MarketingActivityPlan
	if err := h.db.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	startTime := activity.StartTime
	endTime := activity.EndTime

	// 更新字段
	if req.Name != nil {
		activity.Name = *req.Name
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.StartTime != nil {
		startTime = *req.StartTime
		activity.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		endTime = *req.EndTime
		activity.EndTime = *req.EndTime
	}
	if req.IsVisible != nil {
		activity.IsVisible = *req.IsVisible
	}
	if req.Status != nil {
		activity.Status = *req.Status
	}
	if req.Config != nil {
		configBytes, _ := json.Marshal(req.Config)
		activity.Config = string(configBytes)
	}

	// 验证时间
	if startTime.After(endTime) || startTime.Equal(endTime) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "结束时间必须大于开始时间"))
		return
	}

	// 如果时间变化，检查时间冲突
	if req.StartTime != nil || req.EndTime != nil {
		var existingActivity models.MarketingActivityPlan
		if err := h.db.Where("activity_id != ? AND start_time < ? AND end_time > ? AND status IN ?",
			activityID, endTime, startTime, []int16{int16(models.ActivityStatusPending), int16(models.ActivityStatusOngoing)}).First(&existingActivity).Error; err == nil {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "该时间段内已存在其他活动"))
			return
		}

		// 根据时间自动更新状态
		if req.Status == nil {
			now := time.Now()
			if now.Before(startTime) {
				activity.Status = int16(models.ActivityStatusPending)
			} else if now.After(endTime) || now.Equal(endTime) {
				activity.Status = int16(models.ActivityStatusEnded)
			} else {
				activity.Status = int16(models.ActivityStatusOngoing)
			}
		}
	}

	if err := h.db.Save(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新失败: %v", err)))
		return
	}

	// 清理当前活动缓存
	h.clearCurrentActivityCache()

	middleware.Success(c, "活动更新成功", nil)
}

// DeleteActivity delete activity
func (h *MarketingHandler) DeleteActivity(c *gin.Context) {
	activityID := c.Param("activity_id")

	var activity models.MarketingActivityPlan
	if err := h.db.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if err := h.db.Delete(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("删除失败: %v", err)))
		return
	}

	// 清理当前活动缓存
	h.clearCurrentActivityCache()

	middleware.Success(c, "活动删除成功", nil)
}

// PurchaseActivity purchase activity product
func (h *MarketingHandler) PurchaseActivity(c *gin.Context) {
	activityID := c.Param("activity_id")

	var req ActivityPurchaseModel
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err)))
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	var activity models.MarketingActivityPlan
	if err := h.db.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(http.StatusNotFound, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err)))
		return
	}

	if activity.Status != int16(models.ActivityStatusOngoing) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "活动未开始或已结束"))
		return
	}

	// 解析活动配置
	var activityConfig map[string]interface{}
	if err := json.Unmarshal([]byte(activity.Config), &activityConfig); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, "活动配置解析失败"))
		return
	}

	products, ok := activityConfig["products"].([]interface{})
	if !ok || len(products) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "活动没有产品"))
		return
	}

	// 查找匹配的产品
	var targetProduct map[string]interface{}
	for _, p := range products {
		productMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		pid, ok := productMap["product_id"].(float64)
		if ok && int(pid) == req.ProductID {
			targetProduct = productMap
			break
		}
	}

	if targetProduct == nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "产品不匹配"))
		return
	}

	// 检查库存
	stock, ok := targetProduct["stock"].(float64)
	if !ok || int(stock) < req.Quantity {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, "库存不足"))
		return
	}

	// 检查每人限购
	limitPerUser, ok := targetProduct["limit_per_user"].(float64)
	if ok && req.Quantity > int(limitPerUser) {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusBadRequest, fmt.Sprintf("每人限购%d件", int(limitPerUser))))
		return
	}

	// 更新库存
	targetProduct["stock"] = int(stock) - req.Quantity
	activityConfig["products"] = products

	// 保存更新后的配置
	configBytes, _ := json.Marshal(activityConfig)
	activity.Config = string(configBytes)
	if err := h.db.Save(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(http.StatusInternalServerError, fmt.Sprintf("更新库存失败: %v", err)))
		return
	}

	// 清理当前活动缓存
	h.clearCurrentActivityCache()

	middleware.Success(c, "活动产品购买成功, 库存已更新", gin.H{
		"product": targetProduct,
	})
}

// SetupMarketingRoutes setup marketing routes
func SetupMarketingRoutes(r *gin.Engine) {
	handler := NewMarketingHandler()

	marketingGroup := r.Group("/api/v1/marketing")
	{
		marketingGroup.GET("/current-activity", handler.GetCurrentActivity)
		marketingGroup.POST("/activities", middleware.JWTAuth(), handler.CreateActivity)
		marketingGroup.GET("/activities", middleware.JWTAuth(), handler.ListActivities)
		marketingGroup.GET("/activities/:activity_id", middleware.JWTAuth(), handler.GetActivity)
		marketingGroup.PUT("/activities/:activity_id", middleware.JWTAuth(), handler.UpdateActivity)
		marketingGroup.DELETE("/activities/:activity_id", middleware.JWTAuth(), handler.DeleteActivity)
		marketingGroup.POST("/activities/:activity_id/purchase", middleware.JWTAuth(), handler.PurchaseActivity)
	}
}


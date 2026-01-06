package admin

import (
	"encoding/json"
	"fmt"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// parseTimeString 解析时间字符串，支持多种格式
func parseTimeString(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,                  // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,              // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05",         // 不带时区
		"2006-01-02T15:04:05.000Z",    // 带毫秒和Z
		"2006-01-02T15:04:05.000000Z", // 带微秒和Z
		"2006-01-02T15:04:05Z",        // 简单Z格式
		"2006-01-02 15:04:05",         // 空格分隔
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// GetMarketingActivities 获取营销活动列表
func (h *AdminHandler) GetMarketingActivities(c *gin.Context) {
	var req struct {
		Status *int `form:"status"` // 0-待开始, 1-进行中, 2-已结束
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	query := repository.DB.Model(&models.MarketingActivityPlan{})
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var activities []models.MarketingActivityPlan
	if err := query.Order("created_at DESC").Find(&activities).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(activities))
	for _, activity := range activities {
		var config map[string]interface{}
		if activity.Config != "" {
			json.Unmarshal([]byte(activity.Config), &config)
		}

		result = append(result, gin.H{
			"activity_id": activity.ActivityID,
			"name":        activity.Name,
			"description": activity.Description,
			"start_time":  activity.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			"end_time":    activity.EndTime.Format("2006-01-02T15:04:05Z07:00"),
			"status":      activity.Status,
			"is_visible":  activity.IsVisible,
			"config":      config,
			"created_at":  activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":  activity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "获取活动列表成功", result)
}

// GetMarketingActivityDetail 获取营销活动详情
func (h *AdminHandler) GetMarketingActivityDetail(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var config map[string]interface{}
	if activity.Config != "" {
		json.Unmarshal([]byte(activity.Config), &config)
	}

	result := gin.H{
		"activity_id": activity.ActivityID,
		"name":        activity.Name,
		"description": activity.Description,
		"start_time":  activity.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		"end_time":    activity.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		"config":      config,
		"status":      activity.Status,
		"is_visible":  activity.IsVisible,
		"created_at":  activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":  activity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	middleware.Success(c, "获取活动详情成功", result)
}

// CreateMarketingActivity 创建营销活动
func (h *AdminHandler) CreateMarketingActivity(c *gin.Context) {
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description" binding:"required"`
		StartTime   string                 `json:"start_time" binding:"required"` // ISO 8601 格式
		EndTime     string                 `json:"end_time" binding:"required"`   // ISO 8601 格式
		Config      map[string]interface{} `json:"config" binding:"required"`
		IsVisible   *bool                  `json:"is_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 解析时间
	startTime, err := parseTimeString(req.StartTime)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误: "+err.Error()))
		return
	}

	endTime, err := parseTimeString(req.EndTime)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误: "+err.Error()))
		return
	}

	if startTime.After(endTime) || startTime.Equal(endTime) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间必须大于开始时间"))
		return
	}

	// 检查时间段内是否已有其他活动
	var existingActivity models.MarketingActivityPlan
	if err := repository.DB.Where("start_time < ? AND end_time > ? AND status IN ?", endTime, startTime, []int16{0, 1}).
		First(&existingActivity).Error; err == nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "该时间段内已存在其他活动"))
		return
	}

	// 转换 config 为 JSON 字符串
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "配置格式错误"))
		return
	}

	// 根据时间自动设置状态
	now := time.Now()
	status := int16(models.ActivityStatusPending)
	if startTime.Before(now) && endTime.After(now) {
		status = int16(models.ActivityStatusOngoing)
	}

	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	// 生成活动ID
	activityID := fmt.Sprintf("ACT%d", time.Now().Unix())

	activity := models.MarketingActivityPlan{
		ActivityID:  activityID,
		Name:        req.Name,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Config:      string(configJSON),
		Status:      status,
		IsVisible:   isVisible,
	}

	if err := repository.DB.Create(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动创建成功", gin.H{
		"activity_id": activity.ActivityID,
		"status":      activity.Status,
	})
}

// UpdateMarketingActivity 更新营销活动
func (h *AdminHandler) UpdateMarketingActivity(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var req struct {
		Name        *string                `json:"name"`
		Description *string                `json:"description"`
		StartTime   *string                `json:"start_time"`
		EndTime     *string                `json:"end_time"`
		Config      map[string]interface{} `json:"config"`
		IsVisible   *bool                  `json:"is_visible"`
		Status      *int16                 `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	updates := make(map[string]interface{})
	hasUpdate := false

	if req.Name != nil {
		updates["name"] = *req.Name
		hasUpdate = true
	}

	if req.Description != nil {
		updates["description"] = *req.Description
		hasUpdate = true
	}

	startTime := activity.StartTime
	endTime := activity.EndTime

	if req.StartTime != nil {
		st, err := parseTimeString(*req.StartTime)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误: "+err.Error()))
			return
		}
		startTime = st
		updates["start_time"] = startTime
		hasUpdate = true
	}

	if req.EndTime != nil {
		et, err := parseTimeString(*req.EndTime)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误: "+err.Error()))
			return
		}
		endTime = et
		updates["end_time"] = endTime
		hasUpdate = true
	}

	if req.StartTime != nil || req.EndTime != nil {
		if startTime.After(endTime) || startTime.Equal(endTime) {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间必须大于开始时间"))
			return
		}

		// 检查时间冲突
		var existingActivity models.MarketingActivityPlan
		if err := repository.DB.Where("activity_id != ? AND start_time < ? AND end_time > ? AND status IN ?", activityID, endTime, startTime, []int16{0, 1}).
			First(&existingActivity).Error; err == nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "该时间段内已存在其他活动"))
			return
		}

		// 根据时间自动更新状态
		if req.Status == nil {
			now := time.Now()
			if now.Before(startTime) {
				updates["status"] = int16(models.ActivityStatusPending)
			} else if now.After(endTime) {
				updates["status"] = int16(models.ActivityStatusEnded)
			} else {
				updates["status"] = int16(models.ActivityStatusOngoing)
			}
		}
	}

	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "配置格式错误"))
			return
		}
		updates["config"] = string(configJSON)
		hasUpdate = true
	}

	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
		hasUpdate = true
	}

	if req.Status != nil {
		updates["status"] = *req.Status
		hasUpdate = true
	}

	if !hasUpdate {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	if err := repository.DB.Model(&activity).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动更新成功", gin.H{})
}

// DeleteMarketingActivity 删除营销活动
func (h *AdminHandler) DeleteMarketingActivity(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	if err := repository.DB.Delete(&activity).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动删除成功", gin.H{})
}

// PurchaseActivityProduct 购买活动产品
func (h *AdminHandler) PurchaseActivityProduct(c *gin.Context) {
	activityID := c.Param("activity_id")
	if activityID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动ID不能为空"))
		return
	}

	var req struct {
		ProductID int `json:"product_id" binding:"required"`
		Quantity  int `json:"quantity" binding:"min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认数量
	if req.Quantity == 0 {
		req.Quantity = 1
	}

	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("activity_id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "活动不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 检查活动状态
	if activity.Status != int16(models.ActivityStatusOngoing) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动未开始或已结束"))
		return
	}

	// 解析配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(activity.Config), &config); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "活动配置解析失败"))
		return
	}

	products, ok := config["products"].([]interface{})
	if !ok || len(products) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "活动没有产品"))
		return
	}

	// 查找匹配的产品
	var targetProduct map[string]interface{}
	for _, p := range products {
		product, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		pid, ok := product["product_id"].(float64)
		if ok && int(pid) == req.ProductID {
			targetProduct = product
			break
		}
	}

	if targetProduct == nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "产品不匹配"))
		return
	}

	// 检查库存
	stock, ok := targetProduct["stock"].(float64)
	if !ok || int(stock) < req.Quantity {
		middleware.HandleError(c, middleware.NewBusinessError(400, "库存不足"))
		return
	}

	// 检查每人限购
	if limitPerUser, ok := targetProduct["limit_per_user"].(float64); ok {
		if req.Quantity > int(limitPerUser) {
			middleware.HandleError(c, middleware.NewBusinessError(400, fmt.Sprintf("每人限购%d件", int(limitPerUser))))
			return
		}
	}

	// 更新库存
	targetProduct["stock"] = stock - float64(req.Quantity)

	// 保存更新后的配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新配置失败"))
		return
	}

	if err := repository.DB.Model(&activity).Update("config", string(configJSON)).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新库存失败: "+err.Error()))
		return
	}

	middleware.Success(c, "活动产品购买成功, 库存已更新", gin.H{
		"product": targetProduct,
	})
}

// GetCurrentActivity 获取当前正在进行的活动
func (h *AdminHandler) GetCurrentActivity(c *gin.Context) {
	now := time.Now()
	var activity models.MarketingActivityPlan
	if err := repository.DB.Where("start_time <= ? AND end_time > ? AND status = ? AND is_visible = ?", now, now, int16(models.ActivityStatusOngoing), true).
		First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "当前没有进行中的活动"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 解析配置
	var config map[string]interface{}
	if activity.Config != "" {
		json.Unmarshal([]byte(activity.Config), &config)
	}

	result := gin.H{
		"activity_id": activity.ActivityID,
		"name":        activity.Name,
		"description": activity.Description,
		"start_time":  activity.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		"end_time":    activity.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		"config":      config,
	}

	middleware.Success(c, "获取当前活动成功", result)
}

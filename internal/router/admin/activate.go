package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *AdminHandler) GetActivateList(c *gin.Context) {

}

func (h *AdminHandler) GetActivateDetail(c *gin.Context) {

}

func (h *AdminHandler) CreateActivate(c *gin.Context) {

}

func (h *AdminHandler) UpdateActivate(c *gin.Context) {

}

// GetActivationCodeList 获取激活码列表
func (h *AdminHandler) GetActivationCodeList(c *gin.Context) {
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

// generateActivationCode 生成兑换码
func generateActivationCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除容易混淆的字符：0, O, 1, I
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// CreateActivationCodes 创建兑换码
func (h *AdminHandler) CreateActivationCodes(c *gin.Context) {
	var req struct {
		CardType  string `json:"card_type" binding:"required"` // "membership" 或 "credits"
		ProductID int    `json:"product_id" binding:"required"`
		Count     int    `json:"count" binding:"min=1,max=1000"`
		Remark    string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 1
	}

	// 验证卡片类型
	if req.CardType != string(models.CardTypeMembership) && req.CardType != string(models.CardTypeCredits) {
		middleware.HandleError(c, middleware.NewBusinessError(400, "无效的卡片类型"))
		return
	}

	// 验证产品是否存在
	var productName *string
	if req.CardType == string(models.CardTypeMembership) {
		var product models.Production
		if err := repository.DB.Where("id = ?", req.ProductID).First(&product).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(400, "会员产品不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询产品失败: "+err.Error()))
			return
		}
		productName = &product.Name
	} else if req.CardType == string(models.CardTypeCredits) {
		var product models.CreditProduct
		if err := repository.DB.Where("id = ?", req.ProductID).First(&product).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(400, "积分产品不存在"))
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询产品失败: "+err.Error()))
			return
		}
		productName = product.Name
	}

	// 批量创建兑换码
	createdCodes := make([]gin.H, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		// 生成唯一的兑换码
		var code string
		attempts := 0
		for attempts < 10 {
			code = generateActivationCode(10)
			// 检查是否已存在
			var existing models.ActivationCode
			if err := repository.DB.Where("code = ?", code).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					break // 代码不存在，可以使用
				}
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询兑换码失败: "+err.Error()))
				return
			}
			attempts++
		}
		if attempts >= 10 {
			middleware.HandleError(c, middleware.NewBusinessError(500, "生成兑换码失败，请重试"))
			return
		}

		// 创建兑换码记录
		remark := req.Remark
		activationCode := models.ActivationCode{
			Code:      code,
			CardType:  req.CardType,
			ProductID: req.ProductID,
			IsUsed:    false,
		}
		if remark != "" {
			activationCode.Remark = &remark
		}

		if err := repository.DB.Create(&activationCode).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建兑换码失败: "+err.Error()))
			return
		}

		createdCodes = append(createdCodes, gin.H{
			"id":         activationCode.ID,
			"code":       activationCode.Code,
			"card_type":  activationCode.CardType,
			"product_id": activationCode.ProductID,
			"remark":     activationCode.Remark,
		})
	}

	middleware.Success(c, fmt.Sprintf("成功创建 %d 个兑换码", len(createdCodes)), gin.H{
		"count":        len(createdCodes),
		"codes":        createdCodes,
		"product_name": productName,
	})
}

package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
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
		Page      int     `form:"page" binding:"min=1"`
		PageSize  int     `form:"page_size" binding:"min=1,max=9999"`
		Search    string  `form:"search"`          // 搜索激活码
		CardType  *string `form:"card_type"`       // 卡片类型筛选：membership/credits
		IsUsed    *bool   `form:"is_used"`         // 是否已使用筛选
		ProductID *int    `form:"product_id"`      // 产品ID筛选
		OrderBy   string  `form:"order_by"`        // 排序字段
		OrderDir  string  `form:"order_direction"` // 排序方向：asc/desc
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}
	if req.OrderDir == "" {
		req.OrderDir = "desc"
	}

	// 构建查询
	query := repository.DB.Model(&models.ActivationCode{})

	// 搜索激活码
	if req.Search != "" {
		query = query.Where("code LIKE ?", "%"+req.Search+"%")
	}

	// 卡片类型筛选
	if req.CardType != nil && *req.CardType != "" {
		query = query.Where("card_type = ?", *req.CardType)
	}

	// 是否已使用筛选
	if req.IsUsed != nil {
		query = query.Where("is_used = ?", *req.IsUsed)
	}

	// 产品ID筛选
	if req.ProductID != nil {
		query = query.Where("product_id = ?", *req.ProductID)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 排序
	orderField := req.OrderBy
	if req.OrderDir == "asc" {
		orderField = orderField + " ASC"
	} else {
		orderField = orderField + " DESC"
	}
	query = query.Order(orderField)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var activationCodes []models.ActivationCode
	if err := query.Offset(offset).Limit(req.PageSize).Find(&activationCodes).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 收集产品ID和用户ID，批量查询关联数据
	productIDs := make([]int, 0)
	userIDs := make([]string, 0)
	for _, code := range activationCodes {
		productIDs = append(productIDs, code.ProductID)
		if code.UsedByID != nil {
			userIDs = append(userIDs, *code.UsedByID)
		}
	}

	// 批量查询产品信息
	productMap := make(map[int]*models.Production)
	if len(productIDs) > 0 {
		var products []models.Production
		repository.DB.Where("id IN ?", productIDs).Find(&products)
		for i := range products {
			productMap[products[i].ID] = &products[i]
		}
	}

	// 批量查询用户信息
	userMap := make(map[string]*models.User)
	if len(userIDs) > 0 {
		var users []models.User
		repository.DB.Where("user_id IN ?", userIDs).Find(&users)
		for i := range users {
			userMap[users[i].UserID] = &users[i]
		}
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(activationCodes))
	for _, code := range activationCodes {
		item := gin.H{
			"id":         code.ID,
			"code":       code.Code,
			"card_type":  code.CardType,
			"product_id": code.ProductID,
			"is_used":    code.IsUsed,
			"remark":     code.Remark,
			"created_at": code.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// 添加产品信息
		if product, ok := productMap[code.ProductID]; ok {
			item["product"] = gin.H{
				"id":           product.ID,
				"name":         product.Name,
				"price":        product.Price,
				"product_type": product.ProductType,
			}
		}

		// 添加使用用户信息
		if code.UsedByID != nil {
			if user, ok := userMap[*code.UsedByID]; ok {
				item["used_by"] = gin.H{
					"user_id":  user.UserID,
					"username": user.Username,
					"phone":    user.Phone,
					"nickname": user.Nickname,
				}
			}
		}

		// 添加交易信息（如果有）
		if code.TradeID != nil {
			item["trade_id"] = *code.TradeID
		}

		result = append(result, item)
	}

	middleware.Success(c, "获取激活码列表成功", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
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

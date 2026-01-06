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

// GetTradeV2List 交易列表查询V2 - 支持复杂查询条件
func (h *AdminHandler) GetTradeV2List(c *gin.Context) {
	var req struct {
		Page           int      `json:"page" binding:"min=1"`
		PageSize       int      `json:"page_size" binding:"min=1"`
		StartDate      *string  `json:"start_date"` // YYYY-MM-DD
		EndDate        *string  `json:"end_date"`   // YYYY-MM-DD
		TradeNo        *string  `json:"trade_no"`
		PaymentStatus  []string `json:"payment_status"`
		TradeType      []string `json:"trade_type"`
		PaymentChannel []string `json:"payment_channel"`
		UserID         *string  `json:"user_id"`
		Username       *string  `json:"username"`
		Phone          *string  `json:"phone"`
		MinAmount      *float64 `json:"min_amount"`
		MaxAmount      *float64 `json:"max_amount"`
		OrderBy        *string  `json:"order_by"`
		OrderDirection *string  `json:"order_direction"` // asc | desc
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.OrderBy == nil || *req.OrderBy == "" {
		orderBy := "created_at"
		req.OrderBy = &orderBy
	}
	if req.OrderDirection == nil || *req.OrderDirection == "" {
		orderDir := "desc"
		req.OrderDirection = &orderDir
	}

	// 如果提供了交易号，直接根据交易号查询
	if req.TradeNo != nil && *req.TradeNo != "" {
		var trade models.Trade
		if err := repository.DB.Where("trade_no = ?", *req.TradeNo).First(&trade).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.Success(c, "success", gin.H{
					"total":     0,
					"items":     []gin.H{},
					"page":      req.Page,
					"page_size": req.PageSize,
				})
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}

		// 加载用户信息
		loadUserInfo(&trade)

		// 构建单个交易记录返回数据
		result := buildTradeResponse(&trade)
		middleware.Success(c, "success", gin.H{
			"total":     1,
			"items":     []gin.H{result},
			"page":      1,
			"page_size": 1,
		})
		return
	}

	// 构建复杂查询
	query := repository.DB.Model(&models.Trade{})

	// 时间范围筛选
	if req.StartDate != nil && *req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		query = query.Where("created_at >= ?", startTime)
	}

	if req.EndDate != nil && *req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		// 设置为当天23:59:59
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
		query = query.Where("created_at <= ?", endTime)
	}

	// 支付状态筛选（数组）
	if len(req.PaymentStatus) > 0 {
		query = query.Where("payment_status IN ?", req.PaymentStatus)
	}

	// 交易类型筛选（数组）
	if len(req.TradeType) > 0 {
		query = query.Where("trade_type IN ?", req.TradeType)
	}

	// 支付渠道筛选（数组）
	if len(req.PaymentChannel) > 0 {
		query = query.Where("payment_channel IN ?", req.PaymentChannel)
	}

	// 用户相关筛选
	if req.UserID != nil && *req.UserID != "" {
		query = query.Where("user_id = ?", *req.UserID)
	}

	// 用户名和手机号筛选需要 JOIN users 表
	needJoinUser := (req.Username != nil && *req.Username != "") || (req.Phone != nil && *req.Phone != "")
	if needJoinUser {
		query = query.Joins("LEFT JOIN users ON trades.user_id = users.user_id")
		if req.Username != nil && *req.Username != "" {
			query = query.Where("users.username LIKE ?", "%"+*req.Username+"%")
		}
		if req.Phone != nil && *req.Phone != "" {
			query = query.Where("users.phone LIKE ?", "%"+*req.Phone+"%")
		}
	}

	// 价格区间筛选
	if req.MinAmount != nil {
		query = query.Where("amount >= ?", *req.MinAmount)
	}

	if req.MaxAmount != nil {
		query = query.Where("amount <= ?", *req.MaxAmount)
	}

	// 价格区间验证
	if req.MinAmount != nil && req.MaxAmount != nil {
		if *req.MinAmount > *req.MaxAmount {
			middleware.HandleError(c, middleware.NewBusinessError(400, "最小金额不能大于最大金额"))
			return
		}
	}

	// 排序 - 如果使用了 JOIN，需要指定表名
	orderByField := *req.OrderBy
	if needJoinUser && orderByField != "" {
		// 检查是否是 trades 表的字段
		if orderByField == "created_at" || orderByField == "updated_at" || orderByField == "amount" {
			orderByField = "trades." + orderByField
		}
	}
	if *req.OrderDirection == "asc" {
		query = query.Order(orderByField + " ASC")
	} else {
		query = query.Order(orderByField + " DESC")
	}

	// 获取总数 - 如果使用了 JOIN，需要去重
	var total int64
	if needJoinUser {
		// 使用子查询获取符合条件的用户ID
		var userIDs []string
		userQuery := repository.DB.Model(&models.User{})
		if req.Username != nil && *req.Username != "" {
			userQuery = userQuery.Where("username LIKE ?", "%"+*req.Username+"%")
		}
		if req.Phone != nil && *req.Phone != "" {
			userQuery = userQuery.Where("phone LIKE ?", "%"+*req.Phone+"%")
		}
		userQuery.Select("user_id").Find(&userIDs)

		if len(userIDs) > 0 {
			baseQuery := repository.DB.Model(&models.Trade{})
			if req.StartDate != nil && *req.StartDate != "" {
				startTime, _ := time.Parse("2006-01-02", *req.StartDate)
				baseQuery = baseQuery.Where("created_at >= ?", startTime)
			}
			if req.EndDate != nil && *req.EndDate != "" {
				endTime, _ := time.Parse("2006-01-02", *req.EndDate)
				endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
				baseQuery = baseQuery.Where("created_at <= ?", endTime)
			}
			if len(req.PaymentStatus) > 0 {
				baseQuery = baseQuery.Where("payment_status IN ?", req.PaymentStatus)
			}
			if len(req.TradeType) > 0 {
				baseQuery = baseQuery.Where("trade_type IN ?", req.TradeType)
			}
			if len(req.PaymentChannel) > 0 {
				baseQuery = baseQuery.Where("payment_channel IN ?", req.PaymentChannel)
			}
			if req.UserID != nil && *req.UserID != "" {
				baseQuery = baseQuery.Where("user_id = ?", *req.UserID)
			}
			if req.MinAmount != nil {
				baseQuery = baseQuery.Where("amount >= ?", *req.MinAmount)
			}
			if req.MaxAmount != nil {
				baseQuery = baseQuery.Where("amount <= ?", *req.MaxAmount)
			}
			baseQuery = baseQuery.Where("user_id IN ?", userIDs)

			if err := baseQuery.Count(&total).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
				return
			}
		} else {
			// 如果没有匹配的用户，返回空结果
			middleware.Success(c, "success", gin.H{
				"total":     0,
				"items":     []gin.H{},
				"page":      req.Page,
				"page_size": req.PageSize,
			})
			return
		}
	} else {
		if err := query.Count(&total).Error; err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
			return
		}
	}

	// 分页查询 - 如果使用了 JOIN，需要指定表名
	offset := (req.Page - 1) * req.PageSize
	var trades []models.Trade
	queryFind := query
	if needJoinUser {
		queryFind = queryFind.Select("trades.*")
	}
	if err := queryFind.Offset(offset).Limit(req.PageSize).Find(&trades).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 批量加载用户信息
	userIDs := make([]string, 0, len(trades))
	for _, trade := range trades {
		if trade.UserID != "" {
			userIDs = append(userIDs, trade.UserID)
		}
	}

	// 查询所有相关用户
	usersMap := make(map[string]models.User)
	if len(userIDs) > 0 {
		var users []models.User
		repository.DB.Where("user_id IN ?", userIDs).Find(&users)
		for _, user := range users {
			usersMap[user.UserID] = user
		}
	}

	// 构建返回数据
	items := make([]gin.H, 0, len(trades))
	for _, trade := range trades {
		// 设置用户信息
		if user, ok := usersMap[trade.UserID]; ok {
			trade.User = user
		}
		items = append(items, buildTradeResponse(&trade))
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     items,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// loadUserInfo 加载用户信息
func loadUserInfo(trade *models.Trade) {
	if trade.UserID != "" {
		var user models.User
		if err := repository.DB.Where("user_id = ?", trade.UserID).First(&user).Error; err == nil {
			trade.User = user
		}
	}
}

// buildTradeResponse 构建交易响应数据
func buildTradeResponse(trade *models.Trade) gin.H {
	result := gin.H{
		"id":              trade.ID,
		"trade_no":        trade.TradeNo,
		"user_id":         trade.UserID,
		"amount":          trade.Amount,
		"trade_type":      trade.TradeType,
		"payment_channel": trade.PaymentChannel,
		"payment_status":  trade.PaymentStatus,
		"payment_id":      trade.PaymentID,
		"title":           trade.Title,
		"created_at":      trade.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 如果有用户信息
	if trade.User.UserID != "" {
		result["username"] = trade.User.Username
		result["phone"] = trade.User.Phone
		result["nickname"] = trade.User.Nickname
		result["avatar"] = trade.User.Avatar
	}

	// 解析 metadata
	if trade.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*trade.Metadata), &metadata); err == nil {
			result["metadata"] = metadata
		}
	}

	if trade.PaidAt != nil {
		result["paid_at"] = trade.PaidAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		result["paid_at"] = nil
	}

	return result
}

// GetTradeDetail 获取交易详情
func (h *AdminHandler) GetTradeDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	var trade models.Trade
	if err := repository.DB.Where("id = ?", id).First(&trade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "交易不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 加载用户信息
	loadUserInfo(&trade)

	result := buildTradeResponse(&trade)
	middleware.Success(c, "success", result)
}

// CreateTrade 创建交易
func (h *AdminHandler) CreateTrade(c *gin.Context) {
	var req struct {
		TradeNo        string                 `json:"trade_no" binding:"required"`
		UserID         string                 `json:"user_id" binding:"required"`
		Amount         float64                `json:"amount" binding:"required,min=0"`
		TradeType      string                 `json:"trade_type" binding:"required"`
		PaymentChannel string                 `json:"payment_channel" binding:"required"`
		PaymentStatus  string                 `json:"payment_status"`
		Title          string                 `json:"title" binding:"required"`
		Metadata       map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认支付状态
	if req.PaymentStatus == "" {
		req.PaymentStatus = "pending"
	}

	// 转换 metadata 为 JSON 字符串
	var metadataStr *string
	if req.Metadata != nil {
		if metadataJSON, err := json.Marshal(req.Metadata); err == nil {
			metaStr := string(metadataJSON)
			metadataStr = &metaStr
		}
	}

	trade := models.Trade{
		TradeNo:        req.TradeNo,
		UserID:         req.UserID,
		Amount:         req.Amount,
		TradeType:      req.TradeType,
		PaymentChannel: req.PaymentChannel,
		PaymentStatus:  req.PaymentStatus,
		Title:          req.Title,
		Metadata:       metadataStr,
	}

	if err := repository.DB.Create(&trade).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建失败: "+err.Error()))
		return
	}

	// 加载用户信息
	loadUserInfo(&trade)

	result := buildTradeResponse(&trade)
	middleware.Success(c, "success", result)
}

// UpdateTrade 更新交易
func (h *AdminHandler) UpdateTrade(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	var trade models.Trade
	if err := repository.DB.Where("id = ?", id).First(&trade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "交易不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	var req struct {
		Amount         *float64               `json:"amount"`
		TradeType      *string                `json:"trade_type"`
		PaymentChannel *string                `json:"payment_channel"`
		PaymentStatus  *string                `json:"payment_status"`
		Title          *string                `json:"title"`
		Metadata       map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.TradeType != nil {
		updates["trade_type"] = *req.TradeType
	}
	if req.PaymentChannel != nil {
		updates["payment_channel"] = *req.PaymentChannel
	}
	if req.PaymentStatus != nil {
		updates["payment_status"] = *req.PaymentStatus
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Metadata != nil {
		if metadataJSON, err := json.Marshal(req.Metadata); err == nil {
			metaStr := string(metadataJSON)
			updates["metadata"] = &metaStr
		}
	}

	if len(updates) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "至少需要提供一个要更新的字段"))
		return
	}

	if err := repository.DB.Model(&trade).Updates(updates).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 重新加载
	repository.DB.First(&trade, id)
	// 加载用户信息
	loadUserInfo(&trade)

	result := buildTradeResponse(&trade)
	middleware.Success(c, "success", result)
}

// DeleteTrade 删除交易
func (h *AdminHandler) DeleteTrade(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID不能为空"))
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "ID格式错误"))
		return
	}

	var trade models.Trade
	if err := repository.DB.Where("id = ?", id).First(&trade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "交易不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	if err := repository.DB.Delete(&trade).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败: "+err.Error()))
		return
	}

	middleware.Success(c, "success", gin.H{})
}

// RepairIncompleteTrades 修复未完成订单
func (h *AdminHandler) RepairIncompleteTrades(c *gin.Context) {
	var req struct {
		TradeNos []string `json:"trade_nos" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	fixList := make([]gin.H, 0)

	// 处理每个 trade_no
	for _, tradeNo := range req.TradeNos {
		var trade models.Trade
		if err := repository.DB.Where("trade_no = ?", tradeNo).First(&trade).Error; err != nil {
			continue
		}

		// 加载用户信息
		loadUserInfo(&trade)

		// 如果已经是成功状态，跳过
		if trade.PaymentStatus == "success" {
			continue
		}

		// 更新为成功状态
		now := time.Now()
		if err := repository.DB.Model(&trade).Updates(map[string]interface{}{
			"payment_status": "success",
			"paid_at":        &now,
		}).Error; err != nil {
			continue
		}

		// 解析 metadata 获取产品信息
		var metadata map[string]interface{}
		if trade.Metadata != nil {
			json.Unmarshal([]byte(*trade.Metadata), &metadata)
		}

		productID, _ := metadata["product_id"].(float64)
		if productID > 0 {
			var product models.Production
			if err := repository.DB.Where("id = ?", int(productID)).First(&product).Error; err == nil {
				fixList = append(fixList, gin.H{
					"user_id":         trade.UserID,
					"user_phone":      trade.User.Phone,
					"user_name":       trade.User.Username,
					"product_name":    product.Name,
					"product_type":    product.ProductType,
					"benefit_changes": gin.H{}, // 这里可以添加权益变更逻辑
				})
			}
		}
	}

	middleware.Success(c, "修复完成", fixList)
}

// GetProductList 获取产品列表
func (h *AdminHandler) GetProductList(c *gin.Context) {
	var req struct {
		Page           int     `form:"page" binding:"min=1"`
		PageSize       int     `form:"page_size" binding:"min=1,max=9999"`
		Search         string  `form:"search"`
		OrderBy        string  `form:"order_by"`
		Status         *int    `form:"status"`
		ProductType    *string `form:"product_type"`
		IsCustom       bool    `form:"is_custom"`
		OrderDirection string  `form:"order_direction"`
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
		req.PageSize = 10
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}

	// 构建查询
	query := repository.DB.Model(&models.Production{})

	// 产品类型筛选
	if req.ProductType != nil && *req.ProductType != "" {
		query = query.Where("product_type = ?", *req.ProductType)
	}

	// 状态筛选
	if req.Status != nil && (*req.Status == 0 || *req.Status == 1) {
		query = query.Where("status = ?", *req.Status)
	}

	// 搜索
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 排序
	orderField := req.OrderBy
	if req.OrderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var products []models.Production
	if err := query.Offset(offset).Limit(req.PageSize).Find(&products).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(products))
	for _, item := range products {
		result = append(result, gin.H{
			"id":              item.ID,
			"name":            item.Name,
			"description":     item.Description,
			"price":           item.Price,
			"original_price":  item.OriginalPrice,
			"extra_info":      item.ExtraInfo,
			"status":          item.Status,
			"validity_period": item.ValidityPeriod,
			"product_type":    item.ProductType,
			"created_at":      item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// 如果是管理员且is_custom为true，添加"专业版开通测试"产品
	if req.IsCustom {
		var customProduct models.Production
		if err := repository.DB.Where("name = ?", "专业版开通测试").First(&customProduct).Error; err == nil {
			result = append(result, gin.H{
				"id":              customProduct.ID,
				"name":            customProduct.Name,
				"description":     customProduct.Description,
				"price":           customProduct.Price,
				"original_price":  customProduct.OriginalPrice,
				"extra_info":      customProduct.ExtraInfo,
				"status":          customProduct.Status,
				"validity_period": customProduct.ValidityPeriod,
				"product_type":    customProduct.ProductType,
				"created_at":      customProduct.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"updated_at":      customProduct.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

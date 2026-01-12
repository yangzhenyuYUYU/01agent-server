package admin

import (
	"encoding/json"
	"fmt"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"
	"01agent_server/internal/tools"

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

	// 用户名和手机号筛选需要 JOIN user 表
	needJoinUser := (req.Username != nil && *req.Username != "") || (req.Phone != nil && *req.Phone != "")
	if needJoinUser {
		query = query.Joins("LEFT JOIN user ON trades.user_id = user.user_id")
		if req.Username != nil && *req.Username != "" {
			query = query.Where("user.username LIKE ?", "%"+*req.Username+"%")
		}
		if req.Phone != nil && *req.Phone != "" {
			query = query.Where("user.phone LIKE ?", "%"+*req.Phone+"%")
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
	benefitService := service.NewBenefitService()
	processedUserIDs := make(map[string]bool) // 记录已处理的用户ID，避免重复清理缓存

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
			// 即使已经是成功状态，也记录用户ID以便清理缓存
			if trade.UserID != "" {
				processedUserIDs[trade.UserID] = true
			}
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

		// 记录用户ID（订单状态已更新，需要清理缓存）
		if trade.UserID != "" {
			processedUserIDs[trade.UserID] = true
		}

		// 重新加载 trade 以获取更新后的 paid_at
		repository.DB.First(&trade, trade.ID)

		// 解析 metadata 获取产品信息
		var metadata map[string]interface{}
		if trade.Metadata != nil {
			json.Unmarshal([]byte(*trade.Metadata), &metadata)
		}

		productID, _ := metadata["product_id"].(float64)
		if productID > 0 {
			var product models.Production
			if err := repository.DB.Where("id = ?", int(productID)).First(&product).Error; err == nil {
				// 处理权益变更
				benefitChanges := gin.H{}
				if trade.User.UserID != "" {
					user := trade.User
					changes, err := benefitService.ProcessBenefitChanges(&user, &product, &trade)
					if err == nil {
						benefitChanges = gin.H{
							"old_credits":            changes["old_credits"],
							"new_credits":            changes["new_credits"],
							"old_vip_level":          changes["old_vip_level"],
							"new_vip_level":          changes["new_vip_level"],
							"monthly_credits_issued": changes["monthly_credits_issued"],
							"total_timed_credits":    changes["total_timed_credits"],
							"total_monthly_credits":  changes["total_monthly_credits"],
							"total_credits":          changes["total_credits"],
							"changes":                changes["changes"],
						}
					}
				}

				fixList = append(fixList, gin.H{
					"user_id":         trade.UserID,
					"user_phone":      trade.User.Phone,
					"user_name":       trade.User.Username,
					"product_name":    product.Name,
					"product_type":    product.ProductType,
					"benefit_changes": benefitChanges,
				})

				// 用户ID已在订单状态更新时记录，这里不需要重复记录
			}
		}
	}

	// 清理所有相关用户的 Redis 缓存（DB3）
	for userID := range processedUserIDs {
		tools.ClearUserCacheAsync(userID)
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

// GetUserProductionList 获取用户的产品列表（带关联信息）
func (h *AdminHandler) GetUserProductionList(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	var req struct {
		Page     int    `form:"page" binding:"min=1"`
		PageSize int    `form:"page_size" binding:"min=1,max=100"`
		Status   string `form:"status"`
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

	// 构建查询
	query := repository.DB.Model(&models.UserProduction{}).
		Where("user_id = ?", userID).
		Preload("User").
		Preload("Production").
		Preload("Trade")

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var userProductions []models.UserProduction
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&userProductions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	items := make([]gin.H, 0, len(userProductions))
	for _, up := range userProductions {
		item := gin.H{
			"id":            up.ID,
			"user_id":       up.UserID,
			"production_id": up.ProductionID,
			"trade_id":      up.TradeID,
			"status":        up.Status,
			"created_at":    up.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":    up.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// 添加用户信息
		if up.User != nil {
			item["user"] = gin.H{
				"user_id":  up.User.UserID,
				"username": up.User.Username,
				"phone":    up.User.Phone,
				"nickname": up.User.Nickname,
				"avatar":   up.User.Avatar,
			}
		}

		// 添加产品信息
		if up.Production != nil {
			item["production"] = gin.H{
				"id":           up.Production.ID,
				"name":         up.Production.Name,
				"description":  up.Production.Description,
				"price":        up.Production.Price,
				"product_type": up.Production.ProductType,
			}
		}

		// 添加交易信息
		if up.Trade != nil {
			item["trade"] = gin.H{
				"id":             up.Trade.ID,
				"trade_no":       up.Trade.TradeNo,
				"amount":         up.Trade.Amount,
				"payment_status": up.Trade.PaymentStatus,
				"created_at":     up.Trade.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		items = append(items, item)
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     items,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetUserProductionDetail 获取用户产品详情（带关联信息）
func (h *AdminHandler) GetUserProductionDetail(c *gin.Context) {
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

	var userProduction models.UserProduction
	if err := repository.DB.Where("id = ?", id).
		Preload("User").
		Preload("Production").
		Preload("Trade").
		First(&userProduction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户产品不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := gin.H{
		"id":            userProduction.ID,
		"user_id":       userProduction.UserID,
		"production_id": userProduction.ProductionID,
		"trade_id":      userProduction.TradeID,
		"status":        userProduction.Status,
		"created_at":    userProduction.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":    userProduction.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 添加用户信息
	if userProduction.User != nil {
		result["user"] = gin.H{
			"user_id":  userProduction.User.UserID,
			"username": userProduction.User.Username,
			"phone":    userProduction.User.Phone,
			"nickname": userProduction.User.Nickname,
			"avatar":   userProduction.User.Avatar,
		}
	}

	// 添加产品信息
	if userProduction.Production != nil {
		result["production"] = gin.H{
			"id":              userProduction.Production.ID,
			"name":            userProduction.Production.Name,
			"description":     userProduction.Production.Description,
			"price":           userProduction.Production.Price,
			"original_price":  userProduction.Production.OriginalPrice,
			"product_type":    userProduction.Production.ProductType,
			"validity_period": userProduction.Production.ValidityPeriod,
			"status":          userProduction.Production.Status,
		}
	}

	// 添加交易信息
	if userProduction.Trade != nil {
		result["trade"] = gin.H{
			"id":              userProduction.Trade.ID,
			"trade_no":        userProduction.Trade.TradeNo,
			"amount":          userProduction.Trade.Amount,
			"trade_type":      userProduction.Trade.TradeType,
			"payment_channel": userProduction.Trade.PaymentChannel,
			"payment_status":  userProduction.Trade.PaymentStatus,
			"title":           userProduction.Trade.Title,
			"created_at":      userProduction.Trade.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if userProduction.Trade.PaidAt != nil {
			result["paid_at"] = userProduction.Trade.PaidAt.Format("2006-01-02T15:04:05Z07:00")
		}
	}

	middleware.Success(c, "success", result)
}

// GetUserProductionListForCRUD 获取用户产品列表（用于标准CRUD接口，带关联信息）
func (h *AdminHandler) GetUserProductionListForCRUD(c *gin.Context) {
	var req struct {
		Page           int    `form:"page" binding:"min=1"`
		PageSize       int    `form:"page_size" binding:"min=1,max=9999"`
		Search         string `form:"search"`
		OrderBy        string `form:"order_by"`
		OrderDirection string `form:"order_direction"`
		UserID         string `form:"user_id"`
		ProductionID   string `form:"production_id"`
		TradeID        string `form:"trade_id"`
		Status         string `form:"status"`
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
	} else if req.OrderDirection != "asc" && req.OrderDirection != "desc" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: order_direction 只能是 'asc' 或 'desc'"))
		return
	}

	// 构建查询，使用Preload批量加载关联数据，避免N+1查询
	query := repository.DB.Model(&models.UserProduction{}).
		Preload("User").
		Preload("Production").
		Preload("Trade")

	// 动态筛选
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.ProductionID != "" {
		query = query.Where("production_id = ?", req.ProductionID)
	}
	if req.TradeID != "" {
		query = query.Where("trade_id = ?", req.TradeID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 搜索功能
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("user_id LIKE ? OR production_id LIKE ? OR trade_id LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 排序
	orderBy := req.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	if req.OrderDirection == "asc" {
		query = query.Order(orderBy + " ASC")
	} else {
		query = query.Order(orderBy + " DESC")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var userProductions []models.UserProduction
	if err := query.Offset(offset).Limit(req.PageSize).Find(&userProductions).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据，将user_id替换为完整的user对象
	items := make([]gin.H, 0, len(userProductions))
	for _, up := range userProductions {
		item := gin.H{
			"id":            up.ID,
			"production_id": up.ProductionID,
			"trade_id":      up.TradeID,
			"status":        up.Status,
			"created_at":    up.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":    up.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// 添加完整的用户信息（替换user_id）
		if up.User != nil {
			item["user"] = gin.H{
				"user_id":  up.User.UserID,
				"username": up.User.Username,
				"phone":    up.User.Phone,
				"nickname": up.User.Nickname,
				"avatar":   up.User.Avatar,
			}
		} else {
			// 如果关联数据未加载，至少返回user_id
			item["user"] = gin.H{
				"user_id": up.UserID,
			}
		}

		// 添加产品信息
		if up.Production != nil {
			item["production"] = gin.H{
				"id":           up.Production.ID,
				"name":         up.Production.Name,
				"description":  up.Production.Description,
				"price":        up.Production.Price,
				"product_type": up.Production.ProductType,
			}
		}

		// 添加交易信息
		if up.Trade != nil {
			item["trade"] = gin.H{
				"id":             up.Trade.ID,
				"trade_no":       up.Trade.TradeNo,
				"amount":         up.Trade.Amount,
				"payment_status": up.Trade.PaymentStatus,
			}
		}

		items = append(items, item)
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     items,
		"page":      req.Page,
		"page_size": req.PageSize,
		"ordering":  orderBy,
		"direction": req.OrderDirection,
	})
}

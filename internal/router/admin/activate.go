package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

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
	// 防御性限制，避免一次性拉取过大导致 DB 压力或响应过大
	if req.PageSize > 1000 {
		req.PageSize = 1000
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
	// 仅允许白名单字段，避免 order_by 注入
	allowedOrderBy := map[string]bool{
		"id":         true,
		"code":       true,
		"card_type":  true,
		"product_id": true,
		"is_used":    true,
		"created_at": true,
	}
	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	if !allowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	orderDir := strings.ToLower(strings.TrimSpace(req.OrderDir))
	if orderDir != "asc" {
		orderDir = "desc"
	}
	query = query.Order(orderBy + " " + orderDir)

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
	productIDSet := make(map[int]struct{})
	userIDSet := make(map[string]struct{})
	for _, code := range activationCodes {
		if _, ok := productIDSet[code.ProductID]; !ok {
			productIDs = append(productIDs, code.ProductID)
			productIDSet[code.ProductID] = struct{}{}
		}
		if code.UsedByID != nil {
			if _, ok := userIDSet[*code.UsedByID]; !ok {
				userIDs = append(userIDs, *code.UsedByID)
				userIDSet[*code.UsedByID] = struct{}{}
			}
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

// ExportActivationCodes 导出兑换码CSV表（POST方式，支持大量ID和字段选择）
func (h *AdminHandler) ExportActivationCodes(c *gin.Context) {
	var req struct {
		CardType      *string  `json:"card_type"`       // 卡片类型筛选
		IsUsed        *bool    `json:"is_used"`         // 是否已使用
		UserID        *string  `json:"user_id"`         // 用户ID筛选
		ProductID     *int     `json:"product_id"`      // 产品ID筛选
		ProductName   *string  `json:"product_name"`    // 产品名称模糊查询
		Remark        *string  `json:"remark"`          // 备注模糊查询
		StartDate     *string  `json:"start_date"`      // 创建开始时间 YYYY-MM-DD
		EndDate       *string  `json:"end_date"`        // 创建结束时间 YYYY-MM-DD
		UsedStartDate *string  `json:"used_start_date"` // 使用开始时间 YYYY-MM-DD
		UsedEndDate   *string  `json:"used_end_date"`   // 使用结束时间 YYYY-MM-DD
		IDs           []int    `json:"ids"`             // 指定要导出的ID数组
		Fields        []string `json:"fields"`          // 指定要导出的字段列表
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 构建查询条件
	query := repository.DB.Model(&models.ActivationCode{})
	queryConditions := []string{} // 用于生成文件名描述

	// 卡片类型筛选
	if req.CardType != nil && *req.CardType != "" {
		query = query.Where("card_type = ?", *req.CardType)
		cardTypeMap := map[string]string{
			string(models.CardTypeMembership): "会员卡",
			string(models.CardTypeCredits):    "积分卡",
		}
		if name, ok := cardTypeMap[*req.CardType]; ok {
			queryConditions = append(queryConditions, name)
		} else {
			queryConditions = append(queryConditions, *req.CardType)
		}
	}

	// 是否已使用筛选
	if req.IsUsed != nil {
		query = query.Where("is_used = ?", *req.IsUsed)
		if *req.IsUsed {
			queryConditions = append(queryConditions, "已使用")
		} else {
			queryConditions = append(queryConditions, "未使用")
		}
	}

	// 用户ID筛选
	if req.UserID != nil && *req.UserID != "" {
		query = query.Where("used_by_id = ?", *req.UserID)
		queryConditions = append(queryConditions, fmt.Sprintf("用户%s", *req.UserID))
	}

	// 产品ID筛选
	if req.ProductID != nil {
		query = query.Where("product_id = ?", *req.ProductID)
		queryConditions = append(queryConditions, fmt.Sprintf("产品ID%d", *req.ProductID))
	}

	// 备注模糊查询
	if req.Remark != nil && *req.Remark != "" {
		query = query.Where("remark LIKE ?", "%"+*req.Remark+"%")
		queryConditions = append(queryConditions, fmt.Sprintf("备注含'%s'", *req.Remark))
	}

	// 创建时间范围筛选
	if req.StartDate != nil && *req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		query = query.Where("created_at >= ?", startTime)
		queryConditions = append(queryConditions, fmt.Sprintf("创建时间>=%s", *req.StartDate))
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
		queryConditions = append(queryConditions, fmt.Sprintf("创建时间<=%s", *req.EndDate))
	}

	// 使用时间范围筛选（如果有used_at字段的话，这里先预留）
	// 注意：当前模型中没有used_at字段，如果需要可以后续添加

	// 如果指定了ID数组，只查询这些ID的记录
	if len(req.IDs) > 0 {
		query = query.Where("id IN ?", req.IDs)
		queryConditions = append(queryConditions, fmt.Sprintf("指定ID%d个", len(req.IDs)))
	}

	// 预加载关联数据并查询
	var codes []models.ActivationCode
	if err := query.Order("created_at DESC").Find(&codes).Error; err != nil {
		repository.Errorf("ExportActivationCodes: failed to query: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 根据产品名称进行后筛选（因为产品在不同表中）
	if req.ProductName != nil && *req.ProductName != "" {
		filteredCodes := []models.ActivationCode{}
		for _, code := range codes {
			var productName string
			if code.CardType == string(models.CardTypeMembership) {
				var product models.Production
				if err := repository.DB.Where("id = ?", code.ProductID).First(&product).Error; err == nil {
					productName = product.Name
				}
			} else if code.CardType == string(models.CardTypeCredits) {
				var product models.CreditProduct
				if err := repository.DB.Where("id = ?", code.ProductID).First(&product).Error; err == nil {
					if product.Name != nil {
						productName = *product.Name
					}
				}
			}

			if strings.Contains(strings.ToLower(productName), strings.ToLower(*req.ProductName)) {
				filteredCodes = append(filteredCodes, code)
			}
		}
		codes = filteredCodes
		queryConditions = append(queryConditions, fmt.Sprintf("产品名称含'%s'", *req.ProductName))
	}

	// 收集需要查询的产品ID和用户ID
	productIDs := make([]int, 0)
	userIDs := make([]string, 0)
	productIDSet := make(map[int]bool)
	userIDSet := make(map[string]bool)

	for _, code := range codes {
		if !productIDSet[code.ProductID] {
			productIDs = append(productIDs, code.ProductID)
			productIDSet[code.ProductID] = true
		}
		if code.UsedByID != nil && !userIDSet[*code.UsedByID] {
			userIDs = append(userIDs, *code.UsedByID)
			userIDSet[*code.UsedByID] = true
		}
	}

	// 并行查询产品信息和用户信息
	type productResult struct {
		membershipProducts []models.Production
		creditProducts     []models.CreditProduct
	}
	type userResult struct {
		users []models.User
	}

	productChan := make(chan productResult, 1)
	userChan := make(chan userResult, 1)

	// 并行查询会员产品和积分产品
	go func() {
		var membershipProducts []models.Production
		var creditProducts []models.CreditProduct

		if len(productIDs) > 0 {
			// 并行查询两种产品
			membershipChan := make(chan []models.Production, 1)
			creditChan := make(chan []models.CreditProduct, 1)

			go func() {
				var products []models.Production
				repository.DB.Where("id IN ?", productIDs).Find(&products)
				membershipChan <- products
			}()

			go func() {
				var products []models.CreditProduct
				repository.DB.Where("id IN ?", productIDs).Find(&products)
				creditChan <- products
			}()

			membershipProducts = <-membershipChan
			creditProducts = <-creditChan
		}

		productChan <- productResult{
			membershipProducts: membershipProducts,
			creditProducts:     creditProducts,
		}
	}()

	// 并行查询用户信息
	go func() {
		var users []models.User
		if len(userIDs) > 0 {
			repository.DB.Where("user_id IN ?", userIDs).Find(&users)
		}
		userChan <- userResult{users: users}
	}()

	// 等待查询结果
	productRes := <-productChan
	userRes := <-userChan

	// 构建产品映射
	productMap := make(map[int]string)
	for _, product := range productRes.membershipProducts {
		productMap[product.ID] = product.Name
	}
	for _, product := range productRes.creditProducts {
		if product.Name != nil {
			productMap[product.ID] = *product.Name
		}
	}

	// 构建用户映射
	userMap := make(map[string]*models.User)
	for i := range userRes.users {
		userMap[userRes.users[i].UserID] = &userRes.users[i]
	}

	// 定义所有可用字段映射
	fieldMap := map[string]string{
		"id":            "序号",
		"code":          "兑换码",
		"card_type":     "卡片类型",
		"product_id":    "产品ID",
		"product_name":  "产品名称",
		"is_used":       "是否已使用",
		"remark":        "备注",
		"created_at":    "创建时间",
		"user_id":       "使用者ID",
		"user_phone":    "使用者手机号",
		"user_nickname": "使用者昵称",
	}

	// 动态生成表头
	var headers []string
	if len(req.Fields) > 0 {
		// 如果指定了字段，只导出这些字段
		for _, field := range req.Fields {
			if header, ok := fieldMap[field]; ok {
				headers = append(headers, header)
			}
		}
		// 如果没有有效的字段，使用默认字段
		if len(headers) == 0 {
			headers = []string{"序号", "兑换码", "卡片类型", "产品ID", "产品名称", "是否已使用", "备注", "创建时间"}
		}
	} else {
		// 默认字段
		headers = []string{"序号", "兑换码", "卡片类型", "产品ID", "产品名称", "是否已使用", "备注", "创建时间"}

		// 如果查询了特定用户或包含已使用的兑换码，添加用户相关字段
		hasUsedCodes := false
		for _, code := range codes {
			if code.IsUsed {
				hasUsedCodes = true
				break
			}
		}
		if (req.UserID != nil && *req.UserID != "") || (req.IsUsed != nil && *req.IsUsed) || hasUsedCodes {
			headers = append(headers, "使用者ID", "使用者手机号", "使用者昵称")
		}
	}

	// 创建CSV内容
	var csvBuffer strings.Builder
	writer := csv.NewWriter(&csvBuffer)

	// 写入BOM头，确保Excel正确识别UTF-8编码
	csvBuffer.WriteString("\ufeff")

	// 写入表头
	if err := writer.Write(headers); err != nil {
		repository.Errorf("ExportActivationCodes: failed to write headers: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "生成CSV失败"))
		return
	}

	// 卡片类型中文映射
	cardTypeMap := map[string]string{
		string(models.CardTypeMembership): "会员卡",
		string(models.CardTypeCredits):    "积分卡",
	}

	// 安全的字符串处理函数
	safeStr := func(value interface{}) string {
		if value == nil {
			return ""
		}
		str := fmt.Sprintf("%v", value)
		// 移除换行符
		str = strings.ReplaceAll(str, "\n", " ")
		str = strings.ReplaceAll(str, "\r", " ")
		return strings.TrimSpace(str)
	}

	// 创建字段到索引的映射，用于快速查找
	headerIndexMap := make(map[string]int)
	for i, header := range headers {
		// 反向查找字段名
		for field, h := range fieldMap {
			if h == header {
				headerIndexMap[field] = i
				break
			}
		}
	}

	// 写入数据
	for rowIdx, code := range codes {
		// 初始化行数据
		rowData := make([]string, len(headers))

		// 根据字段填充数据
		for field, index := range headerIndexMap {
			switch field {
			case "id":
				rowData[index] = fmt.Sprintf("%d", rowIdx+1)
			case "code":
				rowData[index] = safeStr(code.Code)
			case "card_type":
				rowData[index] = safeStr(cardTypeMap[code.CardType])
			case "product_id":
				rowData[index] = fmt.Sprintf("%d", code.ProductID)
			case "product_name":
				rowData[index] = safeStr(productMap[code.ProductID])
			case "is_used":
				if code.IsUsed {
					rowData[index] = "已使用"
				} else {
					rowData[index] = "未使用"
				}
			case "remark":
				rowData[index] = safeStr(code.Remark)
			case "created_at":
				rowData[index] = code.CreatedAt.Format("2006-01-02 15:04:05")
			case "user_id":
				if code.UsedByID != nil {
					rowData[index] = safeStr(*code.UsedByID)
				} else {
					rowData[index] = ""
				}
			case "user_phone":
				if code.UsedByID != nil {
					if user, ok := userMap[*code.UsedByID]; ok && user.Phone != nil {
						rowData[index] = safeStr(*user.Phone)
					} else {
						rowData[index] = ""
					}
				} else {
					rowData[index] = ""
				}
			case "user_nickname":
				if code.UsedByID != nil {
					if user, ok := userMap[*code.UsedByID]; ok && user.Nickname != nil {
						rowData[index] = safeStr(*user.Nickname)
					} else {
						rowData[index] = ""
					}
				} else {
					rowData[index] = ""
				}
			}
		}

		if err := writer.Write(rowData); err != nil {
			repository.Errorf("ExportActivationCodes: failed to write row: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "生成CSV失败"))
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		repository.Errorf("ExportActivationCodes: failed to flush CSV: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "生成CSV失败"))
		return
	}

	csvContent := csvBuffer.String()

	// 生成动态文件名
	timestamp := time.Now().Format("20060102_150405")
	var filename string
	if len(queryConditions) > 0 {
		conditionStr := strings.Join(queryConditions[:min(3, len(queryConditions))], "_") // 最多取前3个条件
		// 移除文件名中的特殊字符，避免编码问题
		safeCondition := ""
		for _, r := range conditionStr {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' || r >= 0x4e00 && r <= 0x9fff {
				safeCondition += string(r)
			}
		}
		filename = fmt.Sprintf("activation_codes_%s_%s.csv", safeCondition, timestamp)
	} else {
		filename = fmt.Sprintf("activation_codes_all_%s.csv", timestamp)
	}

	// 文件名长度限制，防止过长
	if utf8.RuneCountInString(filename) > 80 {
		filename = fmt.Sprintf("activation_codes_filtered_%s.csv", timestamp)
	}

	// 对于包含中文的查询条件描述，使用URL编码
	conditionDesc := strings.Join(queryConditions, ",")
	if conditionDesc == "" {
		conditionDesc = "无筛选条件"
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	c.Header("X-Total-Count", fmt.Sprintf("%d", len(codes)))
	c.Header("X-Query-Conditions", url.QueryEscape(conditionDesc))

	// 返回文件流响应
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(csvContent))
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

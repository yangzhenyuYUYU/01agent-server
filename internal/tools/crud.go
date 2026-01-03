package tools

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CRUDConfig CRUD配置
type CRUDConfig struct {
	Model          interface{} // 模型实例（用于反射）
	SearchFields   []string    // 可搜索字段列表
	DefaultOrderBy string      // 默认排序字段
	RequireAdmin   bool        // 是否需要管理员权限
	PrimaryKey     string      // 主键字段名，默认为 "id"
}

// CRUDHandler CRUD处理器
type CRUDHandler struct {
	config CRUDConfig
	db     *gorm.DB
}

// NewCRUDHandler 创建CRUD处理器
func NewCRUDHandler(config CRUDConfig, db *gorm.DB) *CRUDHandler {
	if config.PrimaryKey == "" {
		config.PrimaryKey = "id"
	}
	return &CRUDHandler{config: config, db: db}
}

// List 列表查询
func (h *CRUDHandler) List(c *gin.Context) {
	var req struct {
		Page           int    `form:"page" binding:"min=1"`
		PageSize       int    `form:"page_size" binding:"min=1,max=9999"`
		Search         string `form:"search"`
		OrderBy        string `form:"order_by"`
		OrderDirection string `form:"order_direction"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	} else if req.OrderDirection != "asc" && req.OrderDirection != "desc" {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: order_direction 只能是 'asc' 或 'desc'"})
		return
	}

	// 构建查询
	query := h.db.Model(h.config.Model)

	// 动态筛选：从查询参数中提取模型字段进行筛选
	modelType := reflect.TypeOf(h.config.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 系统参数，不用于筛选
	systemParams := map[string]bool{
		"page":            true,
		"page_size":       true,
		"search":          true,
		"order_by":        true,
		"order_direction": true,
		"relations":       true,
		"relation_depth":  true,
	}

	// 遍历查询参数，动态构建筛选条件
	for key, values := range c.Request.URL.Query() {
		if systemParams[key] || len(values) == 0 {
			continue
		}

		value := values[0]
		if value == "" {
			continue
		}

		// 检查字段是否存在
		field, found := modelType.FieldByNameFunc(func(name string) bool {
			// 尝试匹配 gorm tag 中的 column 名称
			field, _ := modelType.FieldByName(name)
			if field.Tag.Get("gorm") != "" {
				column := getColumnName(field.Tag.Get("gorm"))
				return column == key || strings.ToLower(name) == strings.ToLower(key)
			}
			return strings.ToLower(name) == strings.ToLower(key)
		})

		if !found {
			continue
		}

		// 根据字段类型转换值
		switch field.Type.Kind() {
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", key), boolValue)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", key), intValue)
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", key), floatValue)
			}
		default:
			// 字符串类型，使用模糊查询
			query = query.Where(fmt.Sprintf("%s LIKE ?", key), "%"+value+"%")
		}
	}

	// 搜索功能
	if req.Search != "" && len(h.config.SearchFields) > 0 {
		var conditions []string
		var args []interface{}
		for _, field := range h.config.SearchFields {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
			args = append(args, "%"+req.Search+"%")
		}
		if len(conditions) > 0 {
			query = query.Where(strings.Join(conditions, " OR "), args...)
		}
	}

	// 排序
	orderBy := req.OrderBy
	if orderBy == "" {
		orderBy = h.config.DefaultOrderBy
	}
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
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败: " + err.Error()})
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	// 使用反射创建切片（modelType 已在前面声明）
	sliceType := reflect.SliceOf(reflect.PtrTo(modelType))
	items := reflect.New(sliceType).Interface()

	if err := query.Offset(offset).Limit(req.PageSize).Find(items).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败: " + err.Error()})
		return
	}

	// 转换为 JSON 兼容的格式
	itemsSlice := reflect.ValueOf(items).Elem()
	resultItems := make([]interface{}, itemsSlice.Len())
	for i := 0; i < itemsSlice.Len(); i++ {
		resultItems[i] = itemsSlice.Index(i).Interface()
	}

	// 构建响应
	response := gin.H{
		"total":     total,
		"items":     resultItems,
		"page":      req.Page,
		"page_size": req.PageSize,
		"ordering":  orderBy,
		"direction": req.OrderDirection,
	}

	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": response})
}

// Detail 详情查询
func (h *CRUDHandler) Detail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "缺少ID参数"})
		return
	}

	// 使用反射创建模型实例
	modelType := reflect.TypeOf(h.config.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	item := reflect.New(modelType).Interface()

	// 构建查询
	query := h.db.Model(h.config.Model)

	// 尝试转换为整数ID
	if intID, err := strconv.Atoi(id); err == nil {
		query = query.Where(fmt.Sprintf("%s = ?", h.config.PrimaryKey), intID)
	} else {
		query = query.Where(fmt.Sprintf("%s = ?", h.config.PrimaryKey), id)
	}

	if err := query.First(item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"code": 404, "msg": "记录不存在"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": item})
}

// Create 创建
func (h *CRUDHandler) Create(c *gin.Context) {
	// 使用反射创建模型实例
	modelType := reflect.TypeOf(h.config.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	modelValue := reflect.New(modelType).Interface()

	if err := c.ShouldBindJSON(modelValue); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 创建记录
	if err := h.db.Create(modelValue).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": modelValue})
}

// Update 更新
func (h *CRUDHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "缺少ID参数"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 构建查询
	query := h.db.Model(h.config.Model)

	// 尝试转换为整数ID
	if intID, err := strconv.Atoi(id); err == nil {
		query = query.Where(fmt.Sprintf("%s = ?", h.config.PrimaryKey), intID)
	} else {
		query = query.Where(fmt.Sprintf("%s = ?", h.config.PrimaryKey), id)
	}

	// 更新记录
	if err := query.Updates(data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"code": 404, "msg": "记录不存在"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}

	// 查询更新后的记录
	modelType := reflect.TypeOf(h.config.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	item := reflect.New(modelType).Interface()
	if err := query.First(item).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": item})
}

// Delete 删除
func (h *CRUDHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "缺少ID参数"})
		return
	}

	// 使用反射创建模型实例
	modelType := reflect.TypeOf(h.config.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	item := reflect.New(modelType).Interface()

	// 先查询记录是否存在
	query := h.db.Model(h.config.Model)

	// 尝试转换为整数ID
	if intID, err := strconv.Atoi(id); err == nil {
		query = query.Where(fmt.Sprintf("%s = ?", h.config.PrimaryKey), intID)
	} else {
		query = query.Where(fmt.Sprintf("%s = ?", h.config.PrimaryKey), id)
	}

	// 查询记录是否存在
	if err := query.First(item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"code": 404, "msg": "记录不存在"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败: " + err.Error()})
		return
	}

	// 物理删除记录（使用Unscoped确保真正删除，即使有软删除字段）
	result := query.Unscoped().Delete(item)
	if result.Error != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败: " + result.Error.Error()})
		return
	}

	// 检查是否真的删除了记录
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"code": 404, "msg": "记录不存在或已被删除"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "删除成功", "data": nil})
}

// getColumnName 从gorm tag中提取column名称
func getColumnName(gormTag string) string {
	// 简单的解析，查找 column:xxx
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

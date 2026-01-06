package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/service/cache"

	"github.com/gin-gonic/gin"
)

// CacheHandler 缓存管理处理器
type CacheHandler struct {
	cacheService *cache.CacheService
}

// NewCacheHandler 创建缓存管理处理器
func NewCacheHandler() *CacheHandler {
	return &CacheHandler{
		cacheService: cache.NewCacheService(),
	}
}

// ListCache 列出Redis缓存
// @Summary 列出Redis缓存
// @Description 分页列出指定Redis数据库中的缓存键，支持模式匹配
// @Tags admin-cache
// @Accept json
// @Produce json
// @Param db_index query int false "Redis数据库索引，默认为0"
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20，最大100"
// @Param pattern query string false "键匹配模式，默认为*"
// @Param keyword query string false "关键词，用于模糊查询key（如果提供，会覆盖pattern）"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/cache/list [get]
func (h *CacheHandler) ListCache(c *gin.Context) {
	var req cache.ListCacheRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	response, err := h.cacheService.ListCache(&req)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, err.Error()))
		return
	}

	// 转换为 gin.H 格式返回
	list := make([]gin.H, 0, len(response.List))
	for _, item := range response.List {
		listItem := gin.H{
			"key":  item.Key,
			"type": item.Type,
			"ttl":  item.TTL,
		}
		if item.Type == "string" {
			listItem["value"] = item.Value
			listItem["size"] = item.Size
		} else {
			listItem["length"] = item.ValueLen
		}
		list = append(list, listItem)
	}

	middleware.Success(c, "获取缓存列表成功", gin.H{
		"db_index":  response.DBIndex,
		"db_size":   response.DBSize,
		"total":     response.Total,
		"page":      response.Page,
		"page_size": response.PageSize,
		"list":      list,
	})
}

// GetCacheDetail 获取缓存详情
// @Summary 获取缓存详情
// @Description 获取指定Redis键的详细信息，包括类型、TTL、值等
// @Tags admin-cache
// @Accept json
// @Produce json
// @Param db_index query int true "Redis数据库索引"
// @Param key query string true "缓存键名"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/cache/detail [get]
func (h *CacheHandler) GetCacheDetail(c *gin.Context) {
	var req cache.GetCacheDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if req.Key == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "键名不能为空"))
		return
	}

	response, err := h.cacheService.GetCacheDetail(&req)
	if err != nil {
		// 判断是否是键不存在的错误
		if err.Error() == "键不存在" {
			middleware.HandleError(c, middleware.NewBusinessError(404, err.Error()))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, err.Error()))
		return
	}

	result := gin.H{
		"key":  response.Key,
		"type": response.Type,
		"ttl":  response.TTL,
	}

	if response.Type == "string" {
		result["value"] = response.Value
		result["size"] = response.Size
	} else {
		result["length"] = response.ValueLen
	}

	middleware.Success(c, "获取缓存详情成功", result)
}

// ClearCache 清除Redis缓存
// @Summary 清除Redis缓存
// @Description 清除指定Redis数据库的缓存，支持按模式清除或清空整个数据库
// @Tags admin-cache
// @Accept json
// @Produce json
// @Param body body cache.ClearCacheRequest true "清除缓存请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/cache/clear [post]
func (h *CacheHandler) ClearCache(c *gin.Context) {
	var req cache.ClearCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	response, err := h.cacheService.ClearCache(&req)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, err.Error()))
		return
	}

	middleware.Success(c, "清除缓存成功", gin.H{
		"db_index":      response.DBIndex,
		"pattern":       response.Pattern,
		"deleted_count": response.DeletedCount,
	})
}

// UpdateCache 更新Redis缓存
// @Summary 更新Redis缓存
// @Description 更新指定Redis键的值和/或TTL，仅支持string类型。如果value为空，则只更新TTL；如果提供了value，则同时更新值和TTL。支持ttl字段（优先）或expiration字段设置过期时间。
// @Tags admin-cache
// @Accept json
// @Produce json
// @Param body body cache.UpdateCacheRequest true "更新缓存请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/cache/update [put]
func (h *CacheHandler) UpdateCache(c *gin.Context) {
	var req cache.UpdateCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if req.Key == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "键名不能为空"))
		return
	}

	response, err := h.cacheService.UpdateCache(&req)
	if err != nil {
		// 判断是否是键不存在或类型不匹配的错误
		if err.Error() == "键不存在" || err.Error()[:len("只支持更新")] == "只支持更新" {
			middleware.HandleError(c, middleware.NewBusinessError(400, err.Error()))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, err.Error()))
		return
	}

	middleware.Success(c, "更新缓存成功", gin.H{
		"key":        response.Key,
		"value":      response.Value,
		"expiration": response.Expiration,
		"updated":    response.Updated,
	})
}

// DeleteCache 删除Redis缓存
// @Summary 删除Redis缓存
// @Description 删除指定的Redis缓存键
// @Tags admin-cache
// @Accept json
// @Produce json
// @Param body body cache.DeleteCacheRequest true "删除缓存请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/cache/delete [delete]
func (h *CacheHandler) DeleteCache(c *gin.Context) {
	var req cache.DeleteCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if len(req.Keys) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "键列表不能为空"))
		return
	}

	response, err := h.cacheService.DeleteCache(&req)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, err.Error()))
		return
	}

	middleware.Success(c, "删除缓存成功", gin.H{
		"db_index":      response.DBIndex,
		"keys":          response.Keys,
		"deleted_count": response.DeletedCount,
	})
}

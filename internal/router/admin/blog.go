package admin

import (
	"strconv"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
)

type BlogAdminHandler struct {
	blogService *service.BlogService
}

func NewBlogAdminHandler() *BlogAdminHandler {
	return &BlogAdminHandler{
		blogService: service.NewBlogService(),
	}
}

// GetBlogList 获取博客列表（管理后台）
// GET /admin/blog/list
func (h *BlogAdminHandler) GetBlogList(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	tag := c.Query("tag")
	keyword := c.Query("keyword")
	status := c.Query("status") // 管理后台可以查看所有状态
	sort := c.DefaultQuery("sort", "latest")

	// 处理is_featured参数
	var isFeatured *bool
	if featuredStr := c.Query("is_featured"); featuredStr != "" {
		featured := featuredStr == "true"
		isFeatured = &featured
	}

	// 构建查询参数
	params := repository.BlogListParams{
		Page:       page,
		PageSize:   pageSize,
		Category:   category,
		Tag:        tag,
		Keyword:    keyword,
		IsFeatured: isFeatured,
		Sort:       sort,
		Status:     status, // 管理后台可以按状态筛选
	}

	// 获取列表
	result, err := h.blogService.GetBlogList(params)
	if err != nil {
		repository.Errorf("GetBlogList failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取博客列表失败"))
		return
	}

	middleware.Success(c, "success", result)
}

// GetBlogDetail 获取博客详情（管理后台）
// GET /admin/blog/:id
func (h *BlogAdminHandler) GetBlogDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	post, err := h.blogService.GetBlogByID(id)
	if err != nil {
		repository.Errorf("GetBlogDetail failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(404, "文章不存在"))
		return
	}

	middleware.Success(c, "success", post)
}

// CreateBlog 创建博客文章
// POST /admin/blog/create
func (h *BlogAdminHandler) CreateBlog(c *gin.Context) {
	var req models.BlogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	post, err := h.blogService.CreateBlogPost(&req)
	if err != nil {
		repository.Errorf("CreateBlog failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建文章失败: "+err.Error()))
		return
	}

	middleware.Success(c, "创建成功", post)
}

// UpdateBlog 更新博客文章
// PUT /admin/blog/:id
func (h *BlogAdminHandler) UpdateBlog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	var req models.BlogUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	post, err := h.blogService.UpdateBlogPost(id, &req)
	if err != nil {
		repository.Errorf("UpdateBlog failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新文章失败: "+err.Error()))
		return
	}

	middleware.Success(c, "更新成功", post)
}

// DeleteBlog 删除博客文章
// DELETE /admin/blog/:id
func (h *BlogAdminHandler) DeleteBlog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	err := h.blogService.DeleteBlogPost(id)
	if err != nil {
		repository.Errorf("DeleteBlog failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除文章失败: "+err.Error()))
		return
	}

	middleware.Success(c, "删除成功", nil)
}

// BatchDeleteBlogs 批量删除博客文章
// POST /admin/blog/batch-delete
func (h *BlogAdminHandler) BatchDeleteBlogs(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if len(req.IDs) == 0 {
		middleware.HandleError(c, middleware.NewBusinessError(400, "请选择要删除的文章"))
		return
	}

	successCount := 0
	failCount := 0

	for _, id := range req.IDs {
		err := h.blogService.DeleteBlogPost(id)
		if err != nil {
			repository.Errorf("Delete blog %s failed: %v", id, err)
			failCount++
		} else {
			successCount++
		}
	}

	middleware.Success(c, "批量删除完成", gin.H{
		"success_count": successCount,
		"fail_count":    failCount,
		"total":         len(req.IDs),
	})
}

// UpdateBlogStatus 更新博客状态
// PUT /admin/blog/:id/status
func (h *BlogAdminHandler) UpdateBlogStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 验证状态
	if req.Status != models.BlogStatusDraft &&
		req.Status != models.BlogStatusPublished &&
		req.Status != models.BlogStatusArchived {
		middleware.HandleError(c, middleware.NewBusinessError(400, "无效的状态值"))
		return
	}

	updateReq := models.BlogUpdateRequest{
		Status: &req.Status,
	}

	post, err := h.blogService.UpdateBlogPost(id, &updateReq)
	if err != nil {
		repository.Errorf("UpdateBlogStatus failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新状态失败: "+err.Error()))
		return
	}

	middleware.Success(c, "状态更新成功", post)
}

// ToggleFeatured 切换精选状态
// PUT /admin/blog/:id/featured
func (h *BlogAdminHandler) ToggleFeatured(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	var req struct {
		IsFeatured bool `json:"is_featured"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	updateReq := models.BlogUpdateRequest{
		IsFeatured: &req.IsFeatured,
	}

	post, err := h.blogService.UpdateBlogPost(id, &updateReq)
	if err != nil {
		repository.Errorf("ToggleFeatured failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新精选状态失败: "+err.Error()))
		return
	}

	middleware.Success(c, "精选状态更新成功", post)
}

// GetBlogStats 获取博客统计信息
// GET /admin/blog/stats
func (h *BlogAdminHandler) GetBlogStats(c *gin.Context) {
	stats, err := h.blogService.GetBlogStats()
	if err != nil {
		repository.Errorf("GetBlogStats failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取统计信息失败"))
		return
	}

	middleware.Success(c, "success", stats)
}

// SetupBlogAdminRoutes 注册博客管理路由
func SetupBlogAdminRoutes(r *gin.Engine) {
	handler := NewBlogAdminHandler()

	// 博客管理路由（需要管理员权限）
	adminBlog := r.Group("/api/v1/admin/blog")
	// TODO: 添加管理员认证中间件
	// adminBlog.Use(middleware.JWTAuth(), middleware.AdminAuth())
	{
		// 列表和详情
		adminBlog.GET("/list", handler.GetBlogList)       // 获取列表
		adminBlog.GET("/:id", handler.GetBlogDetail)      // 获取详情
		adminBlog.GET("/stats", handler.GetBlogStats)     // 获取统计

		// 创建、更新、删除
		adminBlog.POST("/create", handler.CreateBlog)     // 创建文章
		adminBlog.PUT("/:id", handler.UpdateBlog)         // 更新文章
		adminBlog.DELETE("/:id", handler.DeleteBlog)      // 删除文章

		// 批量操作
		adminBlog.POST("/batch-delete", handler.BatchDeleteBlogs) // 批量删除

		// 状态管理
		adminBlog.PUT("/:id/status", handler.UpdateBlogStatus)    // 更新状态
		adminBlog.PUT("/:id/featured", handler.ToggleFeatured)    // 切换精选
	}
}


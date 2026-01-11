package router

import (
	"strconv"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
)

type BlogHandler struct {
	blogService *service.BlogService
}

// NewBlogHandler 创建博客处理器
func NewBlogHandler() *BlogHandler {
	return &BlogHandler{
		blogService: service.NewBlogService(),
	}
}

// GetBlogList 获取博客文章列表
// GET /blog/list
func (h *BlogHandler) GetBlogList(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	tag := c.Query("tag")
	keyword := c.Query("keyword")
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

// GetBlogPost 获取单篇博客文章详情
// GET /blog/:slug
func (h *BlogHandler) GetBlogPost(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章标识不能为空"))
		return
	}

	// 过滤掉特殊路由（sitemap, list等）
	if slug == "sitemap" || slug == "list" {
		middleware.HandleError(c, middleware.NewBusinessError(404, "文章不存在"))
		return
	}

	post, err := h.blogService.GetBlogBySlug(slug)
	if err != nil {
		repository.Errorf("GetBlogPost failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(404, "文章不存在"))
		return
	}

	middleware.Success(c, "success", post)
}

// GetSitemap 获取所有文章URL（用于sitemap）
// GET /blog/sitemap
func (h *BlogHandler) GetSitemap(c *gin.Context) {
	data, err := h.blogService.GetSitemapData()
	if err != nil {
		repository.Errorf("GetSitemap failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取sitemap数据失败"))
		return
	}

	middleware.Success(c, "success", data)
}

// GetRelatedPosts 获取相关文章推荐
// GET /blog/:slug/related
func (h *BlogHandler) GetRelatedPosts(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章标识不能为空"))
		return
	}

	// 先获取文章信息
	post, err := h.blogService.GetBlogBySlug(slug)
	if err != nil {
		repository.Errorf("GetBlogBySlug failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(404, "文章不存在"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))

	posts, err := h.blogService.GetRelatedPosts(post.ID, limit)
	if err != nil {
		repository.Errorf("GetRelatedPosts failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取相关文章失败"))
		return
	}

	middleware.Success(c, "success", posts)
}

// IncrementViews 增加文章浏览量统计
// POST /blog/:slug/view
func (h *BlogHandler) IncrementViews(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章标识不能为空"))
		return
	}

	// 先获取文章信息
	post, err := h.blogService.GetBlogBySlug(slug)
	if err != nil {
		// 浏览量统计失败不影响用户体验，返回成功
		repository.Errorf("GetBlogBySlug failed: %v", err)
		middleware.Success(c, "success", nil)
		return
	}

	// 浏览量统计失败不影响用户体验，返回成功
	err = h.blogService.IncrementViews(post.ID)
	if err != nil {
		repository.Errorf("IncrementViews failed: %v", err)
	}

	middleware.Success(c, "success", nil)
}

// CreateBlogPost 创建博客文章
// POST /blog/create
func (h *BlogHandler) CreateBlogPost(c *gin.Context) {
	var req models.BlogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	post, err := h.blogService.CreateBlogPost(&req)
	if err != nil {
		repository.Errorf("CreateBlogPost failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "创建文章失败: "+err.Error()))
		return
	}

	middleware.Success(c, "创建成功", post)
}

// UpdateBlogPost 更新博客文章
// PUT /blog/:id
func (h *BlogHandler) UpdateBlogPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	var req models.BlogUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	post, err := h.blogService.UpdateBlogPost(postID, &req)
	if err != nil {
		repository.Errorf("UpdateBlogPost failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新文章失败: "+err.Error()))
		return
	}

	middleware.Success(c, "更新成功", post)
}

// DeleteBlogPost 删除博客文章
// DELETE /blog/:id
func (h *BlogHandler) DeleteBlogPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "文章ID不能为空"))
		return
	}

	err := h.blogService.DeleteBlogPost(postID)
	if err != nil {
		repository.Errorf("DeleteBlogPost failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除文章失败: "+err.Error()))
		return
	}

	middleware.Success(c, "删除成功", nil)
}

// RegisterBlogRoutes 注册博客路由
func RegisterBlogRoutes(r *gin.Engine) {
	handler := NewBlogHandler()

	// 使用 /api/v1/blog 作为统一前缀
	blog := r.Group("/api/v1/blog")
	{
		// 公开接口 - 不需要认证
		blog.GET("/list", handler.GetBlogList)              // 文章列表
		blog.GET("/sitemap", handler.GetSitemap)            // Sitemap数据
		blog.GET("/:slug/related", handler.GetRelatedPosts) // 相关文章（必须在 /:slug 之前）
		blog.POST("/:slug/view", handler.IncrementViews)    // 浏览量统计
		blog.GET("/:slug", handler.GetBlogPost)             // 文章详情（放在最后，作为兜底路由）

		// 管理接口 - 需要认证（后续可添加JWT中间件）
		blog.POST("/create", handler.CreateBlogPost) // 创建文章
		blog.PUT("/:id", handler.UpdateBlogPost)     // 更新文章
		blog.DELETE("/:id", handler.DeleteBlogPost)  // 删除文章
	}
}

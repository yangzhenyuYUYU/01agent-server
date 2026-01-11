package router

import (
	"strconv"

	"01agent_server/internal/middleware"
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

// RegisterBlogRoutes 注册博客路由
func RegisterBlogRoutes(r *gin.Engine) {
	handler := NewBlogHandler()

	blog := r.Group("/blog")
	{
		// 公开接口 - 不需要认证
		blog.GET("/list", handler.GetBlogList)              // 文章列表
		blog.GET("/sitemap", handler.GetSitemap)            // Sitemap数据
		blog.GET("/:slug/related", handler.GetRelatedPosts) // 相关文章（必须在 /post/:slug 之前）
		blog.POST("/:slug/view", handler.IncrementViews)    // 浏览量统计
		blog.GET("/:slug", handler.GetBlogPost)             // 文章详情（放在最后，作为兜底路由）
	}
}

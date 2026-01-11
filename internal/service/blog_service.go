package service

import (
	"fmt"
	"math"

	"01agent_server/internal/models"
	"01agent_server/internal/repository"
)

type BlogService struct {
	repo *repository.BlogRepository
}

func NewBlogService() *BlogService {
	return &BlogService{
		repo: repository.NewBlogRepository(),
	}
}

// GetBlogList 获取博客列表
func (s *BlogService) GetBlogList(params repository.BlogListParams) (*models.BlogListResponse, error) {
	// 参数验证和默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 10
	}
	if params.Sort == "" {
		params.Sort = "latest"
	}

	// 获取列表
	posts, total, err := s.repo.GetBlogList(params)
	if err != nil {
		return nil, err
	}

	// 转换为响应结构
	items := make([]models.BlogPostResponse, len(posts))
	for i, post := range posts {
		resp := post.ToResponse(false) // 列表不包含content
		items[i] = *resp
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	return &models.BlogListResponse{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetBlogBySlug 通过slug获取博客详情
func (s *BlogService) GetBlogBySlug(slug string) (*models.BlogPostResponse, error) {
	post, err := s.repo.GetBlogBySlug(slug)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, fmt.Errorf("blog post not found")
	}

	return post.ToResponse(true), nil // 详情包含content
}

// GetRelatedPosts 获取相关文章
func (s *BlogService) GetRelatedPosts(postID string, limit int) ([]models.BlogPostResponse, error) {
	if limit < 1 || limit > 10 {
		limit = 3
	}

	posts, err := s.repo.GetRelatedPosts(postID, limit)
	if err != nil {
		return nil, err
	}

	// 转换为响应结构
	items := make([]models.BlogPostResponse, len(posts))
	for i, post := range posts {
		resp := post.ToResponse(false)
		items[i] = *resp
	}

	return items, nil
}

// IncrementViews 增加浏览量
func (s *BlogService) IncrementViews(postID string) error {
	return s.repo.IncrementViews(postID)
}

// GetSitemapData 获取sitemap数据
func (s *BlogService) GetSitemapData() ([]map[string]interface{}, error) {
	return s.repo.GetSitemapData()
}

// CreateBlogPost 创建博客文章
func (s *BlogService) CreateBlogPost(post *models.BlogPost) error {
	return s.repo.CreateBlogPost(post)
}

// UpdateBlogPost 更新博客文章
func (s *BlogService) UpdateBlogPost(post *models.BlogPost) error {
	return s.repo.UpdateBlogPost(post)
}

// DeleteBlogPost 删除博客文章
func (s *BlogService) DeleteBlogPost(id string) error {
	return s.repo.DeleteBlogPost(id)
}


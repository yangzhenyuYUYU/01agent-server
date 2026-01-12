package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// IncrementLikes 增加点赞数
func (s *BlogService) IncrementLikes(postID string) error {
	return s.repo.IncrementLikes(postID)
}

// DecrementLikes 减少点赞数
func (s *BlogService) DecrementLikes(postID string) error {
	return s.repo.DecrementLikes(postID)
}

// GetSitemapData 获取sitemap数据
func (s *BlogService) GetSitemapData() ([]map[string]interface{}, error) {
	return s.repo.GetSitemapData()
}

// CreateBlogPost 创建博客文章
func (s *BlogService) CreateBlogPost(req *models.BlogCreateRequest) (*models.BlogPostResponse, error) {
	// 参数验证
	if req.Slug == "" || req.Title == "" || req.Summary == "" || req.Content == "" {
		return nil, fmt.Errorf("slug, title, summary and content are required")
	}

	// 验证分类
	if _, ok := models.CategoryNames[req.Category]; !ok {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}

	// 验证状态
	if req.Status == "" {
		req.Status = models.BlogStatusPublished
	}
	if req.Status != models.BlogStatusDraft &&
		req.Status != models.BlogStatusPublished &&
		req.Status != models.BlogStatusArchived {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// 设置默认值
	if req.Author == "" {
		req.Author = "01Agent Team"
	}

	// 创建文章对象
	post := &models.BlogPost{
		ID:             uuid.New().String(),
		Slug:           req.Slug,
		Title:          req.Title,
		Summary:        req.Summary,
		Content:        req.Content,
		Category:       req.Category,
		CoverImage:     req.CoverImage,
		Author:         req.Author,
		AuthorAvatar:   req.AuthorAvatar,
		PublishDate:    time.Now(),
		ReadTime:       req.ReadTime,
		IsFeatured:     req.IsFeatured,
		SEODescription: req.SEODescription,
		Status:         req.Status,
		ThemeName:      req.ThemeName,
	}

	// 创建文章（包含标签和SEO关键词）
	err := s.repo.CreateBlogPost(post, req.Tags, req.SEOKeywords)
	if err != nil {
		// 检查是否是 slug 重复错误
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "slug") {
			return nil, fmt.Errorf("slug '%s' 已存在，请使用其他slug", req.Slug)
		}
		return nil, fmt.Errorf("create blog post failed: %w", err)
	}

	// 重新获取完整数据（包含关联的标签）
	createdPost, err := s.repo.GetBlogByID(post.ID)
	if err != nil {
		return nil, fmt.Errorf("get created post failed: %w", err)
	}

	// 加载SEO关键词
	createdPost.SEOKeywords = req.SEOKeywords

	return createdPost.ToResponse(true), nil
}

// UpdateBlogPost 更新博客文章
func (s *BlogService) UpdateBlogPost(postID string, req *models.BlogUpdateRequest) (*models.BlogPostResponse, error) {
	// 检查文章是否存在
	existingPost, err := s.repo.GetBlogByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}
	if existingPost == nil {
		return nil, fmt.Errorf("post not found")
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Category != nil {
		// 验证分类
		if _, ok := models.CategoryNames[*req.Category]; !ok {
			return nil, fmt.Errorf("invalid category: %s", *req.Category)
		}
		updates["category"] = *req.Category
	}
	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}
	if req.Author != nil {
		updates["author"] = *req.Author
	}
	if req.AuthorAvatar != nil {
		updates["author_avatar"] = *req.AuthorAvatar
	}
	if req.ReadTime != nil {
		updates["read_time"] = *req.ReadTime
	}
	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}
	if req.SEODescription != nil {
		updates["seo_description"] = *req.SEODescription
	}
	if req.ThemeName != nil {
		updates["theme_name"] = *req.ThemeName
	}
	if req.Status != nil {
		// 验证状态
		if *req.Status != models.BlogStatusDraft &&
			*req.Status != models.BlogStatusPublished &&
			*req.Status != models.BlogStatusArchived {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		updates["status"] = *req.Status
	}

	// 设置更新时间
	updates["updated_date"] = time.Now()

	// 执行更新
	err = s.repo.UpdateBlogPost(postID, updates, req.Tags, req.SEOKeywords)
	if err != nil {
		return nil, fmt.Errorf("update blog post failed: %w", err)
	}

	// 重新获取更新后的数据
	updatedPost, err := s.repo.GetBlogByID(postID)
	if err != nil {
		return nil, fmt.Errorf("get updated post failed: %w", err)
	}

	// 加载SEO关键词
	if req.SEOKeywords != nil {
		updatedPost.SEOKeywords = req.SEOKeywords
	}

	return updatedPost.ToResponse(true), nil
}

// GetBlogByID 通过ID获取博客文章
func (s *BlogService) GetBlogByID(id string) (*models.BlogPostResponse, error) {
	post, err := s.repo.GetBlogByID(id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, fmt.Errorf("blog post not found")
	}

	return post.ToResponse(true), nil
}

// DeleteBlogPost 删除博客文章
func (s *BlogService) DeleteBlogPost(id string) error {
	return s.repo.DeleteBlogPost(id)
}

// GetBlogStats 获取博客统计信息
func (s *BlogService) GetBlogStats() (map[string]interface{}, error) {
	db := s.repo.GetDB()

	stats := make(map[string]interface{})

	// 总文章数
	var total int64
	db.Model(&models.BlogPost{}).Count(&total)
	stats["total"] = total

	// 已发布文章数
	var published int64
	db.Model(&models.BlogPost{}).Where("status = ?", models.BlogStatusPublished).Count(&published)
	stats["published"] = published

	// 草稿数
	var draft int64
	db.Model(&models.BlogPost{}).Where("status = ?", models.BlogStatusDraft).Count(&draft)
	stats["draft"] = draft

	// 已归档文章数
	var archived int64
	db.Model(&models.BlogPost{}).Where("status = ?", models.BlogStatusArchived).Count(&archived)
	stats["archived"] = archived

	// 精选文章数
	var featured int64
	db.Model(&models.BlogPost{}).Where("is_featured = ?", true).Count(&featured)
	stats["featured"] = featured

	// 总浏览量
	var totalViews int64
	db.Model(&models.BlogPost{}).Select("COALESCE(SUM(views), 0)").Scan(&totalViews)
	stats["total_views"] = totalViews

	// 总点赞数
	var totalLikes int64
	db.Model(&models.BlogPost{}).Select("COALESCE(SUM(likes), 0)").Scan(&totalLikes)
	stats["total_likes"] = totalLikes

	// 分类统计
	var categoryStats []map[string]interface{}
	db.Model(&models.BlogPost{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryStats)
	stats["by_category"] = categoryStats

	// 标签总数
	var totalTags int64
	db.Model(&models.BlogTag{}).Count(&totalTags)
	stats["total_tags"] = totalTags

	return stats, nil
}

// GetThemePreview 获取主题预览配置
func (s *BlogService) GetThemePreview(themeName string) (map[string]interface{}, error) {
	if themeName == "" {
		return nil, fmt.Errorf("theme_name is required")
	}

	// 查询 public_templates 表，按 name 字段匹配
	var template models.PublicTemplate
	db := repository.GetDB()

	err := db.Where("name = ? AND status = ? AND is_public = ?",
		themeName,
		models.TemplateStatusPublished,
		true).
		First(&template).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("theme '%s' not found", themeName)
		}
		return nil, fmt.Errorf("query theme failed: %w", err)
	}

	// 解析 template_data JSON
	var templateData map[string]interface{}
	if template.TemplateData != nil && *template.TemplateData != "" {
		if err := json.Unmarshal([]byte(*template.TemplateData), &templateData); err != nil {
			return nil, fmt.Errorf("parse template_data failed: %w", err)
		}
	} else {
		templateData = make(map[string]interface{})
	}

	// 构建响应
	response := map[string]interface{}{
		"theme_id":      template.TemplateID,
		"theme_name":    template.Name,
		"theme_name_en": template.NameEn,
		"description":   template.Description,
		"author":        template.Author,
		"template_type": template.TemplateType,
		"primary_color": template.PrimaryColor,
		"tags":          template.Tags,
		"config":        templateData,
	}

	return response, nil
}

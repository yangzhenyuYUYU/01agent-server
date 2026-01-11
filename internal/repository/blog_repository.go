package repository

import (
	"fmt"
	"strings"

	"01agent_server/internal/models"

	"gorm.io/gorm"
)

type BlogRepository struct {
	db *gorm.DB
}

func NewBlogRepository() *BlogRepository {
	return &BlogRepository{
		db: GetDB(),
	}
}

// GetDB 获取数据库连接（用于service层的统计查询）
func (r *BlogRepository) GetDB() *gorm.DB {
	return r.db
}

// BlogListParams 博客列表查询参数
type BlogListParams struct {
	Page       int
	PageSize   int
	Category   string
	Tag        string
	Keyword    string
	IsFeatured *bool
	Sort       string // latest, popular, views
	Status     string // 状态筛选（管理后台使用）
}

// GetBlogList 获取博客列表
func (r *BlogRepository) GetBlogList(params BlogListParams) ([]models.BlogPost, int64, error) {
	var posts []models.BlogPost
	var total int64

	offset := (params.Page - 1) * params.PageSize

	// 构建查询
	query := r.db.Model(&models.BlogPost{})

	// 基础条件：状态筛选
	if params.Status != "" {
		// 管理后台可以指定状态
		query = query.Where("status = ?", params.Status)
	} else {
		// 前台只查询已发布的文章
		query = query.Where("status = ?", models.BlogStatusPublished)
	}

	// 分类筛选
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}

	// 精选筛选
	if params.IsFeatured != nil {
		query = query.Where("is_featured = ?", *params.IsFeatured)
	}

	// 关键词搜索
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", keyword, keyword)
	}

	// 标签筛选
	if params.Tag != "" {
		query = query.Joins("INNER JOIN blog_post_tags ON blog_posts.id = blog_post_tags.post_id").
			Joins("INNER JOIN blog_tags ON blog_post_tags.tag_id = blog_tags.id").
			Where("blog_tags.name = ?", params.Tag)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count blog posts failed: %w", err)
	}

	// 排序
	orderBy := "publish_date DESC"
	switch params.Sort {
	case "popular":
		orderBy = "views DESC, likes DESC"
	case "views":
		orderBy = "views DESC"
	default:
		orderBy = "publish_date DESC"
	}

	// 查询列表
	err := query.Order(orderBy).
		Limit(params.PageSize).
		Offset(offset).
		Preload("Tags").
		Find(&posts).Error

	if err != nil {
		return nil, 0, fmt.Errorf("get blog posts failed: %w", err)
	}

	return posts, total, nil
}

// GetBlogBySlug 通过slug获取博客文章
func (r *BlogRepository) GetBlogBySlug(slug string) (*models.BlogPost, error) {
	var post models.BlogPost

	err := r.db.Where("slug = ? AND status = ?", slug, models.BlogStatusPublished).
		Preload("Tags").
		First(&post).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get blog post by slug failed: %w", err)
	}

	// 加载SEO关键词
	var keywords []models.BlogSEOKeyword
	if err := r.db.Where("post_id = ?", post.ID).Find(&keywords).Error; err == nil {
		post.SEOKeywords = make([]string, len(keywords))
		for i, kw := range keywords {
			post.SEOKeywords[i] = kw.Keyword
		}
	}

	return &post, nil
}

// GetBlogByID 通过ID获取博客文章
func (r *BlogRepository) GetBlogByID(id string) (*models.BlogPost, error) {
	var post models.BlogPost

	err := r.db.Where("id = ?", id).
		Preload("Tags").
		First(&post).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get blog post by id failed: %w", err)
	}

	return &post, nil
}

// GetRelatedPosts 获取相关文章
func (r *BlogRepository) GetRelatedPosts(postID string, limit int) ([]models.BlogPost, error) {
	// 获取当前文章的分类
	var currentPost models.BlogPost
	if err := r.db.Select("category").Where("id = ?", postID).First(&currentPost).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get current post failed: %w", err)
	}

	// 查询同分类的其他文章
	var posts []models.BlogPost
	err := r.db.Where("category = ? AND id != ? AND status = ?",
		currentPost.Category, postID, models.BlogStatusPublished).
		Order("views DESC, publish_date DESC").
		Limit(limit).
		Find(&posts).Error

	if err != nil {
		return nil, fmt.Errorf("get related posts failed: %w", err)
	}

	return posts, nil
}

// IncrementViews 增加浏览量
func (r *BlogRepository) IncrementViews(postID string) error {
	return r.db.Model(&models.BlogPost{}).
		Where("id = ?", postID).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

// GetSitemapData 获取sitemap数据
func (r *BlogRepository) GetSitemapData() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	rows, err := r.db.Model(&models.BlogPost{}).
		Select("slug, category, COALESCE(updated_date, publish_date) as updated_date").
		Where("status = ?", models.BlogStatusPublished).
		Order("publish_date DESC").
		Rows()

	if err != nil {
		return nil, fmt.Errorf("get sitemap data failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slug, category string
		var updatedDate interface{}
		if err := rows.Scan(&slug, &category, &updatedDate); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"slug":         slug,
			"category":     category,
			"updated_date": updatedDate,
		})
	}

	return results, nil
}

// CreateBlogPost 创建博客文章
func (r *BlogRepository) CreateBlogPost(post *models.BlogPost, tags []string, seoKeywords []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建文章
		if err := tx.Create(post).Error; err != nil {
			return fmt.Errorf("create blog post failed: %w", err)
		}

		// 2. 处理标签
		if len(tags) > 0 {
			for _, tagName := range tags {
				tag, err := r.GetOrCreateTag(tagName)
				if err != nil {
					return fmt.Errorf("get or create tag failed: %w", err)
				}

				// 关联标签
				if err := tx.Create(&models.BlogPostTag{
					PostID: post.ID,
					TagID:  tag.ID,
				}).Error; err != nil {
					return fmt.Errorf("create post tag relation failed: %w", err)
				}
			}
		}

		// 3. 处理SEO关键词
		if len(seoKeywords) > 0 {
			for _, keyword := range seoKeywords {
				if err := tx.Create(&models.BlogSEOKeyword{
					PostID:  post.ID,
					Keyword: keyword,
				}).Error; err != nil {
					return fmt.Errorf("create seo keyword failed: %w", err)
				}
			}
		}

		return nil
	})
}

// UpdateBlogPost 更新博客文章
func (r *BlogRepository) UpdateBlogPost(postID string, updates map[string]interface{}, tags []string, seoKeywords []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新文章基本信息
		if len(updates) > 0 {
			if err := tx.Model(&models.BlogPost{}).Where("id = ?", postID).Updates(updates).Error; err != nil {
				return fmt.Errorf("update blog post failed: %w", err)
			}
		}

		// 2. 更新标签（如果提供了新标签）
		if tags != nil {
			// 删除旧的标签关联
			if err := tx.Where("post_id = ?", postID).Delete(&models.BlogPostTag{}).Error; err != nil {
				return fmt.Errorf("delete old post tags failed: %w", err)
			}

			// 添加新的标签关联
			for _, tagName := range tags {
				tag, err := r.GetOrCreateTag(tagName)
				if err != nil {
					return fmt.Errorf("get or create tag failed: %w", err)
				}

				if err := tx.Create(&models.BlogPostTag{
					PostID: postID,
					TagID:  tag.ID,
				}).Error; err != nil {
					return fmt.Errorf("create post tag relation failed: %w", err)
				}
			}
		}

		// 3. 更新SEO关键词（如果提供了新关键词）
		if seoKeywords != nil {
			// 删除旧的关键词
			if err := tx.Where("post_id = ?", postID).Delete(&models.BlogSEOKeyword{}).Error; err != nil {
				return fmt.Errorf("delete old seo keywords failed: %w", err)
			}

			// 添加新的关键词
			for _, keyword := range seoKeywords {
				if err := tx.Create(&models.BlogSEOKeyword{
					PostID:  postID,
					Keyword: keyword,
				}).Error; err != nil {
					return fmt.Errorf("create seo keyword failed: %w", err)
				}
			}
		}

		return nil
	})
}

// DeleteBlogPost 删除博客文章
func (r *BlogRepository) DeleteBlogPost(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除文章
		if err := tx.Where("id = ?", id).Delete(&models.BlogPost{}).Error; err != nil {
			return err
		}

		// 删除SEO关键词
		if err := tx.Where("post_id = ?", id).Delete(&models.BlogSEOKeyword{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetOrCreateTag 获取或创建标签
func (r *BlogRepository) GetOrCreateTag(tagName string) (*models.BlogTag, error) {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return nil, fmt.Errorf("tag name cannot be empty")
	}

	var tag models.BlogTag
	err := r.db.Where("name = ?", tagName).First(&tag).Error
	if err == nil {
		return &tag, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新标签
	tag = models.BlogTag{Name: tagName}
	if err := r.db.Create(&tag).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}


package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"01agent_server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
}

// TestBlogListHandler 测试博客列表接口
func TestBlogListHandler(t *testing.T) {
	// 创建测试路由
	r := gin.New()
	handler := NewBlogHandler()
	r.GET("/blog/list", handler.GetBlogList)

	// 测试用例
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "获取默认列表",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(0), resp["code"])
				assert.NotNil(t, resp["data"])
			},
		},
		{
			name:           "按分类筛选",
			queryParams:    "?category=tutorials",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(0), resp["code"])
			},
		},
		{
			name:           "分页参数",
			queryParams:    "?page=1&page_size=5",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(0), resp["code"])
			},
		},
		{
			name:           "关键词搜索",
			queryParams:    "?keyword=快速入门",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(0), resp["code"])
			},
		},
		{
			name:           "精选文章",
			queryParams:    "?is_featured=true",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(0), resp["code"])
			},
		},
		{
			name:           "热门排序",
			queryParams:    "?sort=popular",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(0), resp["code"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建请求
			req, _ := http.NewRequest("GET", "/blog/list"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 检查状态码
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 解析响应
			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			// 执行自定义检查
			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestBlogPostHandler 测试博客详情接口
func TestBlogPostHandler(t *testing.T) {
	r := gin.New()
	handler := NewBlogHandler()
	r.GET("/blog/post/:slug", handler.GetBlogPost)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "获取存在的文章",
			slug:           "getting-started-with-01agent",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				// 根据实际数据库情况调整
				data := resp["data"]
				if data != nil {
					post := data.(map[string]interface{})
					assert.NotEmpty(t, post["id"])
					assert.NotEmpty(t, post["title"])
					assert.NotEmpty(t, post["content"])
				}
			},
		},
		{
			name:           "获取不存在的文章",
			slug:           "non-existent-slug",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.NotEqual(t, float64(0), resp["code"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/blog/post/"+tt.slug, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestBlogSitemapHandler 测试sitemap接口
func TestBlogSitemapHandler(t *testing.T) {
	r := gin.New()
	handler := NewBlogHandler()
	r.GET("/blog/sitemap", handler.GetSitemap)

	req, _ := http.NewRequest("GET", "/blog/sitemap", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

// TestBlogRelatedPostsHandler 测试相关文章接口
func TestBlogRelatedPostsHandler(t *testing.T) {
	r := gin.New()
	handler := NewBlogHandler()
	r.GET("/blog/post/:postId/related", handler.GetRelatedPosts)

	tests := []struct {
		name           string
		postId         string
		limit          string
		expectedStatus int
	}{
		{
			name:           "默认限制",
			postId:         "test-blog-001",
			limit:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "自定义限制",
			postId:         "test-blog-001",
			limit:          "?limit=5",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/blog/post/"+tt.postId+"/related"+tt.limit, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
		})
	}
}

// TestBlogIncrementViewsHandler 测试浏览量统计接口
func TestBlogIncrementViewsHandler(t *testing.T) {
	r := gin.New()
	handler := NewBlogHandler()
	r.POST("/blog/post/:postId/view", handler.IncrementViews)

	req, _ := http.NewRequest("POST", "/blog/post/test-blog-001/view", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// 浏览量统计总是返回成功
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

// TestBlogPostToResponse 测试模型转换
func TestBlogPostToResponse(t *testing.T) {
	now := time.Now()
	content := "# Test Content"
	coverImage := "https://example.com/image.jpg"
	readTime := 5

	post := &models.BlogPost{
		ID:          "test-001",
		Slug:        "test-slug",
		Title:       "Test Title",
		Summary:     "Test Summary",
		Content:     content,
		Category:    "tutorials",
		CoverImage:  &coverImage,
		Author:      "Test Author",
		PublishDate: now,
		ReadTime:    &readTime,
		Views:       100,
		Likes:       10,
		IsFeatured:  true,
		Status:      models.BlogStatusPublished,
		Tags: []models.BlogTag{
			{ID: 1, Name: "测试标签1"},
			{ID: 2, Name: "测试标签2"},
		},
		SEOKeywords: []string{"关键词1", "关键词2"},
	}

	// 测试不包含content
	resp := post.ToResponse(false)
	assert.Equal(t, "test-001", resp.ID)
	assert.Equal(t, "test-slug", resp.Slug)
	assert.Equal(t, "Test Title", resp.Title)
	assert.Equal(t, "使用教程", resp.CategoryName)
	assert.Nil(t, resp.Content)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Equal(t, "测试标签1", resp.Tags[0])

	// 测试包含content
	resp = post.ToResponse(true)
	assert.NotNil(t, resp.Content)
	assert.Equal(t, content, *resp.Content)
	assert.Equal(t, 2, len(resp.SEOKeywords))
}


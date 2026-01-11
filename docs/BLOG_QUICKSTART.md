# 博客功能快速开始指南

## 功能概述

博客功能提供了完整的文章管理和展示系统，包括：

- ✅ 文章列表（支持分页、筛选、搜索、排序）
- ✅ 文章详情（Markdown内容）
- ✅ 分类管理（5个预设分类）
- ✅ 标签系统
- ✅ 相关文章推荐
- ✅ 浏览量统计
- ✅ 精选文章
- ✅ SEO优化（关键词、描述）
- ✅ Sitemap支持

## 快速开始

### 1. 数据库初始化

**方法1：使用一键初始化脚本（推荐）**

```bash
mysql -u root -p 01agent < configs/blog-init.sql
```

这个脚本会自动完成：
- ✅ 创建所有表结构
- ✅ 插入测试数据（3篇文章 + 标签）
- ✅ 显示验证信息

**方法2：分步执行**

如果你只想创建表结构而不导入测试数据：

```bash
# 1. 创建表结构
mysql -u root -p 01agent < configs/blog-schema.sql

# 2. （可选）导入测试数据
mysql -u root -p 01agent < configs/test-data.sql
```

**注意**：
- `blog-init.sql` = 表结构 + 精简测试数据（推荐）
- `blog-schema.sql` = 仅表结构
- `test-data.sql` = 更多测试数据（5篇文章）

### 2. 启动服务器

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

### 3. 测试接口

#### 方法1: 使用测试脚本

**Linux/Mac:**
```bash
chmod +x scripts/test_blog_api.sh
./scripts/test_blog_api.sh
```

**Windows:**
```bash
scripts\test_blog_api.bat
```

#### 方法2: 手动测试

```bash
# 获取博客列表
curl "http://localhost:8080/blog/list"

# 获取文章详情
curl "http://localhost:8080/blog/getting-started-with-01agent"

# 获取精选文章
curl "http://localhost:8080/blog/list?is_featured=true"

# 搜索文章
curl "http://localhost:8080/blog/list?keyword=快速"

# 获取相关文章
curl "http://localhost:8080/blog/getting-started-with-01agent/related"

# 增加浏览量
curl -X POST "http://localhost:8080/blog/getting-started-with-01agent/view"
```

## 文件结构

```
internal/
├── models/
│   └── blog.go                 # 数据模型定义
├── repository/
│   └── blog_repository.go      # 数据库操作层
├── service/
│   └── blog_service.go         # 业务逻辑层
└── router/
    ├── blog.go                 # 路由处理器
    ├── blog_test.go            # 单元测试
    └── BLOG_API.md             # 详细API文档

configs/
└── blog-schema.sql             # 数据库表结构

scripts/
├── test_blog_api.sh            # Linux/Mac测试脚本
└── test_blog_api.bat           # Windows测试脚本
```

## API 接口一览

### 公开接口（无需认证）

| 接口 | 方法 | 说明 |
|------|------|------|
| `/blog/list` | GET | 获取文章列表 |
| `/blog/:slug` | GET | 获取文章详情 |
| `/blog/sitemap` | GET | 获取sitemap数据 |
| `/blog/:slug/related` | GET | 获取相关文章 |
| `/blog/:slug/view` | POST | 增加浏览量 |

### 管理接口（需要认证）

| 接口 | 方法 | 说明 |
|------|------|------|
| `/blog/create` | POST | 创建文章 |
| `/blog/:id` | PUT | 更新文章 |
| `/blog/:id` | DELETE | 删除文章 |

详细的接口文档：
- 公开接口文档：[`internal/router/BLOG_API.md`](../internal/router/BLOG_API.md)
- 管理接口文档：[`internal/router/BLOG_ADMIN_API.md`](../internal/router/BLOG_ADMIN_API.md)

## 数据模型

### 分类 (Category)

系统预设了5个分类：

- `product-updates`: 产品动态
- `tutorials`: 使用教程
- `tips-and-tricks`: 运营技巧
- `industry-insights`: 行业洞察
- `case-studies`: 案例故事

### 文章状态 (Status)

- `draft`: 草稿
- `published`: 已发布
- `archived`: 已归档

### 标签 (Tags)

标签是动态的，可以为每篇文章添加多个标签。

## 前端集成示例

### React示例

```javascript
import React, { useEffect, useState } from 'react';
import axios from 'axios';

function BlogList() {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    axios.get('http://localhost:8080/blog/list', {
      params: {
        page: 1,
        page_size: 10,
        sort: 'latest'
      }
    })
    .then(response => {
      if (response.data.code === 0) {
        setPosts(response.data.data.items);
      }
      setLoading(false);
    })
    .catch(error => {
      console.error('获取博客列表失败:', error);
      setLoading(false);
    });
  }, []);

  if (loading) return <div>加载中...</div>;

  return (
    <div className="blog-list">
      {posts.map(post => (
        <article key={post.id}>
          <h2>{post.title}</h2>
          <p>{post.summary}</p>
          <div className="meta">
            <span>{post.category_name}</span>
            <span>{post.read_time}分钟阅读</span>
            <span>{post.views}次浏览</span>
          </div>
          <div className="tags">
            {post.tags.map(tag => (
              <span key={tag} className="tag">{tag}</span>
            ))}
          </div>
        </article>
      ))}
    </div>
  );
}

export default BlogList;
```

### Vue示例

```vue
<template>
  <div class="blog-list">
    <div v-if="loading">加载中...</div>
    <article v-for="post in posts" :key="post.id" class="post-card">
      <img v-if="post.cover_image" :src="post.cover_image" :alt="post.title">
      <h2>{{ post.title }}</h2>
      <p>{{ post.summary }}</p>
      <div class="meta">
        <span>{{ post.category_name }}</span>
        <span>{{ post.read_time }}分钟</span>
        <span>{{ post.views }}次浏览</span>
      </div>
      <div class="tags">
        <span v-for="tag in post.tags" :key="tag" class="tag">
          {{ tag }}
        </span>
      </div>
    </article>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'BlogList',
  data() {
    return {
      posts: [],
      loading: true
    };
  },
  mounted() {
    this.fetchPosts();
  },
  methods: {
    async fetchPosts() {
      try {
        const response = await axios.get('http://localhost:8080/blog/list', {
          params: {
            page: 1,
            page_size: 10
          }
        });
        
        if (response.data.code === 0) {
          this.posts = response.data.data.items;
        }
      } catch (error) {
        console.error('获取博客列表失败:', error);
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>
```

## 测试

运行单元测试：

```bash
go test ./internal/router/... -v
```

## 常见问题

### Q: 如何添加新的分类？

A: 在 `internal/models/blog.go` 的 `CategoryNames` 中添加新的分类映射：

```go
var CategoryNames = map[string]string{
    "product-updates":  "产品动态",
    "tutorials":        "使用教程",
    "your-new-category": "你的新分类", // 添加这里
}
```

### Q: 如何修改默认的分页大小？

A: 在 `internal/service/blog_service.go` 的 `GetBlogList` 方法中修改：

```go
if params.PageSize < 1 || params.PageSize > 100 {
    params.PageSize = 10  // 修改这里的默认值
}
```

### Q: 如何添加缓存？

A: 推荐在 `service` 层添加 Redis 缓存：

```go
// 伪代码示例
func (s *BlogService) GetBlogList(params repository.BlogListParams) (*models.BlogListResponse, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("blog:list:%v", params)
    if cached, err := redis.Get(cacheKey); err == nil {
        return cached, nil
    }
    
    // 2. 从数据库获取
    result, err := s.repo.GetBlogList(params)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    redis.Set(cacheKey, result, 5*time.Minute)
    
    return result, nil
}
```

### Q: 如何支持多语言？

A: 可以在表结构中添加 `language` 字段，或创建多语言内容表。

## 性能优化建议

1. **添加索引**: 已在 SQL 中添加了必要的索引
2. **使用缓存**: 对热门文章和列表页添加 Redis 缓存
3. **CDN**: 将图片资源托管到 CDN
4. **分页优化**: 对于大数据量，使用游标分页
5. **数据库连接池**: 已配置，可根据需要调整
6. **全文搜索**: 对于复杂搜索需求，集成 Elasticsearch

## 下一步

- 查看 [`internal/router/BLOG_API.md`](./internal/router/BLOG_API.md) 了解详细的API文档
- 运行测试脚本验证功能
- 根据需求扩展功能（评论、点赞等）

## 技术支持

如有问题，请查看：
- API文档: `internal/router/BLOG_API.md`
- 测试脚本: `scripts/test_blog_api.sh` 或 `scripts/test_blog_api.bat`
- 单元测试: `internal/router/blog_test.go`


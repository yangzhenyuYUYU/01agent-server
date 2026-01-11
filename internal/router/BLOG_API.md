# 博客功能 API 文档

## 概述

博客功能提供了完整的博客文章管理和展示接口，支持文章列表、详情查看、分类筛选、标签管理、相关推荐等功能。

## 数据库初始化

执行 SQL 文件创建表结构：

```bash
mysql -u root -p 01agent < configs/blog-schema.sql
```

## API 接口说明

基础 URL: `http://localhost:8080`

所有接口返回格式：

```json
{
  "code": 0,           // 0表示成功，非0表示失败
  "msg": "success",    // 消息描述
  "data": {}          // 返回数据
}
```

---

## 1. 获取博客文章列表

**接口地址**: `GET /blog/list`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| page | int | 否 | 页码，默认1 | 1 |
| page_size | int | 否 | 每页数量，默认10，最大100 | 10 |
| category | string | 否 | 分类筛选 | tutorials |
| tag | string | 否 | 标签筛选 | AI写作 |
| keyword | string | 否 | 关键词搜索（标题、摘要） | 快速入门 |
| is_featured | boolean | 否 | 是否精选 | true |
| sort | string | 否 | 排序方式: latest(最新), popular(热门), views(浏览量) | latest |

**分类说明**:
- `product-updates`: 产品动态
- `tutorials`: 使用教程
- `tips-and-tricks`: 运营技巧
- `industry-insights`: 行业洞察
- `case-studies`: 案例故事

**请求示例**:

```bash
# 获取第1页，每页10条
curl "http://localhost:8080/blog/list?page=1&page_size=10"

# 按分类筛选
curl "http://localhost:8080/blog/list?category=tutorials"

# 搜索关键词
curl "http://localhost:8080/blog/list?keyword=AI"

# 获取精选文章
curl "http://localhost:8080/blog/list?is_featured=true"

# 按热门排序
curl "http://localhost:8080/blog/list?sort=popular"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "test-blog-001",
        "slug": "getting-started-with-01agent",
        "title": "01Agent 快速入门指南",
        "summary": "这是一篇介绍如何快速开始使用 01Agent 的教程文章...",
        "category": "tutorials",
        "category_name": "使用教程",
        "cover_image": "https://example.com/covers/getting-started.jpg",
        "author": "01Agent Team",
        "author_avatar": "https://example.com/avatars/team.jpg",
        "publish_date": "2026-01-11T10:00:00Z",
        "updated_date": null,
        "read_time": 5,
        "views": 1250,
        "likes": 89,
        "is_featured": true,
        "tags": ["教程", "入门指南"]
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

---

## 2. 获取单篇文章详情

**接口地址**: `GET /blog/:slug`

**路径参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| slug | string | 是 | 文章的 URL 标识符 |

**请求示例**:

```bash
curl "http://localhost:8080/blog/getting-started-with-01agent"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "test-blog-001",
    "slug": "getting-started-with-01agent",
    "title": "01Agent 快速入门指南",
    "summary": "这是一篇介绍如何快速开始使用 01Agent 的教程文章...",
    "content": "# 01Agent 快速入门\n\n## 简介\n\n01Agent 是一个强大的 AI 代理平台...",
    "category": "tutorials",
    "category_name": "使用教程",
    "cover_image": "https://example.com/covers/getting-started.jpg",
    "author": "01Agent Team",
    "author_avatar": "https://example.com/avatars/team.jpg",
    "publish_date": "2026-01-11T10:00:00Z",
    "updated_date": null,
    "read_time": 5,
    "views": 1250,
    "likes": 89,
    "is_featured": true,
    "tags": ["教程", "入门指南"],
    "seo_keywords": ["01Agent", "快速入门", "AI代理"],
    "seo_description": "学习如何快速开始使用 01Agent，5分钟内完成基础配置和首次运行"
  }
}
```

---

## 3. 获取 Sitemap 数据

**接口地址**: `GET /blog/sitemap`

**说明**: 获取所有已发布文章的 URL 信息，用于生成网站地图。

**请求示例**:

```bash
curl "http://localhost:8080/blog/sitemap"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "slug": "getting-started-with-01agent",
      "category": "tutorials",
      "updated_date": "2026-01-11T10:00:00Z"
    },
    {
      "slug": "10-tips-for-ai-content-creation",
      "category": "tips-and-tricks",
      "updated_date": "2026-01-08T10:00:00Z"
    }
  ]
}
```

---

## 4. 获取相关文章推荐

**接口地址**: `GET /blog/:slug/related`

**路径参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| slug | string | 是 | 文章的 URL 标识符 |

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| limit | int | 否 | 返回数量 | 3 |

**请求示例**:

```bash
curl "http://localhost:8080/blog/getting-started-with-01agent/related?limit=3"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "id": "test-blog-002",
      "slug": "advanced-tutorials",
      "title": "进阶教程",
      "summary": "更深入的使用技巧...",
      "category": "tutorials",
      "category_name": "使用教程",
      "cover_image": "https://example.com/covers/advanced.jpg",
      "publish_date": "2026-01-10T10:00:00Z",
      "read_time": 10
    }
  ]
}
```

---

## 5. 增加文章浏览量

**接口地址**: `POST /blog/:slug/view`

**说明**: 记录文章浏览量，通常在用户打开文章详情页时调用。

**路径参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| slug | string | 是 | 文章的 URL 标识符 |

**请求示例**:

```bash
curl -X POST "http://localhost:8080/blog/getting-started-with-01agent/view"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

**注意**: 即使浏览量统计失败，接口也会返回成功，不影响用户体验。

---

## 使用场景示例

### 1. 博客首页展示

```javascript
// 获取精选文章（最多6篇）
fetch('/blog/list?is_featured=true&page_size=6')
  .then(res => res.json())
  .then(data => {
    // 渲染精选文章
  });

// 获取最新文章
fetch('/blog/list?sort=latest&page_size=10')
  .then(res => res.json())
  .then(data => {
    // 渲染最新文章列表
  });
```

### 2. 分类页面

```javascript
// 获取特定分类的文章
fetch('/blog/list?category=tutorials&page=1&page_size=12')
  .then(res => res.json())
  .then(data => {
    // 渲染分类文章列表
  });
```

### 3. 文章详情页

```javascript
// 获取文章详情
fetch('/blog/getting-started-with-01agent')
  .then(res => res.json())
  .then(data => {
    const post = data.data;
    // 渲染文章内容
    renderMarkdown(post.content);
    
    // 记录浏览量
    fetch(`/blog/${post.slug}/view`, { method: 'POST' });
    
    // 获取相关推荐
    fetch(`/blog/${post.slug}/related?limit=3`)
      .then(res => res.json())
      .then(data => {
        // 渲染相关文章
      });
  });
```

### 4. 搜索功能

```javascript
// 搜索文章
const keyword = '快速入门';
fetch(`/blog/list?keyword=${encodeURIComponent(keyword)}`)
  .then(res => res.json())
  .then(data => {
    // 渲染搜索结果
  });
```

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 404 | 文章不存在 |
| 500 | 服务器内部错误 |

---

## 前端集成建议

### 1. Markdown 渲染

推荐使用以下库渲染 Markdown 内容：
- `marked` - JavaScript Markdown 解析器
- `highlight.js` - 代码高亮
- `DOMPurify` - XSS 防护

```javascript
import marked from 'marked';
import hljs from 'highlight.js';
import DOMPurify from 'dompurify';

// 配置 marked
marked.setOptions({
  highlight: (code, lang) => {
    return hljs.highlight(code, { language: lang }).value;
  }
});

// 渲染 Markdown
const html = DOMPurify.sanitize(marked(post.content));
```

### 2. SEO 优化

```javascript
// 设置页面 meta 标签
const post = data.data;

document.title = post.title;
document.querySelector('meta[name="description"]').content = post.seo_description;
document.querySelector('meta[name="keywords"]').content = post.seo_keywords.join(', ');

// Open Graph 标签
document.querySelector('meta[property="og:title"]').content = post.title;
document.querySelector('meta[property="og:description"]').content = post.summary;
document.querySelector('meta[property="og:image"]').content = post.cover_image;
```

### 3. 阅读时间计算

如果需要前端计算阅读时间：

```javascript
function calculateReadTime(content) {
  const wordsPerMinute = 200; // 中文约 200-300 字/分钟
  const wordCount = content.length;
  return Math.ceil(wordCount / wordsPerMinute);
}
```

---

## 测试

启动服务器：

```bash
go run main.go
```

测试接口：

```bash
# 测试列表接口
curl "http://localhost:8080/blog/list"

# 测试详情接口
curl "http://localhost:8080/blog/getting-started-with-01agent"

# 测试sitemap
curl "http://localhost:8080/blog/sitemap"

# 测试相关文章
curl "http://localhost:8080/blog/getting-started-with-01agent/related"

# 测试浏览量统计
curl -X POST "http://localhost:8080/blog/getting-started-with-01agent/view"
```

---

## 注意事项

1. **性能优化**: 对于高流量的博客，建议添加 Redis 缓存
2. **图片处理**: 封面图和作者头像建议使用 CDN
3. **内容安全**: Markdown 内容在前端渲染时需要进行 XSS 防护
4. **分页优化**: 大量文章时建议使用游标分页而非偏移分页
5. **搜索优化**: 生产环境建议使用 Elasticsearch 等专业搜索引擎

---

## 后续扩展功能

可以考虑添加以下功能：

- [ ] 文章点赞功能
- [ ] 评论系统
- [ ] 文章收藏
- [ ] 分享统计
- [ ] 文章草稿和预览
- [ ] 定时发布
- [ ] 多作者管理
- [ ] 内容审核工作流
- [ ] 全文搜索（集成 Elasticsearch）
- [ ] 阅读进度追踪
- [ ] RSS 订阅
- [ ] 文章归档功能



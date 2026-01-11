# 博客管理后台API文档

## 接口列表

所有接口前缀：`/api/v1/admin/blog`

| 功能 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 获取列表 | GET | `/list` | 支持所有状态筛选 |
| 获取详情 | GET | `/:id` | 通过ID获取 |
| 获取统计 | GET | `/stats` | 统计信息 |
| 创建文章 | POST | `/create` | 创建新文章 |
| 更新文章 | PUT | `/:id` | 更新文章 |
| 删除文章 | DELETE | `/:id` | 删除文章 |
| 批量删除 | POST | `/batch-delete` | 批量删除 |
| 更新状态 | PUT | `/:id/status` | 发布/归档 |
| 切换精选 | PUT | `/:id/featured` | 设置精选 |

---

## 1. 获取博客列表

**接口**: `GET /api/v1/admin/blog/list`

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |
| category | string | 否 | 分类筛选 |
| tag | string | 否 | 标签筛选 |
| keyword | string | 否 | 关键词搜索 |
| status | string | 否 | 状态筛选（draft/published/archived） |
| is_featured | bool | 否 | 是否精选 |
| sort | string | 否 | 排序方式（latest/popular/views） |

**示例**:

```bash
# 获取所有草稿
curl "http://localhost:8080/api/v1/admin/blog/list?status=draft"

# 获取精选文章
curl "http://localhost:8080/api/v1/admin/blog/list?is_featured=true"

# 搜索包含"AI"的文章
curl "http://localhost:8080/api/v1/admin/blog/list?keyword=AI"
```

**响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [...],
    "total": 50,
    "page": 1,
    "page_size": 10,
    "total_pages": 5
  }
}
```

---

## 2. 获取文章详情

**接口**: `GET /api/v1/admin/blog/:id`

**示例**:

```bash
curl "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001"
```

---

## 3. 获取统计信息

**接口**: `GET /api/v1/admin/blog/stats`

**示例**:

```bash
curl "http://localhost:8080/api/v1/admin/blog/stats"
```

**响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 100,
    "published": 85,
    "draft": 10,
    "archived": 5,
    "featured": 15,
    "total_views": 125680,
    "total_likes": 3420,
    "total_tags": 25,
    "by_category": [
      {"category": "tutorials", "count": 40},
      {"category": "tips-and-tricks", "count": 30},
      {"category": "industry-insights", "count": 20},
      {"category": "product-updates", "count": 10}
    ]
  }
}
```

---

## 4. 创建文章

**接口**: `POST /api/v1/admin/blog/create`

**请求体**:

```json
{
  "slug": "my-new-article",
  "title": "文章标题",
  "summary": "文章摘要",
  "content": "# 正文\n\nMarkdown内容...",
  "category": "tutorials",
  "cover_image": "https://example.com/image.jpg",
  "author": "作者名",
  "author_avatar": "https://example.com/avatar.jpg",
  "read_time": 5,
  "is_featured": false,
  "seo_description": "SEO描述",
  "status": "draft",
  "tags": ["AI", "教程"],
  "seo_keywords": ["AI写作", "教程"]
}
```

**示例**:

```bash
curl -X POST "http://localhost:8080/api/v1/admin/blog/create" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "test-article",
    "title": "测试文章",
    "summary": "这是摘要",
    "content": "# 标题\n\n内容",
    "category": "tutorials",
    "status": "draft"
  }'
```

---

## 5. 更新文章

**接口**: `PUT /api/v1/admin/blog/:id`

**请求体**（所有字段可选）:

```json
{
  "title": "新标题",
  "summary": "新摘要",
  "content": "新内容",
  "category": "tips-and-tricks",
  "is_featured": true,
  "status": "published",
  "tags": ["新标签1", "新标签2"]
}
```

**示例**:

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新后的标题",
    "is_featured": true
  }'
```

---

## 6. 删除文章

**接口**: `DELETE /api/v1/admin/blog/:id`

**示例**:

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001"
```

**响应**:

```json
{
  "code": 0,
  "msg": "删除成功",
  "data": null
}
```

---

## 7. 批量删除

**接口**: `POST /api/v1/admin/blog/batch-delete`

**请求体**:

```json
{
  "ids": [
    "550e8400-e29b-41d4-a716-446655440001",
    "550e8400-e29b-41d4-a716-446655440002",
    "550e8400-e29b-41d4-a716-446655440003"
  ]
}
```

**示例**:

```bash
curl -X POST "http://localhost:8080/api/v1/admin/blog/batch-delete" \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["id1", "id2", "id3"]
  }'
```

**响应**:

```json
{
  "code": 0,
  "msg": "批量删除完成",
  "data": {
    "success_count": 2,
    "fail_count": 1,
    "total": 3
  }
}
```

---

## 8. 更新文章状态

**接口**: `PUT /api/v1/admin/blog/:id/status`

**请求体**:

```json
{
  "status": "published"
}
```

**状态值**:
- `draft` - 草稿
- `published` - 已发布
- `archived` - 已归档

**示例**:

```bash
# 发布文章
curl -X PUT "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "published"}'

# 归档文章
curl -X PUT "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "archived"}'
```

---

## 9. 切换精选状态

**接口**: `PUT /api/v1/admin/blog/:id/featured`

**请求体**:

```json
{
  "is_featured": true
}
```

**示例**:

```bash
# 设为精选
curl -X PUT "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001/featured" \
  -H "Content-Type: application/json" \
  -d '{"is_featured": true}'

# 取消精选
curl -X PUT "http://localhost:8080/api/v1/admin/blog/550e8400-e29b-41d4-a716-446655440001/featured" \
  -H "Content-Type: application/json" \
  -d '{"is_featured": false}'
```

---

## 常见使用场景

### 场景1：发布一篇文章的完整流程

```bash
# 1. 创建草稿
ARTICLE_ID=$(curl -X POST "http://localhost:8080/api/v1/admin/blog/create" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "new-article",
    "title": "新文章",
    "summary": "摘要",
    "content": "# 内容",
    "category": "tutorials",
    "status": "draft"
  }' | jq -r '.data.id')

# 2. 编辑和更新
curl -X PUT "http://localhost:8080/api/v1/admin/blog/$ARTICLE_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "# 更新后的内容\n\n详细内容...",
    "tags": ["AI", "教程"]
  }'

# 3. 发布文章
curl -X PUT "http://localhost:8080/api/v1/admin/blog/$ARTICLE_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "published"}'

# 4. 设为精选
curl -X PUT "http://localhost:8080/api/v1/admin/blog/$ARTICLE_ID/featured" \
  -H "Content-Type: application/json" \
  -d '{"is_featured": true}'
```

### 场景2：内容管理

```bash
# 查看所有草稿
curl "http://localhost:8080/api/v1/admin/blog/list?status=draft"

# 查看待审核的文章（假设使用草稿状态）
curl "http://localhost:8080/api/v1/admin/blog/list?status=draft&sort=latest"

# 批量发布多篇文章
for id in id1 id2 id3; do
  curl -X PUT "http://localhost:8080/api/v1/admin/blog/$id/status" \
    -H "Content-Type: application/json" \
    -d '{"status": "published"}'
done
```

### 场景3：数据分析

```bash
# 获取统计数据
curl "http://localhost:8080/api/v1/admin/blog/stats"

# 查看热门文章
curl "http://localhost:8080/api/v1/admin/blog/list?sort=popular&page_size=10"

# 查看最新文章
curl "http://localhost:8080/api/v1/admin/blog/list?sort=latest&page_size=10"
```

---

## 权限说明

**当前状态**: 接口暂未启用认证（开发阶段）

**生产环境建议**: 在 `SetupBlogAdminRoutes` 中启用中间件：

```go
adminBlog.Use(middleware.JWTAuth(), middleware.AdminAuth())
```

---

## 错误响应

```json
{
  "code": 400,
  "msg": "参数错误: slug is required",
  "data": null
}
```

**常见错误码**:
- `400` - 请求参数错误
- `401` - 未授权
- `403` - 无权限
- `404` - 资源不存在
- `500` - 服务器错误

---

## 前端集成示例

### React Admin Panel

```javascript
import axios from 'axios';

const API_BASE = 'http://localhost:8080/api/v1/admin/blog';

// 获取文章列表
export const getBlogList = async (params) => {
  const response = await axios.get(`${API_BASE}/list`, { params });
  return response.data;
};

// 创建文章
export const createBlog = async (data) => {
  const response = await axios.post(`${API_BASE}/create`, data);
  return response.data;
};

// 更新文章
export const updateBlog = async (id, data) => {
  const response = await axios.put(`${API_BASE}/${id}`, data);
  return response.data;
};

// 删除文章
export const deleteBlog = async (id) => {
  const response = await axios.delete(`${API_BASE}/${id}`);
  return response.data;
};

// 批量删除
export const batchDelete = async (ids) => {
  const response = await axios.post(`${API_BASE}/batch-delete`, { ids });
  return response.data;
};

// 更新状态
export const updateStatus = async (id, status) => {
  const response = await axios.put(`${API_BASE}/${id}/status`, { status });
  return response.data;
};

// 切换精选
export const toggleFeatured = async (id, isFeatured) => {
  const response = await axios.put(`${API_BASE}/${id}/featured`, { is_featured: isFeatured });
  return response.data;
};

// 获取统计
export const getStats = async () => {
  const response = await axios.get(`${API_BASE}/stats`);
  return response.data;
};
```

---

## 测试建议

1. **创建测试文章**: 使用 draft 状态测试
2. **更新测试**: 修改各种字段验证
3. **状态流转**: draft → published → archived
4. **批量操作**: 创建多篇后批量删除
5. **统计验证**: 每次操作后查看统计数据变化

---

## 相关文档

- [博客公开API文档](../BLOG_API.md)
- [快速开始指南](../../../docs/BLOG_QUICKSTART.md)


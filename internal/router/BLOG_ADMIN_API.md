# 博客管理接口文档

本文档说明博客文章的创建、更新、删除接口。

## 接口列表

| 接口 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/v1/blog/create` | POST | 创建文章 | 需要认证 |
| `/api/v1/blog/:id` | PUT | 更新文章 | 需要认证 |
| `/api/v1/blog/:id` | DELETE | 删除文章 | 需要认证 |

---

## 1. 创建文章

**接口地址**: `POST /api/v1/blog/create`

**请求头**:
```
Content-Type: application/json
Authorization: Bearer <token>  // 如果需要认证
```

**请求体**:

```json
{
  "slug": "my-blog-post",
  "title": "我的博客文章标题",
  "summary": "文章摘要，简短描述文章内容",
  "content": "# 文章正文\n\n这是Markdown格式的正文内容...",
  "category": "tutorials",
  "cover_image": "https://example.com/image.jpg",
  "author": "作者名称",
  "author_avatar": "https://example.com/avatar.jpg",
  "read_time": 5,
  "is_featured": false,
  "seo_description": "SEO描述，用于搜索引擎优化",
  "status": "published",
  "tags": ["标签1", "标签2", "标签3"],
  "seo_keywords": ["关键词1", "关键词2"]
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| slug | string | 是 | URL友好标识符，唯一 | "my-blog-post" |
| title | string | 是 | 文章标题 | "如何使用01Agent" |
| summary | string | 是 | 文章摘要 | "本文介绍..." |
| content | string | 是 | Markdown正文 | "# 标题\n\n内容..." |
| category | string | 是 | 分类 | "tutorials" |
| cover_image | string | 否 | 封面图URL | "https://..." |
| author | string | 否 | 作者，默认"01Agent Team" | "张三" |
| author_avatar | string | 否 | 作者头像URL | "https://..." |
| read_time | int | 否 | 阅读时间（分钟） | 5 |
| is_featured | bool | 否 | 是否精选，默认false | true |
| seo_description | string | 否 | SEO描述 | "..." |
| status | string | 否 | 状态，默认published | "draft" |
| tags | []string | 否 | 标签列表 | ["AI", "教程"] |
| seo_keywords | []string | 否 | SEO关键词 | ["AI写作"] |

**分类说明**:
- `product-updates`: 产品动态
- `tutorials`: 使用教程
- `tips-and-tricks`: 运营技巧
- `industry-insights`: 行业洞察
- `case-studies`: 案例故事

**状态说明**:
- `draft`: 草稿
- `published`: 已发布
- `archived`: 已归档

**请求示例**:

```bash
curl -X POST "http://localhost:8080/api/v1/api/v1/blog/create" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "ai-writing-guide-2026",
    "title": "2026年AI写作完全指南",
    "summary": "全面介绍AI写作工具的使用技巧和最佳实践",
    "content": "# 2026年AI写作完全指南\n\n## 什么是AI写作\n\nAI写作是...",
    "category": "tutorials",
    "cover_image": "https://example.com/cover.jpg",
    "author": "AI专家",
    "read_time": 10,
    "is_featured": true,
    "status": "published",
    "tags": ["AI写作", "教程", "工具"],
    "seo_keywords": ["AI写作", "人工智能", "内容创作"]
  }'
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "slug": "ai-writing-guide-2026",
    "title": "2026年AI写作完全指南",
    "summary": "全面介绍AI写作工具的使用技巧和最佳实践",
    "content": "# 2026年AI写作完全指南...",
    "category": "tutorials",
    "category_name": "使用教程",
    "cover_image": "https://example.com/cover.jpg",
    "author": "AI专家",
    "author_avatar": null,
    "publish_date": "2026-01-11T10:30:00Z",
    "updated_date": null,
    "read_time": 10,
    "views": 0,
    "likes": 0,
    "is_featured": true,
    "status": "published",
    "tags": ["AI写作", "教程", "工具"],
    "seo_keywords": ["AI写作", "人工智能", "内容创作"],
    "seo_description": null
  }
}
```

---

## 2. 更新文章

**接口地址**: `PUT /api/v1/blog/:id`

**路径参数**:
- `id`: 文章ID（UUID格式）

**请求体**:

所有字段都是可选的，只需要传入需要更新的字段。

```json
{
  "title": "更新后的标题",
  "summary": "更新后的摘要",
  "content": "更新后的正文",
  "category": "tips-and-tricks",
  "cover_image": "https://example.com/new-image.jpg",
  "author": "新作者",
  "read_time": 8,
  "is_featured": true,
  "seo_description": "新的SEO描述",
  "status": "published",
  "tags": ["新标签1", "新标签2"],
  "seo_keywords": ["新关键词1", "新关键词2"]
}
```

**注意**:
- 如果传入 `tags`，会完全替换原有标签
- 如果传入 `seo_keywords`，会完全替换原有关键词
- 如果不传这些字段，则保持原值不变

**请求示例**:

```bash
curl -X PUT "http://localhost:8080/api/v1/api/v1/blog/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "2026年AI写作完全指南（更新版）",
    "is_featured": false,
    "tags": ["AI", "写作", "2026"]
  }'
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "更新成功",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "slug": "ai-writing-guide-2026",
    "title": "2026年AI写作完全指南（更新版）",
    "updated_date": "2026-01-11T11:00:00Z",
    ...
  }
}
```

---

## 3. 删除文章

**接口地址**: `DELETE /api/v1/blog/:id`

**路径参数**:
- `id`: 文章ID（UUID格式）

**请求示例**:

```bash
curl -X DELETE "http://localhost:8080/api/v1/api/v1/blog/550e8400-e29b-41d4-a716-446655440000"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "删除成功",
  "data": null
}
```

**注意**:
- 删除文章会级联删除相关的标签关联和SEO关键词
- 删除操作不可恢复，请谨慎使用

---

## 错误响应

所有接口在出错时返回统一的错误格式：

```json
{
  "code": 400,  // 错误码
  "msg": "参数错误: slug is required",
  "data": null
}
```

**常见错误码**:

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权（需要登录） |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 完整示例：创建一篇博客文章

```javascript
// JavaScript/Node.js 示例
const axios = require('axios');

async function createBlogPost() {
  try {
    const response = await axios.post('http://localhost:8080/api/v1/api/v1/blog/create', {
      slug: 'getting-started-with-ai',
      title: 'AI创作入门指南',
      summary: '从零开始学习AI辅助内容创作',
      content: `# AI创作入门指南

## 什么是AI创作

AI创作是利用人工智能技术辅助内容创作的过程...

## 如何开始

1. 选择合适的AI工具
2. 学习基本的提示词技巧
3. 开始你的第一次创作

## 推荐工具

- 01Agent - 专业的AI内容创作平台
- ChatGPT - 通用AI助手
- Midjourney - AI图像生成

## 总结

AI创作能显著提升创作效率，但需要人类的指导和优化。`,
      category: 'tutorials',
      cover_image: 'https://example.com/ai-cover.jpg',
      author: '李四',
      read_time: 8,
      is_featured: true,
      status: 'published',
      tags: ['AI创作', '入门教程', '工具推荐'],
      seo_keywords: ['AI创作', '人工智能', '内容创作', '01Agent']
    });

    console.log('创建成功:', response.data);
    return response.data.data;
  } catch (error) {
    console.error('创建失败:', error.response?.data || error.message);
  }
}

createBlogPost();
```

```python
# Python 示例
import requests
import json

def create_blog_post():
    url = "http://localhost:8080/api/v1/api/v1/blog/create"
    
    data = {
        "slug": "getting-started-with-ai",
        "title": "AI创作入门指南",
        "summary": "从零开始学习AI辅助内容创作",
        "content": """# AI创作入门指南

## 什么是AI创作

AI创作是利用人工智能技术辅助内容创作的过程...""",
        "category": "tutorials",
        "cover_image": "https://example.com/ai-cover.jpg",
        "author": "李四",
        "read_time": 8,
        "is_featured": True,
        "status": "published",
        "tags": ["AI创作", "入门教程", "工具推荐"],
        "seo_keywords": ["AI创作", "人工智能", "内容创作", "01Agent"]
    }
    
    response = requests.post(url, json=data)
    
    if response.status_code == 200:
        result = response.json()
        print("创建成功:", json.dumps(result, indent=2, ensure_ascii=False))
        return result['data']
    else:
        print("创建失败:", response.text)
        return None

if __name__ == "__main__":
    create_blog_post()
```

---

## 批量操作

### 批量创建文章

```bash
# 使用循环批量创建
for i in {1..5}; do
  curl -X POST "http://localhost:8080/api/v1/api/v1/blog/create" \
    -H "Content-Type: application/json" \
    -d "{
      \"slug\": \"test-post-$i\",
      \"title\": \"测试文章 $i\",
      \"summary\": \"这是第 $i 篇测试文章\",
      \"content\": \"# 测试文章 $i\\n\\n这是测试内容\",
      \"category\": \"tutorials\",
      \"status\": \"draft\"
    }"
done
```

---

## 最佳实践

1. **Slug设计**
   - 使用小写字母和连字符
   - 简短且有意义
   - 包含关键词
   - 示例：`ai-writing-tips-2026`

2. **内容格式**
   - 使用Markdown格式
   - 合理使用标题层级
   - 添加代码块和列表
   - 插入图片链接

3. **SEO优化**
   - 填写完整的SEO描述
   - 添加相关关键词
   - 标题包含关键词
   - 摘要突出核心价值

4. **标签管理**
   - 每篇文章2-5个标签
   - 标签要有意义
   - 保持标签体系一致

5. **状态管理**
   - 草稿：正在编辑的文章
   - 已发布：对外公开的文章
   - 已归档：不再显示但保留的文章

---

## 相关文档

- [博客查询接口](./BLOG_API.md)
- [快速开始指南](../docs/BLOG_QUICKSTART.md)
- [数据库脚本说明](../configs/README.md)


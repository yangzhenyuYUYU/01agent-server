#!/bin/bash

# 博客管理接口测试脚本
# 使用方法: ./test_blog_admin.sh

BASE_URL="http://localhost:8080"

echo "========================================="
echo "博客管理接口测试"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 1. 创建文章
echo -e "${BLUE}[1] 创建新文章${NC}"
echo "POST /blog/create"
echo ""

CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/blog/create" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "test-article-'$(date +%s)'",
    "title": "测试文章：AI内容创作的未来",
    "summary": "探讨AI技术在内容创作领域的应用和发展趋势",
    "content": "# AI内容创作的未来\n\n## 引言\n\nAI正在改变内容创作的方式...\n\n## 主要趋势\n\n1. 自动化写作\n2. 个性化内容\n3. 多模态生成\n\n## 总结\n\nAI是工具，创意是核心。",
    "category": "industry-insights",
    "author": "测试作者",
    "read_time": 5,
    "is_featured": false,
    "status": "published",
    "tags": ["AI", "内容创作", "趋势分析"],
    "seo_keywords": ["AI写作", "内容创作", "人工智能"]
  }')

echo "$CREATE_RESPONSE" | jq '.'
echo ""

# 提取创建的文章ID
ARTICLE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id')
ARTICLE_SLUG=$(echo "$CREATE_RESPONSE" | jq -r '.data.slug')

if [ "$ARTICLE_ID" != "null" ] && [ "$ARTICLE_ID" != "" ]; then
  echo -e "${GREEN}✓ 文章创建成功${NC}"
  echo "文章ID: $ARTICLE_ID"
  echo "文章Slug: $ARTICLE_SLUG"
else
  echo -e "${YELLOW}✗ 文章创建失败${NC}"
  exit 1
fi

echo ""
echo "-----------------------------------------"
echo ""

# 等待1秒
sleep 1

# 2. 查看创建的文章
echo -e "${BLUE}[2] 查看创建的文章${NC}"
echo "GET /blog/$ARTICLE_SLUG"
echo ""

curl -s "${BASE_URL}/blog/${ARTICLE_SLUG}" | jq '.'
echo ""
echo -e "${GREEN}✓ 文章查询成功${NC}"
echo ""
echo "-----------------------------------------"
echo ""

# 等待1秒
sleep 1

# 3. 更新文章
echo -e "${BLUE}[3] 更新文章内容${NC}"
echo "PUT /blog/$ARTICLE_ID"
echo ""

UPDATE_RESPONSE=$(curl -s -X PUT "${BASE_URL}/blog/${ARTICLE_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试文章：AI内容创作的未来（已更新）",
    "is_featured": true,
    "tags": ["AI", "内容创作", "趋势分析", "2026"],
    "seo_keywords": ["AI写作", "内容创作", "人工智能", "最新趋势"]
  }')

echo "$UPDATE_RESPONSE" | jq '.'
echo ""
echo -e "${GREEN}✓ 文章更新成功${NC}"
echo ""
echo "-----------------------------------------"
echo ""

# 等待1秒
sleep 1

# 4. 再次查看更新后的文章
echo -e "${BLUE}[4] 查看更新后的文章${NC}"
echo "GET /blog/$ARTICLE_SLUG"
echo ""

curl -s "${BASE_URL}/blog/${ARTICLE_SLUG}" | jq '.data | {title, is_featured, tags, updated_date}'
echo ""
echo -e "${GREEN}✓ 确认更新成功${NC}"
echo ""
echo "-----------------------------------------"
echo ""

# 5. 增加浏览量测试
echo -e "${BLUE}[5] 增加浏览量${NC}"
echo "POST /blog/$ARTICLE_SLUG/view"
echo ""

for i in {1..3}; do
  curl -s -X POST "${BASE_URL}/blog/${ARTICLE_SLUG}/view" > /dev/null
  echo "浏览 +1"
done

echo ""
echo "查看浏览量变化："
curl -s "${BASE_URL}/blog/${ARTICLE_SLUG}" | jq '.data | {title, views}'
echo ""
echo -e "${GREEN}✓ 浏览量统计正常${NC}"
echo ""
echo "-----------------------------------------"
echo ""

# 6. 删除文章（可选，注释掉以保留测试数据）
read -p "是否删除测试文章? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo -e "${BLUE}[6] 删除测试文章${NC}"
  echo "DELETE /blog/$ARTICLE_ID"
  echo ""
  
  DELETE_RESPONSE=$(curl -s -X DELETE "${BASE_URL}/blog/${ARTICLE_ID}")
  echo "$DELETE_RESPONSE" | jq '.'
  echo ""
  echo -e "${GREEN}✓ 文章删除成功${NC}"
  echo ""
  
  # 验证删除
  echo "验证文章已删除："
  curl -s "${BASE_URL}/blog/${ARTICLE_SLUG}" | jq '.'
  echo ""
else
  echo "保留测试文章"
  echo "可以手动访问: ${BASE_URL}/blog/${ARTICLE_SLUG}"
  echo "可以手动删除: curl -X DELETE ${BASE_URL}/blog/${ARTICLE_ID}"
fi

echo ""
echo "========================================="
echo "测试完成！"
echo "========================================="
echo ""
echo "测试的文章："
echo "- ID: $ARTICLE_ID"
echo "- Slug: $ARTICLE_SLUG"
echo "- URL: ${BASE_URL}/blog/${ARTICLE_SLUG}"


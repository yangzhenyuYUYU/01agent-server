#!/bin/bash

# 博客功能测试脚本
# 使用方法: ./test_blog_api.sh

BASE_URL="http://localhost:8080"

echo "========================================="
echo "博客 API 测试脚本"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local description=$3
    
    echo -e "${BLUE}测试: ${description}${NC}"
    echo "请求: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "${BASE_URL}${endpoint}")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X "$method" "${BASE_URL}${endpoint}")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE/d')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "$body"
    fi
    
    echo ""
    echo "-----------------------------------------"
    echo ""
}

# 检查服务器是否运行
echo "检查服务器状态..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo -e "${RED}错误: 服务器未运行！${NC}"
    echo "请先启动服务器: go run main.go"
    exit 1
fi
echo -e "${GREEN}✓ 服务器正在运行${NC}"
echo ""

# 1. 测试博客列表接口
echo "========================================="
echo "1. 测试博客列表接口"
echo "========================================="
echo ""

test_api "GET" "/blog/list" "获取默认列表（第1页，每页10条）"
test_api "GET" "/blog/list?page=1&page_size=5" "获取列表（第1页，每页5条）"
test_api "GET" "/blog/list?category=tutorials" "按分类筛选（教程类）"
test_api "GET" "/blog/list?is_featured=true" "获取精选文章"
test_api "GET" "/blog/list?keyword=快速" "关键词搜索"
test_api "GET" "/blog/list?sort=popular" "按热门排序"
test_api "GET" "/blog/list?sort=views" "按浏览量排序"

# 2. 测试文章详情接口
echo "========================================="
echo "2. 测试文章详情接口"
echo "========================================="
echo ""

test_api "GET" "/blog/getting-started-with-01agent" "获取文章详情（存在）"
test_api "GET" "/blog/non-existent-slug" "获取文章详情（不存在）"

# 3. 测试 Sitemap 接口
echo "========================================="
echo "3. 测试 Sitemap 接口"
echo "========================================="
echo ""

test_api "GET" "/blog/sitemap" "获取 sitemap 数据"

# 4. 测试相关文章接口
echo "========================================="
echo "4. 测试相关文章推荐"
echo "========================================="
echo ""

test_api "GET" "/blog/getting-started-with-01agent/related" "获取相关文章（默认3条）"
test_api "GET" "/blog/getting-started-with-01agent/related?limit=5" "获取相关文章（5条）"
test_api "GET" "/blog/non-existent-slug/related" "获取相关文章（文章不存在）"

# 5. 测试浏览量统计接口
echo "========================================="
echo "5. 测试浏览量统计"
echo "========================================="
echo ""

test_api "POST" "/blog/getting-started-with-01agent/view" "增加浏览量"
test_api "POST" "/blog/non-existent-slug/view" "增加浏览量（文章不存在）"

# 6. 综合测试
echo "========================================="
echo "6. 综合测试场景"
echo "========================================="
echo ""

echo -e "${BLUE}场景: 用户浏览博客流程${NC}"
echo "1) 访问首页，获取精选文章"
test_api "GET" "/blog/list?is_featured=true&page_size=3" "获取首页精选"

echo "2) 查看某个分类的文章列表"
test_api "GET" "/blog/list?category=tutorials&page=1&page_size=5" "浏览教程分类"

echo "3) 点击文章查看详情"
test_api "GET" "/blog/getting-started-with-01agent" "查看文章详情"

echo "4) 记录浏览量"
test_api "POST" "/blog/getting-started-with-01agent/view" "统计浏览量"

echo "5) 获取相关推荐"
test_api "GET" "/blog/getting-started-with-01agent/related?limit=3" "获取相关文章"

echo ""
echo "========================================="
echo "测试完成！"
echo "========================================="


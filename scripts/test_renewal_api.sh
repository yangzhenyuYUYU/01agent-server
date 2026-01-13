#!/bin/bash

# 续费用户统计API测试脚本
# 使用方法: ./test_renewal_api.sh

# ========== 配置 ==========
BASE_URL="http://localhost:8099/api/v1/admin"
ADMIN_TOKEN="your_admin_token_here"  # 替换为实际的管理员token

# ========== 颜色定义 ==========
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ========== 函数定义 ==========
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# 测试API
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    
    echo -e "${YELLOW}测试: $name${NC}"
    echo "请求: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" \
            -X $method \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo "状态码: $http_code"
    echo "响应:"
    echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        print_success "测试通过"
    else
        print_error "测试失败"
    fi
    echo ""
}

# ========== 主测试流程 ==========

print_header "续费用户统计API测试"

# 检查依赖
if ! command -v curl &> /dev/null; then
    print_error "curl 未安装，请先安装 curl"
    exit 1
fi

if ! command -v python3 &> /dev/null; then
    print_info "python3 未安装，JSON格式化将不可用"
fi

# 提示用户配置token
if [ "$ADMIN_TOKEN" = "your_admin_token_here" ]; then
    print_error "请先在脚本中配置 ADMIN_TOKEN"
    echo "可以通过以下方式获取 token:"
    echo "1. 使用管理员账号登录"
    echo "2. curl -X POST $BASE_URL/auth/login -H \"Content-Type: application/json\" -d '{\"username\":\"admin\",\"password\":\"your_password\"}'"
    exit 1
fi

# ========== 测试1: 获取续费用户排行榜（按续费次数排序） ==========
print_header "测试1: 获取续费用户排行榜（按续费次数排序）"
test_api \
    "续费次数排行榜（默认前100名）" \
    "GET" \
    "/analytics/renewal-ranking?sort_by=count&limit=10" \
    ""

# ========== 测试2: 获取续费用户排行榜（按续费金额排序） ==========
print_header "测试2: 获取续费用户排行榜（按续费金额排序）"
test_api \
    "续费金额排行榜（前10名）" \
    "GET" \
    "/analytics/renewal-ranking?sort_by=amount&limit=10" \
    ""

# ========== 测试3: 获取续费用户排行榜（指定时间范围） ==========
print_header "测试3: 获取续费用户排行榜（指定时间范围）"
START_DATE="2025-01-01"
END_DATE="2025-12-31"
test_api \
    "2025年续费排行榜" \
    "GET" \
    "/analytics/renewal-ranking?start_date=$START_DATE&end_date=$END_DATE&sort_by=count&limit=20" \
    ""

# ========== 测试4: 获取续费统计汇总 ==========
print_header "测试4: 获取续费统计汇总"
test_api \
    "续费统计汇总（全部数据）" \
    "GET" \
    "/analytics/renewal-summary" \
    ""

# ========== 测试5: 获取续费统计汇总（指定时间范围） ==========
print_header "测试5: 获取续费统计汇总（指定时间范围）"
test_api \
    "最近30天续费统计" \
    "GET" \
    "/analytics/renewal-summary?start_date=2025-11-01&end_date=2025-12-31" \
    ""

# ========== 测试6: 获取单个用户的续费详情 ==========
print_header "测试6: 获取单个用户的续费详情"
# 注意：这里需要替换为实际的用户ID
print_info "提示：请替换USER_ID为实际的续费用户ID"
USER_ID="test_user_id"  # 替换为实际的用户ID
test_api \
    "用户续费详情" \
    "GET" \
    "/analytics/renewal-detail/$USER_ID" \
    ""

# ========== 测试7: 错误处理测试 ==========
print_header "测试7: 错误处理测试"

test_api \
    "无效的排序参数" \
    "GET" \
    "/analytics/renewal-ranking?sort_by=invalid" \
    ""

test_api \
    "无效的日期格式" \
    "GET" \
    "/analytics/renewal-summary?start_date=invalid-date" \
    ""

test_api \
    "不存在的用户ID" \
    "GET" \
    "/analytics/renewal-detail/non_existent_user" \
    ""

# ========== 测试完成 ==========
print_header "测试完成"
print_success "所有测试已完成"

echo -e "\n${BLUE}API端点总结:${NC}"
echo "1. GET /analytics/renewal-ranking - 续费用户排行榜"
echo "   参数: start_date, end_date, sort_by(count|amount), limit"
echo ""
echo "2. GET /analytics/renewal-summary - 续费统计汇总"
echo "   参数: start_date, end_date"
echo ""
echo "3. GET /analytics/renewal-detail/:user_id - 用户续费详情"
echo "   参数: user_id (路径参数)"
echo ""

print_info "注意事项:"
echo "1. 续费用户定义：购买订阅服务次数 ≥ 2次的用户"
echo "2. 时间范围：使用 YYYY-MM-DD 格式"
echo "3. 排序方式：count(续费次数) 或 amount(续费金额)"
echo "4. 限制数量：默认100，最大1000"

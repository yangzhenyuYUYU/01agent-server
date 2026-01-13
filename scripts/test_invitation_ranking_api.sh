#!/bin/bash

# 邀请用户排名API测试脚本
# 用于测试基于缓存表的邀请排名功能

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
BASE_URL="${BASE_URL:-http://localhost:8099}"
ADMIN_TOKEN="${1:-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIwMWFnZW50X3NlcnZlciIsInN1YiI6ImFkbWluIiwiZXhwIjoxODkzNDU2MDAwLCJuYmYiOjE3MzY4NTEyMDAsImlhdCI6MTczNjg1MTIwMCwianRpIjoiODdkMjA5MGEtYzJiMi00OGVjLWI4ZTUtNmEzZWI2ZDYwNGNiIn0.V8ZJGfTvOJvLXOg2DQFDlPLz4yOpOyp3f4oa3IZiSB8}"

# 打印分隔线
print_separator() {
    echo -e "${BLUE}============================================${NC}"
}

# 打印标题
print_title() {
    echo -e "${BLUE}▶ $1${NC}"
}

# 打印成功信息
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 打印错误信息
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# 打印警告信息
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 检查API响应
check_response() {
    local response=$1
    local test_name=$2
    
    if echo "$response" | jq -e '.code == 200' > /dev/null 2>&1; then
        print_success "$test_name - 成功"
        return 0
    else
        print_error "$test_name - 失败"
        echo "$response" | jq '.'
        return 1
    fi
}

echo ""
print_separator
echo -e "${BLUE}🎯 邀请用户排名API测试${NC}"
echo -e "${BLUE}📡 服务地址: $BASE_URL${NC}"
echo -e "${BLUE}🔑 Token: ${ADMIN_TOKEN:0:50}...${NC}"
print_separator
echo ""

# ============================================
# 测试1: 获取邀请排名（按综合分）
# ============================================
print_title "测试1: 获取邀请排名（按综合分排序）"
echo "GET /api/v1/admin/analytics/invitation-ranking-v2?sort_by=score&limit=10"
echo ""

RESPONSE=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-ranking-v2?sort_by=score&limit=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")

if check_response "$RESPONSE" "获取邀请排名（综合分）"; then
    echo ""
    print_title "📊 排名数据预览"
    echo "$RESPONSE" | jq -r '.data.rankings[0:3][] | "  排名\(.rank): \(.nickname // .user_id) - 总邀请\(.total_invitations)人, 有效邀请\(.paid_invitations)人, 综合分\(.ranking_score)"'
    echo ""
    print_title "📈 系统指标"
    echo "$RESPONSE" | jq -r '.data.metrics | "  总用户数: \(.total_users)\n  邀请用户数: \(.active_inviters)\n  分享率: \(.share_rate)%\n  总邀请人数: \(.total_invitations)\n  有效邀请人数: \(.paid_invitations)\n  平均裂变系数: \(.avg_viral_coefficient)\n  转化率: \(.conversion_rate)%\n  总佣金: ¥\(.total_commission)"'
fi

echo ""
print_separator
echo ""

# ============================================
# 测试2: 获取邀请排名（按总邀请数）
# ============================================
print_title "测试2: 获取邀请排名（按总邀请数排序）"
echo "GET /api/v1/admin/analytics/invitation-ranking-v2?sort_by=total&limit=5"
echo ""

RESPONSE=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-ranking-v2?sort_by=total&limit=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")

if check_response "$RESPONSE" "获取邀请排名（总邀请数）"; then
    echo ""
    print_title "🏆 Top 5 邀请达人"
    echo "$RESPONSE" | jq -r '.data.rankings[] | "  \(.rank). \(.nickname // .user_id)\n     总邀请: \(.total_invitations)人\n     有效邀请: \(.paid_invitations)人\n     质量分: \(.invitation_quality_score)分\n     活跃度: \(.activity_score)分"'
fi

echo ""
print_separator
echo ""

# ============================================
# 测试3: 获取邀请排名（按有效邀请数）
# ============================================
print_title "测试3: 获取邀请排名（按有效邀请数排序）"
echo "GET /api/v1/admin/analytics/invitation-ranking-v2?sort_by=paid&limit=5"
echo ""

RESPONSE=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-ranking-v2?sort_by=paid&limit=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")

if check_response "$RESPONSE" "获取邀请排名（有效邀请数）"; then
    echo ""
    print_title "💎 Top 5 高质量邀请"
    echo "$RESPONSE" | jq -r '.data.rankings[] | "  \(.rank). \(.nickname // .user_id) - 有效邀请 \(.paid_invitations)/\(.total_invitations)人 (\(.invitation_quality_score)%)"'
fi

echo ""
print_separator
echo ""

# ============================================
# 测试4: 获取邀请排名（按佣金）
# ============================================
print_title "测试4: 获取邀请排名（按佣金排序）"
echo "GET /api/v1/admin/analytics/invitation-ranking-v2?sort_by=commission&limit=5"
echo ""

RESPONSE=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-ranking-v2?sort_by=commission&limit=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")

if check_response "$RESPONSE" "获取邀请排名（佣金）"; then
    echo ""
    print_title "💰 Top 5 佣金收入"
    echo "$RESPONSE" | jq -r '.data.rankings[] | "  \(.rank). \(.nickname // .user_id) - 总佣金 ¥\(.total_commission)"'
fi

echo ""
print_separator
echo ""

# ============================================
# 测试5: 获取系统级邀请指标
# ============================================
print_title "测试5: 获取系统级邀请指标"
echo "GET /api/v1/admin/analytics/invitation-metrics"
echo ""

RESPONSE=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-metrics" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")

if check_response "$RESPONSE" "获取系统级邀请指标"; then
    echo ""
    print_title "📊 详细指标"
    echo "$RESPONSE" | jq '.data'
fi

echo ""
print_separator
echo ""

# ============================================
# 测试6: 获取用户邀请详情（需要真实用户ID）
# ============================================
print_title "测试6: 获取用户邀请详情"

# 先从排名中获取一个用户ID
USER_ID=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-ranking-v2?sort_by=total&limit=1" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" | jq -r '.data.rankings[0].user_id // empty')

if [ -n "$USER_ID" ]; then
    echo "GET /api/v1/admin/analytics/invitation-detail/$USER_ID?page=1&page_size=5"
    echo ""
    
    RESPONSE=$(curl -s -X GET \
      "$BASE_URL/api/v1/admin/analytics/invitation-detail/$USER_ID?page=1&page_size=5" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json")
    
    if check_response "$RESPONSE" "获取用户邀请详情"; then
        echo ""
        print_title "👥 邀请列表（前5个）"
        echo "$RESPONSE" | jq -r '.data.details[]? | "  • \(.nickname // .invitee_id)\n    邀请时间: \(.invited_date)\n    是否付费: \(if .is_paid then "是" else "否" end)\n    订单数: \(.order_count)\n    总支付: ¥\(.total_payment)"'
    fi
else
    print_warning "没有找到邀请用户，跳过此测试"
fi

echo ""
print_separator
echo ""

# ============================================
# 测试7: 获取缓存状态
# ============================================
print_title "测试7: 获取缓存状态"
echo "GET /api/v1/admin/analytics/invitation-ranking/cache-status"
echo ""

RESPONSE=$(curl -s -X GET \
  "$BASE_URL/api/v1/admin/analytics/invitation-ranking/cache-status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json")

if check_response "$RESPONSE" "获取缓存状态"; then
    echo ""
    print_title "🗄️ 缓存信息"
    echo "$RESPONSE" | jq -r '.data | "  总记录数: \(.total_records)\n  最后更新: \(.last_updated)\n  最旧更新: \(.oldest_updated)\n  状态: \(.status)"'
fi

echo ""
print_separator
echo ""

# ============================================
# 测试8: 手动刷新缓存
# ============================================
print_title "测试8: 手动刷新缓存（可选）"
echo "POST /api/v1/admin/analytics/invitation-ranking/refresh"
echo ""

read -p "$(echo -e ${YELLOW}是否执行缓存刷新？这可能需要几秒钟 [y/N]: ${NC})" -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_warning "正在刷新缓存，请稍候..."
    
    RESPONSE=$(curl -s -X POST \
      "$BASE_URL/api/v1/admin/analytics/invitation-ranking/refresh" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json")
    
    if check_response "$RESPONSE" "手动刷新缓存"; then
        echo ""
        print_success "缓存刷新完成"
        echo "$RESPONSE" | jq '.data'
    fi
else
    print_warning "跳过缓存刷新"
fi

echo ""
print_separator
echo ""

# ============================================
# 测试总结
# ============================================
print_title "✨ 测试完成"
echo ""
echo "📖 更多信息请查看文档:"
echo "  - 详细指南: docs/INVITATION_RANKING_GUIDE.md"
echo "  - 快速开始: INVITATION_RANKING_QUICKSTART.md"
echo "  - 设计文档: docs/INVITATION_RANKING_DESIGN.md"
echo ""
print_separator

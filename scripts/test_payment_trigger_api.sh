#!/bin/bash

# 首充触发点分析API测试脚本
# 使用方法: ./test_payment_trigger_api.sh <admin_token>

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
API_HOST="${API_HOST:-http://localhost:8080}"
ADMIN_TOKEN="${1:-YOUR_ADMIN_TOKEN}"

if [ "$ADMIN_TOKEN" = "YOUR_ADMIN_TOKEN" ]; then
    echo -e "${RED}错误: 请提供管理员Token${NC}"
    echo "使用方法: $0 <admin_token>"
    exit 1
fi

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  首充触发点分析API测试${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# 测试1: 获取首充触发点洞察（最简单）
echo -e "${GREEN}[测试1] 获取首充触发点洞察（最近30天）${NC}"
echo -e "${YELLOW}接口: GET /api/v1/admin/analytics/payment-trigger/insights${NC}"
echo ""

RESPONSE=$(curl -s -X GET "${API_HOST}/api/v1/admin/analytics/payment-trigger/insights" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}")

echo "$RESPONSE" | jq '.'

# 提取关键指标
TOTAL_USERS=$(echo "$RESPONSE" | jq -r '.data.total_paying_users // 0')
AVG_CREDITS=$(echo "$RESPONSE" | jq -r '.data.avg_credits_before_payment // 0')
MEDIAN_CREDITS=$(echo "$RESPONSE" | jq -r '.data.median_credits_before_payment // 0')

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}核心洞察:${NC}"
echo -e "  📊 付费用户数: ${BLUE}${TOTAL_USERS}${NC} 人"
echo -e "  💰 平均首充前消耗: ${BLUE}${AVG_CREDITS}${NC} 积分"
echo -e "  📈 中位数消耗: ${BLUE}${MEDIAN_CREDITS}${NC} 积分"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo ""

# 测试2: 获取完整分析（指定时间范围）
echo -e "${GREEN}[测试2] 获取完整的首充触发点分析${NC}"
echo -e "${YELLOW}接口: GET /api/v1/admin/analytics/payment-trigger${NC}"
echo -e "${YELLOW}参数: start_date=2026-01-01, end_date=2026-01-31${NC}"
echo ""

RESPONSE=$(curl -s -X GET "${API_HOST}/api/v1/admin/analytics/payment-trigger?start_date=2026-01-01&end_date=2026-01-31" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}")

echo "$RESPONSE" | jq '.data.summary'

# 提取Top场景
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Top转化场景:${NC}"
echo "$RESPONSE" | jq -r '.data.summary.top_scenes[]? | "  🔥 \(.service_name): \(.user_count)人, 总消耗\(.total_credits)积分, 平均\(.avg_credits)积分"'
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 提取产品分布
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}产品选择偏好:${NC}"
echo "$RESPONSE" | jq -r '.data.summary.product_distribution[]? | "  🎯 \(.product_name) (\(.product_type)): \(.percentage)%, \(.user_count)人"'
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 提取积分区间分布
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}积分消耗区间分布:${NC}"
echo "$RESPONSE" | jq -r '.data.summary.credit_range_distribution[]? | "  📊 \(.range_start)-\(.range_end)积分: \(.user_count)人 (\(.percentage)%)"'
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo ""

# 测试3: 获取单个用户的首充分析（需要用户输入）
echo -e "${GREEN}[测试3] 获取单个用户的首充分析${NC}"
echo -e "${YELLOW}请输入要查询的用户ID (直接回车跳过): ${NC}"
read -r USER_ID

if [ -n "$USER_ID" ]; then
    echo -e "${YELLOW}接口: GET /api/v1/admin/analytics/payment-trigger/user?user_id=${USER_ID}${NC}"
    echo ""
    
    RESPONSE=$(curl -s -X GET "${API_HOST}/api/v1/admin/analytics/payment-trigger/user?user_id=${USER_ID}" \
      -H "Authorization: Bearer ${ADMIN_TOKEN}")
    
    echo "$RESPONSE" | jq '.'
    
    # 提取关键信息
    CREDITS=$(echo "$RESPONSE" | jq -r '.data.credits_before_payment // 0')
    PRODUCT=$(echo "$RESPONSE" | jq -r '.data.product_name // "未知"')
    FIRST_PAY_TIME=$(echo "$RESPONSE" | jq -r '.data.first_payment_time // "未知"')
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}用户${USER_ID}的首充分析:${NC}"
    echo -e "  👤 首次充值时间: ${BLUE}${FIRST_PAY_TIME}${NC}"
    echo -e "  💰 首充前消耗: ${BLUE}${CREDITS}${NC} 积分"
    echo -e "  🎁 购买产品: ${BLUE}${PRODUCT}${NC}"
    echo -e "  📊 场景消耗分布:"
    echo "$RESPONSE" | jq -r '.data.scene_consumption | to_entries[] | "      • \(.key): \(.value)积分"'
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
else
    echo -e "${YELLOW}跳过单用户查询测试${NC}"
    echo ""
fi

# 业务洞察总结
echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  业务洞察总结${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "${GREEN}根据以上数据，您可以：${NC}"
echo ""
echo -e "1️⃣  ${YELLOW}优化免费积分额度${NC}"
echo -e "   建议将新用户免费积分设置为接近中位数的值（约${BLUE}${MEDIAN_CREDITS}${NC}积分）"
echo ""
echo -e "2️⃣  ${YELLOW}识别核心转化场景${NC}"
echo -e "   在新用户引导中优先推荐Top场景"
echo ""
echo -e "3️⃣  ${YELLOW}优化产品推荐策略${NC}"
echo -e "   首充优惠主推最受欢迎的产品"
echo ""
echo -e "4️⃣  ${YELLOW}设计精准营销触达${NC}"
echo -e "   在用户消耗约${BLUE}${MEDIAN_CREDITS}${NC}积分时触发付费转化提示"
echo ""
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "${GREEN}✅ 测试完成！${NC}"
echo ""
echo -e "详细文档: docs/PAYMENT_TRIGGER_QUICKSTART.md"
echo ""

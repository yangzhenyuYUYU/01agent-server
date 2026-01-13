#!/bin/bash

# 测试用户使用排名API
# 用法: ./test_user_ranking_api.sh [admin_token]

# 设置颜色
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1/admin/analytics"

# 从参数获取token，或使用默认token
ADMIN_TOKEN=${1:-"your_admin_token_here"}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  用户使用排名API测试${NC}"
echo -e "${BLUE}========================================${NC}\n"

# 测试1: 获取每日用户排名（JSON格式）
echo -e "${YELLOW}测试1: 获取最近7天的每日用户排名（JSON格式）${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=daily&days=7&top=10&format=json"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=daily&days=7&top=10&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo -e "\n"

# 测试2: 获取每周用户排名（JSON格式）
echo -e "${YELLOW}测试2: 获取最近30天的每周用户排名（JSON格式）${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=weekly&days=30&top=10&format=json"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=weekly&days=30&top=10&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo -e "\n"

# 测试3: 获取每月用户排名（JSON格式）
echo -e "${YELLOW}测试3: 获取最近90天的每月用户排名（JSON格式）${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=monthly&days=90&top=10&format=json"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=monthly&days=90&top=10&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo -e "\n"

# 测试4: 获取每日用户排名（HTML格式）
echo -e "${YELLOW}测试4: 获取最近7天的每日用户排名（HTML格式）${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=daily&days=7&top=10&format=html"
echo -e "${BLUE}HTML输出已保存到: user_ranking_daily.html${NC}"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=daily&days=7&top=10&format=html" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" > user_ranking_daily.html
echo -e "\n"

# 测试5: 获取每周用户排名（HTML格式）
echo -e "${YELLOW}测试5: 获取最近30天的每周用户排名（HTML格式）${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=weekly&days=30&top=10&format=html"
echo -e "${BLUE}HTML输出已保存到: user_ranking_weekly.html${NC}"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=weekly&days=30&top=10&format=html" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" > user_ranking_weekly.html
echo -e "\n"

# 测试6: 获取每月用户排名（HTML格式）
echo -e "${YELLOW}测试6: 获取最近90天的每月用户排名（HTML格式）${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=monthly&days=90&top=20&format=html"
echo -e "${BLUE}HTML输出已保存到: user_ranking_monthly.html${NC}"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=monthly&days=90&top=20&format=html" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" > user_ranking_monthly.html
echo -e "\n"

# 测试7: 测试参数验证 - 无效的周期类型
echo -e "${YELLOW}测试7: 测试参数验证 - 无效的周期类型${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=yearly&days=30&top=10&format=json"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=yearly&days=30&top=10&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo -e "\n"

# 测试8: 测试参数验证 - 超出范围的天数
echo -e "${YELLOW}测试8: 测试参数验证 - 超出范围的天数${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking?period_type=daily&days=400&top=10&format=json"
curl -s -X GET "${BASE_URL}/user-ranking?period_type=daily&days=400&top=10&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo -e "\n"

# 测试9: 测试默认参数
echo -e "${YELLOW}测试9: 测试默认参数${NC}"
echo -e "${GREEN}请求URL:${NC} ${BASE_URL}/user-ranking"
curl -s -X GET "${BASE_URL}/user-ranking" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.'
echo -e "\n"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  测试完成！${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "\n${GREEN}提示:${NC}"
echo -e "1. JSON格式的结果已输出到终端"
echo -e "2. HTML格式的结果已保存到以下文件："
echo -e "   - user_ranking_daily.html"
echo -e "   - user_ranking_weekly.html"
echo -e "   - user_ranking_monthly.html"
echo -e "3. 使用浏览器打开HTML文件查看可视化报告"
echo -e "\n${YELLOW}API说明:${NC}"
echo -e "接口: GET ${BASE_URL}/user-ranking"
echo -e "参数:"
echo -e "  - period_type: 周期类型（daily/weekly/monthly）默认: daily"
echo -e "  - days: 统计天数（1-365）默认: 30"
echo -e "  - top: 每个时期显示的用户数量（1-100）默认: 10"
echo -e "  - format: 返回格式（json/html）默认: json"

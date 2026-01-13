#!/bin/bash

# 场景排名API测试脚本
# 使用方法: ./test_scene_ranking_api.sh [ADMIN_TOKEN]

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取管理员Token
ADMIN_TOKEN=$1
if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}错误: 请提供管理员Token${NC}"
    echo "使用方法: $0 <ADMIN_TOKEN>"
    exit 1
fi

# API基础URL
BASE_URL="http://localhost:8080/api/v1/admin/analytics"

# 输出目录
OUTPUT_DIR="./scene_ranking_reports"
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    场景排名API测试${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 测试1: 每日排名（JSON格式）
echo -e "${YELLOW}测试1: 获取最近7天的每日排名（JSON格式）${NC}"
curl -s -X GET "${BASE_URL}/scene-ranking?period_type=daily&days=7&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.' > "${OUTPUT_DIR}/daily_7days.json"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 测试通过${NC}"
    echo -e "  报告已保存至: ${OUTPUT_DIR}/daily_7days.json"
    
    # 显示第一个时期的数据
    echo -e "\n${BLUE}数据预览:${NC}"
    cat "${OUTPUT_DIR}/daily_7days.json" | jq '.data.rankings[0] | {period, paid_users: .paid_users[:3], free_users: .free_users[:3]}'
else
    echo -e "${RED}✗ 测试失败${NC}"
fi
echo ""

# 测试2: 每周排名（JSON格式）
echo -e "${YELLOW}测试2: 获取最近4周的每周排名（JSON格式）${NC}"
curl -s -X GET "${BASE_URL}/scene-ranking?period_type=weekly&days=28&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.' > "${OUTPUT_DIR}/weekly_4weeks.json"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 测试通过${NC}"
    echo -e "  报告已保存至: ${OUTPUT_DIR}/weekly_4weeks.json"
    
    # 显示数据概览
    echo -e "\n${BLUE}数据概览:${NC}"
    cat "${OUTPUT_DIR}/weekly_4weeks.json" | jq '{
        period_type: .data.period_type,
        start_date: .data.start_date,
        end_date: .data.end_date,
        total_periods: (.data.rankings | length)
    }'
else
    echo -e "${RED}✗ 测试失败${NC}"
fi
echo ""

# 测试3: 每月排名（JSON格式）
echo -e "${YELLOW}测试3: 获取最近3个月的每月排名（JSON格式）${NC}"
curl -s -X GET "${BASE_URL}/scene-ranking?period_type=monthly&days=90&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" | jq '.' > "${OUTPUT_DIR}/monthly_3months.json"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 测试通过${NC}"
    echo -e "  报告已保存至: ${OUTPUT_DIR}/monthly_3months.json"
    
    # 统计每个月的数据量
    echo -e "\n${BLUE}每月数据量:${NC}"
    cat "${OUTPUT_DIR}/monthly_3months.json" | jq '.data.rankings[] | {
        period: .period,
        paid_scenes: (.paid_users | length),
        free_scenes: (.free_users | length),
        all_scenes: (.all_users | length)
    }'
else
    echo -e "${RED}✗ 测试失败${NC}"
fi
echo ""

# 测试4: HTML报告 - 每日
echo -e "${YELLOW}测试4: 生成每日排名HTML报告${NC}"
curl -s -X GET "${BASE_URL}/scene-ranking?period_type=daily&days=7&format=html" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" > "${OUTPUT_DIR}/daily_report.html"

if [ $? -eq 0 ] && [ -s "${OUTPUT_DIR}/daily_report.html" ]; then
    echo -e "${GREEN}✓ 测试通过${NC}"
    echo -e "  报告已保存至: ${OUTPUT_DIR}/daily_report.html"
    echo -e "  在浏览器中打开查看: file://${PWD}/${OUTPUT_DIR}/daily_report.html"
else
    echo -e "${RED}✗ 测试失败${NC}"
fi
echo ""

# 测试5: HTML报告 - 每周
echo -e "${YELLOW}测试5: 生成每周排名HTML报告${NC}"
curl -s -X GET "${BASE_URL}/scene-ranking?period_type=weekly&days=28&format=html" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" > "${OUTPUT_DIR}/weekly_report.html"

if [ $? -eq 0 ] && [ -s "${OUTPUT_DIR}/weekly_report.html" ]; then
    echo -e "${GREEN}✓ 测试通过${NC}"
    echo -e "  报告已保存至: ${OUTPUT_DIR}/weekly_report.html"
    echo -e "  在浏览器中打开查看: file://${PWD}/${OUTPUT_DIR}/weekly_report.html"
else
    echo -e "${RED}✗ 测试失败${NC}"
fi
echo ""

# 测试6: HTML报告 - 每月
echo -e "${YELLOW}测试6: 生成每月排名HTML报告${NC}"
curl -s -X GET "${BASE_URL}/scene-ranking?period_type=monthly&days=180&format=html" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" > "${OUTPUT_DIR}/monthly_report.html"

if [ $? -eq 0 ] && [ -s "${OUTPUT_DIR}/monthly_report.html" ]; then
    echo -e "${GREEN}✓ 测试通过${NC}"
    echo -e "  报告已保存至: ${OUTPUT_DIR}/monthly_report.html"
    echo -e "  在浏览器中打开查看: file://${PWD}/${OUTPUT_DIR}/monthly_report.html"
else
    echo -e "${RED}✗ 测试失败${NC}"
fi
echo ""

# 测试7: 参数验证 - 无效周期类型
echo -e "${YELLOW}测试7: 参数验证 - 无效周期类型${NC}"
response=$(curl -s -X GET "${BASE_URL}/scene-ranking?period_type=invalid&days=7&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}")
error_msg=$(echo "$response" | jq -r '.message')

if [[ "$error_msg" == *"周期类型无效"* ]]; then
    echo -e "${GREEN}✓ 测试通过 - 正确返回错误信息${NC}"
    echo -e "  错误信息: $error_msg"
else
    echo -e "${RED}✗ 测试失败 - 未正确处理无效参数${NC}"
fi
echo ""

# 测试8: 参数验证 - 无效天数
echo -e "${YELLOW}测试8: 参数验证 - 无效天数${NC}"
response=$(curl -s -X GET "${BASE_URL}/scene-ranking?period_type=daily&days=500&format=json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}")
error_msg=$(echo "$response" | jq -r '.message')

if [[ "$error_msg" == *"天数参数无效"* ]]; then
    echo -e "${GREEN}✓ 测试通过 - 正确返回错误信息${NC}"
    echo -e "  错误信息: $error_msg"
else
    echo -e "${RED}✗ 测试失败 - 未正确处理无效参数${NC}"
fi
echo ""

# 测试9: 付费用户 vs 免费用户对比分析
echo -e "${YELLOW}测试9: 付费用户 vs 免费用户TOP场景对比${NC}"
if [ -f "${OUTPUT_DIR}/weekly_4weeks.json" ]; then
    echo -e "${BLUE}分析最近一周的数据:${NC}"
    cat "${OUTPUT_DIR}/weekly_4weeks.json" | jq '.data.rankings[0] | {
        period: .period,
        "付费用户TOP3": [.paid_users[:3][] | {rank, scene_name, usage_count}],
        "免费用户TOP3": [.free_users[:3][] | {rank, scene_name, usage_count}]
    }'
    echo -e "${GREEN}✓ 对比分析完成${NC}"
else
    echo -e "${RED}✗ 无数据可分析${NC}"
fi
echo ""

# 测试总结
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    测试总结${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}所有报告已保存至: ${OUTPUT_DIR}/${NC}"
echo ""
echo -e "JSON报告:"
ls -lh "${OUTPUT_DIR}"/*.json 2>/dev/null | awk '{print "  - " $9 " (" $5 ")"}'
echo ""
echo -e "HTML报告:"
ls -lh "${OUTPUT_DIR}"/*.html 2>/dev/null | awk '{print "  - " $9 " (" $5 ")"}'
echo ""
echo -e "${YELLOW}提示: 在浏览器中打开HTML报告查看精美的可视化效果${NC}"
echo ""

# 生成快速访问链接
echo -e "${BLUE}快速访问链接:${NC}"
echo -e "  每日报告: file://${PWD}/${OUTPUT_DIR}/daily_report.html"
echo -e "  每周报告: file://${PWD}/${OUTPUT_DIR}/weekly_report.html"
echo -e "  每月报告: file://${PWD}/${OUTPUT_DIR}/monthly_report.html"
echo ""

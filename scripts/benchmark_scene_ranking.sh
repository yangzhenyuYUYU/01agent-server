#!/bin/bash

# 场景排名API性能基准测试脚本
# 用于对比优化前后的性能差异

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     场景排名API性能基准测试                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
echo ""

# 测试函数
run_benchmark() {
    local test_name=$1
    local period_type=$2
    local days=$3
    local rounds=${4:-5}  # 默认测试5次
    
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}测试: $test_name${NC}"
    echo -e "${CYAN}参数: period_type=$period_type, days=$days${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    local total_time=0
    local min_time=999999
    local max_time=0
    
    for i in $(seq 1 $rounds); do
        echo -ne "${BLUE}  第 $i/$rounds 轮...${NC} "
        
        # 执行请求并记录时间
        local start=$(date +%s%N)
        response=$(curl -s -w "\n%{time_total}" \
            "${BASE_URL}/scene-ranking?period_type=$period_type&days=$days&format=json" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" \
            2>/dev/null)
        local end=$(date +%s%N)
        
        # 计算耗时（毫秒）
        local elapsed=$(( (end - start) / 1000000 ))
        total_time=$((total_time + elapsed))
        
        # 更新最小和最大时间
        if [ $elapsed -lt $min_time ]; then
            min_time=$elapsed
        fi
        if [ $elapsed -gt $max_time ]; then
            max_time=$elapsed
        fi
        
        # 检查是否成功
        if echo "$response" | grep -q '"code":200'; then
            echo -e "${GREEN}✓ ${elapsed}ms${NC}"
        else
            echo -e "${RED}✗ ${elapsed}ms (错误)${NC}"
        fi
        
        # 避免请求过快
        sleep 0.5
    done
    
    # 计算平均时间
    local avg_time=$((total_time / rounds))
    
    echo ""
    echo -e "${GREEN}  结果统计:${NC}"
    echo -e "    最小时间: ${CYAN}${min_time}ms${NC}"
    echo -e "    最大时间: ${CYAN}${max_time}ms${NC}"
    echo -e "    平均时间: ${YELLOW}${avg_time}ms${NC}"
    echo ""
    
    # 评估性能
    if [ $avg_time -lt 200 ]; then
        echo -e "${GREEN}  ⭐ 性能评级: 优秀 (< 200ms)${NC}"
    elif [ $avg_time -lt 500 ]; then
        echo -e "${YELLOW}  ⭐ 性能评级: 良好 (200-500ms)${NC}"
    elif [ $avg_time -lt 1000 ]; then
        echo -e "${YELLOW}  ⚠️  性能评级: 一般 (500-1000ms)${NC}"
    else
        echo -e "${RED}  ❌ 性能评级: 需优化 (> 1000ms)${NC}"
    fi
    echo ""
}

# ============================================
# 基准测试套件
# ============================================

echo -e "${BLUE}开始基准测试...${NC}"
echo ""

# 测试1: 每日排名 - 7天
run_benchmark "每日排名 (7天)" "daily" 7 5

# 测试2: 每日排名 - 30天
run_benchmark "每日排名 (30天)" "daily" 30 5

# 测试3: 每周排名 - 4周
run_benchmark "每周排名 (4周)" "weekly" 28 5

# 测试4: 每周排名 - 12周
run_benchmark "每周排名 (12周)" "weekly" 84 5

# 测试5: 每月排名 - 3个月
run_benchmark "每月排名 (3个月)" "monthly" 90 5

# 测试6: 每月排名 - 6个月
run_benchmark "每月排名 (6个月)" "monthly" 180 5

# ============================================
# 并发测试（可选）
# ============================================

if command -v ab &> /dev/null; then
    echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║     并发性能测试 (Apache Bench)                   ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # 测试并发性能
    echo -e "${YELLOW}测试: 10并发，50次请求 (每日7天)${NC}"
    ab -n 50 -c 10 \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        "${BASE_URL}/scene-ranking?period_type=daily&days=7&format=json" \
        2>/dev/null | grep -E "Requests per second|Time per request|Failed requests"
    
    echo ""
    
    echo -e "${YELLOW}测试: 20并发，100次请求 (每日7天)${NC}"
    ab -n 100 -c 20 \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        "${BASE_URL}/scene-ranking?period_type=daily&days=7&format=json" \
        2>/dev/null | grep -E "Requests per second|Time per request|Failed requests"
    
    echo ""
else
    echo -e "${YELLOW}提示: 安装Apache Bench (ab)可进行并发测试${NC}"
    echo -e "  macOS: brew install apr-util"
    echo -e "  Linux: sudo apt-get install apache2-utils"
    echo ""
fi

# ============================================
# 总结
# ============================================

echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     测试完成                                       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}性能基准测试已完成！${NC}"
echo ""
echo -e "${CYAN}性能标准参考:${NC}"
echo -e "  ⭐ 优秀: < 200ms"
echo -e "  ⭐ 良好: 200-500ms"
echo -e "  ⚠️  一般: 500-1000ms"
echo -e "  ❌ 需优化: > 1000ms"
echo ""
echo -e "${YELLOW}提示:${NC}"
echo -e "  1. 如果性能不达标，请执行索引优化脚本:"
echo -e "     ${CYAN}mysql < scripts/optimize_scene_ranking_indexes.sql${NC}"
echo ""
echo -e "  2. 查看详细优化文档:"
echo -e "     ${CYAN}docs/SCENE_RANKING_PERFORMANCE.md${NC}"
echo ""
echo -e "  3. 查看优化总结:"
echo -e "     ${CYAN}SCENE_RANKING_OPTIMIZATION.md${NC}"
echo ""

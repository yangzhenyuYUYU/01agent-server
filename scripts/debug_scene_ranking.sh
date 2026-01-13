#!/bin/bash

# 场景排名调试脚本
# 用于检查为什么没有显示文章场景类型

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     场景排名调试工具                              ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
echo ""

# 获取数据库配置
DB_USER=${1:-"root"}
DB_NAME=${2:-"your_database"}

if [ -z "$2" ]; then
    echo -e "${YELLOW}使用方法: $0 <数据库用户名> <数据库名>${NC}"
    echo -e "${YELLOW}示例: $0 root 01agent_db${NC}"
    echo ""
    read -p "数据库用户名 [root]: " DB_USER
    DB_USER=${DB_USER:-root}
    read -p "数据库名: " DB_NAME
    echo ""
fi

echo -e "${CYAN}连接数据库: ${DB_USER}@localhost/${DB_NAME}${NC}"
echo ""

# 1. 检查文章表总数据
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}1. 检查文章表总数据量${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "SELECT COUNT(*) as total_articles FROM article_edit_tasks;"
echo ""

# 2. 检查最近7天的文章数据
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}2. 检查最近7天的文章数据${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "
SELECT 
    COUNT(*) as count_7days,
    MIN(created_at) as earliest,
    MAX(created_at) as latest
FROM article_edit_tasks
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY);
"
echo ""

# 3. 查看文章scene_type分布
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}3. 文章scene_type分布（全部数据）${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "
SELECT 
    IFNULL(scene_type, 'NULL') as scene_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as users,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM article_edit_tasks), 2) as percentage
FROM article_edit_tasks
GROUP BY scene_type
ORDER BY count DESC;
"
echo ""

# 4. 查看最近7天的scene_type分布
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}4. 最近7天文章scene_type分布${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "
SELECT 
    IFNULL(scene_type, 'NULL') as scene_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as users
FROM article_edit_tasks
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY scene_type
ORDER BY count DESC;
"
echo ""

# 5. 查看短文项目数据（对比）
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}5. 短文项目最近7天数据（对比）${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "
SELECT 
    project_type as scene_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as users
FROM short_post_projects
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY project_type
ORDER BY count DESC;
"
echo ""

# 6. 合并查询测试（模拟API）
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}6. 合并查询测试（模拟API查询）${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "
SELECT 
    scene_type,
    SUM(count) as total_count,
    GROUP_CONCAT(DISTINCT source) as sources
FROM (
    SELECT 
        project_type as scene_type,
        COUNT(*) as count,
        'short_post' as source
    FROM short_post_projects
    WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
    GROUP BY project_type
    
    UNION ALL
    
    SELECT 
        IFNULL(scene_type, 'other') as scene_type,
        COUNT(*) as count,
        'article' as source
    FROM article_edit_tasks
    WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
    GROUP BY scene_type
) combined
GROUP BY scene_type
ORDER BY total_count DESC;
"
echo ""

# 7. 最近的文章记录
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}7. 最近5条文章记录${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
mysql -u "$DB_USER" -p "$DB_NAME" -e "
SELECT 
    id,
    LEFT(user_id, 10) as user_id,
    IFNULL(scene_type, 'NULL') as scene_type,
    LEFT(title, 30) as title,
    status,
    created_at
FROM article_edit_tasks
ORDER BY created_at DESC
LIMIT 5;
"
echo ""

# 总结和建议
echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     诊断总结                                       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${CYAN}根据以上数据，请检查：${NC}"
echo ""
echo -e "${GREEN}1. 文章表是否有数据${NC}"
echo -e "   如果 total_articles = 0，说明没有文章数据"
echo ""
echo -e "${GREEN}2. 最近7天是否有文章创建${NC}"
echo -e "   如果 count_7days = 0，说明最近7天没有文章活动"
echo ""
echo -e "${GREEN}3. scene_type字段是否为NULL${NC}"
echo -e "   如果大部分scene_type是NULL或'other'，则会被归类为'其他'"
echo ""
echo -e "${GREEN}4. 对比短文项目数据${NC}"
echo -e "   短文项目应该有'xiaohongshu'、'poster'等类型"
echo ""
echo -e "${YELLOW}可能的原因：${NC}"
echo -e "  • 文章功能较新，用户还未大量使用"
echo -e "  • 文章创建时scene_type字段未正确设置"
echo -e "  • 用户更倾向于使用短文功能而非长文章"
echo ""
echo -e "${CYAN}解决方案：${NC}"
echo -e "  1. 如果scene_type为NULL，需要在创建文章时设置scene_type"
echo -e "  2. 如果没有数据，可以创建测试文章验证功能"
echo -e "  3. 检查文章创建流程，确保scene_type被正确保存"
echo ""

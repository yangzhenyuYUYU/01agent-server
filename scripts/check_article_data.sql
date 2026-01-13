-- 检查文章数据脚本

-- 1. 检查文章表是否有数据
SELECT '=== 文章表总数据量 ===' as info;
SELECT COUNT(*) as total_count FROM article_edit_tasks;

-- 2. 检查最近7天的文章数据
SELECT '=== 最近7天的文章数据 ===' as info;
SELECT 
    COUNT(*) as count_7days,
    MIN(created_at) as earliest,
    MAX(created_at) as latest
FROM article_edit_tasks
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY);

-- 3. 查看文章的scene_type分布
SELECT '=== 文章scene_type分布 ===' as info;
SELECT 
    scene_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as users
FROM article_edit_tasks
GROUP BY scene_type
ORDER BY count DESC;

-- 4. 查看最近7天的scene_type分布
SELECT '=== 最近7天文章scene_type分布 ===' as info;
SELECT 
    scene_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as users,
    MIN(created_at) as earliest,
    MAX(created_at) as latest
FROM article_edit_tasks
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY scene_type
ORDER BY count DESC;

-- 5. 检查短文项目的数据（对比）
SELECT '=== 短文项目最近7天数据 ===' as info;
SELECT 
    project_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as users
FROM short_post_projects
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY project_type
ORDER BY count DESC;

-- 6. 查看NULL的scene_type数量
SELECT '=== NULL scene_type数量 ===' as info;
SELECT 
    COUNT(*) as null_count,
    COUNT(*) * 100.0 / (SELECT COUNT(*) FROM article_edit_tasks) as null_percentage
FROM article_edit_tasks
WHERE scene_type IS NULL OR scene_type = '';

-- 7. 查看最近的10条文章记录（包含scene_type）
SELECT '=== 最近10条文章记录 ===' as info;
SELECT 
    id,
    user_id,
    scene_type,
    title,
    status,
    created_at
FROM article_edit_tasks
ORDER BY created_at DESC
LIMIT 10;

-- 8. 合并查询（模拟API的查询逻辑）
SELECT '=== 合并查询测试（最近7天）===' as info;
SELECT 
    scene_type,
    COUNT(*) as count,
    'short_post' as source
FROM short_post_projects
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY scene_type

UNION ALL

SELECT 
    IFNULL(scene_type, 'other') as scene_type,
    COUNT(*) as count,
    'article' as source
FROM article_edit_tasks
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY scene_type
ORDER BY count DESC;

-- 场景排名API性能优化 - 数据库索引
-- 用于提升场景排名查询性能

-- ============================================
-- 1. 短文项目表 (short_post_projects)
-- ============================================

-- 检查现有索引
SHOW INDEX FROM short_post_projects;

-- 创建复合索引：created_at + project_type (用于时间范围和场景类型过滤)
CREATE INDEX IF NOT EXISTS idx_spp_created_at_project_type 
ON short_post_projects(created_at, project_type);

-- 创建复合索引：created_at + user_id (用于时间范围和用户关联)
CREATE INDEX IF NOT EXISTS idx_spp_created_at_user_id 
ON short_post_projects(created_at, user_id);

-- 创建单独索引：user_id (用于JOIN user表)
CREATE INDEX IF NOT EXISTS idx_spp_user_id 
ON short_post_projects(user_id);

-- ============================================
-- 2. 文章编辑任务表 (article_edit_tasks)
-- ============================================

-- 检查现有索引
SHOW INDEX FROM article_edit_tasks;

-- 创建复合索引：created_at + scene_type
CREATE INDEX IF NOT EXISTS idx_aet_created_at_scene_type 
ON article_edit_tasks(created_at, scene_type);

-- 创建复合索引：created_at + user_id
CREATE INDEX IF NOT EXISTS idx_aet_created_at_user_id 
ON article_edit_tasks(created_at, user_id);

-- 创建单独索引：user_id (用于JOIN user表)
CREATE INDEX IF NOT EXISTS idx_aet_user_id 
ON article_edit_tasks(user_id);

-- ============================================
-- 3. 用户表 (user)
-- ============================================

-- 检查现有索引
SHOW INDEX FROM user;

-- 创建索引：vip_level (用于区分付费/免费用户)
CREATE INDEX IF NOT EXISTS idx_user_vip_level 
ON user(vip_level);

-- 创建复合索引：user_id + vip_level (用于JOIN时同时过滤)
CREATE INDEX IF NOT EXISTS idx_user_id_vip_level 
ON user(user_id, vip_level);

-- ============================================
-- 查询索引创建结果
-- ============================================

-- 查看所有创建的索引
SELECT 
    table_name,
    index_name,
    column_name,
    seq_in_index,
    index_type
FROM information_schema.statistics
WHERE table_schema = DATABASE()
AND table_name IN ('short_post_projects', 'article_edit_tasks', 'user')
AND index_name LIKE 'idx_%'
ORDER BY table_name, index_name, seq_in_index;

-- ============================================
-- 索引大小统计
-- ============================================

SELECT 
    table_name,
    index_name,
    ROUND(stat_value * @@innodb_page_size / 1024 / 1024, 2) AS size_mb
FROM mysql.innodb_index_stats
WHERE database_name = DATABASE()
AND table_name IN ('short_post_projects', 'article_edit_tasks', 'user')
AND stat_name = 'size'
ORDER BY size_mb DESC;

-- ============================================
-- 测试查询性能
-- ============================================

-- 测试1: 查询最近30天的短文项目
EXPLAIN SELECT 
    spp.created_at,
    spp.project_type as scene_type,
    spp.user_id,
    IFNULL(u.vip_level, 0) as vip_level
FROM short_post_projects spp
LEFT JOIN user u ON spp.user_id = u.user_id
WHERE spp.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW();

-- 测试2: 查询最近30天的文章任务
EXPLAIN SELECT 
    aet.created_at,
    IFNULL(aet.scene_type, 'other') as scene_type,
    aet.user_id,
    IFNULL(u.vip_level, 0) as vip_level
FROM article_edit_tasks aet
LEFT JOIN user u ON aet.user_id = u.user_id
WHERE aet.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW();

-- 测试3: 完整的UNION ALL查询
EXPLAIN SELECT 
    created_at,
    scene_type,
    user_id,
    vip_level
FROM (
    -- 短文项目
    SELECT 
        spp.created_at,
        spp.project_type as scene_type,
        spp.user_id,
        IFNULL(u.vip_level, 0) as vip_level
    FROM short_post_projects spp
    LEFT JOIN user u ON spp.user_id = u.user_id
    WHERE spp.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    
    UNION ALL
    
    -- 长文章
    SELECT 
        aet.created_at,
        IFNULL(aet.scene_type, 'other') as scene_type,
        aet.user_id,
        IFNULL(u.vip_level, 0) as vip_level
    FROM article_edit_tasks aet
    LEFT JOIN user u ON aet.user_id = u.user_id
    WHERE aet.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
) combined
ORDER BY created_at;

-- ============================================
-- 性能对比查询
-- ============================================

-- 查询执行时间对比（运行优化前后）
-- 记录优化前的执行时间
SET @before_time = NOW(6);

-- 执行查询
SELECT COUNT(*) FROM (
    SELECT 
        spp.created_at,
        spp.project_type as scene_type,
        spp.user_id,
        IFNULL(u.vip_level, 0) as vip_level
    FROM short_post_projects spp
    LEFT JOIN user u ON spp.user_id = u.user_id
    WHERE spp.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    
    UNION ALL
    
    SELECT 
        aet.created_at,
        IFNULL(aet.scene_type, 'other') as scene_type,
        aet.user_id,
        IFNULL(u.vip_level, 0) as vip_level
    FROM article_edit_tasks aet
    LEFT JOIN user u ON aet.user_id = u.user_id
    WHERE aet.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
) combined;

-- 记录优化后的执行时间
SET @after_time = NOW(6);

-- 计算时间差
SELECT 
    TIMESTAMPDIFF(MICROSECOND, @before_time, @after_time) / 1000 AS execution_time_ms;

-- ============================================
-- 索引维护建议
-- ============================================

-- 1. 定期分析表统计信息
ANALYZE TABLE short_post_projects;
ANALYZE TABLE article_edit_tasks;
ANALYZE TABLE user;

-- 2. 定期优化表（整理碎片）
-- 注意：OPTIMIZE TABLE会锁表，建议在低峰期执行
-- OPTIMIZE TABLE short_post_projects;
-- OPTIMIZE TABLE article_edit_tasks;
-- OPTIMIZE TABLE user;

-- 3. 查看表的碎片率
SELECT 
    table_name,
    ROUND(data_length / 1024 / 1024, 2) AS data_mb,
    ROUND(index_length / 1024 / 1024, 2) AS index_mb,
    ROUND(data_free / 1024 / 1024, 2) AS free_mb,
    ROUND(data_free / (data_length + index_length) * 100, 2) AS fragmentation_pct
FROM information_schema.tables
WHERE table_schema = DATABASE()
AND table_name IN ('short_post_projects', 'article_edit_tasks', 'user');

-- ============================================
-- 清理说明
-- ============================================

/*
如果需要删除这些索引，执行以下命令：

-- 删除短文项目表索引
DROP INDEX idx_spp_created_at_project_type ON short_post_projects;
DROP INDEX idx_spp_created_at_user_id ON short_post_projects;
DROP INDEX idx_spp_user_id ON short_post_projects;

-- 删除文章任务表索引
DROP INDEX idx_aet_created_at_scene_type ON article_edit_tasks;
DROP INDEX idx_aet_created_at_user_id ON article_edit_tasks;
DROP INDEX idx_aet_user_id ON article_edit_tasks;

-- 删除用户表索引
DROP INDEX idx_user_vip_level ON user;
DROP INDEX idx_user_id_vip_level ON user;
*/

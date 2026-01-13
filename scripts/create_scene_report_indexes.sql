-- ====================================
-- 场景使用报告性能优化索引
-- ====================================
-- 用途：提升场景使用分析报告的查询性能
-- 预期提升：查询速度提升 50-100%
-- ====================================

-- 检查当前索引
SELECT 
    TABLE_NAME,
    INDEX_NAME,
    COLUMN_NAME,
    SEQ_IN_INDEX
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('short_post_projects', 'user_productions', 'user', 'short_post_export_records')
ORDER BY TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX;

-- ====================================
-- 1. short_post_projects 表索引
-- ====================================

-- 基础索引：按创建时间查询
CREATE INDEX IF NOT EXISTS idx_spp_created_at 
ON short_post_projects(created_at);

-- 基础索引：按用户ID查询
CREATE INDEX IF NOT EXISTS idx_spp_user_id 
ON short_post_projects(user_id);

-- 基础索引：按场景类型查询
CREATE INDEX IF NOT EXISTS idx_spp_project_type 
ON short_post_projects(project_type);

-- 基础索引：按状态查询
CREATE INDEX IF NOT EXISTS idx_spp_status 
ON short_post_projects(status);

-- 复合索引：覆盖场景报告的主要查询
-- 这个索引可以覆盖大部分场景报告查询，大幅提升性能
CREATE INDEX IF NOT EXISTS idx_spp_report_covering 
ON short_post_projects(created_at, project_type, user_id, status, id);

-- ====================================
-- 2. user_productions 表索引
-- ====================================

-- 复合索引：用户ID + 状态（用于JOIN）
CREATE INDEX IF NOT EXISTS idx_up_user_status 
ON user_productions(user_id, status);

-- 复合索引：状态 + 用户ID + 产品ID（覆盖索引）
CREATE INDEX IF NOT EXISTS idx_up_status_covering 
ON user_productions(status, user_id, production_id);

-- ====================================
-- 3. user 表索引
-- ====================================

-- 基础索引：VIP等级（用于区分付费/免费用户）
CREATE INDEX IF NOT EXISTS idx_user_vip_level 
ON user(vip_level);

-- ====================================
-- 4. short_post_export_records 表索引
-- ====================================

-- 复合索引：创建时间 + 用户ID
CREATE INDEX IF NOT EXISTS idx_export_created_user 
ON short_post_export_records(created_at, user_id);

-- 基础索引：导出格式
CREATE INDEX IF NOT EXISTS idx_export_format 
ON short_post_export_records(export_format);

-- ====================================
-- 5. AI功能表索引
-- ====================================

-- ai_format_records
CREATE INDEX IF NOT EXISTS idx_ai_format_created_user 
ON ai_format_records(created_at, user_id, status);

-- ai_rewrite_records
CREATE INDEX IF NOT EXISTS idx_ai_rewrite_created_user 
ON ai_rewrite_records(created_at, user_id, status);

-- ai_topic_polish_records
CREATE INDEX IF NOT EXISTS idx_ai_polish_created_user 
ON ai_topic_polish_records(created_at, user_id, status);

-- ====================================
-- 验证索引创建结果
-- ====================================

SELECT 
    TABLE_NAME as '表名',
    INDEX_NAME as '索引名',
    GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) as '索引列',
    INDEX_TYPE as '索引类型',
    NON_UNIQUE as '是否唯一'
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('short_post_projects', 'user_productions', 'user', 'short_post_export_records')
  AND INDEX_NAME LIKE 'idx_%'
GROUP BY TABLE_NAME, INDEX_NAME, INDEX_TYPE, NON_UNIQUE
ORDER BY TABLE_NAME, INDEX_NAME;

-- ====================================
-- 性能测试查询
-- ====================================

-- 测试1：场景统计查询（应该使用 idx_spp_report_covering）
EXPLAIN SELECT 
    project_type,
    COUNT(*) as usage_count,
    COUNT(DISTINCT user_id) as user_count
FROM short_post_projects
WHERE created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
GROUP BY project_type;

-- 测试2：用户类型对比查询（应该使用多个索引）
EXPLAIN SELECT 
    CASE WHEN IFNULL(u.vip_level, 0) = 0 THEN '免费用户' ELSE '付费用户' END as user_type,
    spp.project_type,
    COUNT(spp.id) as usage_count
FROM short_post_projects spp
INNER JOIN user u ON spp.user_id = u.user_id
WHERE spp.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
GROUP BY user_type, project_type;

-- 测试3：产品对比查询（最复杂的查询）
EXPLAIN SELECT 
    COALESCE(p.name, '免费用户') as product_name,
    spp.project_type,
    COUNT(spp.id) as usage_count
FROM short_post_projects spp
LEFT JOIN user u ON spp.user_id = u.user_id
LEFT JOIN user_productions up ON u.user_id = up.user_id AND up.status = 'active'
LEFT JOIN productions p ON up.production_id = p.id
WHERE spp.created_at BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
GROUP BY product_name, project_type;

-- ====================================
-- 索引大小统计
-- ====================================

SELECT 
    TABLE_NAME as '表名',
    INDEX_NAME as '索引名',
    ROUND(STAT_VALUE * @@innodb_page_size / 1024 / 1024, 2) as '索引大小(MB)'
FROM mysql.innodb_index_stats
WHERE DATABASE_NAME = DATABASE()
  AND TABLE_NAME IN ('short_post_projects', 'user_productions', 'user', 'short_post_export_records')
  AND STAT_NAME = 'size'
ORDER BY STAT_VALUE DESC;

-- ====================================
-- 维护建议
-- ====================================

-- 定期分析表（更新统计信息）
-- 建议每周执行一次
ANALYZE TABLE short_post_projects;
ANALYZE TABLE user_productions;
ANALYZE TABLE user;
ANALYZE TABLE short_post_export_records;

-- 定期优化表（重建索引）
-- 建议每月执行一次
-- 注意：OPTIMIZE TABLE 会锁表，建议在业务低峰期执行
-- OPTIMIZE TABLE short_post_projects;
-- OPTIMIZE TABLE user_productions;

-- ====================================
-- 回滚脚本（如果需要删除索引）
-- ====================================

/*
-- 删除创建的索引（谨慎使用）

DROP INDEX idx_spp_created_at ON short_post_projects;
DROP INDEX idx_spp_user_id ON short_post_projects;
DROP INDEX idx_spp_project_type ON short_post_projects;
DROP INDEX idx_spp_status ON short_post_projects;
DROP INDEX idx_spp_report_covering ON short_post_projects;

DROP INDEX idx_up_user_status ON user_productions;
DROP INDEX idx_up_status_covering ON user_productions;

DROP INDEX idx_user_vip_level ON user;

DROP INDEX idx_export_created_user ON short_post_export_records;
DROP INDEX idx_export_format ON short_post_export_records;

DROP INDEX idx_ai_format_created_user ON ai_format_records;
DROP INDEX idx_ai_rewrite_created_user ON ai_rewrite_records;
DROP INDEX idx_ai_polish_created_user ON ai_topic_polish_records;
*/

-- ====================================
-- 完成提示
-- ====================================

SELECT '索引创建完成！建议执行 ANALYZE TABLE 更新统计信息' as '提示';

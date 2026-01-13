-- 为 article_edit_tasks 表添加 scene_type 字段
-- 用于区分不同平台的文章场景

-- 1. 添加 scene_type 字段
ALTER TABLE article_edit_tasks 
ADD COLUMN scene_type VARCHAR(20) DEFAULT 'other' COMMENT '场景类型' AFTER theme;

-- 2. 添加索引以提升查询性能
ALTER TABLE article_edit_tasks 
ADD INDEX idx_scene_type (scene_type);

-- 3. 根据 theme 字段推断并更新现有数据的 scene_type
-- 微信公众号相关主题
UPDATE article_edit_tasks 
SET scene_type = 'weixin' 
WHERE theme LIKE '%weixin%' 
   OR theme LIKE '%wechat%' 
   OR theme LIKE '%公众号%'
   OR theme = 'official';

-- 知乎相关主题
UPDATE article_edit_tasks 
SET scene_type = 'zhihu' 
WHERE theme LIKE '%zhihu%' 
   OR theme LIKE '%知乎%';

-- 今日头条相关主题
UPDATE article_edit_tasks 
SET scene_type = 'toutiao' 
WHERE theme LIKE '%toutiao%' 
   OR theme LIKE '%头条%';

-- 简书相关主题
UPDATE article_edit_tasks 
SET scene_type = 'jianshu' 
WHERE theme LIKE '%jianshu%' 
   OR theme LIKE '%简书%';

-- CSDN相关主题
UPDATE article_edit_tasks 
SET scene_type = 'csdn' 
WHERE theme LIKE '%csdn%';

-- 掘金相关主题
UPDATE article_edit_tasks 
SET scene_type = 'juejin' 
WHERE theme LIKE '%juejin%' 
   OR theme LIKE '%掘金%';

-- 个人博客相关主题
UPDATE article_edit_tasks 
SET scene_type = 'blog' 
WHERE theme LIKE '%blog%' 
   OR theme LIKE '%博客%'
   OR theme = 'minimal'
   OR theme = 'elegant';

-- 4. 验证数据更新结果
SELECT 
    scene_type,
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as user_count
FROM article_edit_tasks
GROUP BY scene_type
ORDER BY count DESC;

-- 5. 显示各主题的分布情况（用于后续优化）
SELECT 
    theme,
    scene_type,
    COUNT(*) as count
FROM article_edit_tasks
GROUP BY theme, scene_type
ORDER BY count DESC
LIMIT 20;

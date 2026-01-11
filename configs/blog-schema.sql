-- 博客功能表结构
-- 创建时间: 2026-01-11

-- 如果表已存在，先删除
DROP TABLE IF EXISTS blog_seo_keywords;
DROP TABLE IF EXISTS blog_post_tags;
DROP TABLE IF EXISTS blog_tags;
DROP TABLE IF EXISTS blog_posts;

-- 博客文章表
CREATE TABLE blog_posts (
    id VARCHAR(36) PRIMARY KEY,
    slug VARCHAR(255) UNIQUE NOT NULL COMMENT 'URL友好的标识符',
    title VARCHAR(500) NOT NULL COMMENT '文章标题',
    summary TEXT NOT NULL COMMENT '文章摘要',
    content LONGTEXT NOT NULL COMMENT 'Markdown格式的正文',
    category VARCHAR(50) NOT NULL COMMENT '分类: tutorials, tips-and-tricks等',
    cover_image VARCHAR(500) COMMENT '封面图URL',
    author VARCHAR(100) DEFAULT '01Agent Team' COMMENT '作者',
    author_avatar VARCHAR(500) COMMENT '作者头像URL',
    publish_date DATETIME NOT NULL COMMENT '发布时间',
    updated_date DATETIME COMMENT '更新时间',
    read_time INT COMMENT '阅读时间（分钟）',
    views INT DEFAULT 0 COMMENT '浏览量',
    likes INT DEFAULT 0 COMMENT '点赞数',
    is_featured BOOLEAN DEFAULT FALSE COMMENT '是否精选',
    seo_description VARCHAR(500) COMMENT 'SEO描述',
    status VARCHAR(20) DEFAULT 'published' COMMENT '状态: draft/published/archived',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_slug (slug),
    INDEX idx_publish_date (publish_date),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='博客文章表';

-- 标签表
CREATE TABLE blog_tags (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL COMMENT '标签名',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='博客标签表';

-- 文章标签关联表
CREATE TABLE blog_post_tags (
    post_id VARCHAR(36) NOT NULL,
    tag_id INT NOT NULL,
    PRIMARY KEY (post_id, tag_id),
    FOREIGN KEY (post_id) REFERENCES blog_posts(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES blog_tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章标签关联表';

-- SEO关键词表（可选）
CREATE TABLE blog_seo_keywords (
    id INT AUTO_INCREMENT PRIMARY KEY,
    post_id VARCHAR(36) NOT NULL,
    keyword VARCHAR(100) NOT NULL,
    FOREIGN KEY (post_id) REFERENCES blog_posts(id) ON DELETE CASCADE,
    INDEX idx_post_id (post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SEO关键词表';

-- 插入测试数据
INSERT INTO blog_posts (id, slug, title, summary, content, category, cover_image, author, author_avatar, publish_date, read_time, views, likes, is_featured, seo_description, status) 
VALUES 
(
    'test-blog-001',
    'getting-started-with-01agent',
    '01Agent 快速入门指南',
    '这是一篇介绍如何快速开始使用 01Agent 的教程文章，帮助你在5分钟内完成基础配置。',
    '# 01Agent 快速入门\n\n## 简介\n\n01Agent 是一个强大的 AI 代理平台...\n\n## 安装\n\n```bash\nnpm install @01agent/core\n```\n\n## 配置\n\n...',
    'tutorials',
    'https://example.com/covers/getting-started.jpg',
    '01Agent Team',
    'https://example.com/avatars/team.jpg',
    NOW(),
    5,
    1250,
    89,
    TRUE,
    '学习如何快速开始使用 01Agent，5分钟内完成基础配置和首次运行',
    'published'
),
(
    'test-blog-002',
    '10-tips-for-ai-content-creation',
    'AI内容创作的10个实用技巧',
    '分享10个使用AI工具进行内容创作的实用技巧，帮助你提升创作效率和内容质量。',
    '# AI内容创作的10个实用技巧\n\n## 1. 明确创作目标\n\n在使用AI工具前，先明确你的创作目标...\n\n## 2. 优化提示词\n\n...',
    'tips-and-tricks',
    'https://example.com/covers/ai-tips.jpg',
    '张三',
    'https://example.com/avatars/zhangsan.jpg',
    DATE_SUB(NOW(), INTERVAL 3 DAY),
    8,
    2340,
    156,
    TRUE,
    '掌握AI内容创作的10个实用技巧，提升创作效率和内容质量',
    'published'
),
(
    'test-blog-003',
    'product-update-2026-01',
    '产品更新：2026年1月新功能发布',
    '本月我们发布了多项重要功能更新，包括智能写作助手、多语言支持等。',
    '# 2026年1月产品更新\n\n## 新增功能\n\n### 1. 智能写作助手\n\n我们很高兴地宣布...\n\n### 2. 多语言支持\n\n...',
    'product-updates',
    'https://example.com/covers/update-jan.jpg',
    '产品团队',
    'https://example.com/avatars/product-team.jpg',
    DATE_SUB(NOW(), INTERVAL 1 DAY),
    6,
    890,
    67,
    FALSE,
    '查看01Agent 2026年1月的最新产品更新和功能发布',
    'published'
);

-- 插入测试标签
INSERT INTO blog_tags (name) VALUES 
('AI写作'),
('教程'),
('技巧'),
('产品更新'),
('入门指南');

-- 关联标签
INSERT INTO blog_post_tags (post_id, tag_id) VALUES
('test-blog-001', 2), -- 教程
('test-blog-001', 5), -- 入门指南
('test-blog-002', 1), -- AI写作
('test-blog-002', 3), -- 技巧
('test-blog-003', 4); -- 产品更新

-- 插入SEO关键词
INSERT INTO blog_seo_keywords (post_id, keyword) VALUES
('test-blog-001', '01Agent'),
('test-blog-001', '快速入门'),
('test-blog-001', 'AI代理'),
('test-blog-002', 'AI内容创作'),
('test-blog-002', '写作技巧'),
('test-blog-003', '产品更新'),
('test-blog-003', '新功能');


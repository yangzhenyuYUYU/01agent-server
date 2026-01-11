-- 博客系统完整初始化脚本
-- 包含表结构 + 测试数据
-- 使用方法：mysql -u root -p 01agent < configs/blog-init.sql

-- ============================================
-- 第一部分：创建表结构
-- ============================================

-- 临时禁用外键检查
SET FOREIGN_KEY_CHECKS = 0;

-- 如果表已存在，先删除
DROP TABLE IF EXISTS blog_seo_keywords;
DROP TABLE IF EXISTS blog_post_tags;
DROP TABLE IF EXISTS blog_tags;
DROP TABLE IF EXISTS blog_posts;

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

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

-- SEO关键词表
CREATE TABLE blog_seo_keywords (
    id INT AUTO_INCREMENT PRIMARY KEY,
    post_id VARCHAR(36) NOT NULL,
    keyword VARCHAR(100) NOT NULL,
    FOREIGN KEY (post_id) REFERENCES blog_posts(id) ON DELETE CASCADE,
    INDEX idx_post_id (post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SEO关键词表';

-- ============================================
-- 第二部分：插入测试数据
-- ============================================

-- 插入测试文章
INSERT INTO blog_posts (id, slug, title, summary, content, category, cover_image, author, publish_date, read_time, views, is_featured, seo_description, status) VALUES

-- 文章1：快速入门
(
  '550e8400-e29b-41d4-a716-446655440001',
  'getting-started-with-01agent',
  '01Agent 快速入门指南：5分钟开启AI创作之旅',
  '本指南将带你快速了解01Agent的核心功能，从注册账号到发布第一篇AI生成的内容，只需5分钟。',
  '# 01Agent 快速入门指南

## 什么是01Agent？

01Agent是一款强大的AI内容创作平台，帮助创作者高效产出高质量内容。

### 核心功能

1. **AI文章创作** - 一键生成完整文章
2. **小红书笔记** - 专业图文排版
3. **公众号排版** - 智能美化样式
4. **多平台适配** - 一次创作，多平台发布

## 快速开始

### 步骤1：注册账号

访问 https://01agent.net 注册账号：
- 支持邮箱注册
- 支持手机号注册
- 支持第三方登录（微信、QQ）

### 步骤2：选择创作类型

进入工作台，选择你需要的创作类型：
- 📝 长文章创作
- 📱 小红书笔记
- 💬 公众号图文
- 🎬 短视频脚本

### 步骤3：输入关键词

输入你的主题关键词：
```
示例：如何提升小红书笔记阅读量
```

### 步骤4：AI生成内容

点击"生成"按钮：
- AI会自动生成标题
- 生成结构化大纲
- 完成正文撰写
- 配备合适的图片

### 步骤5：编辑优化

根据需要调整：
- 修改标题
- 调整段落
- 更换图片
- 优化排版

### 步骤6：一键发布

完成后可以：
- 直接复制到目标平台
- 导出为图片
- 保存为草稿

## 进阶技巧

### 提示词优化

好的提示词能显著提升内容质量：

**❌ 不好的提示词：**
```
写一篇文章
```

**✅ 好的提示词：**
```
写一篇关于小红书运营的干货文章，
目标受众是新手创作者，
包含实用技巧和案例分析，
字数2000字左右
```

### 批量创作

使用批量创作功能：
1. 准备多个主题关键词
2. 批量导入
3. 一键生成
4. 节省80%时间

## 常见问题

### Q: AI生成的内容原创吗？
A: 是的，每次生成的内容都是独特的。

### Q: 支持哪些平台？
A: 支持小红书、公众号、知乎、抖音等主流平台。

### Q: 有免费额度吗？
A: 新用户注册即送10次免费生成次数。

## 下一步

- 查看[进阶教程](#)
- 加入[用户社群](#)
- 观看[视频教程](#)

---

💡 **立即开始**：https://01agent.net',
  'tutorials',
  'https://images.unsplash.com/photo-1499750310107-5fef28a66643?w=800',
  '01Agent Team',
  '2026-01-10 10:00:00',
  5,
  3250,
  1,
  '01Agent快速入门指南，5分钟学会使用AI创作工具，从注册到发布一站式教程。',
  'published'
),

-- 文章2：小红书笔记创作
(
  '550e8400-e29b-41d4-a716-446655440002',
  'xiaohongshu-content-strategy',
  '小红书涨粉攻略：从0到10万的完整路线图',
  '深度解析小红书涨粉的底层逻辑，包含账号定位、内容策略、数据分析的全套方法论。',
  '# 小红书涨粉完整攻略

## 账号定位

### 找准垂直领域

选择你擅长的领域：
- 美妆护肤
- 穿搭时尚
- 美食探店
- 旅行攻略
- 知识干货

### 人设打造

确定你的人设标签：
- 专业型：领域专家
- 生活型：邻家女孩
- 搞笑型：段子手

## 内容策略

### 爆款公式

**标题**
- 数字型：《5个技巧让你...》
- 疑问型：《为什么你的...》
- 干货型：《史上最全...》

**封面**
- 3秒吸引眼球
- 文字清晰可读
- 色彩鲜明对比

**正文**
- 开头痛点引入
- 分点阐述干货
- 结尾引导互动

### 发布节奏

建议发布频率：
- 新手：每天1篇
- 进阶：每天2-3篇
- 高手：根据数据调整

最佳发布时间：
- 早上 7-9点
- 中午 12-14点  
- 晚上 18-22点

## 使用01Agent提效

### 快速生成笔记

1. 输入主题关键词
2. 选择笔记类型
3. AI生成图文
4. 微调后发布

### 批量创作

- 准备10个选题
- 批量生成内容
- 一周内容储备

## 数据分析

### 关键指标

- 完读率：>60%为优秀
- 互动率：>5%为优秀
- 涨粉率：关注/浏览>1%

### 优化方向

根据数据调整：
- 标题优化
- 封面优化
- 内容优化
- 发布时间

## 变现路径

### 1. 品牌合作

粉丝要求：
- 5000+粉丝可接推广
- 1万+粉丝价格200-500元/篇
- 10万+粉丝价格2000-5000元/篇

### 2. 知识付费

- 专栏课程
- 1对1咨询
- 社群运营

### 3. 引流变现

- 导流到私域
- 销售产品/服务

## 避坑指南

### 常见错误

❌ 内容杂乱无章
❌ 更新不稳定
❌ 不看数据盲目发布
❌ 标题党但内容不行

### 正确做法

✅ 垂直深耕
✅ 稳定更新
✅ 数据驱动
✅ 内容为王

## 总结

小红书涨粉的核心：
1. 精准定位
2. 优质内容
3. 稳定更新
4. 数据优化

配合01Agent工具，让涨粉更高效！

---

**点赞收藏，开始你的涨粉之旅！**',
  'tutorials',
  'https://images.unsplash.com/photo-1611162617474-5b21e879e113?w=800',
  '运营达人Amy',
  '2026-01-09 14:30:00',
  10,
  5680,
  1,
  '小红书涨粉完整攻略，从0到10万粉丝的实战方法论，包含账号定位、内容策略、数据分析。',
  'published'
),

-- 文章3：AI创作技巧
(
  '550e8400-e29b-41d4-a716-446655440003',
  'ai-writing-best-practices',
  'AI写作的10个黄金技巧：让AI更懂你的需求',
  '掌握这10个技巧，让AI生成的内容质量提升300%，真正实现高效创作。',
  '# AI写作的10个黄金技巧

## 技巧1：精准的提示词

### 对比示例

**❌ 模糊提示词**
```
写一篇文章
```

**✅ 精准提示词**
```
写一篇关于小红书运营的实战教程，
目标读者是0-1岁新手创作者，
包含账号定位、内容创作、数据分析三个部分，
每部分提供3-5个可落地的方法，
整体字数2000-3000字，
语气轻松易懂，多举实例
```

## 技巧2：结构化输入

使用清晰的结构：
```
主题：XXX
目标读者：XXX
核心要点：
1. XXX
2. XXX
3. XXX
风格：XXX
字数：XXX
```

## 技巧3：迭代优化

不要期望一次生成完美：
1. 第一次：生成初稿
2. 第二次：优化标题和开头
3. 第三次：丰富细节和案例
4. 第四次：润色语言和排版

## 技巧4：参考范文

提供优质范文：
```
请参考以下风格生成：
[粘贴范文]

主题改为：XXX
```

## 技巧5：分段生成

长文章分段生成：
1. 先生成大纲
2. 逐段展开
3. 最后统一润色

## 技巧6：限定格式

明确输出格式：
```
请用以下格式输出：
# 大标题
## 小标题
- 要点1
- 要点2
```

## 技巧7：添加约束

设置明确限制：
- 字数限制
- 段落数量
- 不要出现XXX
- 必须包含XXX

## 技巧8：角色扮演

让AI扮演特定角色：
```
你是一个10年经验的小红书运营专家，
用你的专业知识和经验，
写一篇...
```

## 技巧9：提供背景

给AI足够的背景信息：
```
背景：
- 目标平台：小红书
- 受众特征：18-25岁女性
- 竞品分析：XXX
- 我的优势：XXX

基于以上背景，请生成...
```

## 技巧10：人工审核

AI生成后必须：
- ✅ 检查事实准确性
- ✅ 优化语言表达
- ✅ 添加个人观点
- ✅ 确保内容原创

## 实战案例

### 案例1：生成小红书笔记

**提示词：**
```
主题：夏季防晒攻略
平台：小红书
受众：20-30岁女性
要求：
1. 标题吸引人
2. 开头用痛点引入
3. 推荐5款防晒产品
4. 每款简短说明
5. 配使用技巧
6. 结尾引导互动
字数：500-800字
```

### 案例2：生成公众号文章

**提示词：**
```
主题：职场新人生存指南
受众：刚毕业的大学生
风格：亲切、实用
结构：
1. 开头：职场困惑
2. 正文：5大生存法则
3. 每条配案例说明
4. 结尾：总结升华
字数：2000-3000字
```

## 使用01Agent的优势

### 智能理解

- 自动识别内容类型
- 智能匹配写作风格
- 精准把握受众需求

### 高效产出

- 3分钟生成初稿
- 快速迭代优化
- 一键多平台适配

### 质量保证

- 内容原创性检测
- 语言质量把控
- 结构逻辑优化

## 进阶技巧

### 建立个人提示词库

整理常用提示词模板：
- 小红书笔记模板
- 公众号文章模板
- 知乎回答模板

### 积累优质案例

收集爆款内容：
- 分析成功要素
- 提炼写作模式
- 应用到提示词

## 总结

AI写作的本质是：
- 清晰的需求表达
- 合理的期望管理
- 有效的迭代优化

01Agent让这一切变得更简单！

---

**开始体验：https://01agent.net**',
  'tips-and-tricks',
  'https://images.unsplash.com/photo-1455390582262-044cdead277a?w=800',
  'AI创作导师Mike',
  '2026-01-08 16:00:00',
  12,
  4230,
  1,
  'AI写作10大黄金技巧，提升AI生成内容质量的实战方法，让AI更懂你的创作需求。',
  'published'
);

-- 插入标签
INSERT INTO blog_tags (name) VALUES
('新手教程'), ('小红书'), ('公众号'), ('AI创作'), 
('运营技巧'), ('涨粉攻略'), ('内容营销'), ('工具推荐'),
('实战案例'), ('数据分析'), ('变现方法'), ('写作技巧');

-- 关联文章和标签
INSERT INTO blog_post_tags (post_id, tag_id) VALUES
-- 文章1的标签
('550e8400-e29b-41d4-a716-446655440001', 1), -- 新手教程
('550e8400-e29b-41d4-a716-446655440001', 4), -- AI创作
('550e8400-e29b-41d4-a716-446655440001', 8), -- 工具推荐

-- 文章2的标签
('550e8400-e29b-41d4-a716-446655440002', 2), -- 小红书
('550e8400-e29b-41d4-a716-446655440002', 5), -- 运营技巧
('550e8400-e29b-41d4-a716-446655440002', 6), -- 涨粉攻略
('550e8400-e29b-41d4-a716-446655440002', 10), -- 数据分析

-- 文章3的标签
('550e8400-e29b-41d4-a716-446655440003', 4), -- AI创作
('550e8400-e29b-41d4-a716-446655440003', 12), -- 写作技巧
('550e8400-e29b-41d4-a716-446655440003', 8); -- 工具推荐

-- 插入SEO关键词
INSERT INTO blog_seo_keywords (post_id, keyword) VALUES
-- 文章1
('550e8400-e29b-41d4-a716-446655440001', '01Agent'),
('550e8400-e29b-41d4-a716-446655440001', 'AI创作工具'),
('550e8400-e29b-41d4-a716-446655440001', '快速入门'),
('550e8400-e29b-41d4-a716-446655440001', '新手教程'),

-- 文章2
('550e8400-e29b-41d4-a716-446655440002', '小红书涨粉'),
('550e8400-e29b-41d4-a716-446655440002', '小红书运营'),
('550e8400-e29b-41d4-a716-446655440002', '内容策略'),
('550e8400-e29b-41d4-a716-446655440002', '账号定位'),

-- 文章3
('550e8400-e29b-41d4-a716-446655440003', 'AI写作'),
('550e8400-e29b-41d4-a716-446655440003', '提示词技巧'),
('550e8400-e29b-41d4-a716-446655440003', 'AI创作'),
('550e8400-e29b-41d4-a716-446655440003', '内容生成');

-- 验证数据
SELECT '=== 文章列表 ===' as '';
SELECT 
  id, slug, title, category, views, is_featured, status
FROM blog_posts
ORDER BY publish_date DESC;

SELECT '=== 标签列表 ===' as '';
SELECT * FROM blog_tags;

SELECT '=== 文章标签关联 ===' as '';
SELECT 
  bp.title,
  GROUP_CONCAT(bt.name) as tags
FROM blog_posts bp
LEFT JOIN blog_post_tags bpt ON bp.id = bpt.post_id
LEFT JOIN blog_tags bt ON bpt.tag_id = bt.id
GROUP BY bp.id, bp.title;

SELECT '=== 初始化完成 ===' as '';



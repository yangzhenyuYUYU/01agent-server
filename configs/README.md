# 博客功能 - 数据库脚本说明

## 📁 文件列表

### 1. `blog-init.sql` - 一键初始化脚本 ⭐推荐

**包含内容：**
- 完整的表结构（4张表）
- 精简的测试数据（3篇高质量文章）
- 数据验证查询

**使用场景：**
- 首次安装
- 快速演示
- 开发测试

**执行命令：**
```bash
mysql -u root -p 01agent < configs/blog-init.sql
```

**特点：**
- ✅ 一键完成所有初始化
- ✅ 自动处理外键约束
- ✅ 包含完整的测试数据
- ✅ 显示验证信息

---

### 2. `blog-schema.sql` - 表结构定义

**包含内容：**
- 4张表的CREATE语句
- 索引定义
- 外键约束

**使用场景：**
- 只需要表结构
- 自定义数据导入
- 生产环境初始化

**执行命令：**
```bash
mysql -u root -p 01agent < configs/blog-schema.sql
```

---

### 3. `test-data.sql` - 测试数据

**包含内容：**
- 5篇完整的测试文章
- 多个标签
- 文章标签关联
- SEO关键词

**使用场景：**
- 在已有表结构的基础上添加测试数据
- 批量导入样例内容

**执行命令：**
```bash
# 前提：已经执行过 blog-schema.sql
mysql -u root -p 01agent < configs/test-data.sql
```

**注意事项：**
- ⚠️ 需要先禁用外键检查（脚本中已包含）
- ⚠️ 会清空现有数据（使用TRUNCATE）

---

## 🚀 推荐使用流程

### 新项目初始化

```bash
# 一步完成
mysql -u root -p 01agent < configs/blog-init.sql
```

### 生产环境部署

```bash
# 只创建表结构
mysql -u root -p 01agent < configs/blog-schema.sql
```

### 开发环境重置

```bash
# 重新初始化（包含测试数据）
mysql -u root -p 01agent < configs/blog-init.sql
```

---

## 📊 测试数据说明

### blog-init.sql 包含的文章

1. **01Agent 快速入门指南**
   - 分类：tutorials
   - 精选：是
   - 浏览量：3250

2. **小红书涨粉攻略**
   - 分类：tutorials
   - 精选：是
   - 浏览量：5680

3. **AI写作的10个黄金技巧**
   - 分类：tips-and-tricks
   - 精选：是
   - 浏览量：4230

### test-data.sql 包含的文章

包含5篇文章，涵盖所有分类：
- 小红书笔记创作指南
- 公众号排版教程
- 爆款内容创作技巧
- AI写作趋势分析
- 用户成功案例

---

## 🔧 常见问题

### Q1: 执行脚本报错 "Cannot truncate table"

**原因：** 外键约束导致无法清空表

**解决：**
1. 使用 `blog-init.sql`（已自动处理）
2. 或者手动禁用外键检查：
```sql
SET FOREIGN_KEY_CHECKS = 0;
-- 执行你的操作
SET FOREIGN_KEY_CHECKS = 1;
```

### Q2: 如何重置数据？

```bash
# 方法1：重新执行初始化脚本
mysql -u root -p 01agent < configs/blog-init.sql

# 方法2：手动删除后重建
DROP TABLE IF EXISTS blog_seo_keywords;
DROP TABLE IF EXISTS blog_post_tags;
DROP TABLE IF EXISTS blog_tags;
DROP TABLE IF EXISTS blog_posts;
# 然后执行初始化脚本
```

### Q3: UUID() 函数报错？

**原因：** PostgreSQL使用不同的UUID生成函数

**解决：**
- MySQL: `UUID()`
- PostgreSQL: `gen_random_uuid()` 或使用扩展 `uuid-ossp`

### Q4: 如何导入自己的数据？

```bash
# 1. 创建表结构
mysql -u root -p 01agent < configs/blog-schema.sql

# 2. 准备你的数据SQL文件
# 3. 导入数据
mysql -u root -p 01agent < your-data.sql
```

---

## 📝 脚本执行验证

执行完成后，可以运行以下查询验证：

```sql
-- 查看文章数量
SELECT COUNT(*) as article_count FROM blog_posts;

-- 查看标签数量
SELECT COUNT(*) as tag_count FROM blog_tags;

-- 查看文章列表
SELECT id, slug, title, category, views, is_featured 
FROM blog_posts 
ORDER BY publish_date DESC;

-- 查看文章标签
SELECT 
  bp.title,
  GROUP_CONCAT(bt.name) as tags
FROM blog_posts bp
LEFT JOIN blog_post_tags bpt ON bp.id = bpt.post_id
LEFT JOIN blog_tags bt ON bpt.tag_id = bt.id
GROUP BY bp.id, bp.title;
```

---

## 🎯 最佳实践

1. **开发环境** - 使用 `blog-init.sql`，快速初始化+测试数据
2. **测试环境** - 使用 `blog-init.sql` 或 `test-data.sql`
3. **生产环境** - 只使用 `blog-schema.sql`，数据另行导入
4. **演示环境** - 使用 `blog-init.sql`，展示完整功能

---

## 📌 注意事项

1. **备份数据** - 执行脚本前请备份现有数据
2. **数据库权限** - 确保有CREATE, DROP, INSERT权限
3. **字符集** - 所有表使用utf8mb4字符集
4. **外键约束** - 删除数据时注意外键依赖关系
5. **ID生成** - blog-init.sql使用固定UUID，test-data.sql使用UUID()函数

---

## 🔗 相关文档

- [API文档](../internal/router/BLOG_API.md)
- [快速开始](BLOG_QUICKSTART.md)
- [实现总结](../BLOG_IMPLEMENTATION_SUMMARY.md)



/*
 Navicat Premium Dump SQL

 Source Server         : 01编辑器本地
 Source Server Type    : MySQL
 Source Server Version : 80041 (8.0.41)
 Source Host           : localhost:3306
 Source Schema         : 01editor

 Target Server Type    : MySQL
 Target Server Version : 80041 (8.0.41)
 File Encoding         : 65001

 Date: 26/07/2025 17:15:57
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for activation_codes
-- ----------------------------
DROP TABLE IF EXISTS `activation_codes`;
CREATE TABLE `activation_codes`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '激活码',
  `card_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '类型',
  `product_id` int NOT NULL COMMENT '关联产品ID',
  `is_used` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已使用',
  `remark` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `trade_id` int NULL DEFAULT NULL COMMENT '关联交易',
  `used_by_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '使用用户',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `code`(`code` ASC) USING BTREE,
  INDEX `fk_activati_trades_a641af18`(`trade_id` ASC) USING BTREE,
  INDEX `fk_activati_user_747e4cb5`(`used_by_id` ASC) USING BTREE,
  CONSTRAINT `fk_activati_trades_a641af18` FOREIGN KEY (`trade_id`) REFERENCES `trades` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_activati_user_747e4cb5` FOREIGN KEY (`used_by_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 643 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】激活码表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for ai_format_records
-- ----------------------------
DROP TABLE IF EXISTS `ai_format_records`;
CREATE TABLE `ai_format_records`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '唯一记录ID',
  `original_content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '原始内容',
  `formatted_content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '排版后内容',
  `format_type` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '目标格式类型',
  `status` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '处理状态',
  `tokens` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '消耗tokens',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后更新时间',
  `error_log` json NULL COMMENT '错误日志详情',
  `model_version` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT 'AI模型版本',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_ai_format_r_created_db04fe`(`created_at` ASC) USING BTREE,
  INDEX `idx_ai_format_r_created_d305bf`(`created_at` ASC, `format_type` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 507 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'AI内容排版记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for ai_recommend_topics
-- ----------------------------
DROP TABLE IF EXISTS `ai_recommend_topics`;
CREATE TABLE `ai_recommend_topics`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '唯一主题ID',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '主题标题',
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '主题描述',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '主题分类',
  `tags` json NULL COMMENT '主题标签',
  `status` int NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_ai_recommen_created_8c7a2e`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'AI推荐主题表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for ai_rewrite_records
-- ----------------------------
DROP TABLE IF EXISTS `ai_rewrite_records`;
CREATE TABLE `ai_rewrite_records`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '唯一记录ID',
  `original_text` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '原始文本内容',
  `rewritten_text` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '改写后文本内容',
  `status` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '处理状态',
  `tokens` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '消耗tokens',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后更新时间',
  `error_log` json NULL COMMENT '错误日志详情',
  `model_version` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT 'AI模型版本',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_ai_rewrite__created_59e289`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'AI智能文案改写操作记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for ai_topic_polish_records
-- ----------------------------
DROP TABLE IF EXISTS `ai_topic_polish_records`;
CREATE TABLE `ai_topic_polish_records`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '唯一记录ID',
  `original_topic` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '原始主题',
  `polished_topic` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '润色后主题',
  `status` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '处理状态',
  `tokens` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '消耗tokens',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后更新时间',
  `error_log` json NULL COMMENT '错误日志详情',
  `model_version` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT 'AI模型版本',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_ai_topic_po_created_4a4add`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 683 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'AI主题润色记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for article_edit_tasks
-- ----------------------------
DROP TABLE IF EXISTS `article_edit_tasks`;
CREATE TABLE `article_edit_tasks`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '编辑任务ID',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '文章标题',
  `theme` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '文章主题',
  `params` json NULL COMMENT '编辑参数',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '文章内容',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'editing' COMMENT '编辑状态(editing编辑中/pending待发布/published已发布)',
  `is_public` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否公开',
  `tags` json NULL COMMENT '分类标签',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  `published_at` datetime(6) NULL DEFAULT NULL COMMENT '发布时间',
  `article_task_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联文章任务ID',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_article__article__5ee1b03b`(`article_task_id` ASC) USING BTREE,
  INDEX `fk_article__user_cfaa327f`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_article__article__5ee1b03b` FOREIGN KEY (`article_task_id`) REFERENCES `article_tasks` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_article__user_cfaa327f` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '文章编辑任务表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for article_publish_configs
-- ----------------------------
DROP TABLE IF EXISTS `article_publish_configs`;
CREATE TABLE `article_publish_configs`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '发布配置ID',
  `publish_title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '发布标题',
  `author_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '作者名称',
  `summary` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '文章摘要',
  `cover_image` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '封面图片URL',
  `enable_comments` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否开放评论区',
  `followers_only_comment` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否仅粉丝可评论',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  `edit_task_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联编辑任务ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_article__article__0a6385e9`(`edit_task_id` ASC) USING BTREE,
  CONSTRAINT `fk_article__article__0a6385e9` FOREIGN KEY (`edit_task_id`) REFERENCES `article_edit_tasks` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '文章发布配置表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for article_tasks
-- ----------------------------
DROP TABLE IF EXISTS `article_tasks`;
CREATE TABLE `article_tasks`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务ID',
  `client_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '客户端ID',
  `topic` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '文章主题',
  `is_public` tinyint(1) NOT NULL COMMENT '是否公开',
  `theme` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '文章排版主题',
  `author_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '作者名称',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '任务状态',
  `current_step` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '当前步骤',
  `steps` json NULL COMMENT '所有步骤的详细状态',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '文章标题',
  `snippet` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '文章摘要',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '文章内容',
  `word_count` int NULL DEFAULT NULL COMMENT '文章字数',
  `kb_content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '知识库搜索内容',
  `images` json NULL COMMENT '文章相关图片',
  `is_web_search` tinyint(1) NULL DEFAULT NULL COMMENT '是否进行联网搜索',
  `user_links` json NULL COMMENT '用户上传链接',
  `is_published` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已发布',
  `publish_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '发布URL',
  `start_time` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '开始时间',
  `end_time` datetime(6) NULL DEFAULT NULL COMMENT '结束时间',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  `total_duration` int NULL DEFAULT NULL COMMENT '总执行时间(秒)',
  `search_duration` int NULL DEFAULT NULL COMMENT '搜索步骤执行时间(秒)',
  `parse_duration` int NULL DEFAULT NULL COMMENT '解析步骤执行时间(秒)',
  `generate_duration` int NULL DEFAULT NULL COMMENT '生成步骤执行时间(秒)',
  `complete_duration` int NULL DEFAULT NULL COMMENT '完成步骤执行时间(秒)',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `client_id`(`client_id` ASC) USING BTREE,
  INDEX `fk_article__user_333061ee`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_article__user_333061ee` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '文章生成任务表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for article_topics
-- ----------------------------
DROP TABLE IF EXISTS `article_topics`;
CREATE TABLE `article_topics`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务ID',
  `related_task` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联任务ID',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '文章主题标题',
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '主题描述',
  `author_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '作者名称',
  `category` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '主题分类',
  `tags` json NULL COMMENT '主题标签',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'draft' COMMENT '主题状态(draft草稿/published已发布/scheduled计划发布)',
  `publish_date` date NULL DEFAULT NULL COMMENT '计划发布日期',
  `publish_time` datetime(6) NULL DEFAULT NULL COMMENT '实际发布时间',
  `created_at` datetime(6) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL COMMENT '更新时间',
  `view_count` int NOT NULL DEFAULT 0 COMMENT '浏览次数',
  `like_count` int NOT NULL DEFAULT 0 COMMENT '点赞次数',
  `comment_count` int NOT NULL DEFAULT 0 COMMENT '评论次数',
  `token_usage` json NULL COMMENT 'Token使用情况',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_article__user_88d7a32e`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_article__user_88d7a32e` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '文章主题清单表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for bp_orders
-- ----------------------------
DROP TABLE IF EXISTS `bp_orders`;
CREATE TABLE `bp_orders`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'BP订单ID',
  `trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '交易流水号',
  `product_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '产品名称',
  `payment_channel` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '支付渠道',
  `email` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '邮箱',
  `price` decimal(10, 2) NOT NULL COMMENT '订单价格',
  `payment_status` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '订单状态',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `trade_no`(`trade_no` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】BP订单表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for categories
-- ----------------------------
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '分类名称',
  `key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '分类标识',
  `scene` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '分类场景',
  `order` int NOT NULL DEFAULT 0 COMMENT '排序',
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '分类描述',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】分类表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for chat_records
-- ----------------------------
DROP TABLE IF EXISTS `chat_records`;
CREATE TABLE `chat_records`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '聊天记录ID',
  `session_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '会话ID',
  `message_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '消息类型',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '消息内容',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'completed' COMMENT '消息状态',
  `metadata` json NULL COMMENT '消息元数据(如工具调用信息等)',
  `tokens` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '消耗tokens',
  `model_version` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT 'AI模型版本',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后更新时间',
  `error_log` json NULL COMMENT '错误日志详情',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_chat_record_created_c02af5`(`created_at` ASC) USING BTREE,
  INDEX `idx_chat_record_created_eada27`(`created_at` ASC, `session_id` ASC, `user_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '聊天记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for commission_records
-- ----------------------------
DROP TABLE IF EXISTS `commission_records`;
CREATE TABLE `commission_records`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `amount` decimal(10, 2) NOT NULL COMMENT '佣金金额',
  `status` smallint NOT NULL DEFAULT 0 COMMENT '佣金状态',
  `description` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '佣金说明',
  `issue_time` datetime(6) NULL DEFAULT NULL COMMENT '发放时间',
  `withdrawal_time` datetime(6) NULL DEFAULT NULL COMMENT '提现时间',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `order_id` int NOT NULL COMMENT '关联订单',
  `relation_id` int NOT NULL COMMENT '关联的邀请关系',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '获得佣金的用户',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_commissi_trades_538dbf21`(`order_id` ASC) USING BTREE,
  INDEX `fk_commissi_invitati_f3e1b5ae`(`relation_id` ASC) USING BTREE,
  INDEX `fk_commissi_user_e17554e4`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_commissi_invitati_f3e1b5ae` FOREIGN KEY (`relation_id`) REFERENCES `invitation_relations` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_commissi_trades_538dbf21` FOREIGN KEY (`order_id`) REFERENCES `trades` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_commissi_user_e17554e4` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 12 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】佣金记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for credit_products
-- ----------------------------
DROP TABLE IF EXISTS `credit_products`;
CREATE TABLE `credit_products`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '产品ID',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '产品名称',
  `credits` int NULL DEFAULT NULL COMMENT '积分数量',
  `price` decimal(10, 2) NULL DEFAULT NULL COMMENT '价格',
  `status` tinyint(1) NULL DEFAULT 1 COMMENT '是否有效',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】积分产品表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for credit_recharge_orders
-- ----------------------------
DROP TABLE IF EXISTS `credit_recharge_orders`;
CREATE TABLE `credit_recharge_orders`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `product_id` int NOT NULL COMMENT '关联产品',
  `trade_id` int NOT NULL COMMENT '关联交易',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_credit_r_credit_p_811ec52a`(`product_id` ASC) USING BTREE,
  INDEX `fk_credit_r_trades_65625c67`(`trade_id` ASC) USING BTREE,
  INDEX `fk_credit_r_user_ce6621d5`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_credit_r_credit_p_811ec52a` FOREIGN KEY (`product_id`) REFERENCES `credit_products` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_credit_r_trades_65625c67` FOREIGN KEY (`trade_id`) REFERENCES `trades` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_credit_r_user_ce6621d5` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 12 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】积分充值订单表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for credit_records
-- ----------------------------
DROP TABLE IF EXISTS `credit_records`;
CREATE TABLE `credit_records`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `record_type` smallint NOT NULL COMMENT '记录类型',
  `credits` int NULL DEFAULT NULL COMMENT '积分变动数量',
  `balance` int NULL DEFAULT NULL COMMENT '变动后余额',
  `description` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '变动描述',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_credit_r_user_db6562ee`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_credit_r_user_db6562ee` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 2959 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】积分消耗记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for credit_service_prices
-- ----------------------------
DROP TABLE IF EXISTS `credit_service_prices`;
CREATE TABLE `credit_service_prices`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `service_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '服务代号',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '服务名称',
  `credits` int NULL DEFAULT NULL COMMENT '消耗积分数/unit',
  `unit` smallint NULL DEFAULT NULL COMMENT '计费单位',
  `description` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '服务描述',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否有效',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `service_code`(`service_code` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】积分服务定价表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for feedback
-- ----------------------------
DROP TABLE IF EXISTS `feedback`;
CREATE TABLE `feedback`  (
  `feedback_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '反馈ID',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户ID',
  `type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'other' COMMENT '反馈类型',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '反馈标题',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '反馈内容',
  `contact_info` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '联系方式',
  `images` json NULL COMMENT '相关图片URL列表',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '处理状态',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  PRIMARY KEY (`feedback_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户反馈记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for invitation_codes
-- ----------------------------
DROP TABLE IF EXISTS `invitation_codes`;
CREATE TABLE `invitation_codes`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '邀请码',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `code`(`code` ASC) USING BTREE,
  INDEX `fk_invitati_user_740652ea`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_invitati_user_740652ea` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 850 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】邀请码表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for invitation_relations
-- ----------------------------
DROP TABLE IF EXISTS `invitation_relations`;
CREATE TABLE `invitation_relations`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `code_id` int NOT NULL COMMENT '使用的邀请码',
  `invitee_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '被邀请人',
  `inviter_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '邀请人',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_invitati_invitati_33f6cd47`(`code_id` ASC) USING BTREE,
  INDEX `fk_invitati_user_8ebf7ceb`(`invitee_id` ASC) USING BTREE,
  INDEX `fk_invitati_user_d59b592a`(`inviter_id` ASC) USING BTREE,
  CONSTRAINT `fk_invitati_invitati_33f6cd47` FOREIGN KEY (`code_id`) REFERENCES `invitation_codes` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_invitati_user_8ebf7ceb` FOREIGN KEY (`invitee_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_invitati_user_d59b592a` FOREIGN KEY (`inviter_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 265 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】邀请关系表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for marketing_activity_plans
-- ----------------------------
DROP TABLE IF EXISTS `marketing_activity_plans`;
CREATE TABLE `marketing_activity_plans`  (
  `activity_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '活动ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '活动名称',
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '活动描述',
  `status` smallint NOT NULL DEFAULT 0 COMMENT '活动状态',
  `start_time` datetime(6) NOT NULL COMMENT '活动开始时间',
  `end_time` datetime(6) NOT NULL COMMENT '活动结束时间',
  `config` json NOT NULL COMMENT '活动配置（商品、福利和限制条件）',
  `is_visible` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否可见',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  PRIMARY KEY (`activity_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '营销活动计划表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for productions
-- ----------------------------
DROP TABLE IF EXISTS `productions`;
CREATE TABLE `productions`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '产品ID',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '产品名称',
  `price` decimal(10, 2) NOT NULL COMMENT '产品价格',
  `original_price` decimal(10, 2) NULL DEFAULT NULL COMMENT '原价',
  `product_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '产品类型',
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '产品描述',
  `extra_info` json NULL COMMENT '产品扩展信息',
  `validity_period` int NULL DEFAULT NULL COMMENT '有效期',
  `status` int NULL DEFAULT NULL COMMENT '上架状态',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】产品表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for reservations
-- ----------------------------
DROP TABLE IF EXISTS `reservations`;
CREATE TABLE `reservations`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '预约人姓名',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '预约人手机号',
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '预约人邮箱',
  `notes` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '备注信息',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 85 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for scenes
-- ----------------------------
DROP TABLE IF EXISTS `scenes`;
CREATE TABLE `scenes`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '场景ID',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '场景名称',
  `prompt` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '提示词',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `order` int NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 28 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】场景表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for system_notification
-- ----------------------------
DROP TABLE IF EXISTS `system_notification`;
CREATE TABLE `system_notification`  (
  `notification_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '通知ID',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '接收用户ID，为空表示全体用户',
  `type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'system' COMMENT '通知类型',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '通知标题',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '通知内容',
  `link` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '相关链接',
  `is_important` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否重要',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'unread' COMMENT '通知状态',
  `read_time` datetime(6) NULL DEFAULT NULL COMMENT '阅读时间',
  `expire_time` datetime(6) NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  PRIMARY KEY (`notification_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '系统通知表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for task_error_logs
-- ----------------------------
DROP TABLE IF EXISTS `task_error_logs`;
CREATE TABLE `task_error_logs`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '日志ID',
  `error_message` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '错误信息',
  `error_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '错误类型',
  `error_traceback` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '错误堆栈跟踪',
  `step` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '发生错误的步骤',
  `sub_step` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '发生错误的子步骤',
  `client_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '客户端ID',
  `task_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '任务类型(wx/xhs)',
  `additional_info` json NULL COMMENT '额外信息',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `task_id_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联任务ID',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_task_err_article__df99ea55`(`task_id_id` ASC) USING BTREE,
  INDEX `fk_task_err_user_ac921dca`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_task_err_article__df99ea55` FOREIGN KEY (`task_id_id`) REFERENCES `article_tasks` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_task_err_user_ac921dca` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '任务错误日志表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for task_usages
-- ----------------------------
DROP TABLE IF EXISTS `task_usages`;
CREATE TABLE `task_usages`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '统计ID',
  `task_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务名称',
  `provider` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '服务提供商',
  `model` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '模型名称',
  `act_model` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '实际使用的模型',
  `prompt_tokens` int NOT NULL COMMENT '输入token数量',
  `completion_tokens` int NOT NULL COMMENT '输出token数量',
  `total_tokens` int NOT NULL COMMENT '总token数量',
  `total_cost` double NOT NULL COMMENT '总成本(元)',
  `execution_time` double NOT NULL COMMENT '执行时间(秒)',
  `prompt` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '提示词',
  `response` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '响应内容',
  `created_at` datetime(6) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL COMMENT '更新时间',
  `task_id_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联任务ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_task_usa_article__3d88f3ac`(`task_id_id` ASC) USING BTREE,
  CONSTRAINT `fk_task_usa_article__3d88f3ac` FOREIGN KEY (`task_id_id`) REFERENCES `article_tasks` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '任务具体使用统计表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for total_usage_stats
-- ----------------------------
DROP TABLE IF EXISTS `total_usage_stats`;
CREATE TABLE `total_usage_stats`  (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '统计ID',
  `task_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务类型',
  `total_prompt_tokens` int NOT NULL DEFAULT 0 COMMENT '总输入token数量',
  `total_completion_tokens` int NOT NULL DEFAULT 0 COMMENT '总输出token数量',
  `total_tokens` int NOT NULL DEFAULT 0 COMMENT '总token数量',
  `total_cost` double NOT NULL DEFAULT 0 COMMENT '总成本(元)',
  `total_execution_time` double NOT NULL DEFAULT 0 COMMENT '总执行时间(秒)',
  `created_at` datetime(6) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL COMMENT '更新时间',
  `task_id_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联任务ID',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_total_us_article__2c842872`(`task_id_id` ASC) USING BTREE,
  INDEX `fk_total_us_user_7c4acd0f`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_total_us_article__2c842872` FOREIGN KEY (`task_id_id`) REFERENCES `article_tasks` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_total_us_user_7c4acd0f` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '总体使用统计表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for trades
-- ----------------------------
DROP TABLE IF EXISTS `trades`;
CREATE TABLE `trades`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '交易ID',
  `trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '交易流水号',
  `amount` decimal(10, 2) NOT NULL COMMENT '交易金额',
  `trade_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '交易类型',
  `payment_channel` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '支付渠道',
  `payment_status` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '支付状态',
  `payment_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '支付ID',
  `title` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '交易标题',
  `metadata` json NULL COMMENT '元数据，用于存储特定业务数据',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `paid_at` datetime(6) NULL DEFAULT NULL COMMENT '支付时间',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `trade_no`(`trade_no` ASC) USING BTREE,
  INDEX `fk_trades_user_50ed525e`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_trades_user_50ed525e` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 916 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】交易记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户id',
  `nickname` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '头像URL',
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户名',
  `password_hash` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '密码哈希值',
  `appid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '公众号appid',
  `openid` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '微信openid',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '手机号',
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '邮箱',
  `credits` int NOT NULL DEFAULT 0 COMMENT '积分',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否激活',
  `vip_level` int NOT NULL DEFAULT 0 COMMENT 'VIP等级',
  `role` smallint NOT NULL DEFAULT 1 COMMENT '角色',
  `status` smallint NOT NULL DEFAULT 1 COMMENT '状态',
  `registration_date` datetime(6) NOT NULL COMMENT '注册时间',
  `total_consumption` decimal(10, 2) NULL DEFAULT NULL COMMENT '累计消费金额',
  `last_login_time` datetime(6) NOT NULL COMMENT '最后登录时间',
  `usage_count` int NOT NULL DEFAULT 0 COMMENT '使用次数',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  PRIMARY KEY (`user_id`) USING BTREE,
  UNIQUE INDEX `username`(`username` ASC) USING BTREE,
  UNIQUE INDEX `email`(`email` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for user_parameters
-- ----------------------------
DROP TABLE IF EXISTS `user_parameters`;
CREATE TABLE `user_parameters`  (
  `param_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '参数id',
  `enable_head_info` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用头部信息',
  `enable_knowledge_base` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否绑定知识库',
  `default_theme` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'countryside' COMMENT '默认文章主题风格',
  `is_gzh_bind` tinyint(1) NOT NULL DEFAULT 0,
  `is_wechat_authorized` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已授权微信公众平台',
  `publish_target` int NOT NULL COMMENT '自动发布目标位置',
  `has_auth_reminded` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已提醒过授权',
  `qrcode_data` json NULL COMMENT '文章尾部推广二维码数据',
  `created_time` datetime(6) NOT NULL COMMENT '创建时间',
  `updated_time` datetime(6) NOT NULL COMMENT '更新时间',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户',
  PRIMARY KEY (`param_id`) USING BTREE,
  INDEX `fk_user_par_user_ecb98574`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_user_par_user_ecb98574` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户参数表模型' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for user_productions
-- ----------------------------
DROP TABLE IF EXISTS `user_productions`;
CREATE TABLE `user_productions`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '用户产品ID',
  `status` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'active' COMMENT '用户产品状态',
  `created_at` datetime(6) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NULL DEFAULT NULL COMMENT '更新时间',
  `production_id` int NOT NULL COMMENT '关联产品',
  `trade_id` int NOT NULL COMMENT '关联交易',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_user_pro_producti_3a7d5629`(`production_id` ASC) USING BTREE,
  INDEX `fk_user_pro_trades_84104780`(`trade_id` ASC) USING BTREE,
  INDEX `fk_user_pro_user_3175d8dc`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_user_pro_producti_3a7d5629` FOREIGN KEY (`production_id`) REFERENCES `productions` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_user_pro_trades_84104780` FOREIGN KEY (`trade_id`) REFERENCES `trades` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_user_pro_user_3175d8dc` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 316 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【贝】用户产品表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for user_sessions
-- ----------------------------
DROP TABLE IF EXISTS `user_sessions`;
CREATE TABLE `user_sessions`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `token` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '登录token',
  `login_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'web' COMMENT '登录类型',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '登录IP',
  `device_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '设备ID',
  `status` smallint NOT NULL DEFAULT 1 COMMENT '会话状态',
  `last_active_time` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后活跃时间',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_user_ses_user_32d979cc`(`user_id` ASC) USING BTREE,
  CONSTRAINT `fk_user_ses_user_32d979cc` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 6968 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户登录会话表' ROW_FORMAT = DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;

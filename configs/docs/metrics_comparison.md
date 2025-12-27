# 指标实现对比表

## 当前已实现 vs 图片要求对比

| 维度 | 指标名称 | 图片要求Key | 当前实现状态 | 当前实现位置 | 差异说明 |
|------|---------|------------|------------|------------|---------|
| **1. 核心营收** |
| | MRR (月经常性收入) | `revenue_mrr` | ❌ 未实现 | - | 需要订阅数据，当前只有交易记录 |
| | 新增付费用户数 | `new_paying_users` | ⚠️ 部分实现 | `/business/conversion-funnel` | 当前统计的是首次付费，但未按日统计 |
| | 付费转化率 | `payment_conversion_rate` | ⚠️ 部分实现 | `/business/conversion-funnel` | 有转化率计算，但未按日统计 |
| **2. 用户活跃** |
| | WAU (周活跃用户) | `active_users_weekly` | ⚠️ 部分实现 | `/business/activity-metrics` | 当前基于`last_login_time`，图片要求基于"有效生成" |
| | DAU (日活跃用户) | `active_users_daily` | ⚠️ 部分实现 | `/business/activity-metrics` | 当前基于`last_login_time`，图片要求基于"有效生成" |
| | 核心动作执行次数 | `total_generations` | ❌ 未实现 | - | 需要任务生成记录表 |
| **3. 产品价值** |
| | 生成结果采纳率 | `adoption_rate` | ❌ 未实现 | - | 需要用户行为记录表（复制/下载/保存） |
| | 人均生成任务数 | `avg_tasks_per_user` | ❌ 未实现 | - | 需要任务生成记录表 |
| **4. 留存与生命周期** |
| | 次日留存率 | `retention_day_1` | ⚠️ 部分实现 | `/business/retention` | 当前基于`last_login_time`，图片要求基于"再次访问并操作" |
| | 周留存率 | `retention_week_1` | ⚠️ 部分实现 | `/business/retention` | 当前基于`last_login_time`，图片要求基于"再次访问并操作" |
| **5. 成本监控** |
| | 单用户 Token 消耗 | `cost_per_user_token` | ❌ 未实现 | - | 需要任务生成记录表，记录token消耗 |
| | 任务失败率 | `task_error_rate` | ❌ 未实现 | - | 需要任务生成记录表，记录失败状态 |
| **6. 流量来源** |
| | 注册来源分布 | `user_source_channel` | ❌ 未实现 | - | `User.utm_source`字段已存在，但未统计 |

## 实现状态说明

- ✅ **已完全实现**: 指标已按图片要求完整实现
- ⚠️ **部分实现**: 有类似指标，但计算逻辑或数据源不符合图片要求
- ❌ **未实现**: 完全缺失的指标

## 需要新增的数据表

### 1. task_generate_log (任务生成记录表)
```sql
CREATE TABLE task_generate_log (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(50) NOT NULL,
    task_type VARCHAR(50),  -- 'image_text' | 'text_only'
    status VARCHAR(20),      -- 'success' | 'failed'
    token_consumed INT,      -- Token消耗量
    error_message TEXT,      -- 错误信息（如果失败）
    created_at DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_status (status)
);
```

### 2. user_action_log (用户行为记录表)
```sql
CREATE TABLE user_action_log (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(50) NOT NULL,
    action_type VARCHAR(20), -- 'copy' | 'download' | 'save'
    task_id INT,             -- 关联task_generate_log.id
    created_at DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_task_id (task_id),
    INDEX idx_action_type (action_type),
    INDEX idx_created_at (created_at)
);
```

### 3. subscription_records (订阅记录表)
```sql
CREATE TABLE subscription_records (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(50) NOT NULL,
    subscription_type VARCHAR(50), -- 订阅类型
    monthly_fee DECIMAL(10,2),      -- 月费
    start_date DATETIME,             -- 开始时间
    end_date DATETIME,               -- 结束时间
    status VARCHAR(20),              -- 'active' | 'expired' | 'cancelled'
    created_at DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_start_date (start_date),
    INDEX idx_end_date (end_date)
);
```

## 需要修改的现有实现

### 1. 活跃用户计算逻辑
- **当前**: 基于 `User.last_login_time`
- **需要**: 基于 `task_generate_log` 中的成功生成记录

### 2. 留存率计算逻辑
- **当前**: 基于 `User.last_login_time`
- **需要**: 基于 `task_generate_log` 中的操作记录

### 3. 付费用户统计
- **当前**: 基于 `User.total_consumption > 0`
- **需要**: 基于 `Trade` 表中首次付费时间，按日统计

## 优先级实施建议

### P0 优先级（核心指标，必须实现）
1. MRR (月经常性收入) - 需要创建订阅记录表
2. 新增付费用户数 - 需要修改统计逻辑
3. 付费转化率 - 需要修改统计逻辑
4. WAU (周活跃用户) - 需要创建任务生成记录表
5. 核心动作执行次数 - 需要创建任务生成记录表
6. 生成结果采纳率 - 需要创建用户行为记录表
7. 次日留存率 - 需要修改计算逻辑
8. 单用户 Token 消耗 - 需要创建任务生成记录表
9. 任务失败率 - 需要创建任务生成记录表

### P1 优先级（重要指标，建议实现）
1. DAU (日活跃用户) - 需要创建任务生成记录表
2. 人均生成任务数 - 需要创建任务生成记录表
3. 周留存率 - 需要修改计算逻辑
4. 注册来源分布 - 已有字段，只需添加统计接口


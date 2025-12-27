# 产品指标定义表

## 1. 核心营收 (Core Revenue)

### 1.1 MRR (月经常性收入)
- **指标名称**: MRR (Monthly Recurring Revenue)
- **数据库字段/Key**: `revenue_mrr`
- **定义 & 计算公式**: 当前生效中所有订阅用户的月费总和。(不含单次充值,仅计算订阅)
- **数据来源/埋点建议**: 支付系统(Stripe/微信/支付宝回调)
- **优先级**: P0

### 1.2 新增付费用户数
- **指标名称**: 新增付费用户数 (New Paying Users)
- **数据库字段/Key**: `new_paying_users`
- **定义 & 计算公式**: 当日首次完成付费(订阅或充值)的用户数。
- **数据来源/埋点建议**: 支付系统 pay_success 事件
- **优先级**: P0

### 1.3 付费转化率
- **指标名称**: 付费转化率 (Payment Conversion Rate)
- **数据库字段/Key**: `payment_conversion_rate`
- **定义 & 计算公式**: (当日新增付费用户数 / 当日新增注册用户数) * 100%
- **数据来源/埋点建议**: 用户表+订单表
- **优先级**: P0

---

## 2. 用户活跃 (User Activity)

### 2.1 WAU (周活跃用户)
- **指标名称**: WAU (Weekly Active Users)
- **数据库字段/Key**: `active_users_weekly`
- **定义 & 计算公式**: 过去7天内,至少完成过1次有效生成的用户数。(注意:仅登录不算活跃)
- **数据来源/埋点建议**: 埋点事件: task_generate_success
- **优先级**: P0

### 2.2 DAU (日活跃用户)
- **指标名称**: DAU (Daily Active Users)
- **数据库字段/Key**: `active_users_daily`
- **定义 & 计算公式**: 当日完成过至少1次有效生成的去重用户数。
- **数据来源/埋点建议**: 埋点事件: task_generate_success
- **优先级**: P1

### 2.3 核心动作执行次数
- **指标名称**: 核心动作执行次数 (Core Action Execution Count)
- **数据库字段/Key**: `total_generations`
- **定义 & 计算公式**: 用户点击"生成"并成功返回结果的总次数。(可细分为:图文生成、纯文本生成)
- **数据来源/埋点建议**: 后端日志: API调用成功计数
- **优先级**: P0

---

## 3. 产品价值 (Product Value)

### 3.1 生成结果采纳率
- **指标名称**: 生成结果采纳率 (Generation Result Adoption Rate)
- **数据库字段/Key**: `adoption_rate`
- **定义 & 计算公式**: (用户点击复制、下载、保存的次数 / 总生成次数) * 100%
- **数据来源/埋点建议**: 前端埋点: result_copy, result_download
- **优先级**: P0

### 3.2 人均生成任务数
- **指标名称**: 人均生成任务数 (Average Tasks Generated Per User)
- **数据库字段/Key**: `avg_tasks_per_user`
- **定义 & 计算公式**: 当日总生成次数 / DAU。(监测产品是"玩具"还是"工具")
- **数据来源/埋点建议**: 计算字段
- **优先级**: P1

---

## 4. 留存与生命周期 (Retention & Lifecycle)

### 4.1 次日留存率
- **指标名称**: 次日留存率 (Day 1 Retention Rate)
- **数据库字段/Key**: `retention_day_1`
- **定义 & 计算公式**: (昨日注册且今日再次访问并操作的用户数 / 昨日新增注册数) * 100%
- **数据来源/埋点建议**: 用户行为日志
- **优先级**: P0

### 4.2 周留存率
- **指标名称**: 周留存率 (Weekly Retention Rate)
- **数据库字段/Key**: `retention_week_1`
- **定义 & 计算公式**: (上周注册且本周再次访问并操作的用户数 / 上周新增注册数) * 100%
- **数据来源/埋点建议**: 用户行为日志
- **优先级**: P1

---

## 5. 成本监控 (Cost Monitoring)

### 5.1 单用户 Token 消耗
- **指标名称**: 单用户 Token 消耗 (Token Consumption Per User)
- **数据库字段/Key**: `cost_per_user_token`
- **定义 & 计算公式**: 单个用户当日所有任务消耗的 Token 总量(或折算金额)。
- **数据来源/埋点建议**: 后端日志: LLM API 回调 Token 数
- **优先级**: P0

### 5.2 任务失败率
- **指标名称**: 任务失败率 (Task Failure Rate)
- **数据库字段/Key**: `task_error_rate`
- **定义 & 计算公式**: (生成失败或报错的请求数 / 总请求数) * 100%
- **数据来源/埋点建议**: 后端异常日志(Error Logs)
- **优先级**: P0

---

## 6. 流量来源 (Traffic Source)

### 6.1 注册来源分布
- **指标名称**: 注册来源分布 (Registration Source Distribution)
- **数据库字段/Key**: `user_source_channel`
- **定义 & 计算公式**: 统计 utm_source 字段分布(如:Bilibili, Xiaohongshu, Direct).
- **数据来源/埋点建议**: 注册接口记录 Referer 或 UTM 参数
- **优先级**: P1

---

## 指标优先级说明

- **P0**: 核心指标，必须实现
- **P1**: 重要指标，建议实现

---

## 数据表字段映射建议

### User 表
- `utm_source`: 用户来源渠道（已存在）
- `registration_date`: 注册时间（已存在）
- `last_login_time`: 最后登录时间（已存在）

### Trade 表
- `amount`: 交易金额（已存在）
- `payment_status`: 支付状态（已存在）
- `paid_at`: 支付时间（已存在）
- `trade_type`: 交易类型（已存在，需区分订阅/充值）

### 需要新增的埋点/字段
1. **任务生成记录表** (task_generate_log)
   - `user_id`: 用户ID
   - `task_type`: 任务类型（图文生成/纯文本生成）
   - `status`: 状态（成功/失败）
   - `token_consumed`: Token消耗量
   - `created_at`: 创建时间

2. **用户行为记录表** (user_action_log)
   - `user_id`: 用户ID
   - `action_type`: 行为类型（copy/download/save）
   - `task_id`: 关联任务ID
   - `created_at`: 创建时间

3. **订阅记录表** (subscription_records)
   - `user_id`: 用户ID
   - `subscription_type`: 订阅类型
   - `monthly_fee`: 月费
   - `start_date`: 开始时间
   - `end_date`: 结束时间
   - `status`: 状态（生效中/已过期）


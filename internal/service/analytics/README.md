# 数据分析服务 (Analytics Service)

统一的数据分析指标接口，支持并行计算多个指标，提供高性能的数据分析能力。

## 目录结构

```
internal/service/analytics/
├── metrics.go           # 指标定义和枚举
├── user_activity.go     # 用户活跃度统计服务
├── metrics_service.go    # 统一的数据分析服务
└── README.md            # 本文档
```

## 功能特性

1. **统一接口**: 提供统一的数据分析接口，支持多种指标
2. **并行计算**: 使用Go的goroutine并行计算多个指标，提高性能
3. **灵活配置**: 通过枚举定义指标，可以灵活控制显示/隐藏
4. **日期范围查询**: 支持单日查询和日期范围查询
5. **按维度分组**: 支持按维度分组返回数据

## API接口

### 1. 获取指标数据

**接口**: `GET /api/v1/admin/analytics/metrics`

**参数**:
- `date` (可选): 统计日期，格式：`YYYY-MM-DD`，默认为今天
- `metrics` (可选): 要获取的指标列表，多个指标用逗号分隔，如：`active_users_daily,active_users_weekly`。如果不指定则返回所有启用的指标
- `start_date` (可选): 开始日期（用于日期范围查询），格式：`YYYY-MM-DD`
- `end_date` (可选): 结束日期（用于日期范围查询），格式：`YYYY-MM-DD`

**示例1: 获取今天的DAU和WAU**
```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics?metrics=active_users_daily,active_users_weekly" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**示例2: 获取指定日期的所有指标**
```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics?date=2026-01-05" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**示例3: 获取日期范围内的DAU数据**
```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics?metrics=active_users_daily&start_date=2026-01-01&end_date=2026-01-31" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "获取指标数据成功",
  "data": {
    "date": "2026-01-05",
    "metrics": {
      "active_users_daily": {
        "key": "active_users_daily",
        "value": 1234,
        "unit": "人",
        "timestamp": "2026-01-05T00:00:00+08:00",
        "dimension": "user_activity"
      },
      "active_users_weekly": {
        "key": "active_users_weekly",
        "value": 5678,
        "unit": "人",
        "timestamp": "2026-01-05T00:00:00+08:00",
        "dimension": "user_activity"
      }
    },
    "dimensions": {
      "user_activity": [
        {
          "key": "active_users_daily",
          "value": 1234,
          "unit": "人",
          "timestamp": "2026-01-05T00:00:00+08:00",
          "dimension": "user_activity"
        },
        {
          "key": "active_users_weekly",
          "value": 5678,
          "unit": "人",
          "timestamp": "2026-01-05T00:00:00+08:00",
          "dimension": "user_activity"
        }
      ]
    }
  }
}
```

### 2. 获取指标信息列表

**接口**: `GET /api/v1/admin/analytics/metrics/info`

**参数**: 无

**示例**:
```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics/info" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "获取指标信息成功",
  "data": {
    "metrics": [
      {
        "key": "active_users_daily",
        "name": "DAU (日活跃用户)",
        "name_en": "Daily Active Users",
        "dimension": "user_activity",
        "definition": "当日完成过至少1次有效生成的去重用户数。",
        "formula": "COUNT(DISTINCT 当日有有效操作的用户ID)",
        "priority": "P1",
        "enabled": true
      }
    ],
    "dimensions": {
      "user_activity": [...]
    }
  }
}
```

## 趋势图表接口

### 获取用户活跃度趋势

**接口**: `GET /api/v1/admin/analytics/user/activity-trend`

**参数**:
- `period` (可选): 统计周期，`day`/`week`/`month`，默认为`day`
- `start_date` (必需): 开始日期，格式：`YYYY-MM-DD`
- `end_date` (必需): 结束日期，格式：`YYYY-MM-DD`
- `include_wau` (可选): 是否包含周活数据，默认为`true`
- `include_mau` (可选): 是否包含月活数据，默认为`true`

**示例1: 获取按天统计的趋势数据**
```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/user/activity-trend?period=day&start_date=2025-12-05&end_date=2026-01-05" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**示例2: 获取按周统计的趋势数据（只包含DAU和MAU）**
```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/user/activity-trend?period=week&start_date=2025-12-05&end_date=2026-01-05&include_wau=false&include_mau=true" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "获取用户活跃度趋势成功",
  "data": {
    "period": "day",
    "start_date": "2025-12-05",
    "end_date": "2026-01-05",
    "data": [
      {
        "date": "2025-12-05",
        "dau": 150,
        "wau": 1200,
        "mau": 2800,
        "stickiness": 5.36
      },
      {
        "date": "2025-12-06",
        "dau": 165,
        "wau": 1250,
        "mau": 2850,
        "stickiness": 5.79
      }
    ],
    "summary": {
      "avg_dau": 196.5,
      "max_dau": 250,
      "min_dau": 150,
      "current_dau": 196,
      "current_mau": 2910,
      "stickiness": 6.74,
      "growth_rate": 30.67
    }
  }
}
```

**响应字段说明**:
- `data`: 趋势数据点列表，每个数据点包含：
  - `date`: 日期
  - `dau`: 日活跃用户数
  - `wau`: 周活跃用户数（可选）
  - `mau`: 月活跃用户数（可选）
  - `stickiness`: 用户粘性 (DAU/MAU * 100%)
- `summary`: 汇总数据，包含：
  - `avg_dau`: 平均日活
  - `max_dau`: 最大日活
  - `min_dau`: 最小日活
  - `current_dau`: 当前日活
  - `current_mau`: 当前月活
  - `stickiness`: 当前粘性
  - `growth_rate`: 增长率（相比第一个数据点）

## 已实现的指标

### 1. 核心营收 (Core Revenue)

1. **MRR (月经常性收入)** - `revenue_mrr`
   - 定义: 当前生效中所有订阅用户的月费总和
   - 当前实现: 基于`UserProduction`和`Production`表统计
   - 优先级: P0

2. **新增付费用户数** - `new_paying_users`
   - 定义: 当日首次完成付费(订阅或充值)的用户数
   - 当前实现: 基于`Trade`表统计首次付费时间
   - 优先级: P0

3. **付费转化率** - `payment_conversion_rate`
   - 定义: (当日新增付费用户数 / 当日新增注册用户数) * 100%
   - 当前实现: 基于`Trade`和`User`表计算
   - 优先级: P0

### 2. 用户活跃度 (User Activity)

1. **DAU (日活跃用户)** - `active_users_daily`
   - 定义: 当日完成过至少1次有效生成的去重用户数
   - 当前实现: 基于`last_login_time`统计
   - 优先级: P1

2. **WAU (周活跃用户)** - `active_users_weekly`
   - 定义: 过去7天内,至少完成过1次有效生成的用户数
   - 当前实现: 基于`last_login_time`统计
   - 优先级: P0

3. **MAU (月活跃用户)** - `active_users_monthly`
   - 定义: 过去30天内,至少完成过1次有效生成的用户数
   - 当前实现: 基于`last_login_time`统计
   - 优先级: P1

4. **核心动作执行次数** - `total_generations`
   - 定义: 用户点击"生成"并成功返回结果的总次数
   - 当前实现: 基于`CreditRecord`表中消费类型的记录数近似（实际应为task_generate_log表）
   - 优先级: P0
   - 状态: ✅ 已实现（近似）

### 3. 产品价值 (Product Value)

1. **人均生成任务数** - `avg_tasks_per_user`
   - 定义: 当日总生成次数 / DAU
   - 当前实现: 基于`CreditRecord`和DAU计算
   - 优先级: P1
   - 状态: ✅ 已实现

2. **生成结果采纳率** - `adoption_rate`
   - 定义: (用户点击复制、下载、保存的次数 / 总生成次数) * 100%
   - 当前实现: 需要`user_action_log`表，当前返回0
   - 优先级: P0
   - 状态: ⚠️ 待实现（需要新表）

### 4. 留存与生命周期 (Retention & Lifecycle)

1. **次日留存率** - `retention_day_1`
   - 定义: (昨日注册且今日再次访问并操作的用户数 / 昨日新增注册数) * 100%
   - 当前实现: 基于`User`表的`registration_date`和`last_login_time`统计
   - 优先级: P0
   - 状态: ✅ 已实现

2. **周留存率** - `retention_week_1`
   - 定义: (上周注册且本周再次访问并操作的用户数 / 上周新增注册数) * 100%
   - 当前实现: 基于`User`表的`registration_date`和`last_login_time`统计
   - 优先级: P1
   - 状态: ✅ 已实现

### 6. 成本监控 (Cost Monitoring)

1. **单用户 Token 消耗** - `cost_per_user_token`
   - 定义: 单个用户当日所有任务消耗的积分总量（基于积分消费记录近似）
   - 当前实现: 基于`CreditRecord`表中消费类型的积分消耗计算（实际应为task_generate_log表的token_consumed）
   - 优先级: P0
   - 状态: ✅ 已实现（近似）

2. **任务失败率** - `task_error_rate`
   - 定义: (生成失败或报错的请求数 / 总请求数) * 100%
   - 当前实现: 需要`task_generate_log`表，当前返回0
   - 优先级: P0
   - 状态: ⚠️ 待实现（需要新表）

### 5. 成本监控 (Cost Monitoring)

1. **单用户 Token 消耗** - `cost_per_user_token`
   - 定义: 单个用户当日所有任务消耗的积分总量（基于积分消费记录近似）
   - 当前实现: 基于`CreditRecord`表中消费类型的积分消耗计算（实际应为task_generate_log表的token_consumed）
   - 优先级: P0
   - 状态: ✅ 已实现（近似）

2. **任务失败率** - `task_error_rate`
   - 定义: (生成失败或报错的请求数 / 总请求数) * 100%
   - 当前实现: 需要`task_generate_log`表，当前返回0
   - 优先级: P0
   - 状态: ⚠️ 待实现（需要新表）

### 6. 流量来源 (Traffic Source)

1. **注册来源分布** - `user_source_channel`
   - 定义: 统计 utm_source 字段分布
   - 当前实现: 基于`User`表的`utm_source`字段统计
   - 优先级: P1
   - 状态: ✅ 已实现
   - 接口: `GET /api/v1/admin/analytics/traffic/source-distribution`

## 指标实现状态总览

| 维度 | 指标 | Key | 状态 | 备注 |
|------|------|-----|------|------|
| 核心营收 | MRR | `revenue_mrr` | ✅ | 基于UserProduction |
| 核心营收 | 新增付费用户数 | `new_paying_users` | ✅ | 基于Trade表 |
| 核心营收 | 付费转化率 | `payment_conversion_rate` | ✅ | 基于Trade和User表 |
| 用户活跃 | DAU | `active_users_daily` | ✅ | 基于last_login_time |
| 用户活跃 | WAU | `active_users_weekly` | ✅ | 基于last_login_time |
| 用户活跃 | MAU | `active_users_monthly` | ✅ | 基于last_login_time |
| 用户活跃 | 核心动作执行次数 | `total_generations` | ✅ | 基于CreditRecord（近似） |
| 产品价值 | 人均生成任务数 | `avg_tasks_per_user` | ✅ | 基于CreditRecord和DAU |
| 产品价值 | 生成结果采纳率 | `adoption_rate` | ⚠️ | 需要user_action_log表 |
| 留存 | 次日留存率 | `retention_day_1` | ✅ | 基于User表 |
| 留存 | 周留存率 | `retention_week_1` | ✅ | 基于User表 |
| 成本监控 | 单用户Token消耗 | `cost_per_user_token` | ✅ | 基于CreditRecord（近似） |
| 成本监控 | 任务失败率 | `task_error_rate` | ⚠️ | 需要task_generate_log表 |
| 流量来源 | 注册来源分布 | `user_source_channel` | ✅ | 基于User表 |

**说明**:
- ✅ 已实现：可以直接使用
- ⚠️ 待实现：需要新增数据表后才能实现

## 扩展指标

要添加新指标，需要：

1. 在`metrics.go`中添加指标Key和Info定义
2. 在对应的服务文件中实现计算逻辑（如`user_activity.go`、`revenue_service.go`、`product_value_service.go`、`cost_service.go`等）
3. 在`metrics_service.go`的`GetMetrics`方法中添加并行计算逻辑
4. 更新`README.md`文档说明

## 性能优化

1. **并行计算**: 使用goroutine并行计算多个指标
2. **日期范围并行**: 日期范围查询时，每个日期的计算也是并行的（限制最大并发数为5，避免数据库连接池耗尽）
3. **批量查询**: 使用数据库批量查询，避免N+1问题
4. **错误容错**: 日期范围查询时，部分日期失败不影响整体结果，会返回成功的数据
5. **日期范围限制**: 限制最大查询范围为90天，避免查询时间过长

## 注意事项

1. 所有日期时间统一使用北京时间（UTC+8）
2. 当前活跃用户统计基于`last_login_time`，未来可扩展为基于`task_generate_log`表统计有效生成操作
3. 指标计算失败时，会返回错误，不会部分返回数据


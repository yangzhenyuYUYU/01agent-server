package analytics

import "time"

// MetricDimension 指标维度枚举
type MetricDimension string

const (
	DimensionCoreRevenue    MetricDimension = "core_revenue"    // 核心营收
	DimensionUserActivity   MetricDimension = "user_activity"   // 用户活跃
	DimensionProductValue   MetricDimension = "product_value"   // 产品价值
	DimensionRetention      MetricDimension = "retention"       // 留存与生命周期
	DimensionCostMonitoring MetricDimension = "cost_monitoring" // 成本监控
	DimensionTrafficSource  MetricDimension = "traffic_source"  // 流量来源
)

// MetricKey 指标Key枚举
type MetricKey string

const (
	// 核心营收
	MetricMRR               MetricKey = "revenue_mrr"             // MRR (月经常性收入)
	MetricNewPayingUsers    MetricKey = "new_paying_users"        // 新增付费用户数
	MetricPaymentConversion MetricKey = "payment_conversion_rate" // 付费转化率

	// 用户活跃
	MetricDAU              MetricKey = "active_users_daily"   // DAU (日活跃用户)
	MetricWAU              MetricKey = "active_users_weekly"  // WAU (周活跃用户)
	MetricMAU              MetricKey = "active_users_monthly" // MAU (月活跃用户)
	MetricTotalGenerations MetricKey = "total_generations"    // 核心动作执行次数

	// 产品价值
	MetricAdoptionRate    MetricKey = "adoption_rate"      // 生成结果采纳率
	MetricAvgTasksPerUser MetricKey = "avg_tasks_per_user" // 人均生成任务数

	// 留存与生命周期
	MetricRetentionDay1  MetricKey = "retention_day_1"  // 次日留存率
	MetricRetentionWeek1 MetricKey = "retention_week_1" // 周留存率

	// 成本监控
	MetricCostPerUserToken MetricKey = "cost_per_user_token" // 单用户 Token 消耗
	MetricTaskErrorRate    MetricKey = "task_error_rate"     // 任务失败率

	// 流量来源
	MetricUserSourceChannel MetricKey = "user_source_channel" // 注册来源分布
)

// MetricInfo 指标信息
type MetricInfo struct {
	Key        MetricKey       `json:"key"`
	Name       string          `json:"name"`
	NameEN     string          `json:"name_en"`
	Dimension  MetricDimension `json:"dimension"`
	Definition string          `json:"definition"`
	Formula    string          `json:"formula"`
	Priority   string          `json:"priority"` // P0/P1
	Enabled    bool            `json:"enabled"`  // 是否启用
}

// GetMetricInfo 获取指标信息映射
func GetMetricInfo() map[MetricKey]MetricInfo {
	return map[MetricKey]MetricInfo{
		// 核心营收
		MetricMRR: {
			Key:        MetricMRR,
			Name:       "MRR (月经常性收入)",
			NameEN:     "Monthly Recurring Revenue",
			Dimension:  DimensionCoreRevenue,
			Definition: "当前生效中所有订阅用户的月费总和。(不含单次充值,仅计算订阅)",
			Formula:    "SUM(生效中订阅的月费)",
			Priority:   "P0",
			Enabled:    true,
		},
		MetricNewPayingUsers: {
			Key:        MetricNewPayingUsers,
			Name:       "新增付费用户数",
			NameEN:     "New Paying Users",
			Dimension:  DimensionCoreRevenue,
			Definition: "当日首次完成付费(订阅或充值)的用户数。",
			Formula:    "COUNT(DISTINCT 当日首次付费的用户ID)",
			Priority:   "P0",
			Enabled:    true,
		},
		MetricPaymentConversion: {
			Key:        MetricPaymentConversion,
			Name:       "付费转化率",
			NameEN:     "Payment Conversion Rate",
			Dimension:  DimensionCoreRevenue,
			Definition: "(当日新增付费用户数 / 当日新增注册用户数) * 100%",
			Formula:    "(new_paying_users / new_registered_users) * 100%",
			Priority:   "P0",
			Enabled:    true,
		},

		// 用户活跃
		MetricDAU: {
			Key:        MetricDAU,
			Name:       "DAU (日活跃用户)",
			NameEN:     "Daily Active Users",
			Dimension:  DimensionUserActivity,
			Definition: "当日完成过至少1次有效生成的去重用户数。",
			Formula:    "COUNT(DISTINCT 当日有有效操作的用户ID)",
			Priority:   "P1",
			Enabled:    true,
		},
		MetricWAU: {
			Key:        MetricWAU,
			Name:       "WAU (周活跃用户)",
			NameEN:     "Weekly Active Users",
			Dimension:  DimensionUserActivity,
			Definition: "过去7天内,至少完成过1次有效生成的用户数。(注意:仅登录不算活跃)",
			Formula:    "COUNT(DISTINCT 过去7天内有有效操作的用户ID)",
			Priority:   "P0",
			Enabled:    true,
		},
		MetricMAU: {
			Key:        MetricMAU,
			Name:       "MAU (月活跃用户)",
			NameEN:     "Monthly Active Users",
			Dimension:  DimensionUserActivity,
			Definition: "过去30天内,至少完成过1次有效生成的用户数。",
			Formula:    "COUNT(DISTINCT 过去30天内有有效操作的用户ID)",
			Priority:   "P1",
			Enabled:    true,
		},
		MetricTotalGenerations: {
			Key:        MetricTotalGenerations,
			Name:       "核心动作执行次数",
			NameEN:     "Core Action Execution Count",
			Dimension:  DimensionUserActivity,
			Definition: "用户点击\"生成\"并成功返回结果的总次数。",
			Formula:    "COUNT(消费类型的积分记录)",
			Priority:   "P0",
			Enabled:    true, // 基于CreditRecord近似实现
		},

		// 产品价值
		MetricAdoptionRate: {
			Key:        MetricAdoptionRate,
			Name:       "生成结果采纳率",
			NameEN:     "Generation Result Adoption Rate",
			Dimension:  DimensionProductValue,
			Definition: "(用户点击复制、下载、保存的次数 / 总生成次数) * 100%",
			Formula:    "(action_count / generation_count) * 100%",
			Priority:   "P0",
			Enabled:    false, // 暂未实现，需要user_action_log表
		},
		MetricAvgTasksPerUser: {
			Key:        MetricAvgTasksPerUser,
			Name:       "人均生成任务数",
			NameEN:     "Average Tasks Generated Per User",
			Dimension:  DimensionProductValue,
			Definition: "当日总生成次数 / DAU。(监测产品是\"玩具\"还是\"工具\")",
			Formula:    "total_generations / dau",
			Priority:   "P1",
			Enabled:    true, // 基于CreditRecord和DAU计算
		},

		// 留存与生命周期
		MetricRetentionDay1: {
			Key:        MetricRetentionDay1,
			Name:       "次日留存率",
			NameEN:     "Day 1 Retention Rate",
			Dimension:  DimensionRetention,
			Definition: "(昨日注册且今日再次访问并操作的用户数 / 昨日新增注册数) * 100%",
			Formula:    "(retained_users / new_registered_users) * 100%",
			Priority:   "P0",
			Enabled:    true,
		},
		MetricRetentionWeek1: {
			Key:        MetricRetentionWeek1,
			Name:       "周留存率",
			NameEN:     "Weekly Retention Rate",
			Dimension:  DimensionRetention,
			Definition: "(上周注册且本周再次访问并操作的用户数 / 上周新增注册数) * 100%",
			Formula:    "(retained_users / new_registered_users) * 100%",
			Priority:   "P1",
			Enabled:    true,
		},

		// 成本监控
		MetricCostPerUserToken: {
			Key:        MetricCostPerUserToken,
			Name:       "单用户 Token 消耗",
			NameEN:     "Token Consumption Per User",
			Dimension:  DimensionCostMonitoring,
			Definition: "单个用户当日所有任务消耗的积分总量（基于积分消费记录近似）。",
			Formula:    "SUM(ABS(credits)) / user_count",
			Priority:   "P0",
			Enabled:    true, // 基于CreditRecord近似实现
		},
		MetricTaskErrorRate: {
			Key:        MetricTaskErrorRate,
			Name:       "任务失败率",
			NameEN:     "Task Failure Rate",
			Dimension:  DimensionCostMonitoring,
			Definition: "(生成失败或报错的请求数 / 总请求数) * 100%",
			Formula:    "(failed_count / total_count) * 100%",
			Priority:   "P0",
			Enabled:    false, // 暂未实现，需要task_generate_log表
		},

		// 流量来源
		MetricUserSourceChannel: {
			Key:        MetricUserSourceChannel,
			Name:       "注册来源分布",
			NameEN:     "Registration Source Distribution",
			Dimension:  DimensionTrafficSource,
			Definition: "统计 utm_source 字段分布(如:Bilibili, Xiaohongshu, Direct).",
			Formula:    "GROUP BY utm_source",
			Priority:   "P1",
			Enabled:    true,
		},
	}
}

// MetricValue 指标值
type MetricValue struct {
	Key       MetricKey       `json:"key"`
	Value     interface{}     `json:"value"`
	Unit      string          `json:"unit,omitempty"` // 单位，如：人、次、%、元
	Timestamp time.Time       `json:"timestamp"`      // 统计时间
	Dimension MetricDimension `json:"dimension"`      // 所属维度
}

// MetricsResponse 指标响应结构
type MetricsResponse struct {
	Date       string                            `json:"date"`                 // 统计日期
	Metrics    map[MetricKey]MetricValue         `json:"metrics"`              // 指标值映射
	Dimensions map[MetricDimension][]MetricValue `json:"dimensions,omitempty"` // 按维度分组（可选）
}

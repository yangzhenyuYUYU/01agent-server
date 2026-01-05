# 会员统计接口文档

## 概述

会员统计接口用于统计和分析不同会员类型的购买情况，包括购买数量、收入、占比以及趋势数据。支持按日期范围筛选，趋势数据最大跨度限制为1年。

## 基础信息

- **基础路径**: `/api/v1/admin/membership`
- **认证方式**: 需要管理员权限（Bearer Token）
- **时区**: 所有日期时间统一使用北京时间（UTC+8，CST）

## 会员分类

系统将产品按照以下分类进行统计：

| 分类 | 包含产品 |
|------|---------|
| 免费版 | 免费版 |
| 轻量版 | 轻量版、轻量版体验、轻量版年度会员 |
| 专业版 | 专业版、专业版半年订阅升级套餐、专业版年度会员、专业版体验、专业版周体验、专业版开通测试 |
| 种子终身会员 | 种子终身会员 |

---

## 接口列表

### 1. 获取会员购买概览

获取指定时间范围内各会员类型的购买统计信息，包括购买数量、收入、占比等。

**接口地址**: `GET /api/v1/admin/membership/overview`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期，格式：`YYYY-MM-DD`。如果为空，则统计全部数据 |
| end_date | string | 否 | 结束日期，格式：`YYYY-MM-DD`。如果为空，则统计全部数据 |

**请求示例**:

```bash
# 获取全部数据
curl -X GET "http://localhost:8080/api/v1/admin/membership/overview" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 获取指定日期范围的数据
curl -X GET "http://localhost:8080/api/v1/admin/membership/overview?start_date=2025-12-01&end_date=2026-01-31" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "获取会员概览成功",
  "data": {
    "total_count": 1250,
    "total_revenue": 125000.50,
    "total_users": 850,
    "category_stats": [
      {
        "category": "专业版",
        "product_name": "",
        "count": 600,
        "revenue": 75000.00,
        "percentage": 48.0,
        "unique_users": 450
      },
      {
        "category": "轻量版",
        "product_name": "",
        "count": 450,
        "revenue": 35000.00,
        "percentage": 36.0,
        "unique_users": 320
      },
      {
        "category": "种子终身会员",
        "product_name": "",
        "count": 150,
        "revenue": 12000.50,
        "percentage": 12.0,
        "unique_users": 80
      },
      {
        "category": "免费版",
        "product_name": "",
        "count": 50,
        "revenue": 0.00,
        "percentage": 4.0,
        "unique_users": 50
      }
    ],
    "product_stats": [
      {
        "category": "专业版",
        "product_name": "专业版",
        "count": 300,
        "revenue": 45000.00,
        "percentage": 24.0,
        "unique_users": 250
      },
      {
        "category": "专业版",
        "product_name": "专业版年度会员",
        "count": 200,
        "revenue": 25000.00,
        "percentage": 16.0,
        "unique_users": 150
      },
      {
        "category": "轻量版",
        "product_name": "轻量版",
        "count": 250,
        "revenue": 20000.00,
        "percentage": 20.0,
        "unique_users": 180
      },
      {
        "category": "轻量版",
        "product_name": "轻量版年度会员",
        "count": 150,
        "revenue": 12000.00,
        "percentage": 12.0,
        "unique_users": 100
      },
      {
        "category": "轻量版",
        "product_name": "轻量版体验",
        "count": 50,
        "revenue": 3000.00,
        "percentage": 4.0,
        "unique_users": 40
      },
      {
        "category": "种子终身会员",
        "product_name": "种子终身会员",
        "count": 150,
        "revenue": 12000.50,
        "percentage": 12.0,
        "unique_users": 80
      },
      {
        "category": "免费版",
        "product_name": "免费版",
        "count": 50,
        "revenue": 0.00,
        "percentage": 4.0,
        "unique_users": 50
      }
    ]
  }
}
```

**响应字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| total_count | integer | 总购买数量 |
| total_revenue | number | 总收入金额（元） |
| total_users | integer | 总用户数（去重） |
| category_stats | array | 按分类统计的数据 |
| product_stats | array | 按产品统计的数据 |

**category_stats/product_stats 字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| category | string | 会员分类（免费版/轻量版/专业版/种子终身会员） |
| product_name | string | 产品名称（category_stats中为空） |
| count | integer | 购买数量 |
| revenue | number | 收入金额（元） |
| percentage | number | 占比（%） |
| unique_users | integer | 去重用户数 |

---

### 2. 获取会员购买趋势

获取指定时间范围内的会员购买趋势数据，支持按天/周/月统计。**最大跨度限制为1年**。

**接口地址**: `GET /api/v1/admin/membership/trend`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 是 | 开始日期，格式：`YYYY-MM-DD` |
| end_date | string | 是 | 结束日期，格式：`YYYY-MM-DD` |
| period | string | 否 | 统计周期，可选值：`day`（按天）、`week`（按周）、`month`（按月）。默认为 `day` |

**请求示例**:

```bash
# 按天统计（默认）
curl -X GET "http://localhost:8080/api/v1/admin/membership/trend?start_date=2025-12-01&end_date=2026-01-31&period=day" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 按周统计
curl -X GET "http://localhost:8080/api/v1/admin/membership/trend?start_date=2025-12-01&end_date=2026-01-31&period=week" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 按月统计
curl -X GET "http://localhost:8080/api/v1/admin/membership/trend?start_date=2025-01-01&end_date=2025-12-31&period=month" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例**:

```json
{
  "code": 0,
  "msg": "获取会员趋势成功",
  "data": {
    "start_date": "2025-12-01",
    "end_date": "2026-01-31",
    "period": "day",
    "data": [
      {
        "date": "2025-12-01",
        "category_stats": [
          {
            "category": "专业版",
            "product_name": "专业版",
            "count": 15,
            "revenue": 2250.00,
            "percentage": 50.0,
            "unique_users": 12
          },
          {
            "category": "轻量版",
            "product_name": "轻量版",
            "count": 10,
            "revenue": 800.00,
            "percentage": 33.3,
            "unique_users": 8
          },
          {
            "category": "种子终身会员",
            "product_name": "种子终身会员",
            "count": 5,
            "revenue": 400.50,
            "percentage": 16.7,
            "unique_users": 3
          }
        ],
        "total_count": 30,
        "total_revenue": 3450.50
      },
      {
        "date": "2025-12-02",
        "category_stats": [
          {
            "category": "专业版",
            "product_name": "专业版年度会员",
            "count": 8,
            "revenue": 1000.00,
            "percentage": 40.0,
            "unique_users": 6
          },
          {
            "category": "轻量版",
            "product_name": "轻量版年度会员",
            "count": 12,
            "revenue": 960.00,
            "percentage": 60.0,
            "unique_users": 10
          }
        ],
        "total_count": 20,
        "total_revenue": 1960.00
      }
    ],
    "summary": {
      "total_count": 1250,
      "total_revenue": 125000.50,
      "total_users": 850,
      "category_stats": [...],
      "product_stats": [...]
    }
  }
}
```

**响应字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| start_date | string | 开始日期 |
| end_date | string | 结束日期 |
| period | string | 统计周期 |
| data | array | 趋势数据点数组 |
| summary | object | 汇总数据（格式同概览接口） |

**data 数组元素字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| date | string | 日期（格式根据period不同：day为YYYY-MM-DD，week为YYYY-Www，month为YYYY-MM） |
| category_stats | array | 该日期各分类的统计数据 |
| total_count | integer | 该日期总购买数量 |
| total_revenue | number | 该日期总收入金额（元） |

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误（日期格式错误、日期范围超过限制等） |
| 500 | 服务器内部错误 |

**错误响应示例**:

```json
{
  "code": 400,
  "msg": "日期范围不能超过1年",
  "data": null
}
```

---

## 前端集成示例

### JavaScript/Axios 示例

```javascript
import axios from 'axios';

// 配置axios实例
const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1/admin',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  }
});

// 获取会员概览
async function getMembershipOverview(startDate, endDate) {
  try {
    const params = {};
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;
    
    const response = await api.get('/membership/overview', { params });
    return response.data.data;
  } catch (error) {
    console.error('获取会员概览失败:', error);
    throw error;
  }
}

// 获取会员趋势
async function getMembershipTrend(startDate, endDate, period = 'day') {
  try {
    const response = await api.get('/membership/trend', {
      params: {
        start_date: startDate,
        end_date: endDate,
        period: period
      }
    });
    return response.data.data;
  } catch (error) {
    console.error('获取会员趋势失败:', error);
    throw error;
  }
}

// 使用示例
async function loadMembershipData() {
  // 获取最近30天的概览
  const endDate = new Date();
  const startDate = new Date();
  startDate.setDate(startDate.getDate() - 30);
  
  const overview = await getMembershipOverview(
    startDate.toISOString().split('T')[0],
    endDate.toISOString().split('T')[0]
  );
  
  console.log('会员概览:', overview);
  
  // 获取最近7天的趋势（按天）
  const trend = await getMembershipTrend(
    startDate.toISOString().split('T')[0],
    endDate.toISOString().split('T')[0],
    'day'
  );
  
  console.log('会员趋势:', trend);
}
```

### ECharts 图表集成示例

```javascript
// 绘制会员占比饼图
function renderMembershipPieChart(overview) {
  const option = {
    title: {
      text: '会员购买占比',
      left: 'center'
    },
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '会员类型',
        type: 'pie',
        radius: '50%',
        data: overview.category_stats.map(item => ({
          value: item.count,
          name: item.category
        })),
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  };
  
  const chart = echarts.init(document.getElementById('membership-pie-chart'));
  chart.setOption(option);
}

// 绘制会员趋势折线图
function renderMembershipTrendChart(trend) {
  const dates = trend.data.map(item => item.date);
  const categories = ['专业版', '轻量版', '种子终身会员', '免费版'];
  
  const series = categories.map(category => {
    const data = trend.data.map(item => {
      const stat = item.category_stats.find(s => s.category === category);
      return stat ? stat.count : 0;
    });
    
    return {
      name: category,
      type: 'line',
      data: data
    };
  });
  
  const option = {
    title: {
      text: '会员购买趋势',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: categories,
      top: '10%'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: series
  };
  
  const chart = echarts.init(document.getElementById('membership-trend-chart'));
  chart.setOption(option);
}
```

---

## 性能优化说明

1. **批量查询**: 使用SQL的`GROUP BY`和聚合函数，一次性获取所有统计数据，避免N+1查询问题
2. **索引优化**: 查询使用了`user_productions.created_at`、`productions.product_type`等字段的索引
3. **日期范围限制**: 趋势查询最大跨度限制为1年，避免查询过大数据集
4. **并发控制**: 服务层使用高效的SQL查询，减少数据库连接时间

---

## 注意事项

1. **日期格式**: 所有日期参数必须使用 `YYYY-MM-DD` 格式
2. **时区**: 所有日期时间统一使用北京时间（UTC+8）
3. **日期范围**: 趋势查询的日期范围不能超过1年，超过会自动截断
4. **数据来源**: 统计数据基于`user_productions`、`productions`和`trades`表的关联查询
5. **收入计算**: 只统计支付状态为`success`的交易金额
6. **用户去重**: `unique_users`字段统计的是去重后的用户数

---

## 更新日志

- **2026-01-06**: 初始版本发布
  - 支持会员购买概览统计
  - 支持会员购买趋势统计（按天/周/月）
  - 支持日期范围筛选
  - 性能优化：批量查询、索引优化


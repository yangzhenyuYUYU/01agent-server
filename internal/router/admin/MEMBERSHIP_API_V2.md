# 会员统计API文档 V2

## 概述

会员统计API提供会员购买和积分套餐的统计分析功能，包括概览、趋势和产品销售趋势（折线图）数据。

**基础URL**: `/api/v1/admin/membership`

**认证**: 所有接口需要管理员权限（`AdminAuth`中间件）

---

## 1. 获取会员购买概览

### 接口信息
- **URL**: `/api/v1/admin/membership/overview`
- **方法**: `GET`
- **描述**: 获取会员和积分套餐的购买概览，包含分类统计、产品统计和总计数据

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期，格式：YYYY-MM-DD |
| end_date | string | 否 | 结束日期，格式：YYYY-MM-DD |

### 响应示例

```json
{
  "code": 200,
  "msg": "获取会员概览成功",
  "data": {
    "membership_count": 150,        // 会员购买数
    "membership_revenue": 45000.00,  // 会员收入
    "membership_users": 120,          // 会员用户数
    "credit_package_count": 80,      // 积分套餐购买数
    "credit_package_revenue": 12000.00, // 积分套餐收入
    "credit_package_users": 65,      // 积分套餐用户数
    "total_count": 230,              // 总购买数
    "total_revenue": 57000.00,       // 总收入
    "total_users": 185,              // 总用户数
    "category_stats": [              // 按分类统计（仅会员，用于饼状图）
      {
        "category": "专业版",
        "product_name": "",
        "count": 100,
        "revenue": 30000.00,
        "percentage": 66.67,
        "unique_users": 80
      }
    ],
    "credit_category_stats": [       // 按积分套餐分类统计（用于饼状图）
      {
        "category": "600积分",
        "product_name": "600积分",
        "count": 30,
        "revenue": 4500.00,
        "percentage": 37.50,
        "unique_users": 25
      },
      {
        "category": "1500积分",
        "product_name": "1500积分",
        "count": 25,
        "revenue": 3750.00,
        "percentage": 31.25,
        "unique_users": 20
      },
      {
        "category": "3000积分",
        "product_name": "3000积分",
        "count": 25,
        "revenue": 3750.00,
        "percentage": 31.25,
        "unique_users": 20
      }
    ],
    "product_stats": [                // 按产品统计（会员+积分套餐）
      {
        "category": "",
        "product_name": "专业版年度会员",
        "count": 50,
        "revenue": 15000.00,
        "percentage": 21.74,
        "unique_users": 45
      },
      {
        "category": "",
        "product_name": "600积分",
        "count": 30,
        "revenue": 4500.00,
        "percentage": 13.04,
        "unique_users": 25
      }
    ]
  }
}
```

### 字段说明

- **membership_count/revenue/users**: 会员相关统计（从`user_productions`表统计，只统计微信和支付宝渠道，排除兑换码兑换）
- **credit_package_count/revenue/users**: 积分套餐统计（从`user_productions`表统计，关联`productions`表，`product_type = "积分套餐"`，只统计微信和支付宝渠道，排除兑换码兑换）
- **total_count/revenue/users**: 总计（会员+积分套餐）
- **category_stats**: 按会员分类统计（免费版、轻量版、专业版、种子终身会员），用于会员饼状图
- **credit_category_stats**: 按积分套餐分类统计（600积分、1500积分、3000积分），用于积分套餐饼状图
- **product_stats**: 按产品名称统计（包含所有会员产品和积分套餐）

**重要说明**：
- 收入统计与`/api/v1/admin/analytics/payment/overview`接口保持一致
- 只统计微信（`wx_qr`）和支付宝（`alipay_qr`）渠道的支付成功订单
- 日期过滤使用`paid_at`字段（如果没有则使用`created_at`）
- 自动排除兑换码兑换（`trade_type = 'activation'`）的记录和用户

---

## 2. 获取会员购买趋势

### 接口信息
- **URL**: `/api/v1/admin/membership/trend`
- **方法**: `GET`
- **描述**: 获取会员购买趋势数据，按时间周期聚合

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 是 | 开始日期，格式：YYYY-MM-DD |
| end_date | string | 是 | 结束日期，格式：YYYY-MM-DD |
| period | string | 否 | 统计周期：day/week/month，默认为day |

**限制**: 日期范围最大1年

### 响应示例

```json
{
  "code": 200,
  "msg": "获取会员趋势成功",
  "data": {
    "start_date": "2025-12-01",
    "end_date": "2026-01-05",
    "period": "day",
    "data": [
      {
        "date": "2025-12-01",
        "category_stats": [
          {
            "category": "专业版",
            "product_name": "",
            "count": 5,
            "revenue": 1500.00,
            "percentage": 100.00,
            "unique_users": 5
          }
        ],
        "total_count": 5,
        "total_revenue": 1500.00
      }
    ],
    "summary": {
      // 同概览接口的完整数据结构
    }
  }
}
```

---

## 3. 获取产品销售趋势（折线图）

### 接口信息
- **URL**: `/api/v1/admin/membership/product-trend`
- **方法**: `GET`
- **描述**: 获取产品销售趋势数据，每个产品一条折线，用于绘制折线图。支持会员和积分套餐产品。

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 是 | 开始日期，格式：YYYY-MM-DD |
| end_date | string | 是 | 结束日期，格式：YYYY-MM-DD |
| period | string | 否 | 统计周期：day/week/month，默认为day |

**限制**: 日期范围最大1年

### 响应示例

```json
{
  "code": 200,
  "msg": "获取产品销售趋势成功",
  "data": {
    "start_date": "2025-12-01",
    "end_date": "2026-01-05",
    "period": "day",
    "products": [
      {
        "product_name": "专业版年度会员",
        "data": [
          {
            "date": "2025-12-01",
            "count": 2,
            "revenue": 600.00
          },
          {
            "date": "2025-12-02",
            "count": 0,
            "revenue": 0.00
          },
          {
            "date": "2025-12-03",
            "count": 3,
            "revenue": 900.00
          }
        ]
      },
      {
        "product_name": "600积分",
        "data": [
          {
            "date": "2025-12-01",
            "count": 5,
            "revenue": 750.00
          },
          {
            "date": "2025-12-02",
            "count": 2,
            "revenue": 300.00
          }
        ]
      }
    ],
    "summary": {
      // 同概览接口的完整数据结构
    }
  }
}
```

### 字段说明

- **products**: 产品列表，每个产品包含：
  - **product_name**: 产品名称
  - **data**: 趋势数据点数组，每个数据点包含：
    - **date**: 日期（格式根据period而定）
    - **count**: 购买数量
    - **revenue**: 收入金额
- **summary**: 汇总数据（同概览接口）

### 前端使用示例（ECharts）

```javascript
// 假设API返回的数据存储在 response.data 中
const products = response.data.products;
const dates = []; // 从所有产品的data中提取唯一日期

// 构建series数据
const series = products.map(product => {
  return {
    name: product.product_name,
    type: 'line',
    data: product.data.map(point => point.revenue), // 或 point.count
    smooth: true
  };
});

// ECharts配置
const option = {
  title: { text: '产品销售趋势' },
  tooltip: { trigger: 'axis' },
  legend: { data: products.map(p => p.product_name) },
  xAxis: {
    type: 'category',
    data: dates
  },
  yAxis: { type: 'value' },
  series: series
};
```

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误（日期格式、日期范围超限等） |
| 500 | 服务器内部错误 |

---

## 性能说明

- **并行查询**: 概览和趋势接口使用Go的goroutine并行查询会员和积分套餐数据，提升性能
- **批量查询**: 趋势接口使用SQL的`GROUP BY`批量查询，避免N+1问题
- **日期范围限制**: 趋势接口最大支持1年范围，避免查询时间过长

---

## 注意事项

1. **时区**: 所有日期使用CST时区（UTC+8）
2. **统计规则（与`GetPaymentOverview`保持一致）**:
   - 只统计微信和支付宝渠道（`payment_channel = 'wx_qr' OR 'alipay_qr'`）
   - 只统计支付成功的交易（`payment_status = 'success'`）
   - 排除兑换码兑换（`trade_type != 'activation'`）
   - 使用`paid_at`进行日期过滤
   - 兑换码兑换的用户不计入用户数量
3. **数据来源**:
   - 会员统计：从`Trade`表查询，关联`user_productions`和`productions`表，`product_type = "订阅服务"`
   - 积分套餐统计：从`Trade`表查询，关联`user_productions`和`productions`表，`product_type = "积分套餐"`，产品名称匹配`products.py`中的定义（600积分、1500积分、3000积分）
4. **日期格式**: 所有日期参数和返回都使用`YYYY-MM-DD`格式（period=day时）或相应格式（week/month）
5. **饼状图数据**: 
   - `category_stats`: 会员分类统计（免费版、轻量版、专业版、种子终身会员），用于会员购买饼状图
   - `credit_category_stats`: 积分套餐分类统计（600积分、1500积分、3000积分），用于积分套餐购买饼状图
6. **数据一致性**: `total_revenue`应与`/api/v1/admin/analytics/payment/overview`接口的`period_income`保持一致（在相同日期范围内）


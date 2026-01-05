# 数据分析接口使用示例

## 1. 获取用户活跃度趋势（用于图表展示）

### 请求示例

```bash
# 按天统计，获取2025-12-05到2026-01-05的趋势数据
curl -X GET "http://localhost:8080/api/v1/admin/analytics/user/activity-trend?period=day&start_date=2025-12-05&end_date=2026-01-05" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript/Axios 示例

```javascript
// 获取用户活跃度趋势数据
async function getActivityTrend() {
  try {
    const response = await axios.get('/api/v1/admin/analytics/user/activity-trend', {
      params: {
        period: 'day',           // 统计周期：day/week/month
        start_date: '2025-12-05', // 开始日期
        end_date: '2026-01-05',   // 结束日期
        include_wau: true,        // 是否包含周活（可选，默认true）
        include_mau: true         // 是否包含月活（可选，默认true）
      },
      headers: {
        'Authorization': 'Bearer YOUR_TOKEN'
      }
    });
    
    console.log('趋势数据:', response.data.data);
    // 使用 response.data.data.data 绘制图表
    // 使用 response.data.data.summary 显示汇总信息
  } catch (error) {
    console.error('获取趋势数据失败:', error);
  }
}
```

### 响应示例

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
      },
      {
        "date": "2025-12-07",
        "dau": 180,
        "wau": 1300,
        "mau": 2900,
        "stickiness": 6.21
      }
      // ... 更多数据点
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

### 前端图表集成示例（使用 ECharts）

```javascript
// 使用 ECharts 绘制趋势图
function renderTrendChart(trendData) {
  const dates = trendData.data.map(item => item.date);
  const dauData = trendData.data.map(item => item.dau);
  const mauData = trendData.data.map(item => item.mau || 0);
  const stickinessData = trendData.data.map(item => item.stickiness || 0);

  const option = {
    title: {
      text: '用户活跃度趋势'
    },
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['DAU', 'MAU', '用户粘性(%)']
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: [
      {
        type: 'value',
        name: '用户数',
        position: 'left'
      },
      {
        type: 'value',
        name: '粘性(%)',
        position: 'right'
      }
    ],
    series: [
      {
        name: 'DAU',
        type: 'line',
        data: dauData,
        smooth: true
      },
      {
        name: 'MAU',
        type: 'line',
        data: mauData,
        smooth: true
      },
      {
        name: '用户粘性(%)',
        type: 'line',
        yAxisIndex: 1,
        data: stickinessData,
        smooth: true
      }
    ]
  };

  const chart = echarts.init(document.getElementById('trendChart'));
  chart.setOption(option);
}
```

## 2. 获取单日指标数据

### 请求示例

```bash
# 获取2026-01-05的DAU和WAU
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics?date=2026-01-05&metrics=active_users_daily,active_users_weekly" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript 示例

```javascript
async function getDailyMetrics() {
  const response = await axios.get('/api/v1/admin/analytics/metrics', {
    params: {
      date: '2026-01-05',
      metrics: 'active_users_daily,active_users_weekly,active_users_monthly'
    }
  });
  
  const metrics = response.data.data.metrics;
  console.log('今日DAU:', metrics.active_users_daily.value);
  console.log('本周WAU:', metrics.active_users_weekly.value);
  console.log('本月MAU:', metrics.active_users_monthly.value);
}
```

## 3. 获取日期范围内的指标数据

### 请求示例

```bash
# 获取2025-12-05到2026-01-05的DAU数据
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics?metrics=active_users_daily&start_date=2025-12-05&end_date=2026-01-05" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript 示例

```javascript
async function getMetricsRange() {
  const response = await axios.get('/api/v1/admin/analytics/metrics', {
    params: {
      metrics: 'active_users_daily,active_users_monthly',
      start_date: '2025-12-05',
      end_date: '2026-01-05'
    }
  });
  
  // response.data.data.results 是一个数组，包含每天的指标数据
  const results = response.data.data.results;
  results.forEach(dayData => {
    console.log(`${dayData.date}: DAU=${dayData.metrics.active_users_daily.value}`);
  });
}
```

## 4. 获取所有指标信息

### 请求示例

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics/info" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript 示例

```javascript
async function getMetricsInfo() {
  const response = await axios.get('/api/v1/admin/analytics/metrics/info');
  
  // 获取所有启用的指标
  const enabledMetrics = response.data.data.metrics.filter(m => m.enabled);
  console.log('启用的指标:', enabledMetrics);
  
  // 按维度分组
  const byDimension = response.data.data.dimensions;
  console.log('用户活跃度指标:', byDimension.user_activity);
}
```

## 5. 完整的前端集成示例

```javascript
// 完整的用户活跃度仪表盘组件
class ActivityDashboard {
  constructor() {
    this.chart = null;
    this.initChart();
  }

  async loadData(startDate, endDate, period = 'day') {
    try {
      // 1. 获取趋势数据
      const trendResponse = await axios.get('/api/v1/admin/analytics/user/activity-trend', {
        params: { period, start_date: startDate, end_date: endDate }
      });
      
      // 2. 获取今日指标
      const today = new Date().toISOString().split('T')[0];
      const metricsResponse = await axios.get('/api/v1/admin/analytics/metrics', {
        params: { 
          date: today,
          metrics: 'active_users_daily,active_users_weekly,active_users_monthly'
        }
      });

      // 3. 更新图表
      this.updateChart(trendResponse.data.data);
      
      // 4. 更新卡片数据
      this.updateCards(metricsResponse.data.data.metrics, trendResponse.data.data.summary);
      
    } catch (error) {
      console.error('加载数据失败:', error);
    }
  }

  updateChart(trendData) {
    const dates = trendData.data.map(item => item.date);
    const dauData = trendData.data.map(item => item.dau);
    const mauData = trendData.data.map(item => item.mau || 0);

    const option = {
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value' },
      series: [
        { name: 'DAU', type: 'line', data: dauData, smooth: true },
        { name: 'MAU', type: 'line', data: mauData, smooth: true }
      ]
    };

    this.chart.setOption(option);
  }

  updateCards(metrics, summary) {
    // 更新今日DAU卡片
    document.getElementById('today-dau').textContent = metrics.active_users_daily.value;
    
    // 更新本月MAU卡片
    document.getElementById('month-mau').textContent = summary.current_mau;
    
    // 更新本周WAU卡片
    document.getElementById('week-wau').textContent = metrics.active_users_weekly.value;
    
    // 更新用户粘性
    document.getElementById('stickiness').textContent = 
      summary.stickiness.toFixed(2) + '%';
  }

  initChart() {
    this.chart = echarts.init(document.getElementById('trendChart'));
  }
}

// 使用示例
const dashboard = new ActivityDashboard();
dashboard.loadData('2025-12-05', '2026-01-05', 'day');
```

## 6. 获取注册来源分布

### 请求示例

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/traffic/source-distribution?start_date=2025-12-05&end_date=2026-01-05" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript 示例

```javascript
async function getSourceDistribution() {
  const response = await axios.get('/api/v1/admin/analytics/traffic/source-distribution', {
    params: {
      start_date: '2025-12-05',
      end_date: '2026-01-05'
    }
  });
  
  const distributions = response.data.data.data;
  console.log('来源分布:', distributions);
  
  // 可以用于绘制饼图或柱状图
  const chartData = distributions.map(item => ({
    name: item.source,
    value: item.count
  }));
}
```

### 响应示例

```json
{
  "code": 0,
  "msg": "获取注册来源分布成功",
  "data": {
    "start_date": "2025-12-05",
    "end_date": "2026-01-05",
    "total": 1000,
    "data": [
      {
        "source": "Bilibili",
        "count": 500
      },
      {
        "source": "Xiaohongshu",
        "count": 300
      },
      {
        "source": "direct",
        "count": 200
      }
    ]
  }
}
```

## 7. 获取所有指标数据（统一接口）

### 请求示例

```bash
# 获取今天的所有启用指标
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 获取指定日期的特定指标
curl -X GET "http://localhost:8080/api/v1/admin/analytics/metrics?date=2026-01-05&metrics=revenue_mrr,new_paying_users,retention_day_1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript 示例

```javascript
async function getAllMetrics() {
  // 获取今天的所有指标
  const response = await axios.get('/api/v1/admin/analytics/metrics');
  
  const metrics = response.data.data.metrics;
  
  // 按维度分组显示
  const byDimension = response.data.data.dimensions;
  
  console.log('核心营收指标:', byDimension.core_revenue);
  console.log('用户活跃指标:', byDimension.user_activity);
  console.log('留存指标:', byDimension.retention);
}
```

### 响应示例

```json
{
  "code": 0,
  "msg": "获取指标数据成功",
  "data": {
    "date": "2026-01-05",
    "metrics": {
      "revenue_mrr": {
        "key": "revenue_mrr",
        "value": 50000.0,
        "unit": "元",
        "timestamp": "2026-01-05T00:00:00+08:00",
        "dimension": "core_revenue"
      },
      "new_paying_users": {
        "key": "new_paying_users",
        "value": 10,
        "unit": "人",
        "timestamp": "2026-01-05T00:00:00+08:00",
        "dimension": "core_revenue"
      },
      "active_users_daily": {
        "key": "active_users_daily",
        "value": 196,
        "unit": "人",
        "timestamp": "2026-01-05T00:00:00+08:00",
        "dimension": "user_activity"
      },
      "retention_day_1": {
        "key": "retention_day_1",
        "value": 15.5,
        "unit": "%",
        "timestamp": "2026-01-05T00:00:00+08:00",
        "dimension": "retention"
      }
    },
    "dimensions": {
      "core_revenue": [...],
      "user_activity": [...],
      "retention": [...]
    }
  }
}
```

## 注意事项

1. **日期格式**: 所有日期参数必须使用 `YYYY-MM-DD` 格式
2. **日期范围限制**: 最大查询范围为90天，超过会返回400错误
3. **时区**: 所有日期时间统一使用北京时间（UTC+8）
4. **并发控制**: 日期范围查询时，系统会自动控制并发数（最大5个），避免数据库连接池耗尽
5. **错误处理**: 部分日期查询失败时，会返回成功的数据，不会完全失败
6. **并行计算**: 所有指标计算使用goroutine并行执行，充分利用Go的并发优势


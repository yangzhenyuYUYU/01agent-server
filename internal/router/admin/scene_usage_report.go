package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"01agent_server/internal/repository"
	"01agent_server/internal/service/analytics"
	"github.com/gin-gonic/gin"
)

// generateSceneUsageHTML 生成场景使用分析HTML报告
func generateSceneUsageHTML(report *analytics.SceneUsageReport) string {
	// 场景统计表格
	sceneStatsHTML := generateSceneStatsTable(report.SceneStats)

	// 用户类型对比图表
	userComparisonHTML := generateUserTypeComparisonChart(report.UserTypeComparison)

	// 产品对比表格
	productComparisonHTML := generateProductComparisonTable(report.ProductComparison)

	// 导出统计图表
	exportStatsHTML := generateExportStatsChart(report.ExportStats)

	// AI功能统计
	aiStatsHTML := generateAIFeatureStatsTable(report.AIFeatureStats)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>场景使用分析报告 - %s</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB',
                'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            padding: 20px;
            color: #333;
            line-height: 1.6;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 36px;
            margin-bottom: 10px;
            font-weight: 700;
        }

        .header .subtitle {
            font-size: 16px;
            opacity: 0.9;
        }

        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            padding: 30px 40px;
            background: #f8f9fa;
            border-bottom: 2px solid #e9ecef;
        }

        .summary-card {
            background: white;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .summary-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
        }

        .summary-card .label {
            font-size: 14px;
            color: #6c757d;
            margin-bottom: 8px;
        }

        .summary-card .value {
            font-size: 28px;
            font-weight: 700;
            color: #667eea;
        }

        .section {
            padding: 40px;
            border-bottom: 1px solid #e9ecef;
        }

        .section:last-child {
            border-bottom: none;
        }

        .section-title {
            font-size: 24px;
            font-weight: 700;
            margin-bottom: 20px;
            color: #2c3e50;
            padding-bottom: 10px;
            border-bottom: 3px solid #667eea;
        }

        .section-description {
            color: #6c757d;
            margin-bottom: 20px;
            font-size: 14px;
        }

        table {
            width: 100%%;
            border-collapse: collapse;
            margin-top: 20px;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
        }

        thead {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }

        th, td {
            padding: 16px;
            text-align: left;
            border-bottom: 1px solid #e9ecef;
        }

        th {
            font-weight: 600;
            text-transform: uppercase;
            font-size: 12px;
            letter-spacing: 0.5px;
        }

        tbody tr:hover {
            background: #f8f9fa;
        }

        .badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            background: #e9ecef;
            color: #495057;
        }

        .badge-primary {
            background: #667eea;
            color: white;
        }

        .badge-success {
            background: #28a745;
            color: white;
        }

        .badge-warning {
            background: #ffc107;
            color: #212529;
        }

        .chart-container {
            position: relative;
            height: 400px;
            margin: 30px 0;
            padding: 20px;
            background: white;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
        }

        .progress-bar {
            width: 100%%;
            height: 8px;
            background: #e9ecef;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 8px;
        }

        .progress-fill {
            height: 100%%;
            background: linear-gradient(90deg, #667eea 0%%, #764ba2 100%%);
            transition: width 0.3s ease;
        }

        .footer {
            padding: 30px 40px;
            background: #f8f9fa;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
        }

        .highlight {
            background: linear-gradient(120deg, #84fab0 0%%, #8fd3f4 100%%);
            padding: 2px 8px;
            border-radius: 4px;
            font-weight: 600;
        }

        @media print {
            body {
                background: white;
                padding: 0;
            }
            .container {
                box-shadow: none;
            }
        }

        @media (max-width: 768px) {
            .header h1 {
                font-size: 24px;
            }
            .summary {
                grid-template-columns: 1fr;
                padding: 20px;
            }
            .section {
                padding: 20px;
            }
            .chart-container {
                height: 300px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- 头部 -->
        <div class="header">
            <h1>🎯 场景使用分析报告</h1>
            <div class="subtitle">
                数据周期：%s ~ %s | 生成时间：%s
            </div>
        </div>

        <!-- 数据概览 -->
        <div class="summary">
            <div class="summary-card">
                <div class="label">总用户数</div>
                <div class="value">%d</div>
            </div>
            <div class="summary-card">
                <div class="label">免费用户</div>
                <div class="value">%d</div>
            </div>
            <div class="summary-card">
                <div class="label">付费用户</div>
                <div class="value">%d</div>
            </div>
            <div class="summary-card">
                <div class="label">总项目数</div>
                <div class="value">%d</div>
            </div>
            <div class="summary-card">
                <div class="label">总导出数</div>
                <div class="value">%d</div>
            </div>
        </div>

        <!-- 场景使用统计 -->
        <div class="section">
            <h2 class="section-title">📊 场景使用TOP排行</h2>
            <p class="section-description">统计各场景的使用次数、用户数、完成率等核心指标</p>
            %s
        </div>

        <!-- 用户类型对比 -->
        <div class="section">
            <h2 class="section-title">👥 付费 vs 非付费用户场景偏好</h2>
            <p class="section-description">对比不同用户类型在各场景的使用差异</p>
            %s
        </div>

        <!-- 产品套餐对比 -->
        <div class="section">
            <h2 class="section-title">💎 产品套餐场景分析</h2>
            <p class="section-description">分析不同产品套餐用户的场景使用偏好</p>
            %s
        </div>

        <!-- 导出格式统计 -->
        <div class="section">
            <h2 class="section-title">📤 导出格式偏好</h2>
            <p class="section-description">统计用户导出内容时的格式选择分布</p>
            %s
        </div>

        <!-- AI功能使用 -->
        <div class="section">
            <h2 class="section-title">🤖 AI能力使用统计</h2>
            <p class="section-description">分析AI排版、改写、润色等功能的使用情况</p>
            %s
        </div>

        <!-- 页脚 -->
        <div class="footer">
            <p>© 2026 01agent - 数据驱动产品决策</p>
        </div>
    </div>
</body>
</html>`,
		report.ReportDate,
		report.StartDate, report.EndDate, report.ReportDate,
		report.TotalUsers, report.FreeUsers, report.PaidUsers,
		report.TotalProjects, report.TotalExports,
		sceneStatsHTML,
		userComparisonHTML,
		productComparisonHTML,
		exportStatsHTML,
		aiStatsHTML,
	)
}

// generateSceneStatsTable 生成场景统计表格
func generateSceneStatsTable(stats []analytics.SceneUsageStats) string {
	if len(stats) == 0 {
		return `<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>`
	}

	rows := []string{}
	for i, stat := range stats {
		rank := i + 1
		rankBadge := ""
		if rank == 1 {
			rankBadge = `<span class="badge badge-primary">🥇 第1名</span>`
		} else if rank == 2 {
			rankBadge = `<span class="badge badge-success">🥈 第2名</span>`
		} else if rank == 3 {
			rankBadge = `<span class="badge badge-warning">🥉 第3名</span>`
		} else {
			rankBadge = fmt.Sprintf(`<span class="badge">第%d名</span>`, rank)
		}

		sceneNameMap := map[string]string{
			// 短文项目场景类型
			"xiaohongshu": "小红书",
			"poster":      "海报",
			"long_post":   "长图文",
			"short_post":  "短图文",
			// 文章场景类型（所有文章统一为"文章"）
			"article": "文章",
			// 其他
			"other": "其他",
		}
		sceneName := sceneNameMap[stat.SceneType]
		if sceneName == "" {
			sceneName = stat.SceneType
		}

		rows = append(rows, fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td><strong>%s</strong></td>
                <td>%d</td>
                <td>%d</td>
                <td>%.2f</td>
                <td>%.1f%%</td>
                <td>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: %.1f%%"></div>
                    </div>
                </td>
                <td>%.1f%%</td>
            </tr>`,
			rankBadge, sceneName, stat.UsageCount, stat.UserCount,
			stat.AvgPerUser, stat.Percentage, stat.Percentage, stat.CompletionRate,
		))
	}

	return fmt.Sprintf(`
        <table>
            <thead>
                <tr>
                    <th>排名</th>
                    <th>场景类型</th>
                    <th>使用次数</th>
                    <th>用户数</th>
                    <th>人均次数</th>
                    <th>占比</th>
                    <th>使用占比</th>
                    <th>完成率</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>`,
		strings.Join(rows, ""),
	)
}

// generateUserTypeComparisonChart 生成用户类型对比图表
func generateUserTypeComparisonChart(stats []analytics.UserTypeSceneStats) string {
	if len(stats) == 0 {
		return `<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>`
	}

	// 构建场景列表和数据
	sceneMap := make(map[string]map[string]int64) // scene -> userType -> count
	for _, stat := range stats {
		if sceneMap[stat.SceneType] == nil {
			sceneMap[stat.SceneType] = make(map[string]int64)
		}
		sceneMap[stat.SceneType][stat.UserType] = stat.UsageCount
	}

	sceneNameMap := map[string]string{
		// 短文项目场景类型
		"xiaohongshu": "小红书",
		"poster":      "海报",
		"long_post":   "长图文",
		"short_post":  "短图文",
		// 文章场景类型（所有文章统一为"文章"）
		"article": "文章",
		// 其他
		"other": "其他",
	}

	scenes := []string{}
	freeUserData := []int64{}
	paidUserData := []int64{}

	for scene, data := range sceneMap {
		sceneName := sceneNameMap[scene]
		if sceneName == "" {
			sceneName = scene
		}
		scenes = append(scenes, sceneName)
		freeUserData = append(freeUserData, data["免费用户"])
		paidUserData = append(paidUserData, data["付费用户"])
	}

	// 生成表格
	tableRows := []string{}
	for i, scene := range scenes {
		freeCount := freeUserData[i]
		paidCount := paidUserData[i]
		total := freeCount + paidCount
		var ratio float64
		if freeCount > 0 {
			ratio = float64(paidCount) / float64(freeCount)
		}

		tableRows = append(tableRows, fmt.Sprintf(`
            <tr>
                <td><strong>%s</strong></td>
                <td>%d</td>
                <td>%d</td>
                <td>%d</td>
                <td><span class="highlight">%.2fx</span></td>
            </tr>`,
			scene, freeCount, paidCount, total, ratio,
		))
	}

	return fmt.Sprintf(`
        <div class="chart-container">
            <canvas id="userComparisonChart"></canvas>
        </div>
        <table>
            <thead>
                <tr>
                    <th>场景</th>
                    <th>免费用户使用次数</th>
                    <th>付费用户使用次数</th>
                    <th>总计</th>
                    <th>倍数差异</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
        <script>
            const ctx = document.getElementById('userComparisonChart');
            new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: %s,
                    datasets: [
                        {
                            label: '免费用户',
                            data: %s,
                            backgroundColor: 'rgba(108, 117, 125, 0.7)',
                            borderColor: 'rgba(108, 117, 125, 1)',
                            borderWidth: 2
                        },
                        {
                            label: '付费用户',
                            data: %s,
                            backgroundColor: 'rgba(102, 126, 234, 0.7)',
                            borderColor: 'rgba(102, 126, 234, 1)',
                            borderWidth: 2
                        }
                    ]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'top',
                            labels: {
                                font: { size: 14 },
                                padding: 20
                            }
                        },
                        title: {
                            display: true,
                            text: '付费 vs 免费用户场景使用对比',
                            font: { size: 18 }
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                font: { size: 12 }
                            }
                        },
                        x: {
                            ticks: {
                                font: { size: 12 }
                            }
                        }
                    }
                }
            });
        </script>`,
		strings.Join(tableRows, ""),
		toJSONArray(scenes),
		toJSONArrayInt64(freeUserData),
		toJSONArrayInt64(paidUserData),
	)
}

// generateProductComparisonTable 生成产品对比表格
func generateProductComparisonTable(stats []analytics.ProductSceneStats) string {
	if len(stats) == 0 {
		return `<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>`
	}

	sceneNameMap := map[string]string{
		// 短文项目场景类型
		"xiaohongshu": "小红书",
		"poster":      "海报",
		"long_post":   "长图文",
		"short_post":  "短图文",
		// 文章场景类型（所有文章统一为"文章"）
		"article": "文章",
		// 其他
		"other": "其他",
	}

	rows := []string{}
	for _, stat := range stats {
		sceneName := sceneNameMap[stat.SceneType]
		if sceneName == "" {
			sceneName = stat.SceneType
		}

		rows = append(rows, fmt.Sprintf(`
            <tr>
                <td><strong>%s</strong></td>
                <td>%s</td>
                <td>%d</td>
                <td>%d</td>
                <td>%.2f</td>
            </tr>`,
			stat.ProductName, sceneName, stat.UsageCount, stat.UserCount, stat.AvgPerUser,
		))
	}

	return fmt.Sprintf(`
        <table>
            <thead>
                <tr>
                    <th>产品名称</th>
                    <th>场景类型</th>
                    <th>使用次数</th>
                    <th>用户数</th>
                    <th>人均次数</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>`,
		strings.Join(rows, ""),
	)
}

// generateExportStatsChart 生成导出统计图表
func generateExportStatsChart(stats []analytics.ExportFormatStats) string {
	if len(stats) == 0 {
		return `<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>`
	}

	labels := []string{}
	data := []int64{}
	percentages := []float64{}

	for _, stat := range stats {
		labels = append(labels, stat.ExportFormat)
		data = append(data, stat.ExportCount)
		percentages = append(percentages, stat.Percentage)
	}

	rows := []string{}
	for i, stat := range stats {
		rows = append(rows, fmt.Sprintf(`
            <tr>
                <td><strong>%s</strong></td>
                <td>%d</td>
                <td>%d</td>
                <td>%.1f%%</td>
                <td>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: %.1f%%"></div>
                    </div>
                </td>
            </tr>`,
			stat.ExportFormat, stat.ExportCount, stat.UserCount,
			stat.Percentage, percentages[i],
		))
	}

	return fmt.Sprintf(`
        <div class="chart-container">
            <canvas id="exportStatsChart"></canvas>
        </div>
        <table>
            <thead>
                <tr>
                    <th>导出格式</th>
                    <th>导出次数</th>
                    <th>用户数</th>
                    <th>占比</th>
                    <th>占比可视化</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
        <script>
            const ctxExport = document.getElementById('exportStatsChart');
            new Chart(ctxExport, {
                type: 'doughnut',
                data: {
                    labels: %s,
                    datasets: [{
                        data: %s,
                        backgroundColor: [
                            'rgba(102, 126, 234, 0.8)',
                            'rgba(118, 75, 162, 0.8)',
                            'rgba(255, 99, 132, 0.8)',
                            'rgba(54, 162, 235, 0.8)',
                            'rgba(255, 206, 86, 0.8)',
                            'rgba(75, 192, 192, 0.8)',
                            'rgba(153, 102, 255, 0.8)',
                            'rgba(255, 159, 64, 0.8)'
                        ],
                        borderWidth: 2,
                        borderColor: '#fff'
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'right',
                            labels: {
                                font: { size: 14 },
                                padding: 15
                            }
                        },
                        title: {
                            display: true,
                            text: '导出格式分布',
                            font: { size: 18 }
                        }
                    }
                }
            });
        </script>`,
		strings.Join(rows, ""),
		toJSONArray(labels),
		toJSONArrayInt64(data),
	)
}

// generateAIFeatureStatsTable 生成AI功能统计表格
func generateAIFeatureStatsTable(stats []analytics.AIFeatureStats) string {
	if len(stats) == 0 {
		return `<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>`
	}

	rows := []string{}
	for _, stat := range stats {
		successBadge := `<span class="badge badge-success">高</span>`
		if stat.SuccessRate < 90 {
			successBadge = `<span class="badge badge-warning">中</span>`
		}
		if stat.SuccessRate < 70 {
			successBadge = `<span class="badge" style="background:#dc3545;color:white;">低</span>`
		}

		rows = append(rows, fmt.Sprintf(`
            <tr>
                <td><strong>%s</strong></td>
                <td>%d</td>
                <td>%d</td>
                <td>%.1f%% %s</td>
                <td>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: %.1f%%"></div>
                    </div>
                </td>
            </tr>`,
			stat.FeatureType, stat.UsageCount, stat.UserCount,
			stat.SuccessRate, successBadge, stat.SuccessRate,
		))
	}

	return fmt.Sprintf(`
        <table>
            <thead>
                <tr>
                    <th>功能类型</th>
                    <th>使用次数</th>
                    <th>用户数</th>
                    <th>成功率</th>
                    <th>成功率可视化</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>`,
		strings.Join(rows, ""),
	)
}

// toJSONArray 将字符串数组转换为JSON数组格式
func toJSONArray(arr []string) string {
	quoted := []string{}
	for _, s := range arr {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, s))
	}
	return "[" + strings.Join(quoted, ",") + "]"
}

// toJSONArrayInt64 将int64数组转换为JSON数组格式
func toJSONArrayInt64(arr []int64) string {
	strs := []string{}
	for _, n := range arr {
		strs = append(strs, fmt.Sprintf("%d", n))
	}
	return "[" + strings.Join(strs, ",") + "]"
}

// GetSceneRanking 获取场景排名（支持每日/每周/每月）
// @Summary 获取场景使用排名
// @Description 获取每日/每周/每月的场景使用排名，区分付费和非付费用户
// @Tags 数据分析
// @Accept json
// @Produce json,html
// @Param period_type query string false "周期类型：daily/weekly/monthly" default(daily)
// @Param days query int false "统计天数（默认30天）" default(30)
// @Param format query string false "返回格式：json/html" default(json)
// @Success 200 {object} analytics.SceneRankingResponse
// @Router /api/v1/admin/analytics/scene-ranking [get]
func GetSceneRanking(c *gin.Context) {
	// 获取查询参数
	periodType := c.DefaultQuery("period_type", "daily")
	daysStr := c.DefaultQuery("days", "30")
	format := c.DefaultQuery("format", "json")

	// 验证周期类型
	if periodType != "daily" && periodType != "weekly" && periodType != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "周期类型无效，支持：daily/weekly/monthly",
		})
		return
	}

	// 解析天数
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "天数参数无效，范围：1-365",
		})
		return
	}

	// 获取数据库连接 - 使用repository.DB而不是从context获取
	service := analytics.NewSceneUsageService(repository.DB)

	// 获取场景排名数据
	ranking, err := service.GetSceneRanking(periodType, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("获取场景排名失败: %v", err),
		})
		return
	}

	// 根据格式返回数据
	if format == "html" {
		html := generateSceneRankingHTML(ranking)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data":    ranking,
		})
	}
}

// generateSceneRankingHTML 生成场景排名HTML报告
func generateSceneRankingHTML(ranking *analytics.SceneRankingResponse) string {
	// 构建每个时期的HTML
	periodSections := []string{}
	for _, period := range ranking.Rankings {
		periodHTML := generatePeriodRankingHTML(&period)
		periodSections = append(periodSections, periodHTML)
	}

	// 周期类型中文映射
	periodTypeMap := map[string]string{
		"daily":   "每日",
		"weekly":  "每周",
		"monthly": "每月",
	}
	periodTypeName := periodTypeMap[ranking.PeriodType]
	if periodTypeName == "" {
		periodTypeName = ranking.PeriodType
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>场景使用排名报告 - %s</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB',
                'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            padding: 20px;
            color: #333;
            line-height: 1.6;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 36px;
            margin-bottom: 10px;
            font-weight: 700;
        }

        .header .subtitle {
            font-size: 16px;
            opacity: 0.9;
        }

        .period-section {
            padding: 40px;
            border-bottom: 2px solid #e9ecef;
        }

        .period-section:last-child {
            border-bottom: none;
        }

        .period-title {
            font-size: 28px;
            font-weight: 700;
            margin-bottom: 30px;
            color: #2c3e50;
            text-align: center;
            padding: 15px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border-radius: 12px;
        }

        .ranking-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
            gap: 30px;
            margin-top: 20px;
        }

        .ranking-card {
            background: #f8f9fa;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }

        .ranking-card-title {
            font-size: 20px;
            font-weight: 700;
            margin-bottom: 20px;
            color: #495057;
            display: flex;
            align-items: center;
            padding-bottom: 10px;
            border-bottom: 2px solid #dee2e6;
        }

        .ranking-card-title .icon {
            font-size: 24px;
            margin-right: 10px;
        }

        .ranking-item {
            display: flex;
            align-items: center;
            padding: 15px;
            margin-bottom: 10px;
            background: white;
            border-radius: 8px;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .ranking-item:hover {
            transform: translateX(5px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .rank-badge {
            width: 40px;
            height: 40px;
            border-radius: 50%%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 700;
            font-size: 16px;
            margin-right: 15px;
            flex-shrink: 0;
        }

        .rank-1 {
            background: linear-gradient(135deg, #FFD700, #FFA500);
            color: white;
        }

        .rank-2 {
            background: linear-gradient(135deg, #C0C0C0, #A8A8A8);
            color: white;
        }

        .rank-3 {
            background: linear-gradient(135deg, #CD7F32, #B8860B);
            color: white;
        }

        .rank-other {
            background: #e9ecef;
            color: #6c757d;
        }

        .scene-info {
            flex: 1;
        }

        .scene-name {
            font-size: 16px;
            font-weight: 600;
            color: #2c3e50;
            margin-bottom: 5px;
        }

        .scene-stats {
            font-size: 13px;
            color: #6c757d;
        }

        .scene-percentage {
            font-size: 18px;
            font-weight: 700;
            color: #667eea;
            margin-left: 15px;
        }

        .footer {
            padding: 30px 40px;
            background: #f8f9fa;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
        }

        @media (max-width: 768px) {
            .header h1 {
                font-size: 24px;
            }
            .ranking-grid {
                grid-template-columns: 1fr;
            }
            .period-section {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏆 场景使用排名报告</h1>
            <div class="subtitle">
                %s排名 | 数据周期：%s ~ %s
            </div>
        </div>

        %s

        <div class="footer">
            <p>© 2026 01agent - 数据驱动产品决策</p>
        </div>
    </div>
</body>
</html>`,
		periodTypeName,
		periodTypeName, ranking.StartDate, ranking.EndDate,
		strings.Join(periodSections, "\n"),
	)
}

// generatePeriodRankingHTML 生成单个时期的排名HTML
func generatePeriodRankingHTML(period *analytics.PeriodSceneRanking) string {
	// 生成付费用户排名
	paidHTML := generateRankingListHTML(period.PaidUsers)
	
	// 生成免费用户排名
	freeHTML := generateRankingListHTML(period.FreeUsers)
	
	// 生成所有用户排名
	allHTML := generateRankingListHTML(period.AllUsers)

	return fmt.Sprintf(`
        <div class="period-section">
            <h2 class="period-title">📅 %s</h2>
            <div class="ranking-grid">
                <div class="ranking-card">
                    <h3 class="ranking-card-title">
                        <span class="icon">💎</span> 付费用户排名
                    </h3>
                    %s
                </div>
                <div class="ranking-card">
                    <h3 class="ranking-card-title">
                        <span class="icon">👤</span> 免费用户排名
                    </h3>
                    %s
                </div>
                <div class="ranking-card">
                    <h3 class="ranking-card-title">
                        <span class="icon">🌟</span> 总体排名
                    </h3>
                    %s
                </div>
            </div>
        </div>`,
		period.Period,
		paidHTML,
		freeHTML,
		allHTML,
	)
}

// generateRankingListHTML 生成排名列表HTML
func generateRankingListHTML(rankings []analytics.SceneRankingItem) string {
	if len(rankings) == 0 {
		return `<div style="text-align: center; padding: 20px; color: #6c757d;">暂无数据</div>`
	}

	items := []string{}
	for _, item := range rankings {
		rankClass := "rank-other"
		if item.Rank == 1 {
			rankClass = "rank-1"
		} else if item.Rank == 2 {
			rankClass = "rank-2"
		} else if item.Rank == 3 {
			rankClass = "rank-3"
		}

		items = append(items, fmt.Sprintf(`
            <div class="ranking-item">
                <div class="rank-badge %s">%d</div>
                <div class="scene-info">
                    <div class="scene-name">%s</div>
                    <div class="scene-stats">使用 %d 次 | %d 人</div>
                </div>
                <div class="scene-percentage">%.1f%%</div>
            </div>`,
			rankClass,
			item.Rank,
			item.SceneName,
			item.UsageCount,
			item.UserCount,
			item.Percentage,
		))
	}

	return strings.Join(items, "\n")
}

// GetUserUsageRanking 获取用户使用排名
// @Summary 获取用户使用排名
// @Description 获取每日/每周/每月使用次数最多的用户排名，显示用户主要使用的场景占比
// @Tags 数据分析
// @Accept json
// @Produce json,html
// @Param period_type query string false "周期类型：daily/weekly/monthly" default(daily)
// @Param days query int false "统计天数（默认30天）" default(30)
// @Param top query int false "每个时期显示的用户数量（默认10）" default(10)
// @Param format query string false "返回格式：json/html" default(json)
// @Success 200 {object} analytics.UserRankingResponse
// @Router /api/v1/admin/analytics/user-ranking [get]
func GetUserUsageRanking(c *gin.Context) {
	// 获取查询参数
	periodType := c.DefaultQuery("period_type", "daily")
	daysStr := c.DefaultQuery("days", "30")
	topStr := c.DefaultQuery("top", "10")
	format := c.DefaultQuery("format", "json")

	// 验证周期类型
	if periodType != "daily" && periodType != "weekly" && periodType != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "周期类型无效，支持：daily/weekly/monthly",
		})
		return
	}

	// 解析参数
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "天数参数无效，范围：1-365",
		})
		return
	}

	top, err := strconv.Atoi(topStr)
	if err != nil || top <= 0 || top > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "top参数无效，范围：1-100",
		})
		return
	}

	// 获取数据库连接
	service := analytics.NewSceneUsageService(repository.DB)

	// 获取用户使用排名数据
	ranking, err := service.GetUserUsageRanking(periodType, days, top)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("获取用户使用排名失败: %v", err),
		})
		return
	}

	// 根据格式返回数据
	if format == "html" {
		html := generateUserRankingHTML(ranking)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data":    ranking,
		})
	}
}

// generateUserRankingHTML 生成用户排名HTML报告
func generateUserRankingHTML(ranking *analytics.UserRankingResponse) string {
	// 构建每个时期的HTML
	periodSections := []string{}
	for _, period := range ranking.Rankings {
		periodHTML := generatePeriodUserRankingHTML(&period)
		periodSections = append(periodSections, periodHTML)
	}

	// 周期类型中文映射
	periodTypeMap := map[string]string{
		"daily":   "每日",
		"weekly":  "每周",
		"monthly": "每月",
	}
	periodTypeName := periodTypeMap[ranking.PeriodType]
	if periodTypeName == "" {
		periodTypeName = ranking.PeriodType
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>用户使用排名报告 - %s</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB',
                'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            padding: 20px;
            color: #333;
            line-height: 1.6;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 36px;
            margin-bottom: 10px;
            font-weight: 700;
        }

        .header .subtitle {
            font-size: 16px;
            opacity: 0.9;
        }

        .period-section {
            padding: 40px;
            border-bottom: 2px solid #e9ecef;
        }

        .period-section:last-child {
            border-bottom: none;
        }

        .period-title {
            font-size: 28px;
            font-weight: 700;
            margin-bottom: 30px;
            color: #2c3e50;
            text-align: center;
            padding: 15px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border-radius: 12px;
        }

        .ranking-list {
            display: grid;
            gap: 20px;
        }

        .ranking-item {
            display: flex;
            align-items: center;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 12px;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .ranking-item:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .rank-badge {
            width: 50px;
            height: 50px;
            border-radius: 50%%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 700;
            font-size: 20px;
            margin-right: 20px;
            flex-shrink: 0;
        }

        .rank-1 {
            background: linear-gradient(135deg, #FFD700, #FFA500);
            color: white;
        }

        .rank-2 {
            background: linear-gradient(135deg, #C0C0C0, #A8A8A8);
            color: white;
        }

        .rank-3 {
            background: linear-gradient(135deg, #CD7F32, #B8860B);
            color: white;
        }

        .rank-other {
            background: #e9ecef;
            color: #6c757d;
        }

        .user-info {
            flex: 1;
            min-width: 0;
        }

        .user-header {
            display: flex;
            align-items: center;
            margin-bottom: 12px;
        }

        .user-avatar {
            width: 40px;
            height: 40px;
            border-radius: 50%%;
            margin-right: 12px;
            object-fit: cover;
            background: #dee2e6;
        }

        .user-name {
            font-size: 18px;
            font-weight: 600;
            color: #2c3e50;
            margin-right: 10px;
        }

        .user-id {
            font-size: 12px;
            color: #6c757d;
            background: #e9ecef;
            padding: 2px 8px;
            border-radius: 4px;
        }

        .user-meta {
            display: flex;
            gap: 8px;
            align-items: center;
            margin-top: 6px;
            flex-wrap: wrap;
        }

        .user-badge {
            font-size: 11px;
            padding: 2px 8px;
            border-radius: 12px;
            font-weight: 600;
        }

        .badge-vip {
            background: linear-gradient(135deg, #ffd700, #ffa500);
            color: white;
        }

        .badge-free {
            background: #e9ecef;
            color: #6c757d;
        }

        .user-phone {
            font-size: 11px;
            color: #6c757d;
        }

        .usage-count {
            font-size: 24px;
            font-weight: 700;
            color: #667eea;
            margin-left: auto;
            padding: 0 20px;
        }

        .scene-distribution {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }

        .scene-badge {
            display: flex;
            align-items: center;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 13px;
            background: white;
            border: 2px solid #dee2e6;
        }

        .scene-badge .scene-name {
            font-weight: 600;
            margin-right: 8px;
        }

        .scene-badge .scene-count {
            color: #667eea;
            font-weight: 700;
            margin-right: 4px;
        }

        .scene-badge .scene-percentage {
            color: #6c757d;
            font-size: 11px;
        }

        .scene-xiaohongshu {
            border-color: #ff2442;
            color: #ff2442;
        }

        .scene-article {
            border-color: #52c41a;
            color: #52c41a;
        }

        .scene-other {
            border-color: #8c8c8c;
            color: #8c8c8c;
        }

        .footer {
            padding: 30px 40px;
            background: #f8f9fa;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
        }

        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #6c757d;
        }

        .empty-state-icon {
            font-size: 64px;
            margin-bottom: 20px;
        }

        @media (max-width: 768px) {
            .header h1 {
                font-size: 24px;
            }
            .period-section {
                padding: 20px;
            }
            .ranking-item {
                flex-direction: column;
                align-items: flex-start;
            }
            .usage-count {
                margin-left: 0;
                padding: 10px 0;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👥 用户使用排名报告</h1>
            <div class="subtitle">
                %s排名 | 数据周期：%s ~ %s
            </div>
        </div>

        %s

        <div class="footer">
            <p>© 2026 01agent - 数据驱动产品决策</p>
        </div>
    </div>
</body>
</html>`,
		periodTypeName,
		periodTypeName, ranking.StartDate, ranking.EndDate,
		strings.Join(periodSections, "\n"),
	)
}

// generatePeriodUserRankingHTML 生成单个时期的用户排名HTML
func generatePeriodUserRankingHTML(period *analytics.PeriodUserRanking) string {
	if len(period.Rankings) == 0 {
		return fmt.Sprintf(`
        <div class="period-section">
            <h2 class="period-title">📅 %s</h2>
            <div class="empty-state">
                <div class="empty-state-icon">📊</div>
                <div>该时期暂无数据</div>
            </div>
        </div>`,
			period.Period,
		)
	}

	rankingItems := []string{}
	for _, item := range period.Rankings {
		rankClass := "rank-other"
		if item.Rank == 1 {
			rankClass = "rank-1"
		} else if item.Rank == 2 {
			rankClass = "rank-2"
		} else if item.Rank == 3 {
			rankClass = "rank-3"
		}

		// 用户头像
		avatarHTML := `<div class="user-avatar"></div>`
		if item.Avatar != nil && *item.Avatar != "" {
			avatarHTML = fmt.Sprintf(`<img src="%s" class="user-avatar" alt="avatar">`, *item.Avatar)
		}

		// 用户昵称
		displayName := item.Username
		if item.Nickname != nil && *item.Nickname != "" {
			displayName = *item.Nickname
		}

		// 会员状态
		vipBadge := `<span class="user-badge badge-free">免费</span>`
		if item.VipStatus == "vip" {
			vipBadge = fmt.Sprintf(`<span class="user-badge badge-vip">VIP Lv%d</span>`, item.VipLevel)
		}

		// 手机号
		phoneHTML := ""
		if item.Phone != nil && *item.Phone != "" {
			phoneHTML = fmt.Sprintf(`<span class="user-phone">📱 %s</span>`, *item.Phone)
		}

		// 场景分布
		sceneHTML := []string{}
		for _, scene := range item.SceneDistribution {
			sceneClass := "scene-other"
			if scene.SceneType == "xiaohongshu" {
				sceneClass = "scene-xiaohongshu"
			} else if scene.SceneType == "article" {
				sceneClass = "scene-article"
			}

			sceneHTML = append(sceneHTML, fmt.Sprintf(`
                <div class="scene-badge %s">
                    <span class="scene-name">%s</span>
                    <span class="scene-count">%d</span>
                    <span class="scene-percentage">(%.1f%%)</span>
                </div>`,
				sceneClass,
				scene.SceneName,
				scene.Count,
				scene.Percentage,
			))
		}

		rankingItems = append(rankingItems, fmt.Sprintf(`
            <div class="ranking-item">
                <div class="rank-badge %s">%d</div>
                <div class="user-info">
                    <div class="user-header">
                        %s
                        <span class="user-name">%s</span>
                        <span class="user-id">ID: %s</span>
                    </div>
                    <div class="user-meta">
                        %s
                        %s
                    </div>
                    <div class="scene-distribution">
                        %s
                    </div>
                </div>
                <div class="usage-count">%d 次</div>
            </div>`,
			rankClass,
			item.Rank,
			avatarHTML,
			displayName,
			item.UserID,
			vipBadge,
			phoneHTML,
			strings.Join(sceneHTML, "\n"),
			item.TotalUsage,
		))
	}

	return fmt.Sprintf(`
        <div class="period-section">
            <h2 class="period-title">📅 %s</h2>
            <div class="ranking-list">
                %s
            </div>
        </div>`,
		period.Period,
		strings.Join(rankingItems, "\n"),
	)
}

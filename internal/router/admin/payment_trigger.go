package admin

import (
	"01agent_server/internal/service/analytics"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetFirstPaymentTriggerAnalysis 获取首充触发点分析
// @Summary 获取首充触发点分析
// @Description 分析付费用户在首次充值前的积分消耗情况，包括消耗场景和购买的产品
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期 YYYY-MM-DD"
// @Param end_date query string false "结束日期 YYYY-MM-DD"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/payment-trigger [get]
func GetFirstPaymentTriggerAnalysis(c *gin.Context) {
	// 解析日期参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 默认时间范围：最近30天
	loc := time.FixedZone("CST", 8*60*60)
	endDate := time.Now().In(loc)
	startDate := endDate.AddDate(0, 0, -30)

	var err error
	if startDateStr != "" {
		startDate, err = time.ParseInLocation("2006-01-02", startDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "开始日期格式错误，应为 YYYY-MM-DD",
			})
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.ParseInLocation("2006-01-02", endDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "结束日期格式错误，应为 YYYY-MM-DD",
			})
			return
		}
	}

	// 验证日期范围
	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "开始日期不能晚于结束日期",
		})
		return
	}

	// 限制最大查询范围为90天
	if endDate.Sub(startDate).Hours() > 90*24 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "查询范围不能超过90天",
		})
		return
	}

	// 调用服务层
	service := analytics.NewPaymentTriggerService()
	summary, analyses, err := service.GetFirstPaymentAnalysis(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "获取首充触发点分析成功",
		"data": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"summary":    summary,
			"details":    analyses,
		},
	})
}

// GetUserFirstPaymentTrigger 获取单个用户的首充触发点分析
// @Summary 获取单个用户的首充触发点分析
// @Description 查询指定用户在首次充值前的积分消耗情况
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/payment-trigger/user [get]
func GetUserFirstPaymentTrigger(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "用户ID不能为空",
		})
		return
	}

	// 调用服务层
	service := analytics.NewPaymentTriggerService()
	analysis, err := service.GetUserFirstPaymentTrigger(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "获取用户首充触发点分析成功",
		"data": analysis,
	})
}

// GetPaymentTriggerInsights 获取首充触发点洞察（简化版）
// @Summary 获取首充触发点核心洞察
// @Description 返回最近30天首充用户的核心指标，包括平均首充前消耗、Top场景和热门产品
// @Tags 数据分析
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/analytics/payment-trigger/insights [get]
func GetPaymentTriggerInsights(c *gin.Context) {
	// 固定查询最近30天
	loc := time.FixedZone("CST", 8*60*60)
	endDate := time.Now().In(loc)
	startDate := endDate.AddDate(0, 0, -30)

	// 调用服务层
	service := analytics.NewPaymentTriggerService()
	summary, _, err := service.GetFirstPaymentAnalysis(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "分析失败: " + err.Error(),
		})
		return
	}

	// 只返回核心洞察数据（不返回详细列表）
	topScenes := summary.TopScenes
	if len(topScenes) > 5 {
		topScenes = topScenes[:5] // 只取前5个场景
	}

	topProducts := summary.ProductDistribution
	if len(topProducts) > 5 {
		topProducts = topProducts[:5] // 只取前5个产品
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "获取首充触发点洞察成功",
		"data": gin.H{
			"period":                       "最近30天",
			"total_paying_users":           summary.TotalPayingUsers,
			"avg_credits_before_payment":   summary.AvgCreditsBeforePayment,
			"median_credits_before_payment": summary.MedianCreditsBeforePayment,
			"top_scenes":                   topScenes,
			"top_products":                 topProducts,
			"credit_range_distribution":    summary.CreditRangeDistribution,
			"insights": []string{
				formatInsight("平均首充前消耗", summary.AvgCreditsBeforePayment, "积分"),
				formatInsight("中位数消耗", float64(summary.MedianCreditsBeforePayment), "积分"),
				formatTopScene(topScenes),
				formatTopProduct(topProducts),
			},
		},
	})
}

// 辅助函数：格式化洞察文本
func formatInsight(name string, value float64, unit string) string {
	return fmt.Sprintf("%s: %.1f%s", name, value, unit)
}

func formatTopScene(scenes []analytics.SceneConsumption) string {
	if len(scenes) == 0 {
		return "暂无场景数据"
	}
	return "最常用场景: " + scenes[0].ServiceName
}

func formatTopProduct(products []analytics.ProductStats) string {
	if len(products) == 0 {
		return "暂无产品数据"
	}
	return "最受欢迎产品: " + products[0].ProductName
}

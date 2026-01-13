package admin

import (
	"01agent_server/internal/service/analytics"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetRenewalRanking 获取续费用户排行榜
// @Summary 获取续费用户排行榜
// @Description 获取续费用户排行榜（续费次数≥2次的用户）
// @Tags Admin-Analytics
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Param sort_by query string false "排序方式: count(续费次数，默认) | amount(续费金额)"
// @Param limit query int false "返回数量限制，默认100，最大1000"
// @Success 200 {object} map[string]interface{} "成功返回续费排行榜"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /admin/analytics/renewal-ranking [get]
func GetRenewalRanking(c *gin.Context) {
	// 解析查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	sortBy := c.DefaultQuery("sort_by", "count")
	limitStr := c.DefaultQuery("limit", "100")

	// 解析limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	// 解析日期参数
	var startDate, endDate *time.Time
	loc := time.FixedZone("CST", 8*60*60)

	if startDateStr != "" {
		t, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "开始日期格式错误，应为 YYYY-MM-DD",
			})
			return
		}
		startDate = &t
	}

	if endDateStr != "" {
		t, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "结束日期格式错误，应为 YYYY-MM-DD",
			})
			return
		}
		endDate = &t
	}

	// 验证排序方式
	if sortBy != "count" && sortBy != "amount" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "排序方式只能是 count 或 amount",
		})
		return
	}

	// 调用服务
	service := analytics.NewRenewalService()
	rankings, err := service.GetRenewalRanking(startDate, endDate, sortBy, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"rankings":   rankings,
			"total":      len(rankings),
			"sort_by":    sortBy,
			"start_date": startDateStr,
			"end_date":   endDateStr,
		},
	})
}

// GetRenewalSummary 获取续费统计汇总
// @Summary 获取续费统计汇总
// @Description 获取续费率、续费次数分布、续费产品分布等统计信息
// @Tags Admin-Analytics
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "成功返回续费统计汇总"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /admin/analytics/renewal-summary [get]
func GetRenewalSummary(c *gin.Context) {
	// 解析查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 解析日期参数
	var startDate, endDate *time.Time
	loc := time.FixedZone("CST", 8*60*60)

	if startDateStr != "" {
		t, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "开始日期格式错误，应为 YYYY-MM-DD",
			})
			return
		}
		startDate = &t
	}

	if endDateStr != "" {
		t, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "结束日期格式错误，应为 YYYY-MM-DD",
			})
			return
		}
		endDate = &t
	}

	// 调用服务
	service := analytics.NewRenewalService()
	summary, err := service.GetRenewalSummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"summary":    summary,
			"start_date": startDateStr,
			"end_date":   endDateStr,
		},
	})
}

// GetUserRenewalDetail 获取单个用户的续费详情
// @Summary 获取单个用户的续费详情
// @Description 获取指定用户的完整续费历史、时间线和统计信息
// @Tags Admin-Analytics
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} map[string]interface{} "成功返回用户续费详情"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户未找到"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /admin/analytics/renewal-detail/:user_id [get]
func GetUserRenewalDetail(c *gin.Context) {
	userID := c.Param("user_id")
	
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	// 调用服务
	service := analytics.NewRenewalService()
	detail, err := service.GetUserRenewalDetail(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    detail,
	})
}

package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/service/analytics"
	"time"

	"github.com/gin-gonic/gin"
)

// MembershipHandler 会员统计处理器
type MembershipHandler struct {
	membershipService *analytics.MembershipService
}

// NewMembershipHandler 创建会员统计处理器
func NewMembershipHandler() *MembershipHandler {
	return &MembershipHandler{
		membershipService: analytics.NewMembershipService(),
	}
}

// GetMembershipOverview 获取会员购买概览
// @Summary 获取会员购买概览
// @Description 统计各会员类型的购买数量、收入、占比等信息
// @Tags admin-membership
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string false "结束日期，格式：YYYY-MM-DD"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/membership/overview [get]
func (h *MembershipHandler) GetMembershipOverview(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	loc := time.FixedZone("CST", 8*60*60)
	var startDate, endDate *time.Time

	// 解析开始日期，如果没有提供则默认为2025年7月1日（产品上线日期）
	if startDateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始日期格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		startDate = &parsed
	} else {
		// 默认从2025年7月1日开始（产品上线日期，之前的订单都是测试数据）
		defaultStartDate := time.Date(2025, 7, 1, 0, 0, 0, 0, loc)
		startDate = &defaultStartDate
	}

	// 解析结束日期，如果没有提供则默认为当前日期
	if endDateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束日期格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		// 设置为当天的23:59:59
		endOfDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, loc)
		endDate = &endOfDay
	} else {
		// 默认为当前日期
		now := time.Now().In(loc)
		endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
		endDate = &endOfDay
	}

	// 获取概览数据
	overview, err := h.membershipService.GetMembershipOverview(startDate, endDate)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取会员概览失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取会员概览成功", overview)
}

// GetMembershipTrend 获取会员购买趋势
// @Summary 获取会员购买趋势
// @Description 获取指定时间范围内的会员购买趋势数据，支持按天/周/月统计，最大跨度1年
// @Tags admin-membership
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string true "结束日期，格式：YYYY-MM-DD"
// @Param period query string false "统计周期，day/week/month，默认为day"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/membership/trend [get]
func (h *MembershipHandler) GetMembershipTrend(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	period := c.DefaultQuery("period", "day")

	if startDateStr == "" || endDateStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "开始日期和结束日期不能为空"))
		return
	}

	// 验证period参数
	if period != "day" && period != "week" && period != "month" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "统计周期参数错误，请使用day/week/month"))
		return
	}

	// 解析日期
	loc := time.FixedZone("CST", 8*60*60)
	startDate, err1 := time.ParseInLocation("2006-01-02", startDateStr, loc)
	endDate, err2 := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err1 != nil || err2 != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}

	// 验证日期范围（最大1年）
	maxDuration := 365 * 24 * time.Hour
	if endDate.Sub(startDate) > maxDuration {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期范围不能超过1年"))
		return
	}

	// 获取趋势数据
	trend, err := h.membershipService.GetMembershipTrend(startDate, endDate, period)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取会员趋势失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取会员趋势成功", trend)
}

// GetProductSalesTrend 获取产品销售趋势（折线图数据）
// @Summary 获取产品销售趋势
// @Description 获取指定时间范围内的产品销售趋势，每个产品一条折线，支持会员和积分套餐
// @Tags admin-membership
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string true "结束日期，格式：YYYY-MM-DD"
// @Param period query string false "统计周期，可选值：day, week, month，默认为day"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/membership/product-trend [get]
func (h *MembershipHandler) GetProductSalesTrend(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	period := c.DefaultQuery("period", "day")

	if startDateStr == "" || endDateStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "开始日期和结束日期不能为空"))
		return
	}

	loc := time.FixedZone("CST", 8*60*60)
	startDate, err1 := time.ParseInLocation("2006-01-02", startDateStr, loc)
	endDate, err2 := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err1 != nil || err2 != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}

	trend, err := h.membershipService.GetProductSalesTrend(startDate, endDate, period)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取产品销售趋势失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取产品销售趋势成功", trend)
}

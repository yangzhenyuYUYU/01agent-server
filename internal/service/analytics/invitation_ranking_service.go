package analytics

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"fmt"
)

// InvitationRankingService 邀请排名服务
type InvitationRankingService struct{}

// NewInvitationRankingService 创建邀请排名服务
func NewInvitationRankingService() *InvitationRankingService {
	return &InvitationRankingService{}
}

// GetInvitationRanking 获取邀请用户排行榜（实时查询版）
// sortBy: total（按总邀请数，默认）, paid（按有效邀请数）, commission（按佣金）
// limit: 返回数量限制，默认50
func (s *InvitationRankingService) GetInvitationRanking(sortBy string, limit int) ([]models.InvitationRankingResponse, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	// 实时查询邀请统计
	query := repository.DB.Table("invitation_relations as ir").
		Select(`
			ir.inviter_id as user_id,
			u.nickname,
			u.avatar,
			COUNT(DISTINCT ir.invitee_id) as total_invitations,
			COUNT(DISTINCT CASE WHEN comm.relation_id IS NOT NULL THEN ir.invitee_id END) as paid_invitations,
			COUNT(DISTINCT CASE WHEN ir.created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN ir.invitee_id END) as recent_30d_invitations,
			COALESCE(cr.total_commission, 0) as total_commission,
			MIN(ir.created_at) as first_invitation_date,
			MAX(ir.created_at) as last_invitation_date
		`).
		Joins("LEFT JOIN user u ON ir.inviter_id = u.user_id").
		Joins("LEFT JOIN (SELECT DISTINCT relation_id FROM commission_records) comm ON ir.id = comm.relation_id").
		Joins("LEFT JOIN (SELECT user_id, SUM(amount) as total_commission FROM commission_records GROUP BY user_id) cr ON ir.inviter_id = cr.user_id").
		Group("ir.inviter_id, u.nickname, u.avatar, cr.total_commission")

	// 排序（默认按总邀请数）
	switch sortBy {
	case "paid":
		query = query.Order("paid_invitations DESC, total_invitations DESC")
	case "commission":
		query = query.Order("total_commission DESC, total_invitations DESC")
	default: // total - 默认按总邀请数排序
		query = query.Order("total_invitations DESC, paid_invitations DESC")
	}

	query = query.Limit(limit)

	// 执行查询
	var results []struct {
		UserID               string    `gorm:"column:user_id"`
		Nickname             *string   `gorm:"column:nickname"`
		Avatar               *string   `gorm:"column:avatar"`
		TotalInvitations     int       `gorm:"column:total_invitations"`
		PaidInvitations      int       `gorm:"column:paid_invitations"`
		Recent30dInvitations int       `gorm:"column:recent_30d_invitations"`
		TotalCommission      float64   `gorm:"column:total_commission"`
		FirstInvitationDate  *string   `gorm:"column:first_invitation_date"`
		LastInvitationDate   *string   `gorm:"column:last_invitation_date"`
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询邀请排名失败: %w", err)
	}

	// 构建响应
	rankings := make([]models.InvitationRankingResponse, 0, len(results))
	for i, result := range results {
		resp := models.InvitationRankingResponse{
			Rank:                 i + 1,
			UserID:               result.UserID,
			Nickname:             result.Nickname,
			Avatar:               result.Avatar,
			TotalInvitations:     result.TotalInvitations,
			PaidInvitations:      result.PaidInvitations,
			Recent30dInvitations: result.Recent30dInvitations,
			TotalCommission:      result.TotalCommission,
			PersonalViralRate:    float64(result.TotalInvitations),
		}

		// 计算质量分和活跃度分
		if result.TotalInvitations > 0 {
			resp.InvitationQualityScore = float64(result.PaidInvitations) * 100.0 / float64(result.TotalInvitations)
			resp.ActivityScore = float64(result.Recent30dInvitations) * 100.0 / float64(result.TotalInvitations)
		}

		// 处理时间字段
		if result.FirstInvitationDate != nil {
			resp.FirstInvitationDate = *result.FirstInvitationDate
		}
		if result.LastInvitationDate != nil {
			resp.LastInvitationDate = *result.LastInvitationDate
		}

		rankings = append(rankings, resp)
	}

	return rankings, nil
}

// GetSystemMetrics 获取系统级邀请指标（实时查询版）
func (s *InvitationRankingService) GetSystemMetrics() (*models.InvitationSystemMetrics, error) {
	var metrics models.InvitationSystemMetrics

	// 查询总用户数
	var totalUsers int64
	if err := repository.DB.Table("user").Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("查询总用户数失败: %w", err)
	}
	metrics.TotalUsers = int(totalUsers)

	// 实时查询邀请统计数据
	var stats struct {
		ActiveInviters       int     `gorm:"column:active_inviters"`
		TotalInvitations     int     `gorm:"column:total_invitations"`
		PaidInvitations      int     `gorm:"column:paid_invitations"`
		TotalCommission      float64 `gorm:"column:total_commission"`
		Recent30dInvitations int     `gorm:"column:recent_30d_invitations"`
		Recent7dInvitations  int     `gorm:"column:recent_7d_invitations"`
	}

	err := repository.DB.Table("invitation_relations as ir").
		Select(`
			COUNT(DISTINCT ir.inviter_id) as active_inviters,
			COUNT(DISTINCT ir.invitee_id) as total_invitations,
			COUNT(DISTINCT CASE WHEN comm.relation_id IS NOT NULL THEN ir.invitee_id END) as paid_invitations,
			(SELECT COALESCE(SUM(amount), 0) FROM commission_records) as total_commission,
			COUNT(DISTINCT CASE WHEN ir.created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN ir.invitee_id END) as recent_30d_invitations,
			COUNT(DISTINCT CASE WHEN ir.created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) THEN ir.invitee_id END) as recent_7d_invitations
		`).
		Joins("LEFT JOIN (SELECT DISTINCT relation_id FROM commission_records) comm ON ir.id = comm.relation_id").
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("查询邀请统计数据失败: %w", err)
	}

	// 填充指标
	metrics.ActiveInviters = stats.ActiveInviters
	metrics.TotalInvitations = stats.TotalInvitations
	metrics.PaidInvitations = stats.PaidInvitations
	metrics.TotalCommission = stats.TotalCommission
	metrics.Recent30dInvitations = stats.Recent30dInvitations
	metrics.Recent7dInvitations = stats.Recent7dInvitations

	// 计算分享率
	if metrics.TotalUsers > 0 {
		metrics.ShareRate = float64(metrics.ActiveInviters) * 100.0 / float64(metrics.TotalUsers)
	}

	// 计算平均裂变系数
	if metrics.ActiveInviters > 0 {
		metrics.AvgViralCoefficient = float64(metrics.TotalInvitations) / float64(metrics.ActiveInviters)
	}

	// 计算有效邀请转化率
	if metrics.TotalInvitations > 0 {
		metrics.ConversionRate = float64(metrics.PaidInvitations) * 100.0 / float64(metrics.TotalInvitations)
	}

	// 计算平均每用户佣金
	if metrics.ActiveInviters > 0 {
		metrics.AvgCommissionPerUser = metrics.TotalCommission / float64(metrics.ActiveInviters)
	}

	return &metrics, nil
}

// GetUserInvitationDetail 获取用户的邀请详情
func (s *InvitationRankingService) GetUserInvitationDetail(userID string, page, pageSize int) ([]models.InvitationDetailInfo, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 计算总数
	var total int64
	if err := repository.DB.Table("invitation_relations").
		Where("inviter_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询邀请总数失败: %w", err)
	}

	if total == 0 {
		return []models.InvitationDetailInfo{}, 0, nil
	}

	// 查询邀请列表
	offset := (page - 1) * pageSize
	query := repository.DB.Table("invitation_relations as ir").
		Select(`
			ir.invitee_id,
			u.nickname,
			u.avatar,
			ir.created_at as invited_date,
			CASE WHEN MAX(comm.id) IS NOT NULL THEN 1 ELSE 0 END as is_paid,
			COUNT(DISTINCT comm.id) as order_count,
			COALESCE(SUM(comm.amount), 0) as total_payment
		`).
		Joins("LEFT JOIN user u ON ir.invitee_id = u.user_id").
		Joins("LEFT JOIN commission_records comm ON ir.id = comm.relation_id").
		Where("ir.inviter_id = ?", userID).
		Group("ir.invitee_id, u.nickname, u.avatar, ir.created_at").
		Order("ir.created_at DESC").
		Offset(offset).
		Limit(pageSize)

	var results []struct {
		InviteeID    string  `gorm:"column:invitee_id"`
		Nickname     *string `gorm:"column:nickname"`
		Avatar       *string `gorm:"column:avatar"`
		InvitedDate  string  `gorm:"column:invited_date"`
		IsPaid       int     `gorm:"column:is_paid"`
		OrderCount   int     `gorm:"column:order_count"`
		TotalPayment float64 `gorm:"column:total_payment"`
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("查询邀请列表失败: %w", err)
	}

	// 构建响应
	details := make([]models.InvitationDetailInfo, 0, len(results))
	for _, result := range results {
		detail := models.InvitationDetailInfo{
			InviteeID:    result.InviteeID,
			Nickname:     result.Nickname,
			Avatar:       result.Avatar,
			InvitedDate:  result.InvitedDate,
			IsPaid:       result.IsPaid > 0,
			OrderCount:   result.OrderCount,
			TotalPayment: result.TotalPayment,
		}
		details = append(details, detail)
	}

	return details, total, nil
}

// RefreshCache 实时查询版本不需要刷新缓存
func (s *InvitationRankingService) RefreshCache() error {
	// 实时查询版本，数据始终是最新的，无需刷新
	return nil
}

// GetCacheStatus 实时查询版本返回数据统计
func (s *InvitationRankingService) GetCacheStatus() (map[string]interface{}, error) {
	var stats struct {
		TotalInviters int `gorm:"column:total_inviters"`
		TotalRelations int `gorm:"column:total_relations"`
	}

	err := repository.DB.Table("invitation_relations").
		Select(`
			COUNT(DISTINCT inviter_id) as total_inviters,
			COUNT(*) as total_relations
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("查询统计数据失败: %w", err)
	}

	return map[string]interface{}{
		"query_mode":      "realtime",
		"total_inviters":  stats.TotalInviters,
		"total_relations": stats.TotalRelations,
		"status":          "active",
	}, nil
}

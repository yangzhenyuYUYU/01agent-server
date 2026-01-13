package models

import "time"

// InvitationRankingCache 邀请排名缓存表模型
type InvitationRankingCache struct {
	// 基础信息
	UserID string `json:"user_id" gorm:"primaryKey;column:user_id;type:varchar(50)" description:"邀请人用户ID"`

	// 邀请统计
	TotalInvitations      int `json:"total_invitations" gorm:"column:total_invitations;default:0" description:"总邀请人数"`
	PaidInvitations       int `json:"paid_invitations" gorm:"column:paid_invitations;default:0" description:"付费邀请人数（有效邀请）"`
	Recent30dInvitations  int `json:"recent_30d_invitations" gorm:"column:recent_30d_invitations;default:0" description:"近30天邀请人数"`
	Recent7dInvitations   int `json:"recent_7d_invitations" gorm:"column:recent_7d_invitations;default:0" description:"近7天邀请人数"`

	// 裂变指标
	PersonalViralRate      float64 `json:"personal_viral_rate" gorm:"column:personal_viral_rate;type:decimal(10,2);default:0" description:"个人裂变率（=总邀请人数）"`
	InvitationGrowthRate   float64 `json:"invitation_growth_rate" gorm:"column:invitation_growth_rate;type:decimal(10,4);default:0" description:"邀请增长率（近30天/总数）"`

	// 佣金统计
	TotalCommission   float64 `json:"total_commission" gorm:"column:total_commission;type:decimal(10,2);default:0" description:"总佣金金额"`
	PendingCommission float64 `json:"pending_commission" gorm:"column:pending_commission;type:decimal(10,2);default:0" description:"待发放佣金"`
	IssuedCommission  float64 `json:"issued_commission" gorm:"column:issued_commission;type:decimal(10,2);default:0" description:"已发放佣金"`

	// 质量指标
	InvitationQualityScore float64 `json:"invitation_quality_score" gorm:"column:invitation_quality_score;type:decimal(10,2);default:0" description:"邀请质量分（付费率×100）"`
	ActivityScore          float64 `json:"activity_score" gorm:"column:activity_score;type:decimal(10,2);default:0" description:"活跃度分（近30天占比×100）"`

	// 综合排名
	RankingScore float64 `json:"ranking_score" gorm:"column:ranking_score;type:decimal(10,2);default:0" description:"综合排名分数"`

	// 时间信息
	FirstInvitationDate *time.Time `json:"first_invitation_date" gorm:"column:first_invitation_date" description:"首次邀请时间"`
	LastInvitationDate  *time.Time `json:"last_invitation_date" gorm:"column:last_invitation_date" description:"最后邀请时间"`
	LastUpdated         time.Time  `json:"last_updated" gorm:"column:last_updated;autoUpdateTime" description:"最后更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

// TableName 表名设置
func (InvitationRankingCache) TableName() string {
	return "invitation_ranking_cache"
}

// InvitationRankingResponse 邀请排名响应结构
type InvitationRankingResponse struct {
	Rank                   int       `json:"rank"`                      // 排名
	UserID                 string    `json:"user_id"`                   // 用户ID
	Nickname               *string   `json:"nickname"`                  // 昵称
	Avatar                 *string   `json:"avatar"`                    // 头像
	TotalInvitations       int       `json:"total_invitations"`         // 总邀请人数
	PaidInvitations        int       `json:"paid_invitations"`          // 有效邀请人数
	Recent30dInvitations   int       `json:"recent_30d_invitations"`    // 近30天邀请数
	PersonalViralRate      float64   `json:"personal_viral_rate"`       // 个人裂变率
	InvitationGrowthRate   float64   `json:"invitation_growth_rate"`    // 邀请增长率
	TotalCommission        float64   `json:"total_commission"`          // 总佣金
	InvitationQualityScore float64   `json:"invitation_quality_score"`  // 邀请质量分
	ActivityScore          float64   `json:"activity_score"`            // 活跃度分
	RankingScore           float64   `json:"ranking_score"`             // 综合排名分
	FirstInvitationDate    string    `json:"first_invitation_date"`     // 首次邀请时间
	LastInvitationDate     string    `json:"last_invitation_date"`      // 最后邀请时间
	LastUpdated            string    `json:"last_updated"`              // 最后更新时间
}

// InvitationSystemMetrics 邀请系统级指标
type InvitationSystemMetrics struct {
	// 基础指标
	TotalUsers           int     `json:"total_users"`            // 总用户数
	ActiveInviters       int     `json:"active_inviters"`        // 有邀请行为的用户数
	ShareRate            float64 `json:"share_rate"`             // 分享率（%）
	
	// 邀请指标
	TotalInvitations     int     `json:"total_invitations"`      // 总邀请人数
	PaidInvitations      int     `json:"paid_invitations"`       // 付费邀请人数
	AvgViralCoefficient  float64 `json:"avg_viral_coefficient"`  // 平均裂变系数
	ConversionRate       float64 `json:"conversion_rate"`        // 有效邀请转化率（%）
	
	// 佣金指标
	TotalCommission      float64 `json:"total_commission"`       // 总佣金金额
	AvgCommissionPerUser float64 `json:"avg_commission_per_user"` // 平均每用户佣金
	
	// 活跃指标
	Recent30dInvitations int     `json:"recent_30d_invitations"` // 近30天邀请数
	Recent7dInvitations  int     `json:"recent_7d_invitations"`  // 近7天邀请数
}

// InvitationDetailInfo 邀请详细信息
type InvitationDetailInfo struct {
	InviteeID    string  `json:"invitee_id"`    // 被邀请人ID
	Nickname     *string `json:"nickname"`      // 被邀请人昵称
	Avatar       *string `json:"avatar"`        // 被邀请人头像
	IsPaid       bool    `json:"is_paid"`       // 是否付费
	InvitedDate  string  `json:"invited_date"`  // 邀请时间
	OrderCount   int     `json:"order_count"`   // 订单数量
	TotalPayment float64 `json:"total_payment"` // 总支付金额
}

// ToResponse 转换为响应结构
func (irc *InvitationRankingCache) ToResponse(rank int) InvitationRankingResponse {
	resp := InvitationRankingResponse{
		Rank:                   rank,
		UserID:                 irc.UserID,
		TotalInvitations:       irc.TotalInvitations,
		PaidInvitations:        irc.PaidInvitations,
		Recent30dInvitations:   irc.Recent30dInvitations,
		PersonalViralRate:      irc.PersonalViralRate,
		InvitationGrowthRate:   irc.InvitationGrowthRate,
		TotalCommission:        irc.TotalCommission,
		InvitationQualityScore: irc.InvitationQualityScore,
		ActivityScore:          irc.ActivityScore,
		RankingScore:           irc.RankingScore,
		LastUpdated:            irc.LastUpdated.Format("2006-01-02 15:04:05"),
	}

	// 处理用户信息
	if irc.User != nil {
		resp.Nickname = irc.User.Nickname
		resp.Avatar = irc.User.Avatar
	}

	// 处理时间字段
	if irc.FirstInvitationDate != nil {
		resp.FirstInvitationDate = irc.FirstInvitationDate.Format("2006-01-02 15:04:05")
	}
	if irc.LastInvitationDate != nil {
		resp.LastInvitationDate = irc.LastInvitationDate.Format("2006-01-02 15:04:05")
	}

	return resp
}

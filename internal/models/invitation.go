package models

import (
	"time"
)

// InvitationCode 邀请码表 - 对应Python的InvitationCode模型
type InvitationCode struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID    string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Code      string    `json:"code" gorm:"column:code;type:varchar(32);uniqueIndex;not null" description:"邀请码"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`

	// 关联关系 - 移除自动外键约束，避免循环引用
	User      User                 `json:"user,omitempty" gorm:"-"`
	Relations []InvitationRelation `json:"relations,omitempty" gorm:"-"`
}

// InvitationRelation 邀请关系表 - 对应Python的InvitationRelation模型
type InvitationRelation struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	InviterID string    `json:"inviter_id" gorm:"column:inviter_id;type:varchar(50);not null;index" description:"邀请人ID"`
	InviteeID string    `json:"invitee_id" gorm:"column:invitee_id;type:varchar(50);not null;uniqueIndex" description:"被邀请人ID"`
	CodeID    int       `json:"code_id" gorm:"column:code_id;not null;index" description:"使用的邀请码ID"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`

	// 关联关系 - 定义外键关联，支持Preload优化查询
	Inviter     User               `json:"inviter,omitempty" gorm:"foreignKey:InviterID;references:UserID"`
	Invitee     User               `json:"invitee,omitempty" gorm:"foreignKey:InviteeID;references:UserID"`
	Code        InvitationCode     `json:"code,omitempty" gorm:"foreignKey:CodeID;references:ID"`
	Commissions []CommissionRecord `json:"commissions,omitempty" gorm:"foreignKey:RelationID;references:ID"`
}

// CommissionStatus 佣金状态枚举 - 对应Python的CommissionStatus
type CommissionStatus int

const (
	CommissionPending   CommissionStatus = 0 // 待发放
	CommissionIssued    CommissionStatus = 1 // 已发放
	CommissionWithdrawn CommissionStatus = 2 // 已提现
	CommissionRejected  CommissionStatus = 3 // 已拒绝
	CommissionApplying  CommissionStatus = 4 // 申请中
)

// String 返回佣金状态的字符串表示
func (cs CommissionStatus) String() string {
	switch cs {
	case CommissionPending:
		return "待发放"
	case CommissionIssued:
		return "已发放"
	case CommissionWithdrawn:
		return "已提现"
	case CommissionRejected:
		return "已拒绝"
	case CommissionApplying:
		return "申请中"
	default:
		return "未知状态"
	}
}

// CommissionRecord 佣金记录表 - 对应Python的CommissionRecord模型
type CommissionRecord struct {
	ID             int              `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	UserID         string           `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"获得佣金的用户ID"`
	RelationID     int              `json:"relation_id" gorm:"column:relation_id;not null;index" description:"关联的邀请关系ID"`
	OrderID        *int             `json:"order_id" gorm:"column:order_id;index" description:"关联订单ID"`
	Amount         float64          `json:"amount" gorm:"column:amount;type:decimal(10,2);not null" description:"佣金金额"`
	Status         CommissionStatus `json:"status" gorm:"column:status;type:tinyint;default:0" description:"佣金状态"`
	Description    string           `json:"description" gorm:"column:description;type:varchar(256)" description:"佣金说明"`
	IssueTime      *time.Time       `json:"issue_time" gorm:"column:issue_time" description:"发放时间"`
	WithdrawalTime *time.Time       `json:"withdrawal_time" gorm:"column:withdrawal_time" description:"提现时间"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`

	// 关联关系 - 定义外键关联，支持Preload优化查询
	User     User               `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
	Relation InvitationRelation `json:"relation,omitempty" gorm:"foreignKey:RelationID;references:ID"`
	// Order    Trade              `json:"order" gorm:"foreignKey:OrderID"` // 如果有订单模型的话
}

// 表名设置
func (InvitationCode) TableName() string {
	return "invitation_codes"
}

func (InvitationRelation) TableName() string {
	return "invitation_relations"
}

func (CommissionRecord) TableName() string {
	return "commission_records"
}

// InvitationCodeResponse 邀请码响应结构
type InvitationCodeResponse struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	CreatedAt string `json:"created_at"`
}

// InvitationRelationResponse 邀请关系响应结构
type InvitationRelationResponse struct {
	ID        int                     `json:"id"`
	Inviter   *UserResponse           `json:"inviter,omitempty"`
	Invitee   *UserResponse           `json:"invitee,omitempty"`
	Code      *InvitationCodeResponse `json:"code,omitempty"`
	CreatedAt string                  `json:"created_at"`
}

// CommissionRecordResponse 佣金记录响应结构
type CommissionRecordResponse struct {
	ID             int     `json:"id"`
	Amount         float64 `json:"amount"`
	Status         int     `json:"status"`
	StatusText     string  `json:"status_text"`
	Description    string  `json:"description"`
	IssueTime      *string `json:"issue_time"`
	WithdrawalTime *string `json:"withdrawal_time"`
	CreatedAt      string  `json:"created_at"`
}

// ToResponse 转换为响应结构
func (ic *InvitationCode) ToResponse() InvitationCodeResponse {
	return InvitationCodeResponse{
		ID:        ic.ID,
		Code:      ic.Code,
		CreatedAt: ic.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (ir *InvitationRelation) ToResponse() InvitationRelationResponse {
	resp := InvitationRelationResponse{
		ID:        ir.ID,
		CreatedAt: ir.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 处理关联数据
	if ir.Inviter.UserID != "" {
		inviterResp := ir.Inviter.ToResponse()
		resp.Inviter = &inviterResp
	}
	if ir.Invitee.UserID != "" {
		inviteeResp := ir.Invitee.ToResponse()
		resp.Invitee = &inviteeResp
	}
	if ir.Code.ID != 0 {
		codeResp := ir.Code.ToResponse()
		resp.Code = &codeResp
	}

	return resp
}

func (cr *CommissionRecord) ToResponse() CommissionRecordResponse {
	resp := CommissionRecordResponse{
		ID:          cr.ID,
		Amount:      cr.Amount,
		Status:      int(cr.Status),
		StatusText:  cr.Status.String(),
		Description: cr.Description,
		CreatedAt:   cr.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if cr.IssueTime != nil {
		issueTimeStr := cr.IssueTime.Format("2006-01-02T15:04:05Z07:00")
		resp.IssueTime = &issueTimeStr
	}
	if cr.WithdrawalTime != nil {
		withdrawalTimeStr := cr.WithdrawalTime.Format("2006-01-02T15:04:05Z07:00")
		resp.WithdrawalTime = &withdrawalTimeStr
	}

	return resp
}

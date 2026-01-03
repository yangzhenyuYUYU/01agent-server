package models

import (
	"time"
)

// CardType 卡片类型枚举
type CardType string

const (
	CardTypeMembership  CardType = "membership"   // 会员卡
	CardTypeCredits     CardType = "credits"      // 积分卡
	CardTypeCreditsTemp CardType = "credits_temp" // 积分卡临时
)

// ActivationCode 激活码模型
type ActivationCode struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Code      string    `json:"code" gorm:"column:code;type:varchar(32);not null;uniqueIndex" description:"激活码"`
	CardType  string    `json:"card_type" gorm:"column:card_type;type:varchar(32);not null" description:"类型"`
	ProductID int       `json:"product_id" gorm:"column:product_id;not null" description:"关联产品ID"`
	IsUsed    bool      `json:"is_used" gorm:"column:is_used;default:false" description:"是否已使用"`
	Remark    *string   `json:"remark" gorm:"column:remark;type:varchar(256)" description:"备注"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	TradeID   *int      `json:"trade_id" gorm:"column:trade_id;index" description:"关联交易"`
	UsedByID  *string   `json:"used_by_id" gorm:"column:used_by_id;type:varchar(50);index" description:"使用用户"`

	// 关联关系
	Trade *Trade `json:"trade,omitempty" gorm:"-"`
	User  *User  `json:"user,omitempty" gorm:"-"`
}

// 表名设置
func (ActivationCode) TableName() string {
	return "activation_codes"
}

// ActivationCodeResponse 响应结构
type ActivationCodeResponse struct {
	ID        int     `json:"id"`
	Code      string  `json:"code"`
	CardType  string  `json:"card_type"`
	IsUsed    bool    `json:"is_used"`
	Remark    *string `json:"remark"`
	CreatedAt string  `json:"created_at"`
}

// ToResponse 转换方法
func (a *ActivationCode) ToResponse() ActivationCodeResponse {
	return ActivationCodeResponse{
		ID:        a.ID,
		Code:      a.Code,
		CardType:  a.CardType,
		IsUsed:    a.IsUsed,
		Remark:    a.Remark,
		CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

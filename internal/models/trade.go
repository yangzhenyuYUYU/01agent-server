package models

import (
	"time"
)

// PaymentChannel 支付渠道枚举
type PaymentChannel string

const (
	// 支付宝相关支付方式
	PaymentChannelAlipay     PaymentChannel = "alipay"      // 支付宝APP支付
	PaymentChannelAlipayQR   PaymentChannel = "alipay_qr"   // 支付宝正扫
	PaymentChannelAlipayWap  PaymentChannel = "alipay_wap"  // 支付宝H5支付
	PaymentChannelAlipayLite PaymentChannel = "alipay_lite" // 支付宝小程序支付
	PaymentChannelAlipayPub  PaymentChannel = "alipay_pub"  // 支付宝生活号支付
	PaymentChannelAlipayScan PaymentChannel = "alipay_scan" // 支付宝反扫

	// 微信相关支付方式
	PaymentChannelWxQR   PaymentChannel = "wx_qr"   // 微信正扫
	PaymentChannelWxPub  PaymentChannel = "wx_pub"  // 微信公众号支付
	PaymentChannelWxLite PaymentChannel = "wx_lite" // 微信小程序支付
	PaymentChannelWxScan PaymentChannel = "wx_scan" // 微信反扫

	// 银联相关支付方式
	PaymentChannelUnion         PaymentChannel = "union"          // 银联云闪付App支付
	PaymentChannelUnionQR       PaymentChannel = "union_qr"       // 银联云闪付正扫
	PaymentChannelUnionWap      PaymentChannel = "union_wap"      // 银联云闪付H5支付
	PaymentChannelUnionScan     PaymentChannel = "union_scan"     // 银联云闪付反扫
	PaymentChannelUnionOnline   PaymentChannel = "union_online"   // 银联H5支付
	PaymentChannelUnionCheckout PaymentChannel = "union_checkout" // 银联统一收银台支付

	// 其他支付方式
	PaymentChannelFastPay    PaymentChannel = "fast_pay"   // 快捷支付
	PaymentChannelB2C        PaymentChannel = "b2c"        // 个人网银支付
	PaymentChannelB2B        PaymentChannel = "b2b"        // 企业网银支付
	PaymentChannelCardKey    PaymentChannel = "card_key"   // 卡密兑换
	PaymentChannelActivation PaymentChannel = "activation" // 兑换码兑换
	PaymentChannelCredit     PaymentChannel = "credit"     // 积分支付
)

// PaymentStatus 支付状态枚举
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"  // 待支付
	PaymentStatusSuccess  PaymentStatus = "success"  // 支付成功
	PaymentStatusFailed   PaymentStatus = "failed"   // 支付失败
	PaymentStatusRefunded PaymentStatus = "refunded" // 已退款
	PaymentStatusFinished PaymentStatus = "finished" // 已完成，不可退款
)

// TradeType 交易类型枚举
type TradeType string

const (
	TradeTypeRecharge         TradeType = "recharge"          // 充值
	TradeTypeConsume          TradeType = "consume"           // 消费
	TradeTypeRefund           TradeType = "refund"            // 退款
	TradeTypeActivation       TradeType = "activation"        // 兑换码兑换
	TradeTypeActivationRefund TradeType = "activation_refund" // 兑换码兑换退款
	TradeTypeCommission       TradeType = "commission"        // 佣金收入
)

// Trade 交易模型
type Trade struct {
	ID             int        `json:"id" gorm:"primaryKey;column:id" description:"交易ID"`
	TradeNo        string     `json:"trade_no" gorm:"column:trade_no;type:varchar(64);not null;uniqueIndex" description:"交易流水号"`
	UserID         string     `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	Amount         float64    `json:"amount" gorm:"column:amount;type:decimal(10,2);not null" description:"交易金额"`
	TradeType      string     `json:"trade_type" gorm:"column:trade_type;type:varchar(64);not null" description:"交易类型"`
	PaymentChannel string     `json:"payment_channel" gorm:"column:payment_channel;type:varchar(64);not null" description:"支付渠道"`
	PaymentStatus  string     `json:"payment_status" gorm:"column:payment_status;type:varchar(64);not null;default:'pending'" description:"支付状态"`
	PaymentID      *string    `json:"payment_id" gorm:"column:payment_id;type:varchar(64)" description:"支付ID"`
	Title          string     `json:"title" gorm:"column:title;type:varchar(128);not null" description:"交易标题"`
	Metadata       *string    `json:"metadata" gorm:"column:metadata;type:json" description:"元数据，用于存储特定业务数据"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	PaidAt         *time.Time `json:"paid_at" gorm:"column:paid_at" description:"支付时间"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

// BPOrder BP订单模型
type BPOrder struct {
	ID             int       `json:"id" gorm:"primaryKey;column:id" description:"BP订单ID"`
	TradeNo        string    `json:"trade_no" gorm:"column:trade_no;type:varchar(64);not null;uniqueIndex" description:"交易流水号"`
	ProductName    string    `json:"product_name" gorm:"column:product_name;type:varchar(128);not null" description:"产品名称"`
	PaymentChannel string    `json:"payment_channel" gorm:"column:payment_channel;type:varchar(64);not null" description:"支付渠道"`
	Email          string    `json:"email" gorm:"column:email;type:varchar(128);not null" description:"邮箱"`
	Price          float64   `json:"price" gorm:"column:price;type:decimal(10,2);not null" description:"订单价格"`
	PaymentStatus  string    `json:"payment_status" gorm:"column:payment_status;type:varchar(64);not null;default:'pending'" description:"订单状态"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// Production 产品模型
type Production struct {
	ID             int       `json:"id" gorm:"primaryKey;column:id" description:"产品ID"`
	Name           string    `json:"name" gorm:"column:name;type:varchar(128);not null" description:"产品名称"`
	Price          float64   `json:"price" gorm:"column:price;type:decimal(10,2);not null" description:"产品价格"`
	OriginalPrice  *float64  `json:"original_price" gorm:"column:original_price;type:decimal(10,2)" description:"原价"`
	ProductType    string    `json:"product_type" gorm:"column:product_type;type:varchar(64);not null" description:"产品类型"`
	Description    string    `json:"description" gorm:"column:description;type:longtext;not null" description:"产品描述"`
	ExtraInfo      *string   `json:"extra_info" gorm:"column:extra_info;type:json" description:"产品扩展信息"`
	ValidityPeriod *int      `json:"validity_period" gorm:"column:validity_period" description:"有效期"`
	Status         *int      `json:"status" gorm:"column:status" description:"上架状态"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// UserProductionStatus 用户产品状态枚举
type UserProductionStatus string

const (
	UserProductionStatusActive   UserProductionStatus = "active"   // 兑换
	UserProductionStatusInactive UserProductionStatus = "inactive" // 未兑换
	UserProductionStatusExpired  UserProductionStatus = "expired"  // 已过期
)

// UserProduction 用户产品关联模型
type UserProduction struct {
	ID           int       `json:"id" gorm:"primaryKey;column:id" description:"用户产品ID"`
	UserID       string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户"`
	ProductionID int       `json:"production_id" gorm:"column:production_id;not null;index" description:"关联产品"`
	TradeID      int       `json:"trade_id" gorm:"column:trade_id;not null;index" description:"关联交易"`
	Status       *string   `json:"status" gorm:"column:status;type:varchar(10);default:'active'" description:"用户产品状态"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User       *User       `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
	Production *Production `json:"production,omitempty" gorm:"foreignKey:ProductionID;references:ID"`
	Trade      *Trade      `json:"trade,omitempty" gorm:"foreignKey:TradeID;references:ID"`
}

// 表名设置
func (Trade) TableName() string {
	return "trades"
}

func (BPOrder) TableName() string {
	return "bp_orders"
}

func (Production) TableName() string {
	return "productions"
}

func (UserProduction) TableName() string {
	return "user_productions"
}

// 响应结构
type TradeResponse struct {
	ID             int     `json:"id"`
	TradeNo        string  `json:"trade_no"`
	UserID         string  `json:"user_id"`
	Amount         float64 `json:"amount"`
	TradeType      string  `json:"trade_type"`
	PaymentChannel string  `json:"payment_channel"`
	PaymentStatus  string  `json:"payment_status"`
	PaymentID      *string `json:"payment_id"`
	Title          string  `json:"title"`
	CreatedAt      string  `json:"created_at"`
	PaidAt         *string `json:"paid_at"`
}

type BPOrderResponse struct {
	ID             int     `json:"id"`
	TradeNo        string  `json:"trade_no"`
	ProductName    string  `json:"product_name"`
	PaymentChannel string  `json:"payment_channel"`
	Email          string  `json:"email"`
	Price          float64 `json:"price"`
	PaymentStatus  string  `json:"payment_status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type ProductionResponse struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Price          float64  `json:"price"`
	OriginalPrice  *float64 `json:"original_price"`
	ProductType    string   `json:"product_type"`
	Description    string   `json:"description"`
	ValidityPeriod *int     `json:"validity_period"`
	Status         *int     `json:"status"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// ToResponse 转换方法
func (t *Trade) ToResponse() TradeResponse {
	resp := TradeResponse{
		ID:             t.ID,
		TradeNo:        t.TradeNo,
		UserID:         t.UserID,
		Amount:         t.Amount,
		TradeType:      t.TradeType,
		PaymentChannel: t.PaymentChannel,
		PaymentStatus:  t.PaymentStatus,
		PaymentID:      t.PaymentID,
		Title:          t.Title,
		CreatedAt:      t.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if t.PaidAt != nil {
		paidAtStr := t.PaidAt.Format("2006-01-02 15:04:05")
		resp.PaidAt = &paidAtStr
	}
	return resp
}

func (b *BPOrder) ToResponse() BPOrderResponse {
	return BPOrderResponse{
		ID:             b.ID,
		TradeNo:        b.TradeNo,
		ProductName:    b.ProductName,
		PaymentChannel: b.PaymentChannel,
		Email:          b.Email,
		Price:          b.Price,
		PaymentStatus:  b.PaymentStatus,
		CreatedAt:      b.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      b.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (p *Production) ToResponse() ProductionResponse {
	return ProductionResponse{
		ID:             p.ID,
		Name:           p.Name,
		Price:          p.Price,
		OriginalPrice:  p.OriginalPrice,
		ProductType:    p.ProductType,
		Description:    p.Description,
		ValidityPeriod: p.ValidityPeriod,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

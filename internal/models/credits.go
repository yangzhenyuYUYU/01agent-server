package models

import (
	"time"
)

// CreditProduct 积分产品模型
type CreditProduct struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"产品ID"`
	Name      *string   `json:"name" gorm:"column:name;type:varchar(128)" description:"产品名称"`
	Credits   *int      `json:"credits" gorm:"column:credits" description:"积分数量"`
	Price     *float64  `json:"price" gorm:"column:price;type:decimal(10,2)" description:"价格"`
	Status    *bool     `json:"status" gorm:"column:status;default:true" description:"是否有效"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// CreditRecordType 积分记录类型枚举
type CreditRecordType int16

const (
	CreditRecharge    CreditRecordType = 1 // 充值
	CreditConsumption CreditRecordType = 2 // 消费
	CreditReward      CreditRecordType = 3 // 奖励
	CreditExpired     CreditRecordType = 4 // 过期
	CreditRefund      CreditRecordType = 5 // 退款
)

// CreditRecord 积分记录模型
type CreditRecord struct {
	ID          int              `json:"id" gorm:"primaryKey;column:id" description:"记录ID"`
	RecordType  CreditRecordType `json:"record_type" gorm:"column:record_type;type:smallint;not null" description:"记录类型"`
	Credits     *int             `json:"credits" gorm:"column:credits" description:"积分变动数量"`
	Balance     *int             `json:"balance" gorm:"column:balance" description:"变动后余额"`
	Description *string          `json:"description" gorm:"column:description;type:varchar(256)" description:"变动描述"`
	ServiceCode *string          `json:"service_code" gorm:"column:service_code;type:varchar(64)" description:"服务代号"`
	UserID      string           `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户"`
	CreatedAt   time.Time        `json:"created_at" gorm:"column:created_at" description:"创建时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// CreditRechargeOrder 积分充值订单模型
type CreditRechargeOrder struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"订单ID"`
	ProductID int       `json:"product_id" gorm:"column:product_id;not null;index" description:"关联产品"`
	TradeID   int       `json:"trade_id" gorm:"column:trade_id;not null;index" description:"关联交易"`
	UserID    string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`

	// 关联关系
	Product *CreditProduct `json:"product,omitempty" gorm:"-"`
	Trade   *Trade         `json:"trade,omitempty" gorm:"-"`
	User    *User          `json:"user,omitempty" gorm:"-"`
}

// ServiceUnit 服务单位枚举
type ServiceUnit int16

const (
	ServiceUnitCount  ServiceUnit = 1 // 按次数计费
	ServiceUnitMinute ServiceUnit = 2 // 按分钟计费
	ServiceUnitChar   ServiceUnit = 3 // 按字符数计费
	ServiceUnitSecond ServiceUnit = 4 // 按秒数计费
	ServiceUnitToken  ServiceUnit = 5 // 按token数计费
)

// CreditServicePrice 积分服务定价模型
type CreditServicePrice struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	ServiceCode string    `json:"service_code" gorm:"column:service_code;type:varchar(64);not null;uniqueIndex" description:"服务代号"`
	Name        *string   `json:"name" gorm:"column:name;type:varchar(128)" description:"服务名称"`
	Credits     *int      `json:"credits" gorm:"column:credits" description:"消耗积分数/unit"`
	Unit        *int16    `json:"unit" gorm:"column:unit;type:smallint" description:"计费单位"`
	Description *string   `json:"description" gorm:"column:description;type:varchar(256)" description:"服务描述"`
	Status      bool      `json:"status" gorm:"column:status;default:true" description:"是否有效"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
}

// UserDailyBenefit 用户每日权益模型
type UserDailyBenefit struct {
	ID            int       `json:"id" gorm:"primaryKey;column:id" description:"记录ID"`
	UserID        string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户"`
	DailyCredits  int       `json:"daily_credits" gorm:"column:daily_credits;default:30" description:"每日积分额度（剩余可用）"`
	ExtraBenefits *string   `json:"extra_benefits" gorm:"column:extra_benefits;type:json" description:"其他权益额度（预留扩展），如：{\"daily_images\": 10, \"daily_exports\": 5}"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;index" description:"创建时间（按日期判断当日记录）"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"-"`
}

// UserMonthlyBenefit 用户每月权益模型
// 积分类型说明：
// - 用户积分（user.credits）：用户充值购买的永久积分，不会过期
// - 每日积分（UserDailyBenefit）：每日赠送的积分，当日有效，次日清零
// - 每月积分（UserMonthlyBenefit）：会员权益赠送的积分，会员有效期内有效，会员过期时清零
// 消费优先级：每日积分 > 每月积分 > 用户积分
type UserMonthlyBenefit struct {
	ID               int        `json:"id" gorm:"primaryKey;column:id" description:"记录ID"`
	UserID           string     `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户"`
	UserProductionID *int       `json:"user_production_id" gorm:"column:user_production_id;index" description:"关联用户产品（哪个订阅产生的权益）"`
	MonthlyCredits   int        `json:"monthly_credits" gorm:"column:monthly_credits;default:0" description:"每月积分额度（剩余可用）"`
	BenefitMonth     time.Time  `json:"benefit_month" gorm:"column:benefit_month;type:date;not null;index" description:"权益月份（YYYY-MM-01格式，记录是哪个月发放的）"`
	ExpireAt         *time.Time `json:"expire_at" gorm:"column:expire_at;index" description:"会员过期时间（积分在此时间后失效，NULL表示终身会员）"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User           *User           `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
	UserProduction *UserProduction `json:"user_production,omitempty" gorm:"foreignKey:UserProductionID;references:ID"`
}

// TimedCreditSourceType 有期限积分来源类型
const (
	TimedCreditSourceInvite   = "invite"   // 邀请奖励
	TimedCreditSourcePackage  = "package"  // 积分套餐购买
	TimedCreditSourceActivity = "activity" // 活动奖励
	TimedCreditSourceRegister = "register" // 注册奖励
	TimedCreditSourceOther    = "other"    // 其他
)

// UserTimedCredits 用户有期限积分模型
// 用于存储有有效期的积分（如邀请奖励、积分套餐购买的积分）
// 这些积分有固定有效期，过期后自动失效
// 消费优先级：每日积分 > 有期限积分（按过期时间升序） > 每月权益积分 > 永久积分
type UserTimedCredits struct {
	ID              int       `json:"id" gorm:"primaryKey;column:id" description:"记录ID"`
	UserID          string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户"`
	Credits         int       `json:"credits" gorm:"column:credits;default:0" description:"剩余积分"`
	OriginalCredits int       `json:"original_credits" gorm:"column:original_credits;default:0" description:"原始积分"`
	SourceType      string    `json:"source_type" gorm:"column:source_type;type:varchar(50);not null;index" description:"来源类型：invite/package/activity/register/other"`
	SourceDesc      *string   `json:"source_desc" gorm:"column:source_desc;type:varchar(255)" description:"来源描述"`
	ExpireAt        time.Time `json:"expire_at" gorm:"column:expire_at;not null;index" description:"过期时间"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

// 表名设置
func (CreditProduct) TableName() string {
	return "credit_products"
}

func (CreditRecord) TableName() string {
	return "credit_records"
}

func (CreditRechargeOrder) TableName() string {
	return "credit_recharge_orders"
}

func (CreditServicePrice) TableName() string {
	return "credit_service_prices"
}

func (UserDailyBenefit) TableName() string {
	return "user_daily_benefits"
}

func (UserMonthlyBenefit) TableName() string {
	return "user_monthly_benefits"
}

func (UserTimedCredits) TableName() string {
	return "user_timed_credits"
}

// 响应结构
type CreditProductResponse struct {
	ID        int      `json:"id"`
	Name      *string  `json:"name"`
	Credits   *int     `json:"credits"`
	Price     *float64 `json:"price"`
	Status    *bool    `json:"status"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type CreditRecordResponse struct {
	ID          int     `json:"id"`
	RecordType  int16   `json:"record_type"`
	Credits     *int    `json:"credits"`
	Balance     *int    `json:"balance"`
	Description *string `json:"description"`
	ServiceCode *string `json:"service_code"`
	CreatedAt   string  `json:"created_at"`
}

// ToResponse 转换方法
func (cp *CreditProduct) ToResponse() CreditProductResponse {
	return CreditProductResponse{
		ID:        cp.ID,
		Name:      cp.Name,
		Credits:   cp.Credits,
		Price:     cp.Price,
		Status:    cp.Status,
		CreatedAt: cp.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: cp.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (cr *CreditRecord) ToResponse() CreditRecordResponse {
	return CreditRecordResponse{
		ID:          cr.ID,
		RecordType:  int16(cr.RecordType),
		Credits:     cr.Credits,
		Balance:     cr.Balance,
		Description: cr.Description,
		ServiceCode: cr.ServiceCode,
		CreatedAt:   cr.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

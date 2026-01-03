package digital

import (
	"gin_web/internal/models"
	"time"
)

// DigitalTemplateOrder 数字人模板订单模型
type DigitalTemplateOrder struct {
	ID         int       `json:"id" gorm:"primaryKey;column:id" description:"订单ID"`
	UserID     string    `json:"user_id" gorm:"column:user_id;type:varchar(50);not null;index" description:"关联用户ID"`
	TemplateID int       `json:"template_id" gorm:"column:template_id;not null;index" description:"关联模板ID"`
	TradeID    int       `json:"trade_id" gorm:"column:trade_id;not null;index" description:"关联交易ID"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`

	// 关联关系
	User     *models.User     `json:"user,omitempty" gorm:"-"`
	Template *DigitalTemplate `json:"template,omitempty" gorm:"-"`
	Trade    *models.Trade    `json:"trade,omitempty" gorm:"-"`
}

// 表名设置
func (DigitalTemplateOrder) TableName() string {
	return "digital_template_orders"
}

package models

import (
	"time"
)

// Reservation 预约模型
type Reservation struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id" description:"ID"`
	Name      *string   `json:"name" gorm:"column:name;type:varchar(100)" description:"预约人姓名"`
	Phone     *string   `json:"phone" gorm:"column:phone;type:varchar(20)" description:"预约人手机号"`
	Email     *string   `json:"email" gorm:"column:email;type:varchar(100)" description:"预约人邮箱"`
	Notes     *string   `json:"notes" gorm:"column:notes;type:longtext" description:"备注信息"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// 表名设置
func (Reservation) TableName() string {
	return "reservations"
}

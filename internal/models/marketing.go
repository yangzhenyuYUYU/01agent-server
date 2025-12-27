package models

import (
	"time"
)

// ActivityStatus 活动状态枚举
type ActivityStatus int16

const (
	ActivityStatusPending ActivityStatus = 0 // 待开始
	ActivityStatusOngoing ActivityStatus = 1 // 进行中
	ActivityStatusEnded   ActivityStatus = 2 // 已结束
)

// ActivityType 活动类型枚举
type ActivityType int16

const (
	ActivityTypeDiscount ActivityType = 1 // 折扣活动
	ActivityTypeGift     ActivityType = 2 // 赠品活动
	ActivityTypePoints   ActivityType = 3 // 积分活动
	ActivityTypeFlashSale ActivityType = 4 // 限时特卖
	ActivityTypeGroupBuy ActivityType = 5 // 团购活动
)

// MarketingActivityPlan 营销活动计划模型
type MarketingActivityPlan struct {
	ActivityID  string    `json:"activity_id" gorm:"primaryKey;column:activity_id;type:varchar(50)" description:"活动ID"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(100);not null" description:"活动名称"`
	Description string    `json:"description" gorm:"column:description;type:longtext;not null" description:"活动描述"`
	Status      int16     `json:"status" gorm:"column:status;type:smallint;not null;default:0" description:"活动状态"`
	StartTime   time.Time `json:"start_time" gorm:"column:start_time;not null" description:"活动开始时间"`
	EndTime     time.Time `json:"end_time" gorm:"column:end_time;not null" description:"活动结束时间"`
	Config      string    `json:"config" gorm:"column:config;type:json;not null" description:"活动配置（商品、福利和限制条件）"`
	IsVisible   bool      `json:"is_visible" gorm:"column:is_visible;default:true" description:"是否可见"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime" description:"创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime" description:"更新时间"`
}

// 表名设置
func (MarketingActivityPlan) TableName() string {
	return "marketing_activity_plans"
}


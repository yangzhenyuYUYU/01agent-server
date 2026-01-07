package config

import (
	"fmt"
)

// ==================== 每日积分配置 ====================
// GetDefaultDailyCredits 获取默认每日积分，从配置文件读取
func GetDefaultDailyCredits() int {
	if AppConfig != nil && AppConfig.Credits.DailyLoginReward > 0 {
		return AppConfig.Credits.DailyLoginReward
	}
	return 30 // 默认每日积分
}

// ==================== 存储配额配置 ====================
// 单位：字节
var StorageQuotaMap = map[int]int64{
	0: 314572800,   // 免费版：300MB
	1: 1073741824,  // 轻量版：1GB
	2: 5368709120,  // 专业版：5GB
	3: 10737418240, // 种子终身会员：10GB
}

// 存储配额描述
var StorageQuotaDesc = map[int]string{
	0: "300MB",
	1: "1GB",
	2: "5GB",
	3: "5GB",
	4: "10GB",
}

// ==================== 产品配置结构 ====================

// SubscriptionProduct 订阅服务产品配置
type SubscriptionProduct struct {
	Name           string // 产品名称
	VipLevel       int    // VIP等级
	ValidityMonths int    // 有效月份数，-1表示终身
	BonusCredits   int    // 总赠送积分（会按月平摊发放）
	StorageQuota   int64  // 存储配额（字节），默认300MB
	UpgradeDesc    string // 升级描述
	CreditsDesc    string // 积分描述
	StorageDesc    string // 存储描述
}

// MonthlyCredits 计算每月发放的积分数
func (p *SubscriptionProduct) MonthlyCredits() int {
	if p.ValidityMonths <= 0 {
		// 终身会员，一次性发放全部积分
		return p.BonusCredits
	}
	return p.BonusCredits / p.ValidityMonths
}

// GetChanges 获取变更说明列表
func (p *SubscriptionProduct) GetChanges() []string {
	changes := []string{}
	if p.UpgradeDesc != "" {
		changes = append(changes, p.UpgradeDesc)
	}
	if p.CreditsDesc != "" {
		changes = append(changes, p.CreditsDesc)
	}
	if p.StorageDesc != "" {
		changes = append(changes, p.StorageDesc)
	}
	return changes
}

// CreditPackage 积分套餐产品配置
type CreditPackage struct {
	Name    string // 产品名称
	Credits int    // 积分数量
	Desc    string // 描述
}

// GetChanges 获取变更说明列表
func (p *CreditPackage) GetChanges() []string {
	if p.Desc != "" {
		return []string{p.Desc}
	}
	return []string{fmt.Sprintf("购买%d积分", p.Credits)}
}

// ==================== 订阅服务产品配置 ====================
var SubscriptionProducts = map[string]*SubscriptionProduct{
	"免费版": {
		Name:           "免费版",
		VipLevel:       0,
		ValidityMonths: 0, // 免费版无有效期概念
		BonusCredits:   0,
		StorageQuota:   StorageQuotaMap[0],
		UpgradeDesc:    "更新为免费版会员",
	},
	"轻量版": {
		Name:           "轻量版",
		VipLevel:       1,
		ValidityMonths: 1,   // 月度会员
		BonusCredits:   870, // 每月870积分
		StorageQuota:   StorageQuotaMap[1],
		UpgradeDesc:    "升级为轻量版会员",
		CreditsDesc:    "每月赠送870积分",
		StorageDesc:    "存储空间升级为1GB",
	},
	"轻量版体验": {
		Name:           "轻量版体验",
		VipLevel:       1,
		ValidityMonths: 1,   // 体验版1个月
		BonusCredits:   100, // 每月100积分
		StorageQuota:   StorageQuotaMap[1],
		UpgradeDesc:    "升级为轻量版会员",
		CreditsDesc:    "每月赠送100积分",
		StorageDesc:    "存储空间升级为1GB",
	},
	"轻量版年度会员": {
		Name:           "轻量版年度会员",
		VipLevel:       1,
		ValidityMonths: 12,    // 年度会员12个月
		BonusCredits:   10440, // 总计10440积分，每月870积分
		StorageQuota:   StorageQuotaMap[1],
		UpgradeDesc:    "升级为轻量版年度会员",
		CreditsDesc:    "每月赠送870积分（共12期）",
		StorageDesc:    "存储空间升级为1GB",
	},
	"专业版": {
		Name:           "专业版",
		VipLevel:       3,
		ValidityMonths: 1,    // 月度会员
		BonusCredits:   4500, // 每月4500积分
		StorageQuota:   StorageQuotaMap[2],
		UpgradeDesc:    "升级为专业版会员",
		CreditsDesc:    "每月赠送4500积分",
		StorageDesc:    "存储空间升级为5GB",
	},
	"专业版半年订阅升级套餐": {
		Name:           "专业版半年订阅升级套餐",
		VipLevel:       3,
		ValidityMonths: 6, // 半年会员
		BonusCredits:   0,
		StorageQuota:   StorageQuotaMap[2],
		UpgradeDesc:    "升级为专业版半年会员",
		StorageDesc:    "存储空间升级为5GB",
	},
	"专业版年度会员": {
		Name:           "专业版年度会员",
		VipLevel:       3,
		ValidityMonths: 12,    // 年度会员12个月
		BonusCredits:   54000, // 总计54000积分，每月4500积分
		StorageQuota:   StorageQuotaMap[2],
		UpgradeDesc:    "升级为专业版会员",
		CreditsDesc:    "每月赠送4500积分（共12期）",
		StorageDesc:    "存储空间升级为5GB",
	},
	"专业版会员体验": {
		Name:           "专业版会员体验",
		VipLevel:       3,
		ValidityMonths: 1, // 体验版1个月
		BonusCredits:   300,
		StorageQuota:   StorageQuotaMap[2],
		UpgradeDesc:    "升级为专业版会员体验会员",
		CreditsDesc:    "每月赠送300积分",
		StorageDesc:    "存储空间升级为5GB",
	},
	"专业版周体验": {
		Name:           "专业版周体验",
		VipLevel:       3,
		ValidityMonths: 1, // 周体验按1个月算，积分一次性发
		BonusCredits:   500,
		StorageQuota:   StorageQuotaMap[2],
		UpgradeDesc:    "升级为专业版周体验",
		CreditsDesc:    "赠送500积分",
		StorageDesc:    "存储空间升级为5GB",
	},
	"专业版开通测试": {
		Name:           "专业版开通测试",
		VipLevel:       3,
		ValidityMonths: 1, // 测试版1个月
		BonusCredits:   50,
		StorageQuota:   StorageQuotaMap[2],
		UpgradeDesc:    "升级为专业版开通测试",
		CreditsDesc:    "赠送50积分",
		StorageDesc:    "存储空间升级为5GB",
	},
	"种子终身会员": {
		Name:           "种子终身会员",
		VipLevel:       4,
		ValidityMonths: -1,    // 终身会员，-1表示无限期
		BonusCredits:   70000, // 终身会员积分一次性发放
		StorageQuota:   StorageQuotaMap[3],
		UpgradeDesc:    "升级为种子终身会员",
		CreditsDesc:    "赠送70000积分",
		StorageDesc:    "存储空间升级为10GB",
	},
}

// ==================== 积分套餐产品配置 ====================
var CreditPackages = map[string]*CreditPackage{
	"600积分": {
		Name:    "600积分",
		Credits: 600,
		Desc:    "购买600积分",
	},
	"1500积分": {
		Name:    "1500积分",
		Credits: 1500,
		Desc:    "购买1500积分",
	},
	"3000积分": {
		Name:    "3000积分",
		Credits: 3000,
		Desc:    "购买3000积分",
	},
}

// ==================== 辅助函数 ====================

// GetSubscriptionProduct 根据产品名称获取订阅服务配置
func GetSubscriptionProduct(name string) *SubscriptionProduct {
	return SubscriptionProducts[name]
}

// GetCreditPackage 根据产品名称获取积分套餐配置
func GetCreditPackage(name string) *CreditPackage {
	return CreditPackages[name]
}

// GetVipLevelByProductName 根据产品名称获取VIP等级
func GetVipLevelByProductName(name string) int {
	product := SubscriptionProducts[name]
	if product != nil {
		return product.VipLevel
	}
	return 0
}

// GetStorageQuotaByVipLevel 根据VIP等级获取存储配额
func GetStorageQuotaByVipLevel(vipLevel int) int64 {
	if quota, ok := StorageQuotaMap[vipLevel]; ok {
		return quota
	}
	return StorageQuotaMap[0] // 默认返回免费版配额
}

// GetStorageDescByVipLevel 根据VIP等级获取存储配额描述
func GetStorageDescByVipLevel(vipLevel int) string {
	if desc, ok := StorageQuotaDesc[vipLevel]; ok {
		return desc
	}
	return "300MB"
}

// GetMonthlyCreditsByProductName 根据产品名称获取每月发放的积分数
func GetMonthlyCreditsByProductName(name string) int {
	product := SubscriptionProducts[name]
	if product != nil {
		return product.MonthlyCredits()
	}
	return 0
}

// GetValidityMonthsByProductName 根据产品名称获取有效月份数
func GetValidityMonthsByProductName(name string) int {
	product := SubscriptionProducts[name]
	if product != nil {
		return product.ValidityMonths
	}
	return 0
}

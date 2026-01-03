package repository

import (
	"time"

	"01agent_server/internal/models"

	"gorm.io/gorm"
)

type InvitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository 创建邀请相关仓库
func NewInvitationRepository() *InvitationRepository {
	return &InvitationRepository{
		db: DB,
	}
}

// ===== InvitationCode 邀请码相关 =====

// CreateInvitationCode 创建邀请码
func (r *InvitationRepository) CreateInvitationCode(code *models.InvitationCode) error {
	return r.db.Create(code).Error
}

// GetInvitationCodeByUserID 根据用户ID获取邀请码
func (r *InvitationRepository) GetInvitationCodeByUserID(userID string) (*models.InvitationCode, error) {
	var code models.InvitationCode
	err := r.db.Where("user_id = ?", userID).First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// GetInvitationCodeByCode 根据邀请码获取记录
func (r *InvitationRepository) GetInvitationCodeByCode(code string) (*models.InvitationCode, error) {
	var invCode models.InvitationCode
	err := r.db.Where("code = ?", code).First(&invCode).Error
	if err != nil {
		return nil, err
	}
	return &invCode, nil
}

// ===== InvitationRelation 邀请关系相关 =====

// CreateInvitationRelation 创建邀请关系
func (r *InvitationRepository) CreateInvitationRelation(relation *models.InvitationRelation) error {
	return r.db.Create(relation).Error
}

// GetInvitationRelationByInvitee 根据被邀请人获取邀请关系
func (r *InvitationRepository) GetInvitationRelationByInvitee(inviteeID string) (*models.InvitationRelation, error) {
	var relation models.InvitationRelation
	err := r.db.Where("invitee_id = ?", inviteeID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetInvitationCountByInviter 根据邀请人获取邀请数量
func (r *InvitationRepository) GetInvitationCountByInviter(inviterID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.InvitationRelation{}).Where("inviter_id = ?", inviterID).Count(&count).Error
	return count, err
}

// GetInvitationRelationsByInviter 根据邀请人获取邀请关系列表
func (r *InvitationRepository) GetInvitationRelationsByInviter(inviterID string, page, size int) ([]models.InvitationRelation, int64, error) {
	var relations []models.InvitationRelation
	var total int64

	// 计算总数
	if err := r.db.Model(&models.InvitationRelation{}).Where("inviter_id = ?", inviterID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := r.db.Where("inviter_id = ?", inviterID).
		Offset(offset).Limit(size).
		Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}

	return relations, total, nil
}

// ===== CommissionRecord 佣金记录相关 =====

// CreateCommissionRecord 创建佣金记录
func (r *InvitationRepository) CreateCommissionRecord(record *models.CommissionRecord) error {
	return r.db.Create(record).Error
}

// GetCommissionRecordsByUserID 根据用户ID获取佣金记录
func (r *InvitationRepository) GetCommissionRecordsByUserID(userID string, page, size int) ([]models.CommissionRecord, int64, error) {
	var records []models.CommissionRecord
	var total int64

	// 计算总数
	if err := r.db.Model(&models.CommissionRecord{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(size).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetCommissionSummaryByUserID 获取用户佣金汇总
func (r *InvitationRepository) GetCommissionSummaryByUserID(userID string) (map[string]interface{}, error) {
	var result struct {
		TotalAmount     float64 `json:"total_amount"`
		PendingAmount   float64 `json:"pending_amount"`
		IssuedAmount    float64 `json:"issued_amount"`
		WithdrawnAmount float64 `json:"withdrawn_amount"`
		TotalCount      int64   `json:"total_count"`
	}

	// 查询总金额和数量
	if err := r.db.Model(&models.CommissionRecord{}).
		Select("SUM(amount) as total_amount, COUNT(*) as total_count").
		Where("user_id = ?", userID).
		Scan(&result).Error; err != nil {
		return nil, err
	}

	// 查询各状态金额
	r.db.Model(&models.CommissionRecord{}).
		Select("SUM(amount) as pending_amount").
		Where("user_id = ? AND status = ?", userID, models.CommissionPending).
		Scan(&result)

	r.db.Model(&models.CommissionRecord{}).
		Select("SUM(amount) as issued_amount").
		Where("user_id = ? AND status = ?", userID, models.CommissionIssued).
		Scan(&result)

	r.db.Model(&models.CommissionRecord{}).
		Select("SUM(amount) as withdrawn_amount").
		Where("user_id = ? AND status = ?", userID, models.CommissionWithdrawn).
		Scan(&result)

	return map[string]interface{}{
		"total_amount":     result.TotalAmount,
		"pending_amount":   result.PendingAmount,
		"issued_amount":    result.IssuedAmount,
		"withdrawn_amount": result.WithdrawnAmount,
		"total_count":      result.TotalCount,
	}, nil
}

// UpdateCommissionStatus 更新佣金状态
func (r *InvitationRepository) UpdateCommissionStatus(id int, status models.CommissionStatus, updateTime *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if updateTime != nil {
		switch status {
		case models.CommissionIssued:
			updates["issue_time"] = updateTime
		case models.CommissionWithdrawn:
			updates["withdrawal_time"] = updateTime
		}
	}

	return r.db.Model(&models.CommissionRecord{}).Where("id = ?", id).Updates(updates).Error
}

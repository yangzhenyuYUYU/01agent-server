package service

import (
	"fmt"
	"math/rand"
	"time"

	"01agent_server/internal/models"
	"01agent_server/internal/repository"
)

type InvitationService struct {
	invitationRepo *repository.InvitationRepository
	userRepo       *repository.UserRepository
}

// NewInvitationService 创建邀请服务
func NewInvitationService() *InvitationService {
	return &InvitationService{
		invitationRepo: repository.NewInvitationRepository(),
		userRepo:       repository.NewUserRepository(),
	}
}

// ===== 邀请码相关 =====

// GetOrCreateInvitationCode 获取或创建用户的邀请码
func (s *InvitationService) GetOrCreateInvitationCode(userID string) (*models.InvitationCode, error) {
	// 先尝试获取现有邀请码
	code, err := s.invitationRepo.GetInvitationCodeByUserID(userID)
	if err == nil {
		return code, nil
	}

	// 如果不存在，创建新的邀请码
	newCode := &models.InvitationCode{
		UserID:    userID,
		Code:      s.generateInvitationCode(),
		CreatedAt: time.Now(),
	}

	if err := s.invitationRepo.CreateInvitationCode(newCode); err != nil {
		return nil, fmt.Errorf("failed to create invitation code: %w", err)
	}

	return newCode, nil
}

// generateInvitationCode 生成邀请码
func (s *InvitationService) generateInvitationCode() string {
	// 生成8位随机大写字母邀请码
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// ===== 邀请关系相关 =====

// CreateInvitationRelation 创建邀请关系
func (s *InvitationService) CreateInvitationRelation(inviteeID, invitationCode string) error {
	// 获取邀请码记录
	codeRecord, err := s.invitationRepo.GetInvitationCodeByCode(invitationCode)
	if err != nil {
		return fmt.Errorf("invalid invitation code: %w", err)
	}

	// 检查是否已存在邀请关系
	existingRelation, _ := s.invitationRepo.GetInvitationRelationByInvitee(inviteeID)
	if existingRelation != nil {
		return fmt.Errorf("user already has an inviter")
	}

	// 不能邀请自己
	if codeRecord.UserID == inviteeID {
		return fmt.Errorf("cannot invite yourself")
	}

	// 创建邀请关系
	relation := &models.InvitationRelation{
		InviterID: codeRecord.UserID,
		InviteeID: inviteeID,
		CodeID:    codeRecord.ID,
		CreatedAt: time.Now(),
	}

	return s.invitationRepo.CreateInvitationRelation(relation)
}

// GetInvitationInfo 获取用户的邀请信息
func (s *InvitationService) GetInvitationInfo(userID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取邀请码
	code, err := s.GetOrCreateInvitationCode(userID)
	if err != nil {
		return nil, err
	}
	result["invitation_code"] = code.Code

	// 获取邀请数量
	invitationCount, err := s.invitationRepo.GetInvitationCountByInviter(userID)
	if err != nil {
		return nil, err
	}
	result["invitation_count"] = invitationCount

	// 获取邀请人信息
	relation, err := s.invitationRepo.GetInvitationRelationByInvitee(userID)
	if err == nil && relation != nil {
		result["inviter"] = map[string]interface{}{
			"id":       relation.Inviter.UserID,
			"nickname": relation.Inviter.Nickname,
			"avatar":   relation.Inviter.Avatar,
		}
	} else {
		result["inviter"] = nil
	}

	return result, nil
}

// GetInvitationList 获取用户的邀请列表
func (s *InvitationService) GetInvitationList(userID string, page, size int) ([]models.InvitationRelationResponse, int64, error) {
	relations, total, err := s.invitationRepo.GetInvitationRelationsByInviter(userID, page, size)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.InvitationRelationResponse
	for _, relation := range relations {
		responses = append(responses, relation.ToResponse())
	}

	return responses, total, nil
}

// ===== 佣金相关 =====

// CreateCommissionRecord 创建佣金记录
func (s *InvitationService) CreateCommissionRecord(userID string, relationID int, orderID *int, amount float64, description string) error {
	record := &models.CommissionRecord{
		UserID:      userID,
		RelationID:  relationID,
		OrderID:     orderID,
		Amount:      amount,
		Status:      models.CommissionPending,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return s.invitationRepo.CreateCommissionRecord(record)
}

// GetCommissionRecords 获取用户的佣金记录
func (s *InvitationService) GetCommissionRecords(userID string, page, size int) ([]models.CommissionRecordResponse, int64, error) {
	records, total, err := s.invitationRepo.GetCommissionRecordsByUserID(userID, page, size)
	if err != nil {
		return nil, 0, err
	}

	var responses []models.CommissionRecordResponse
	for _, record := range records {
		responses = append(responses, record.ToResponse())
	}

	return responses, total, nil
}

// GetCommissionSummary 获取用户佣金汇总
func (s *InvitationService) GetCommissionSummary(userID string) (map[string]interface{}, error) {
	return s.invitationRepo.GetCommissionSummaryByUserID(userID)
}

// UpdateCommissionStatus 更新佣金状态
func (s *InvitationService) UpdateCommissionStatus(id int, status models.CommissionStatus) error {
	now := time.Now()
	return s.invitationRepo.UpdateCommissionStatus(id, status, &now)
}

// CalculateCommission 计算佣金（示例逻辑）
func (s *InvitationService) CalculateCommission(orderAmount float64, commissionRate float64) float64 {
	return orderAmount * commissionRate
}

// ProcessOrderCommission 处理订单佣金（示例业务逻辑）
func (s *InvitationService) ProcessOrderCommission(buyerID string, orderAmount float64, commissionRate float64) error {
	// 获取买家的邀请关系
	relation, err := s.invitationRepo.GetInvitationRelationByInvitee(buyerID)
	if err != nil {
		// 如果没有邀请关系，不产生佣金
		return nil
	}

	// 计算佣金
	commissionAmount := s.CalculateCommission(orderAmount, commissionRate)
	if commissionAmount <= 0 {
		return nil
	}

	// 创建佣金记录
	description := fmt.Sprintf("邀请用户消费佣金，订单金额：%.2f", orderAmount)
	return s.CreateCommissionRecord(
		relation.InviterID,
		relation.ID,
		nil, // 如果有订单ID可以传入
		commissionAmount,
		description,
	)
}

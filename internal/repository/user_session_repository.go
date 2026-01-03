package repository

import (
	"01agent_server/internal/models"
	"time"

	"gorm.io/gorm"
)

type UserSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository 创建用户会话仓库
func NewUserSessionRepository() *UserSessionRepository {
	return &UserSessionRepository{
		db: DB,
	}
}

// Create 创建会话
func (r *UserSessionRepository) Create(session *models.UserSession) error {
	return r.db.Create(session).Error
}

// GetByUserID 根据用户ID获取活跃会话
func (r *UserSessionRepository) GetByUserID(userID string) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&sessions).Error
	return sessions, err
}

// GetByToken 根据令牌获取会话
func (r *UserSessionRepository) GetByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.Where("token = ? AND is_active = ?", token, true).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeactivateByToken 根据令牌停用会话
func (r *UserSessionRepository) DeactivateByToken(token string) error {
	return r.db.Model(&models.UserSession{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_active": false,
			"status":    0,
			"token":     nil,
		}).Error
}

// DeactivateByUserID 停用用户的所有会话
func (r *UserSessionRepository) DeactivateByUserID(userID string) error {
	return r.db.Model(&models.UserSession{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// DeleteExpired 删除过期会话
func (r *UserSessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < NOW()").Delete(&models.UserSession{}).Error
}

// CountActiveSessionsByUserID 统计用户的活跃会话数量
func (r *UserSessionRepository) CountActiveSessionsByUserID(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserSession{}).
		Where("user_id = ? AND is_active = ? AND status = ?", userID, true, 1).
		Count(&count).Error
	return count, err
}

// GetActiveSessionsByUserID 获取用户的活跃会话列表（按创建时间降序）
func (r *UserSessionRepository) GetActiveSessionsByUserID(userID string) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := r.db.Where("user_id = ? AND is_active = ? AND status = ?", userID, true, 1).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// DeactivateOtherSessions 停用用户的其他会话（保留当前token）
func (r *UserSessionRepository) DeactivateOtherSessions(userID string, currentToken string) error {
	return r.db.Model(&models.UserSession{}).
		Where("user_id = ? AND token != ?", userID, currentToken).
		Updates(map[string]interface{}{
			"is_active": false,
			"status":    0,
			"token":     nil,
		}).Error
}

// CleanupSessionsKeepRecent 清理会话，只保留最近的N个活跃会话
// 返回被删除的会话数量
func (r *UserSessionRepository) CleanupSessionsKeepRecent(userID string, keepCount int) (int64, error) {
	// 获取所有活跃会话，按创建时间降序
	activeSessions, err := r.GetActiveSessionsByUserID(userID)
	if err != nil {
		return 0, err
	}

	// 如果活跃会话数量不超过保留数量，不需要清理
	if len(activeSessions) <= keepCount {
		return 0, nil
	}

	// 获取需要保留的会话ID
	keepIDs := make([]int, 0, keepCount)
	for i := 0; i < keepCount && i < len(activeSessions); i++ {
		keepIDs = append(keepIDs, activeSessions[i].ID)
	}

	// 删除不在保留列表中的会话
	result := r.db.Where("user_id = ?", userID).
		Where("id NOT IN ?", keepIDs).
		Delete(&models.UserSession{})

	return result.RowsAffected, result.Error
}

// UpdateLastActiveTime 更新会话最后活跃时间
func (r *UserSessionRepository) UpdateLastActiveTime(sessionID int) error {
	return r.db.Model(&models.UserSession{}).
		Where("id = ?", sessionID).
		Update("last_active_time", time.Now()).Error
}

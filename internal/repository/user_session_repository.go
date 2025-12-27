package repository

import (
	"gin_web/internal/models"

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
	return r.db.Model(&models.UserSession{}).Where("token = ?", token).Update("is_active", false).Error
}

// DeactivateByUserID 停用用户的所有会话
func (r *UserSessionRepository) DeactivateByUserID(userID string) error {
	return r.db.Model(&models.UserSession{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// DeleteExpired 删除过期会话
func (r *UserSessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < NOW()").Delete(&models.UserSession{}).Error
}

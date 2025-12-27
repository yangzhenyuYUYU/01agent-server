package repository

import (
	"gin_web/internal/models"

	"gorm.io/gorm"
)

type UserParametersRepository struct {
	db *gorm.DB
}

// NewUserParametersRepository 创建用户参数仓库
func NewUserParametersRepository() *UserParametersRepository {
	return &UserParametersRepository{
		db: DB,
	}
}

// Create 创建用户参数
func (r *UserParametersRepository) Create(params *models.UserParameters) error {
	return r.db.Create(params).Error
}

// GetByUserID 根据用户ID获取参数
func (r *UserParametersRepository) GetByUserID(userID string) (*models.UserParameters, error) {
	var params models.UserParameters
	err := r.db.Where("user_id = ?", userID).First(&params).Error
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// Update 更新用户参数
func (r *UserParametersRepository) Update(params *models.UserParameters) error {
	return r.db.Save(params).Error
}

// Delete 删除用户参数
func (r *UserParametersRepository) Delete(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserParameters{}).Error
}

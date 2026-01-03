package repository

import (
	"01agent_server/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: DB,
	}
}

// Create 创建用户
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(userID string) (*models.User, error) {
	var user models.User

	// 添加调试日志
	Infof("UserRepository.GetByID: searching for user_id = '%s'", userID)

	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		Errorf("UserRepository.GetByID: query failed for user_id '%s': %v", userID, err)

		// 检查表中是否有任何用户数据
		var count int64
		r.db.Model(&models.User{}).Count(&count)
		Infof("UserRepository.GetByID: total users in database: %d", count)

		return nil, err
	}

	username := "nil"
	if user.Username != nil {
		username = *user.Username
	}
	Infof("UserRepository.GetByID: found user '%s' with username '%s'", userID, username)
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *UserRepository) Delete(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.User{}).Error
}

// List 获取用户列表
func (r *UserRepository) List(page, size int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算总数
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := r.db.Offset(offset).Limit(size).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLastLoginTime 更新最后登录时间
func (r *UserRepository) UpdateLastLoginTime(userID string) error {
	return r.db.Model(&models.User{}).Where("user_id = ?", userID).Update("last_login_time", gorm.Expr("NOW()")).Error
}

// IsEmailExists 检查邮箱是否存在
func (r *UserRepository) IsEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// IsUsernameExists 检查用户名是否存在
func (r *UserRepository) IsUsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// GetByPhone 根据手机号获取用户
func (r *UserRepository) GetByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByOpenID 根据OpenID获取用户
func (r *UserRepository) GetByOpenID(openID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("openid = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IsPhoneExists 检查手机号是否存在
func (r *UserRepository) IsPhoneExists(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

// IsOpenIDExists 检查OpenID是否存在
func (r *UserRepository) IsOpenIDExists(openID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("openid = ?", openID).Count(&count).Error
	return count > 0, err
}

// UpdateCredits 更新用户积分
func (r *UserRepository) UpdateCredits(userID string, credits int) error {
	return r.db.Model(&models.User{}).Where("user_id = ?", userID).Update("credits", credits).Error
}

// AddCredits 增加用户积分
func (r *UserRepository) AddCredits(userID string, amount int) error {
	return r.db.Model(&models.User{}).Where("user_id = ?", userID).
		Update("credits", gorm.Expr("credits + ?", amount)).Error
}

// DeductCredits 扣减用户积分
func (r *UserRepository) DeductCredits(userID string, amount int) error {
	return r.db.Model(&models.User{}).Where("user_id = ? AND credits >= ?", userID, amount).
		Update("credits", gorm.Expr("credits - ?", amount)).Error
}

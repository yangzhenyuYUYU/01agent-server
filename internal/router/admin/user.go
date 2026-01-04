package admin

import (
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole 用户角色常量
const (
	UserRoleNormal      int16 = 1 // 普通用户
	UserRoleVIP         int16 = 2 // VIP用户
	UserRoleAdmin       int16 = 3 // 管理员
	UserRoleDistributor int16 = 4 // 分销商/合作方
)

// SetDistributorRequest 设置分销商请求
type SetDistributorRequest struct {
	CommissionRate *float64 `json:"commission_rate"` // 佣金比例，可选，默认0.2
	ExtraParams    *string  `json:"extra_params"`    // 额外参数，JSON格式，可选
}

// SetDistributor 设置用户为分销商身份
// @Summary 设置分销商身份
// @Description 将指定用户设置为分销商，如果已是分销商则更新信息
// @Tags admin-user
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param body body SetDistributorRequest false "分销商信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/user/{user_id}/distributor [post]
func (h *AdminHandler) SetDistributor(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	var req SetDistributorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有传body，也允许
		req = SetDistributorRequest{}
	}

	// 检查用户是否存在
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		repository.Errorf("查询用户失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败"))
		return
	}

	// 设置默认佣金比例
	commissionRate := 0.2
	if req.CommissionRate != nil {
		if *req.CommissionRate < 0 || *req.CommissionRate > 1 {
			middleware.HandleError(c, middleware.NewBusinessError(400, "佣金比例必须在0到1之间"))
			return
		}
		commissionRate = *req.CommissionRate
	}

	// 开启事务
	tx := repository.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已存在分销商记录
	var existingDistributor models.Distributor
	err := tx.Where("user_id = ?", userID).First(&existingDistributor).Error

	if err == nil {
		// 已存在，更新分销商信息
		existingDistributor.CommissionRate = commissionRate
		if req.ExtraParams != nil {
			existingDistributor.ExtraParams = req.ExtraParams
		}
		if err := tx.Save(&existingDistributor).Error; err != nil {
			tx.Rollback()
			repository.Errorf("更新分销商信息失败: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新分销商信息失败"))
			return
		}
	} else if err == gorm.ErrRecordNotFound {
		// 不存在，创建分销商记录
		newDistributor := models.Distributor{
			DistributorID:  uuid.New().String(),
			UserID:         userID,
			CommissionRate: commissionRate,
			ExtraParams:    req.ExtraParams,
		}
		if err := tx.Create(&newDistributor).Error; err != nil {
			tx.Rollback()
			repository.Errorf("创建分销商记录失败: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "创建分销商记录失败"))
			return
		}
		existingDistributor = newDistributor
	} else {
		tx.Rollback()
		repository.Errorf("查询分销商记录失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询分销商记录失败"))
		return
	}

	// 更新用户角色为分销商
	if user.Role != UserRoleDistributor {
		if err := tx.Model(&models.User{}).Where("user_id = ?", userID).
			Update("role", UserRoleDistributor).Error; err != nil {
			tx.Rollback()
			repository.Errorf("更新用户角色失败: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新用户角色失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		repository.Errorf("提交事务失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "操作失败"))
		return
	}

	middleware.Success(c, "设置分销商身份成功", gin.H{
		"distributor_id":  existingDistributor.DistributorID,
		"user_id":         existingDistributor.UserID,
		"commission_rate": existingDistributor.CommissionRate,
		"extra_params":    existingDistributor.ExtraParams,
	})
}

// RemoveDistributor 移除用户的分销商身份
// @Summary 移除分销商身份
// @Description 移除指定用户的分销商身份，删除分销商记录并恢复用户角色
// @Tags admin-user
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/user/{user_id}/distributor [delete]
func (h *AdminHandler) RemoveDistributor(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		repository.Errorf("查询用户失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败"))
		return
	}

	// 检查是否存在分销商记录
	var distributor models.Distributor
	if err := repository.DB.Where("user_id = ?", userID).First(&distributor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "该用户不是分销商"))
			return
		}
		repository.Errorf("查询分销商记录失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询分销商记录失败"))
		return
	}

	// 开启事务
	tx := repository.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除分销商记录
	if err := tx.Where("user_id = ?", userID).Delete(&models.Distributor{}).Error; err != nil {
		tx.Rollback()
		repository.Errorf("删除分销商记录失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除分销商记录失败"))
		return
	}

	// 如果用户当前角色是分销商，恢复为普通用户
	if user.Role == UserRoleDistributor {
		if err := tx.Model(&models.User{}).Where("user_id = ?", userID).
			Update("role", UserRoleNormal).Error; err != nil {
			tx.Rollback()
			repository.Errorf("更新用户角色失败: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "更新用户角色失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		repository.Errorf("提交事务失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "操作失败"))
		return
	}

	middleware.Success(c, "移除分销商身份成功", gin.H{
		"user_id": userID,
	})
}

// GetDistributorInfo 获取用户的分销商信息
// @Summary 获取分销商信息
// @Description 获取指定用户的分销商信息
// @Tags admin-user
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/user/{user_id}/distributor [get]
func (h *AdminHandler) GetDistributorInfo(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		repository.Errorf("查询用户失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败"))
		return
	}

	// 查询分销商记录
	var distributor models.Distributor
	if err := repository.DB.Where("user_id = ?", userID).First(&distributor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "该用户不是分销商"))
			return
		}
		repository.Errorf("查询分销商记录失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询分销商记录失败"))
		return
	}

	middleware.Success(c, "获取分销商信息成功", gin.H{
		"distributor_id":  distributor.DistributorID,
		"user_id":         distributor.UserID,
		"commission_rate": distributor.CommissionRate,
		"extra_params":    distributor.ExtraParams,
		"created_at":      distributor.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":      distributor.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetDistributorList 获取分销商列表
// @Summary 获取分销商列表
// @Description 获取所有分销商列表，支持分页
// @Tags admin-user
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/distributor/list [get]
func (h *AdminHandler) GetDistributorList(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"min=0"`
		PageSize int `form:"page_size" binding:"min=0,max=100"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	var distributors []models.Distributor
	var total int64

	// 计算总数
	if err := repository.DB.Model(&models.Distributor{}).Count(&total).Error; err != nil {
		repository.Errorf("查询分销商总数失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询分销商总数失败"))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := repository.DB.Order("created_at DESC").
		Offset(offset).Limit(req.PageSize).
		Find(&distributors).Error; err != nil {
		repository.Errorf("查询分销商列表失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询分销商列表失败"))
		return
	}

	// 获取用户信息
	userIDs := make([]string, len(distributors))
	for i, d := range distributors {
		userIDs[i] = d.UserID
	}

	var users []models.User
	userMap := make(map[string]*models.User)
	if len(userIDs) > 0 {
		if err := repository.DB.Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			repository.Errorf("查询用户信息失败: %v", err)
		} else {
			for i := range users {
				userMap[users[i].UserID] = &users[i]
			}
		}
	}

	// 构建响应
	items := make([]gin.H, len(distributors))
	for i, d := range distributors {
		item := gin.H{
			"distributor_id":  d.DistributorID,
			"user_id":         d.UserID,
			"commission_rate": d.CommissionRate,
			"extra_params":    d.ExtraParams,
			"created_at":      d.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":      d.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 添加用户信息
		if user, ok := userMap[d.UserID]; ok {
			item["username"] = user.Username
			item["nickname"] = user.Nickname
			item["phone"] = user.Phone
			item["email"] = user.Email
			item["avatar"] = user.Avatar
		}

		items[i] = item
	}

	middleware.Success(c, "获取分销商列表成功", gin.H{
		"items":     items,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

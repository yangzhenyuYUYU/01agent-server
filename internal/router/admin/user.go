package admin

import (
	"fmt"
	"strings"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/tools"

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

	// 清除用户Redis缓存
	tools.ClearUserCacheAsync(userID)

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

	// 清除用户Redis缓存
	tools.ClearUserCacheAsync(userID)

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

// UserV2ListRequest 用户列表查询请求模型V2
type UserV2ListRequest struct {
	Page                int      `json:"page" binding:"min=1"`
	PageSize            int      `json:"page_size" binding:"min=1,max=100"`
	UserID              *string  `json:"user_id"`
	Username            *string  `json:"username"`
	Phone               *string  `json:"phone"`
	Email               *string  `json:"email"`
	Nickname            *string  `json:"nickname"`
	Roles               []string `json:"roles"`
	Statuses            []string `json:"statuses"`
	VipLevels           []int    `json:"vip_levels"`
	Channels            []string `json:"channels"`
	MinTotalConsumption *float64 `json:"min_total_consumption"`
	MaxTotalConsumption *float64 `json:"max_total_consumption"`
	MinCredits          *float64 `json:"min_credits"`
	MaxCredits          *float64 `json:"max_credits"`
	StartDate           *string  `json:"start_date"`
	EndDate             *string  `json:"end_date"`
	OrderBy             *string  `json:"order_by"`
	OrderDirection      string   `json:"order_direction"`
}

// UserV2List 用户列表查询接口V2
func (h *AdminHandler) UserV2List(c *gin.Context) {
	var req UserV2ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "desc"
	}
	if req.OrderBy == nil || *req.OrderBy == "" {
		orderBy := "created_at"
		req.OrderBy = &orderBy
	}

	// 如果提供了用户ID，直接根据用户ID查询
	if req.UserID != nil && *req.UserID != "" {
		userRepo := repository.NewUserRepository()
		user, err := userRepo.GetByID(*req.UserID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.Success(c, "success", gin.H{
					"total":     0,
					"items":     []gin.H{},
					"page":      req.Page,
					"page_size": req.PageSize,
				})
				return
			}
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败: "+err.Error()))
			return
		}

		// 计算用户统计信息
		var paymentCount, articleCount, copilotCount, aiEditCount, invitationCount int64

		repository.DB.Model(&models.Trade{}).
			Where("user_id = ? AND payment_status = ?", user.UserID, 1). // PaymentStatus.SUCCESS = 1
			Count(&paymentCount)

		repository.DB.Model(&models.ArticleTask{}).
			Where("user_id = ?", user.UserID).
			Count(&articleCount)

		repository.DB.Model(&models.InvitationRelation{}).
			Where("inviter_id = ?", user.UserID).
			Count(&invitationCount)

		// 获取用户最近登录认证凭证token
		var session models.UserSession
		var token *string
		if err := repository.DB.Where("user_id = ? AND is_active = ?", user.UserID, true).
			Order("created_at DESC").First(&session).Error; err == nil {
			token = session.Token
		}

		totalConsumption := 0.0
		if user.TotalConsumption != nil {
			totalConsumption = *user.TotalConsumption
		}

		// 构建单个用户记录返回数据
		result := []gin.H{{
			"user_id":           user.UserID,
			"username":          user.Username,
			"phone":             user.Phone,
			"email":             user.Email,
			"nickname":          user.Nickname,
			"avatar":            user.Avatar,
			"role":              user.Role,
			"status":            user.Status,
			"vip_level":         user.VipLevel,
			"credits":           float64(user.Credits),
			"total_consumption": totalConsumption,
			"usage_count":       user.UsageCount,
			"payment_count":     paymentCount,
			"article_count":     articleCount,
			"copilot_count":     copilotCount,
			"ai_edit_count":     aiEditCount,
			"utm_source":        user.UtmSource,
			"invitation_count":  invitationCount,
			"token":             token,
			"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}}

		middleware.Success(c, "success", gin.H{
			"total":     1,
			"items":     result,
			"page":      1,
			"page_size": 1,
		})
		return
	}

	// 构建复杂查询
	query := repository.DB.Model(&models.User{})

	// 时间范围筛选
	if req.StartDate != nil && *req.StartDate != "" {
		start, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "开始时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		query = query.Where("created_at >= ?", start)
	}

	if req.EndDate != nil && *req.EndDate != "" {
		end, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "结束时间格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
		query = query.Where("created_at <= ?", end)
	}

	// 用户基本信息筛选
	if req.Username != nil && *req.Username != "" {
		query = query.Where("username LIKE ?", "%"+*req.Username+"%")
	}
	if req.Phone != nil && *req.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+*req.Phone+"%")
	}
	if req.Email != nil && *req.Email != "" {
		query = query.Where("email LIKE ?", "%"+*req.Email+"%")
	}
	if req.Nickname != nil && *req.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+*req.Nickname+"%")
	}

	// 角色筛选
	if len(req.Roles) > 0 {
		roleValues := make([]int16, 0)
		roleMapping := map[string]int16{
			"user":  1, // UserRoleNormal
			"vip":   2, // UserRoleVIP
			"admin": 3, // UserRoleAdmin
		}
		for _, roleStr := range req.Roles {
			if role, ok := roleMapping[strings.ToLower(roleStr)]; ok {
				roleValues = append(roleValues, role)
			} else {
				// 尝试解析为数字
				var role int16
				if _, err := fmt.Sscanf(roleStr, "%d", &role); err == nil {
					roleValues = append(roleValues, role)
				}
			}
		}
		if len(roleValues) > 0 {
			query = query.Where("role IN ?", roleValues)
		}
	}

	// 状态筛选
	if len(req.Statuses) > 0 {
		statusValues := make([]int16, 0)
		statusMapping := map[string]int16{
			"inactive": 0,
			"active":   1,
		}
		for _, statusStr := range req.Statuses {
			if status, ok := statusMapping[strings.ToLower(statusStr)]; ok {
				statusValues = append(statusValues, status)
			} else {
				// 尝试解析为数字
				var status int16
				if _, err := fmt.Sscanf(statusStr, "%d", &status); err == nil {
					statusValues = append(statusValues, status)
				}
			}
		}
		if len(statusValues) > 0 {
			query = query.Where("status IN ?", statusValues)
		}
	}

	// VIP等级筛选
	if len(req.VipLevels) > 0 {
		query = query.Where("vip_level IN ?", req.VipLevels)
	}

	// 渠道筛选
	if len(req.Channels) > 0 {
		query = query.Where("utm_source IN ?", req.Channels)
	}

	// 消费金额区间筛选
	if req.MinTotalConsumption != nil {
		query = query.Where("total_consumption >= ?", *req.MinTotalConsumption)
	}
	if req.MaxTotalConsumption != nil {
		query = query.Where("total_consumption <= ?", *req.MaxTotalConsumption)
	}

	// 积分区间筛选
	if req.MinCredits != nil {
		query = query.Where("credits >= ?", int(*req.MinCredits))
	}
	if req.MaxCredits != nil {
		query = query.Where("credits <= ?", int(*req.MaxCredits))
	}

	// 排序
	orderField := *req.OrderBy
	if req.OrderDirection == "desc" {
		orderField = orderField + " DESC"
	} else {
		orderField = orderField + " ASC"
	}
	query = query.Order(orderField)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var users []models.User
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(users))
	for _, user := range users {
		// 获取用户最近登录认证凭证token
		var session models.UserSession
		var token *string
		if err := repository.DB.Where("user_id = ? AND is_active = ?", user.UserID, true).
			Order("created_at DESC").First(&session).Error; err == nil {
			tokenStr := session.Token
			token = tokenStr
		}

		totalConsumption := 0.0
		if user.TotalConsumption != nil {
			totalConsumption = *user.TotalConsumption
		}

		result = append(result, gin.H{
			"user_id":           user.UserID,
			"username":          user.Username,
			"phone":             user.Phone,
			"email":             user.Email,
			"nickname":          user.Nickname,
			"avatar":            user.Avatar,
			"role":              user.Role,
			"status":            user.Status,
			"vip_level":         user.VipLevel,
			"token":             token,
			"credits":           float64(user.Credits),
			"total_consumption": totalConsumption,
			"usage_count":       user.UsageCount,
			"utm_source":        user.UtmSource,
			"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"total":     total,
		"items":     result,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetUserFeedbackStats 获取用户反馈统计
func (h *AdminHandler) GetUserFeedbackStats(c *gin.Context) {
	// 统计各类型反馈的数量
	var noneCount, satisfiedCount, dissatisfiedCount, totalCount int64

	// 统计未操作类型（0）
	if err := repository.DB.Model(&models.UserFeedback{}).
		Where("feedback_type = ?", models.FeedbackTypeNone).
		Count(&noneCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计满意类型（1）
	if err := repository.DB.Model(&models.UserFeedback{}).
		Where("feedback_type = ?", models.FeedbackTypeSatisfied).
		Count(&satisfiedCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计不满意类型（2）
	if err := repository.DB.Model(&models.UserFeedback{}).
		Where("feedback_type = ?", models.FeedbackTypeUnsatisfied).
		Count(&dissatisfiedCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	// 统计总数
	if err := repository.DB.Model(&models.UserFeedback{}).
		Count(&totalCount).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计失败: "+err.Error()))
		return
	}

	middleware.Success(c, "获取反馈统计成功", gin.H{
		"none_count":         noneCount,
		"satisfied_count":    satisfiedCount,
		"dissatisfied_count": dissatisfiedCount,
		"total_count":        totalCount,
	})
}

// GetPaidUsersWechatAccounts 获取付费用户绑定的公众号信息
func (h *AdminHandler) GetPaidUsersWechatAccounts(c *gin.Context) {
	var req struct {
		Page            int    `form:"page" binding:"min=1"`
		PageSize        int    `form:"page_size" binding:"min=1,max=100"`
		Keyword         string `form:"keyword"`
		HasAppid        *bool  `form:"has_appid"`
		FetchWechatInfo bool   `form:"fetch_wechat_info"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 1. 获取所有VIP和管理员用户ID（role = 2 或 3）
	var vipAdminUserIDs []string
	if err := repository.DB.Model(&models.User{}).
		Where("role IN ?", []int16{2, 3}). // UserRoleVIP = 2, UserRoleAdmin = 3
		Pluck("user_id", &vipAdminUserIDs).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询VIP和管理员用户失败: "+err.Error()))
		return
	}

	// 2. 获取所有有成功充值订单的用户ID（trade_type = "recharge", payment_status = "success"）
	var paidUserIDs []string
	if err := repository.DB.Model(&models.Trade{}).
		Where("payment_status = ? AND trade_type = ?", "success", "recharge").
		Distinct("user_id").
		Pluck("user_id", &paidUserIDs).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询付费用户失败: "+err.Error()))
		return
	}

	// 3. 合并用户ID（去重）
	userIDMap := make(map[string]bool)
	for _, id := range vipAdminUserIDs {
		userIDMap[id] = true
	}
	for _, id := range paidUserIDs {
		userIDMap[id] = true
	}
	var allPaidUserIDs []string
	for id := range userIDMap {
		allPaidUserIDs = append(allPaidUserIDs, id)
	}

	if len(allPaidUserIDs) == 0 {
		middleware.Success(c, "获取成功", gin.H{
			"total":     0,
			"page":      req.Page,
			"page_size": req.PageSize,
			"list":      []gin.H{},
		})
		return
	}

	// 4. 构建查询条件
	query := repository.DB.Model(&models.User{}).Where("user_id IN ?", allPaidUserIDs)

	// 5. 默认只返回已绑定appid的用户（has_appid默认为true）
	hasAppidFilter := req.HasAppid == nil || (req.HasAppid != nil && *req.HasAppid)
	if hasAppidFilter {
		query = query.Where("appid IS NOT NULL AND appid != ''")
	} else {
		// 如果明确要求查看未绑定的用户，重新构建查询
		query = repository.DB.Model(&models.User{}).Where("user_id IN ?", allPaidUserIDs)
		if req.Keyword != "" {
			keyword := "%" + req.Keyword + "%"
			query = query.Where("username LIKE ? OR phone LIKE ? OR nickname LIKE ?", keyword, keyword, keyword)
		}
		query = query.Where("appid IS NULL OR appid = ''")
	}

	// 6. 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("username LIKE ? OR phone LIKE ? OR nickname LIKE ?", keyword, keyword, keyword)
	}

	// 7. 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "统计总数失败: "+err.Error()))
		return
	}

	// 8. 分页查询
	var users []models.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&users).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户失败: "+err.Error()))
		return
	}

	// 9. 构建返回数据
	resultList := []gin.H{}
	for _, user := range users {
		// 获取用户的支付统计（只统计充值订单）
		var paidCount int64
		var totalAmount float64
		if err := repository.DB.Model(&models.Trade{}).
			Where("user_id = ? AND payment_status = ? AND trade_type = ?", user.UserID, "success", "recharge").
			Count(&paidCount).Error; err == nil {
			var amounts []float64
			if err := repository.DB.Model(&models.Trade{}).
				Where("user_id = ? AND payment_status = ? AND trade_type = ?", user.UserID, "success", "recharge").
				Pluck("amount", &amounts).Error; err == nil {
				for _, amount := range amounts {
					totalAmount += amount
				}
			}
		}

		// 获取用户参数
		var userParam models.UserParameters
		repository.DB.Where("user_id = ?", user.UserID).First(&userParam)

		// 确定角色名称
		roleName := "普通用户"
		if user.Role == 3 {
			roleName = "管理员"
		} else if user.Role == 2 {
			roleName = "VIP"
		}

		// 构建用户数据
		userData := gin.H{
			"user_id":       user.UserID,
			"username":      user.Username,
			"nickname":      user.Nickname,
			"phone":         user.Phone,
			"email":         user.Email,
			"avatar":        user.Avatar,
			"role":          user.Role,
			"role_name":     roleName,
			"appid":         user.AppID,
			"has_appid":     user.AppID != nil && *user.AppID != "",
			"is_gzh_bind":   userParam.IsGzhBind,
			"paid_count":    paidCount,
			"total_amount":  totalAmount,
			"wechat_info":   nil,
			"wechat_status": "未绑定",
		}

		// 格式化注册日期
		if !user.RegistrationDate.IsZero() {
			userData["registration_date"] = user.RegistrationDate.Format("2006-01-02 15:04:05")
		} else {
			userData["registration_date"] = nil
		}

		// 如果用户有appid，处理公众号信息
		if user.AppID != nil && *user.AppID != "" {
			if req.FetchWechatInfo {
				// TODO: 调用微信API获取公众号详细信息
				// 这里暂时返回占位符，实际需要调用微信API
				userData["wechat_status"] = "已配置(未获取详情)"
				userData["wechat_info"] = gin.H{
					"success": nil,
					"message": "已配置appid，但未获取详细信息。设置 fetch_wechat_info=true 可获取详情",
				}
			} else {
				userData["wechat_status"] = "已配置(未获取详情)"
				userData["wechat_info"] = gin.H{
					"success": nil,
					"message": "已配置appid，但未获取详细信息。设置 fetch_wechat_info=true 可获取详情",
				}
			}
		}

		resultList = append(resultList, userData)
	}

	middleware.Success(c, "获取成功", gin.H{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      resultList,
	})
}

// GetUserDetail 获取用户详情（包含用户参数）
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "用户ID不能为空"))
		return
	}

	// 查询用户
	var user models.User
	if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 查询用户参数，如果不存在则创建
	var userParam models.UserParameters
	if err := repository.DB.Where("user_id = ?", userID).First(&userParam).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认用户参数
			userParam = models.UserParameters{
				ParamID:             userID, // 使用 user_id 作为 param_id
				UserID:              userID,
				EnableHeadInfo:      false,
				EnableKnowledgeBase: false,
				DefaultTheme:        "countryside",
				IsGzhBind:           false,
				IsWechatAuthorized:  false,
				HasAuthReminded:     false,
				PublishTarget:       0,
				StorageQuota:        314572800, // 300MB
			}
			if err := repository.DB.Create(&userParam).Error; err != nil {
				middleware.HandleError(c, middleware.NewBusinessError(500, "创建用户参数失败: "+err.Error()))
				return
			}
		} else {
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询用户参数失败: "+err.Error()))
			return
		}
	}

	// 统计总邀请人数
	var totalInvitations int64
	if err := repository.DB.Model(&models.InvitationRelation{}).
		Where("inviter_id = ?", userID).
		Count(&totalInvitations).Error; err != nil {
		repository.Errorf("统计邀请人数失败: %v", err)
		totalInvitations = 0
	}

	// 统计积分信息
	now := time.Now()

	// 永久积分（从user表获取）
	permanentCredits := user.Credits
	if permanentCredits < 0 {
		permanentCredits = 0
	}

	// 月度积分（统计所有未过期的月度积分总和）
	var monthlyCredits int
	if err := repository.DB.Model(&models.UserMonthlyBenefit{}).
		Where("user_id = ? AND (expire_at IS NULL OR expire_at > ?)", userID, now).
		Select("COALESCE(SUM(monthly_credits), 0)").
		Scan(&monthlyCredits).Error; err != nil {
		repository.Errorf("统计月度积分失败: %v", err)
		monthlyCredits = 0
	}

	// 限时积分（统计所有未过期的限时积分总和）
	var timedCredits int
	if err := repository.DB.Model(&models.UserTimedCredits{}).
		Where("user_id = ? AND expire_at > ?", userID, now).
		Select("COALESCE(SUM(credits), 0)").
		Scan(&timedCredits).Error; err != nil {
		repository.Errorf("统计限时积分失败: %v", err)
		timedCredits = 0
	}

	// 每日积分（获取当前日期的每日积分）
	var dailyCredits int
	var dailyBenefit models.UserDailyBenefit
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	if err := repository.DB.Model(&models.UserDailyBenefit{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, todayStart, todayEnd).
		First(&dailyBenefit).Error; err == nil {
		dailyCredits = dailyBenefit.DailyCredits
	}

	// 计算总积分
	totalCredits := permanentCredits + monthlyCredits + timedCredits + dailyCredits

	// 构建返回数据
	result := gin.H{
		"user_id":           user.UserID,
		"nickname":          user.Nickname,
		"avatar":            user.Avatar,
		"username":          user.Username,
		"phone":             user.Phone,
		"email":             user.Email,
		"openid":            user.OpenID,
		"credits":           user.Credits, // 永久积分
		"is_active":         user.IsActive,
		"total_consumption": user.TotalConsumption,
		"vip_level":         user.VipLevel,
		"role":              user.Role,
		"status":            user.Status,
		"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"user_parameter": gin.H{
			"param_id":              userParam.ParamID,
			"enable_head_info":      userParam.EnableHeadInfo,
			"enable_knowledge_base": userParam.EnableKnowledgeBase,
			"default_theme":         userParam.DefaultTheme,
			"is_wechat_authorized":  userParam.IsWechatAuthorized,
			"is_gzh_bind":           userParam.IsGzhBind,
			"has_auth_reminded":     userParam.HasAuthReminded,
			"publish_target":        userParam.PublishTarget,
			"qrcode_data":           userParam.QrcodeData,
			"created_time":          userParam.CreatedTime.Format("2006-01-02 15:04:05"),
			"updated_time":          userParam.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
		// 邀请统计
		"total_invitations": totalInvitations,
		// 积分统计
		"credits_detail": gin.H{
			"permanent_credits": permanentCredits, // 永久积分
			"monthly_credits":   monthlyCredits,   // 月度积分
			"timed_credits":     timedCredits,     // 限时积分
			"daily_credits":     dailyCredits,     // 每日积分（当日可用）
			"total_credits":     totalCredits,     // 总积分
		},
	}

	middleware.Success(c, "success", result)
}

// GetInvitationRelationOverview 获取邀请关系概览统计
func (h *AdminHandler) GetInvitationRelationOverview(c *gin.Context) {
	// 总邀请关系数
	var totalRelations int64
	repository.DB.Model(&models.InvitationRelation{}).Count(&totalRelations)

	// 唯一邀请人数
	var uniqueInviters int64
	repository.DB.Model(&models.InvitationRelation{}).
		Select("COUNT(DISTINCT inviter_id)").
		Scan(&uniqueInviters)

	// 唯一被邀请人数
	var uniqueInvitees int64
	repository.DB.Model(&models.InvitationRelation{}).
		Select("COUNT(DISTINCT invitee_id)").
		Scan(&uniqueInvitees)

	// 今日新增邀请关系
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var todayCount int64
	repository.DB.Model(&models.InvitationRelation{}).
		Where("created_at >= ?", todayStart).
		Count(&todayCount)

	// 本周新增邀请关系
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	if now.Weekday() == time.Sunday {
		weekStart = todayStart.AddDate(0, 0, -6)
	}
	var weekCount int64
	repository.DB.Model(&models.InvitationRelation{}).
		Where("created_at >= ?", weekStart).
		Count(&weekCount)

	// 本月新增邀请关系
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthCount int64
	repository.DB.Model(&models.InvitationRelation{}).
		Where("created_at >= ?", monthStart).
		Count(&monthCount)

	// 平均每个邀请人邀请的人数
	var avgInvitationsPerInviter float64
	if uniqueInviters > 0 {
		avgInvitationsPerInviter = float64(totalRelations) / float64(uniqueInviters)
	}

	// 邀请人数最多的前10个邀请人
	type TopInviter struct {
		InviterID string `gorm:"column:inviter_id"`
		Count     int64  `gorm:"column:count"`
	}
	var topInviters []TopInviter
	repository.DB.Model(&models.InvitationRelation{}).
		Select("inviter_id, COUNT(*) as count").
		Group("inviter_id").
		Order("count DESC").
		Limit(10).
		Scan(&topInviters)

	// 构建返回数据
	topInvitersData := make([]gin.H, 0, len(topInviters))
	for _, inviter := range topInviters {
		// 获取邀请人信息
		var user models.User
		repository.DB.Where("user_id = ?", inviter.InviterID).First(&user)

		topInvitersData = append(topInvitersData, gin.H{
			"inviter_id": inviter.InviterID,
			"count":      inviter.Count,
			"user": gin.H{
				"user_id":  user.UserID,
				"nickname": user.Nickname,
				"username": user.Username,
			},
		})
	}

	middleware.Success(c, "获取成功", gin.H{
		"total_relations":             totalRelations,
		"unique_inviters":             uniqueInviters,
		"unique_invitees":             uniqueInvitees,
		"today_count":                 todayCount,
		"week_count":                  weekCount,
		"month_count":                 monthCount,
		"avg_invitations_per_inviter": avgInvitationsPerInviter,
		"top_inviters":                topInvitersData,
	})
}

// GetUserCustomSizeList 获取用户自定义尺寸列表
func (h *AdminHandler) GetUserCustomSizeList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"min=1"`
		PageSize int    `form:"page_size" binding:"min=1"`
		Status   *int   `form:"status"`
		Name     string `form:"name"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	// 构建查询
	query := repository.DB.Model(&models.UserCustomSize{})

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 名称搜索
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var sizes []models.UserCustomSize
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&sizes).Error; err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败: "+err.Error()))
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0, len(sizes))
	for _, size := range sizes {
		result = append(result, gin.H{
			"id":         size.ID,
			"user_id":    size.UserID,
			"name":       size.Name,
			"data":       size.Data,
			"status":     size.Status,
			"is_default": size.IsDefault,
			"sort_order": size.SortOrder,
			"created_at": size.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": size.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "success", gin.H{
		"items":     result,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

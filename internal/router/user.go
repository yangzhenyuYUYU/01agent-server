package router

import (
	"encoding/json"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService       *service.UserService
	invitationService *service.InvitationService
}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService:       service.NewUserService(),
		invitationService: service.NewInvitationService(),
	}
}

// Hello 简单的hello接口
func (h *UserHandler) Hello(c *gin.Context) {
	// 使用新的成功响应格式 (code = 0)
	middleware.Success(c, "hello", gin.H{
		"message": "Hello from Gin Web Server!",
		"version": "1.0.0",
	})
}

// GetUserInfo 获取当前用户信息 - 对应Python的/user/info接口
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		repository.Errorf("GetUserInfo: userID not found in context")
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	// 获取详情模式参数
	detail := c.Query("detail")

	repository.Infof("GetUserInfo: attempting to find user with ID: %s, detail: %s", userID, detail)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		repository.Errorf("GetUserInfo: failed to get user by ID %s: %v", userID, err)
		middleware.HandleError(c, middleware.NewBusinessError(404, "未找到用户记录"))
		return
	}

	// 获取用户参数
	userParams, err := h.userService.GetUserParameters(userID)
	if err != nil {
		repository.Warnf("GetUserInfo: failed to get user parameters for %s: %v", userID, err)
		userParams = nil
	}

	// 构建基础用户信息
	info := h.buildUserInfo(user, userParams)

	// 如果需要完整信息
	if detail == "all" {
		h.addDetailedInfo(info, user)
	}

	repository.Infof("GetUserInfo: successfully found user: %s", userID)
	// 成功响应 (code = 0)
	middleware.Success(c, "获取用户信息成功", info)
}

// buildUserInfo 构建基础用户信息
func (h *UserHandler) buildUserInfo(user *models.User, userParams *models.UserParameters) gin.H {
	// 处理用户参数
	var userParamDict gin.H
	if userParams != nil {
		// 计算综合发布状态：0=不同步，1=同步到草稿箱，2=直接发布上线
		publishStatus := 0
		if userParams.IsWechatAuthorized {
			if userParams.PublishTarget == 1 { // 假设1为草稿箱
				publishStatus = 1
			} else { // 直接发布
				publishStatus = 2
			}
		}

		userParamDict = gin.H{
			"enable_head_info":      userParams.EnableHeadInfo,
			"enable_knowledge_base": userParams.EnableKnowledgeBase,
			"default_theme":         userParams.DefaultTheme,
			"is_wechat_authorized":  userParams.IsWechatAuthorized,
			"has_auth_reminded":     userParams.HasAuthReminded,
			"is_gzh_bind":           userParams.IsGzhBind,
			"publish_target":        userParams.PublishTarget,
			"publish_status":        publishStatus,
		}
	}

	// 安全地处理指针字段
	var nickname, avatar, openid, username, phone, email, appid interface{}
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	if user.Avatar != nil {
		avatar = *user.Avatar
	}
	if user.OpenID != nil {
		openid = *user.OpenID
	}
	if user.Username != nil {
		username = *user.Username
	}
	if user.Phone != nil {
		phone = *user.Phone
	}
	if user.Email != nil {
		email = *user.Email
	}
	if user.AppID != nil {
		appid = *user.AppID
	}

	// 构建基础返回信息
	info := gin.H{
		"id":                user.UserID,
		"nickname":          nickname,
		"avatar":            avatar,
		"openid":            openid,
		"username":          username,
		"phone":             phone,
		"email":             email,
		"appid":             appid,
		"credits":           user.Credits,
		"is_active":         user.IsActive,
		"vip_level":         user.VipLevel,
		"role":              user.Role,
		"status":            user.Status,
		"registration_date": user.RegistrationDate.Format("2006-01-02T15:04:05Z07:00"),
		"last_login_time":   user.LastLoginTime.Format("2006-01-02T15:04:05Z07:00"),
		"user_param":        userParamDict,
		"created_at":        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return info
}

// addDetailedInfo 添加详细信息 (对应Python中的detail="all"逻辑)
func (h *UserHandler) addDetailedInfo(info gin.H, user *models.User) {
	// 这里可以添加更多详细信息，如：
	// - 邀请码信息
	// - 邀请人信息
	// - 使用次数统计
	// - 权益信息等

	// 暂时添加一些基础的详细信息
	info["invitation_code"] = nil         // 邀请码
	info["invitation_count"] = 0          // 邀请数量
	info["inviter"] = nil                 // 邀请人信息
	info["usage_count"] = user.UsageCount // 使用次数
	info["benefit_info"] = gin.H{         // 权益信息
		"membership_name": "",
		"is_active":       false,
		"expire_time":     nil,
		"production_info": nil,
	}
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 调用服务层更新
	user, err := h.userService.Update(userID, &req)
	if err != nil {
		repository.Errorf("UpdateUserInfo failed: %v", err)
		if err.Error() == "user not found" {
			middleware.HandleError(c, middleware.NewBusinessError(404, "用户不存在"))
			return
		}
		if err.Error() == "email already exists" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "邮箱已被使用"))
			return
		}
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败: "+err.Error()))
		return
	}

	// 构建响应
	response := gin.H{
		"user_id":  user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"phone":    user.Phone,
		"message":  "更新成功",
	}

	middleware.Success(c, "更新成功", response)
}

// GetUserParameters 获取用户参数配置
func (h *UserHandler) GetUserParameters(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	params, err := h.userService.GetUserParameters(userID)
	if err != nil {
		repository.Errorf("GetUserParameters failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取用户参数失败: "+err.Error()))
		return
	}

	// 构建响应
	response := gin.H{
		"enable_head_info":      params.EnableHeadInfo,
		"enable_knowledge_base": params.EnableKnowledgeBase,
		"default_theme":         params.DefaultTheme,
		"is_wechat_authorized":  params.IsWechatAuthorized,
		"has_auth_reminded":     params.HasAuthReminded,
		"is_gzh_bind":           params.IsGzhBind,
		"publish_target":        params.PublishTarget,
	}

	middleware.Success(c, "获取用户参数成功", response)
}

// UpdateUserParameters 更新用户参数配置
func (h *UserHandler) UpdateUserParameters(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req models.UserParameters
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	// 调用服务层更新
	if err := h.userService.UpdateUserParameters(userID, &req); err != nil {
		repository.Errorf("UpdateUserParameters failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新用户参数失败: "+err.Error()))
		return
	}

	middleware.Success(c, "更新用户参数成功", gin.H{
		"message": "更新成功",
	})
}

// GetUserSessions 获取用户活跃会话
func (h *UserHandler) GetUserSessions(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	sessions, err := h.userService.GetActiveSessions(userID)
	if err != nil {
		repository.Errorf("GetUserSessions failed: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "获取会话列表失败: "+err.Error()))
		return
	}

	// 构建响应
	var sessionList []gin.H
	for _, session := range sessions {
		sessionList = append(sessionList, gin.H{
			"id":         session.ID,
			"ip_address": session.IPAddress,
			"login_type": session.LoginType,
			"created_at": session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	middleware.Success(c, "获取会话列表成功", gin.H{
		"sessions": sessionList,
		"count":    len(sessionList),
	})
}

// GetConfigTemplateList 获取配置模板列表（客户端接口）
// @Summary 获取配置模板列表
// @Description 获取当前用户的配置模板列表，支持分页和筛选
// @Tags user-config-template
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param template_type query int false "模板类型"
// @Param status query int false "状态"
// @Param name query string false "名称搜索"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config-template/list [get]
func (h *UserHandler) GetConfigTemplateList(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req struct {
		Page         int                              `form:"page" binding:"min=1"`
		PageSize     int                              `form:"page_size" binding:"min=1,max=100"`
		TemplateType *models.ConfigTemplateType       `form:"template_type"`
		Status       *models.ConfigTemplateStatus     `form:"status"`
		Name         string                           `form:"name"`
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
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// 构建查询，只查询当前用户的模板
	query := repository.DB.Model(&models.UserConfigTemplate{}).Where("user_id = ?", userID)

	// 筛选条件
	if req.TemplateType != nil {
		query = query.Where("template_type = ?", *req.TemplateType)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		repository.Errorf("查询配置模板总数失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 获取列表
	var templates []models.UserConfigTemplate
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&templates).Error; err != nil {
		repository.Errorf("查询配置模板列表失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 序列化结果
	items := make([]map[string]interface{}, 0, len(templates))
	for _, t := range templates {
		var configData interface{}
		var metadata interface{}

		if t.ConfigData != "" {
			json.Unmarshal([]byte(t.ConfigData), &configData)
		}
		if t.Metadata != nil && *t.Metadata != "" {
			json.Unmarshal([]byte(*t.Metadata), &metadata)
		}

		items = append(items, map[string]interface{}{
			"id":            t.ID,
			"user_id":       t.UserID,
			"name":          t.Name,
			"template_type": t.TemplateType,
			"config_data":   configData,
			"metadata":      metadata,
			"status":        t.Status,
			"created_at":    t.CreatedAt.Format("2006-01-02 %H:%M:%S"),
			"updated_at":    t.UpdatedAt.Format("2006-01-02 %H:%M:%S"),
		})
	}

	middleware.Success(c, "获取配置模板列表成功", gin.H{
		"items":     items,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// SaveConfigTemplate 创建或更新配置模板（客户端接口）
// @Summary 创建或更新配置模板
// @Description 创建或更新当前用户的配置模板
// @Tags user-config-template
// @Accept json
// @Produce json
// @Param body body models.SaveConfigTemplateRequest true "配置模板信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config-template/save [post]
func (h *UserHandler) SaveConfigTemplate(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var req models.SaveConfigTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	var template models.UserConfigTemplate
	var isUpdate bool

	if req.ID != nil {
		// 更新模式，验证模板属于当前用户
		if err := repository.DB.Where("id = ? AND user_id = ?", *req.ID, userID).First(&template).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				middleware.HandleError(c, middleware.NewBusinessError(404, "未找到配置模板"))
				return
			}
			repository.Errorf("查询配置模板失败: %v", err)
			middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
			return
		}
		isUpdate = true
	} else {
		// 创建模式，使用当前登录用户ID
		template.UserID = userID
		template.CreatedAt = time.Now()
	}

	// 更新字段
	if req.Name != nil {
		if len(*req.Name) > 128 {
			middleware.HandleError(c, middleware.NewBusinessError(400, "模板名称不能超过128个字符"))
			return
		}
		template.Name = *req.Name
	} else if !isUpdate {
		template.Name = "未命名模板"
	}

	if req.TemplateType != nil {
		template.TemplateType = *req.TemplateType
	} else if !isUpdate {
		template.TemplateType = models.ConfigTemplateTypeCustom
	}

	if req.ConfigData != nil {
		configDataBytes, err := json.Marshal(req.ConfigData)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "config_data格式错误: "+err.Error()))
			return
		}
		template.ConfigData = string(configDataBytes)
	} else if !isUpdate {
		template.ConfigData = "{}"
	}

	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			middleware.HandleError(c, middleware.NewBusinessError(400, "metadata格式错误: "+err.Error()))
			return
		}
		metadataStr := string(metadataBytes)
		template.Metadata = &metadataStr
	} else if !isUpdate {
		emptyMetadata := "{}"
		template.Metadata = &emptyMetadata
	}

	if req.Status != nil {
		template.Status = *req.Status
	} else if !isUpdate {
		template.Status = models.ConfigTemplateStatusActive
	}

	template.UpdatedAt = time.Now()

	// 保存
	if err := repository.DB.Save(&template).Error; err != nil {
		repository.Errorf("保存配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "保存失败"))
		return
	}

	// 返回结果
	var configData interface{}
	var metadata interface{}

	if template.ConfigData != "" {
		json.Unmarshal([]byte(template.ConfigData), &configData)
	}
	if template.Metadata != nil && *template.Metadata != "" {
		json.Unmarshal([]byte(*template.Metadata), &metadata)
	}

	middleware.Success(c, "保存配置模板成功", gin.H{
		"id":            template.ID,
		"user_id":       template.UserID,
		"name":          template.Name,
		"template_type": template.TemplateType,
		"config_data":   configData,
		"metadata":      metadata,
		"status":        template.Status,
		"created_at":    template.CreatedAt.Format("2006-01-02 %H:%M:%S"),
		"updated_at":    template.UpdatedAt.Format("2006-01-02 %H:%M:%S"),
	})
}

// GetConfigTemplateDetail 获取配置模板详情（客户端接口）
// @Summary 获取配置模板详情
// @Description 获取当前用户的配置模板详情
// @Tags user-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config-template/{id} [get]
func (h *UserHandler) GetConfigTemplateDetail(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var id int
	if err := c.ShouldBindUri(&id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.UserConfigTemplate
	if err := repository.DB.Where("id = ? AND user_id = ?", id, userID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "未找到对应配置模板"))
			return
		}
		repository.Errorf("查询配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	var configData interface{}
	var metadata interface{}

	if template.ConfigData != "" {
		json.Unmarshal([]byte(template.ConfigData), &configData)
	}
	if template.Metadata != nil && *template.Metadata != "" {
		json.Unmarshal([]byte(*template.Metadata), &metadata)
	}

	middleware.Success(c, "获取配置模板详情成功", gin.H{
		"id":            template.ID,
		"user_id":       template.UserID,
		"name":          template.Name,
		"template_type": template.TemplateType,
		"config_data":   configData,
		"metadata":      metadata,
		"status":        template.Status,
		"created_at":    template.CreatedAt.Format("2006-01-02 %H:%M:%S"),
		"updated_at":    template.UpdatedAt.Format("2006-01-02 %H:%M:%S"),
	})
}

// DeleteConfigTemplate 删除配置模板（客户端接口）
// @Summary 删除配置模板
// @Description 删除当前用户的配置模板
// @Tags user-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config-template/{id} [delete]
func (h *UserHandler) DeleteConfigTemplate(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var id int
	if err := c.ShouldBindUri(&id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var template models.UserConfigTemplate
	if err := repository.DB.Where("id = ? AND user_id = ?", id, userID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "未找到配置模板，无法删除"))
			return
		}
		repository.Errorf("查询配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	if err := repository.DB.Delete(&template).Error; err != nil {
		repository.Errorf("删除配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "删除失败"))
		return
	}

	middleware.Success(c, "删除配置模板成功", nil)
}

// UpdateConfigTemplateStatus 更新配置模板状态（客户端接口）
// @Summary 更新配置模板状态
// @Description 更新当前用户的配置模板状态（启用/禁用）
// @Tags user-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param status query int true "状态: 0=禁用, 1=启用"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config-template/{id}/status [put]
func (h *UserHandler) UpdateConfigTemplateStatus(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var id int
	if err := c.ShouldBindUri(&id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var req struct {
		Status models.ConfigTemplateStatus `form:"status" binding:"required"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "状态参数不能为空"))
		return
	}

	var template models.UserConfigTemplate
	if err := repository.DB.Where("id = ? AND user_id = ?", id, userID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "未找到配置模板"))
			return
		}
		repository.Errorf("查询配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	template.Status = req.Status
	template.UpdatedAt = time.Now()

	if err := repository.DB.Save(&template).Error; err != nil {
		repository.Errorf("更新配置模板状态失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "更新失败"))
		return
	}

	var configData interface{}
	var metadata interface{}

	if template.ConfigData != "" {
		json.Unmarshal([]byte(template.ConfigData), &configData)
	}
	if template.Metadata != nil && *template.Metadata != "" {
		json.Unmarshal([]byte(*template.Metadata), &metadata)
	}

	middleware.Success(c, "更新配置模板状态成功", gin.H{
		"id":            template.ID,
		"user_id":       template.UserID,
		"name":          template.Name,
		"template_type": template.TemplateType,
		"config_data":   configData,
		"metadata":      metadata,
		"status":        template.Status,
		"created_at":    template.CreatedAt.Format("2006-01-02 %H:%M:%S"),
		"updated_at":    template.UpdatedAt.Format("2006-01-02 %H:%M:%S"),
	})
}

// DuplicateConfigTemplate 复制配置模板（客户端接口）
// @Summary 复制配置模板
// @Description 复制当前用户的一个配置模板
// @Tags user-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config-template/duplicate/{id} [post]
func (h *UserHandler) DuplicateConfigTemplate(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.HandleError(c, middleware.NewBusinessError(401, "未授权访问"))
		return
	}

	var id int
	if err := c.ShouldBindUri(&id); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}

	var original models.UserConfigTemplate
	if err := repository.DB.Where("id = ? AND user_id = ?", id, userID).First(&original).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.NewBusinessError(404, "未找到原配置模板"))
			return
		}
		repository.Errorf("查询配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "查询失败"))
		return
	}

	// 创建新模板
	newTemplate := models.UserConfigTemplate{
		UserID:       userID,
		Name:         original.Name + " (副本)",
		TemplateType: original.TemplateType,
		ConfigData:   original.ConfigData,
		Metadata:     original.Metadata,
		Status:       models.ConfigTemplateStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repository.DB.Create(&newTemplate).Error; err != nil {
		repository.Errorf("复制配置模板失败: %v", err)
		middleware.HandleError(c, middleware.NewBusinessError(500, "复制失败"))
		return
	}

	var configData interface{}
	var metadata interface{}

	if newTemplate.ConfigData != "" {
		json.Unmarshal([]byte(newTemplate.ConfigData), &configData)
	}
	if newTemplate.Metadata != nil && *newTemplate.Metadata != "" {
		json.Unmarshal([]byte(*newTemplate.Metadata), &metadata)
	}

	middleware.Success(c, "复制配置模板成功", gin.H{
		"id":            newTemplate.ID,
		"user_id":       newTemplate.UserID,
		"name":          newTemplate.Name,
		"template_type": newTemplate.TemplateType,
		"config_data":   configData,
		"metadata":      metadata,
		"status":        newTemplate.Status,
		"created_at":    newTemplate.CreatedAt.Format("2006-01-02 %H:%M:%S"),
		"updated_at":    newTemplate.UpdatedAt.Format("2006-01-02 %H:%M:%S"),
	})
}

// SetupUserRoutes 设置用户路由
func SetupUserRoutes(r *gin.Engine, userHandler *UserHandler) {
	// 公开路由
	public := r.Group("/api/v1")
	{
		// 基础路由
		public.GET("/hello", userHandler.Hello)
		public.GET("/", userHandler.Hello)
	}

	// 需要认证的用户路由 - 放在 /api/v1/user 下
	userGroup := r.Group("/api/v1/user")
	userGroup.Use(middleware.JWTAuth())
	{
		// 用户信息接口 - /api/v1/user/info
		userGroup.GET("/info", userHandler.GetUserInfo)
		// 更新用户信息 - /api/v1/user/info
		userGroup.PUT("/info", userHandler.UpdateUserInfo)
		// 获取用户参数 - /api/v1/user/parameters
		userGroup.GET("/parameters", userHandler.GetUserParameters)
		// 更新用户参数 - /api/v1/user/parameters
		userGroup.PUT("/parameters", userHandler.UpdateUserParameters)
		// 获取用户会话列表 - /api/v1/user/sessions
		userGroup.GET("/sessions", userHandler.GetUserSessions)
	}

	// 配置模板路由 - 放在 /api/v1/config-template 下（与Python路径一致）
	configTemplateGroup := r.Group("/api/v1/config-template")
	configTemplateGroup.Use(middleware.JWTAuth())
	{
		configTemplateGroup.GET("/list", userHandler.GetConfigTemplateList)
		configTemplateGroup.POST("/save", userHandler.SaveConfigTemplate)
		configTemplateGroup.GET("/:id", userHandler.GetConfigTemplateDetail)
		configTemplateGroup.DELETE("/:id", userHandler.DeleteConfigTemplate)
		configTemplateGroup.PUT("/:id/status", userHandler.UpdateConfigTemplateStatus)
		configTemplateGroup.POST("/duplicate/:id", userHandler.DuplicateConfigTemplate)
	}
}

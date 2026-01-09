package admin

import (
	"encoding/json"
	"strconv"
	"time"

	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetConfigTemplateList 获取配置模板列表（管理员接口）
// @Summary 获取配置模板列表
// @Description 管理员获取配置模板列表，支持分页和筛选
// @Tags admin-config-template
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param template_type query int false "模板类型"
// @Param status query int false "状态"
// @Param name query string false "名称搜索"
// @Param user_id query string false "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/config_template/list [get]
func (h *AdminHandler) GetConfigTemplateList(c *gin.Context) {
	var req models.ConfigTemplateListRequest
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

	// 构建查询
	query := repository.DB.Model(&models.UserConfigTemplate{})

	// 筛选条件
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
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
			"created_at":    t.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":    t.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.Success(c, "获取配置模板列表成功", gin.H{
		"items":     items,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetConfigTemplateDetail 获取配置模板详情（管理员接口）
// @Summary 获取配置模板详情
// @Description 管理员获取配置模板详情
// @Tags admin-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/config_template/{id} [get]
func (h *AdminHandler) GetConfigTemplateDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID格式错误"))
		return
	}

	var template models.UserConfigTemplate
	if err := repository.DB.Where("id = ?", id).First(&template).Error; err != nil {
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
		"created_at":    template.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":    template.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// SaveConfigTemplate 创建或更新配置模板（管理员接口）
// @Summary 创建或更新配置模板
// @Description 管理员创建或更新配置模板
// @Tags admin-config-template
// @Accept json
// @Produce json
// @Param body body models.SaveConfigTemplateRequest true "配置模板信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/config_template/save [post]
func (h *AdminHandler) SaveConfigTemplate(c *gin.Context) {
	var req models.SaveConfigTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "参数错误: "+err.Error()))
		return
	}

	var template models.UserConfigTemplate
	var isUpdate bool

	if req.ID != nil {
		// 更新模式
		if err := repository.DB.Where("id = ?", *req.ID).First(&template).Error; err != nil {
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
		// 创建模式，需要user_id
		if req.UserID == "" {
			// 尝试从query参数获取
			req.UserID = c.Query("user_id")
		}
		if req.UserID == "" {
			middleware.HandleError(c, middleware.NewBusinessError(400, "创建模板时user_id不能为空"))
			return
		}
		template.UserID = req.UserID
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
		"created_at":    template.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":    template.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DeleteConfigTemplate 删除配置模板（管理员接口）
// @Summary 删除配置模板
// @Description 管理员删除配置模板
// @Tags admin-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/config_template/{id} [delete]
func (h *AdminHandler) DeleteConfigTemplate(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID格式错误"))
		return
	}

	var template models.UserConfigTemplate
	if err := repository.DB.Where("id = ?", id).First(&template).Error; err != nil {
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

// UpdateConfigTemplateStatus 更新配置模板状态（管理员接口）
// @Summary 更新配置模板状态
// @Description 管理员更新配置模板状态（启用/禁用）
// @Tags admin-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param status query int true "状态: 0=禁用, 1=启用"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/config_template/{id}/status [put]
func (h *AdminHandler) UpdateConfigTemplateStatus(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID格式错误"))
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
	if err := repository.DB.Where("id = ?", id).First(&template).Error; err != nil {
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
		"created_at":    template.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":    template.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// DuplicateConfigTemplate 复制配置模板（管理员接口）
// @Summary 复制配置模板
// @Description 管理员复制一个配置模板
// @Tags admin-config-template
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/config_template/{id}/duplicate [post]
func (h *AdminHandler) DuplicateConfigTemplate(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID不能为空"))
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.HandleError(c, middleware.NewBusinessError(400, "模板ID格式错误"))
		return
	}

	var original models.UserConfigTemplate
	if err := repository.DB.Where("id = ?", id).First(&original).Error; err != nil {
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
		UserID:       original.UserID,
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
		"created_at":    newTemplate.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":    newTemplate.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

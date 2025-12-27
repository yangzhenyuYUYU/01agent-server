package router

import (
	"net/http"

	"gin_web/internal/models"
	"gin_web/internal/service"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configService: service.NewConfigService(),
	}
}

// GetConfig 获取当前配置（敏感信息已隐藏）
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	config := h.configService.GetSafeConfig()
	if config == nil {
		c.JSON(http.StatusInternalServerError, models.NewResponse(500, "Configuration not loaded", nil))
		return
	}

	c.JSON(http.StatusOK, models.NewResponse(200, "Success", config))
}

// GetThemes 获取主题配置
func (h *ConfigHandler) GetThemes(c *gin.Context) {
	themes := h.configService.GetThemes()
	c.JSON(http.StatusOK, models.NewResponse(200, "Success", themes))
}

// GetCreditsConfig 获取积分配置
func (h *ConfigHandler) GetCreditsConfig(c *gin.Context) {
	credits := h.configService.GetCreditsConfig()
	c.JSON(http.StatusOK, models.NewResponse(200, "Success", credits))
}

// SetupConfigRoutes 设置配置路由
func SetupConfigRoutes(r *gin.Engine, configHandler *ConfigHandler) {
	// 公开路由
	public := r.Group("/api/v1")
	{
		// 配置信息（公开）
		public.GET("/config", configHandler.GetConfig)
		public.GET("/config/themes", configHandler.GetThemes)
		public.GET("/config/credits", configHandler.GetCreditsConfig)
	}
}

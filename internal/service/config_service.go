package service

import (
	"01agent_server/internal/config"
)

type ConfigService struct{}

// NewConfigService 创建配置服务
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// GetSafeConfig 获取安全的配置信息（隐藏敏感信息）
func (s *ConfigService) GetSafeConfig() map[string]interface{} {
	cfg := config.AppConfig
	if cfg == nil {
		return nil
	}

	return map[string]interface{}{
		"server": map[string]interface{}{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		},
		"database": map[string]interface{}{
			"type": cfg.Database.Type,
		},
		"jwt": map[string]interface{}{
			"expire": cfg.JWT.Expire.String(),
		},
		"credits": cfg.Credits,
		"themes":  cfg.Themes,
		"bp": map[string]interface{}{
			"documentPath":      cfg.BP.DocumentPath,
			"emailTemplatePath": cfg.BP.EmailTemplatePath,
			"weixinQRPath":      cfg.BP.WeixinQRPath,
			"websiteURL":        cfg.BP.WebsiteURL,
		},
	}
}

// GetThemes 获取主题配置
func (s *ConfigService) GetThemes() map[string]string {
	if config.AppConfig == nil {
		return make(map[string]string)
	}
	return config.AppConfig.Themes
}

// GetCreditsConfig 获取积分配置
func (s *ConfigService) GetCreditsConfig() config.CreditsConfig {
	if config.AppConfig == nil {
		return config.CreditsConfig{}
	}
	return config.AppConfig.Credits
}

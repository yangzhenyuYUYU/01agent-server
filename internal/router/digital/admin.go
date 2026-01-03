package digital

import (
	"gin_web/internal/middleware"
	"gin_web/internal/models/digital"
	"gin_web/internal/repository"
	"gin_web/internal/tools"

	"github.com/gin-gonic/gin"
)

// SetupDigitalAdminRoutes 设置数字人管理端路由
func SetupDigitalAdminRoutes(r *gin.Engine) {
	digitalAdmin := r.Group("/admin")
	digitalAdmin.Use(middleware.AdminAuth())

	// 数字人分类管理 CRUD
	categoryCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.DigitalCategory{},
		SearchFields:   []string{"name"},
		DefaultOrderBy: "position",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	categoryGroup := digitalAdmin.Group("/digital/category")
	{
		categoryGroup.GET("/list", categoryCRUD.List)
		categoryGroup.GET("/:id", categoryCRUD.Detail)
		categoryGroup.POST("", categoryCRUD.Create)
		categoryGroup.PUT("/:id", categoryCRUD.Update)
		categoryGroup.DELETE("/:id", categoryCRUD.Delete)
	}

	// 数字人国家管理 CRUD
	countryCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.DigitalCountry{},
		SearchFields:   []string{"name", "english_name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	countryGroup := digitalAdmin.Group("/digital/country")
	{
		countryGroup.GET("/list", countryCRUD.List)
		countryGroup.GET("/:id", countryCRUD.Detail)
		countryGroup.POST("", countryCRUD.Create)
		countryGroup.PUT("/:id", countryCRUD.Update)
		countryGroup.DELETE("/:id", countryCRUD.Delete)
	}

	// 数字人提示词管理 CRUD
	promptCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.DigitalPrompt{},
		SearchFields:   []string{"name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	promptGroup := digitalAdmin.Group("/digital/prompt")
	{
		promptGroup.GET("/list", promptCRUD.List)
		promptGroup.GET("/:id", promptCRUD.Detail)
		promptGroup.POST("", promptCRUD.Create)
		promptGroup.PUT("/:id", promptCRUD.Update)
		promptGroup.DELETE("/:id", promptCRUD.Delete)
	}

	// 数字人模板管理 CRUD
	templateCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.DigitalTemplate{},
		SearchFields:   []string{"name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	templateGroup := digitalAdmin.Group("/digital-template")
	{
		templateGroup.GET("/list", templateCRUD.List)
		templateGroup.GET("/:id", templateCRUD.Detail)
		templateGroup.POST("", templateCRUD.Create)
		templateGroup.PUT("/:id", templateCRUD.Update)
		templateGroup.DELETE("/:id", templateCRUD.Delete)
	}

	// 语音模型管理 CRUD
	voiceModelCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.VoiceModel{},
		SearchFields:   []string{"name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	voiceModelGroup := digitalAdmin.Group("/voice-model")
	{
		voiceModelGroup.GET("/list", voiceModelCRUD.List)
		voiceModelGroup.GET("/:id", voiceModelCRUD.Detail)
		voiceModelGroup.POST("", voiceModelCRUD.Create)
		voiceModelGroup.PUT("/:id", voiceModelCRUD.Update)
		voiceModelGroup.DELETE("/:id", voiceModelCRUD.Delete)
	}

	// 语音合成记录管理 CRUD
	voiceSynthesisCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.VoiceSynthesisRecord{},
		SearchFields:   []string{"text_content"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	voiceSynthesisGroup := digitalAdmin.Group("/voice-synthesis-record")
	{
		voiceSynthesisGroup.GET("/list", voiceSynthesisCRUD.List)
		voiceSynthesisGroup.GET("/:id", voiceSynthesisCRUD.Detail)
		voiceSynthesisGroup.POST("", voiceSynthesisCRUD.Create)
		voiceSynthesisGroup.PUT("/:id", voiceSynthesisCRUD.Update)
		voiceSynthesisGroup.DELETE("/:id", voiceSynthesisCRUD.Delete)
	}

	// 语音训练任务管理 CRUD
	voiceTrainTaskCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.VoiceTrainTask{},
		SearchFields:   []string{"task_name"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	voiceTrainTaskGroup := digitalAdmin.Group("/voice-train-task")
	{
		voiceTrainTaskGroup.GET("/list", voiceTrainTaskCRUD.List)
		voiceTrainTaskGroup.GET("/:id", voiceTrainTaskCRUD.Detail)
		voiceTrainTaskGroup.POST("", voiceTrainTaskCRUD.Create)
		voiceTrainTaskGroup.PUT("/:id", voiceTrainTaskCRUD.Update)
		voiceTrainTaskGroup.DELETE("/:id", voiceTrainTaskCRUD.Delete)
	}

	// 语音训练音频管理 CRUD
	voiceTrainAudioCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.VoiceTrainAudio{},
		SearchFields:   []string{"text_content"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	voiceTrainAudioGroup := digitalAdmin.Group("/voice-train-audio")
	{
		voiceTrainAudioGroup.GET("/list", voiceTrainAudioCRUD.List)
		voiceTrainAudioGroup.GET("/:id", voiceTrainAudioCRUD.Detail)
		voiceTrainAudioGroup.POST("", voiceTrainAudioCRUD.Create)
		voiceTrainAudioGroup.PUT("/:id", voiceTrainAudioCRUD.Update)
		voiceTrainAudioGroup.DELETE("/:id", voiceTrainAudioCRUD.Delete)
	}

	// 数字人模板订单管理 CRUD
	templateOrderCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
		Model:          &digital.DigitalTemplateOrder{},
		SearchFields:   []string{"user_id"},
		DefaultOrderBy: "created_at",
		RequireAdmin:   true,
		PrimaryKey:     "id",
	}, repository.DB)
	templateOrderGroup := digitalAdmin.Group("/digital-template-order")
	{
		templateOrderGroup.GET("/list", templateOrderCRUD.List)
		templateOrderGroup.GET("/:id", templateOrderCRUD.Detail)
		templateOrderGroup.POST("", templateOrderCRUD.Create)
		templateOrderGroup.PUT("/:id", templateOrderCRUD.Update)
		templateOrderGroup.DELETE("/:id", templateOrderCRUD.Delete)
	}

	// 系统数字人列表接口 - /admin/digital/system/digital-human/list
	systemGroup := digitalAdmin.Group("/digital")
	{
		systemGroup.GET("/system/digital-human/list", GetDigitalHumanList)
	}

	// 系统语音训练公开列表接口 - /admin/system/voice/train/public/list
	systemVoiceGroup := r.Group("/admin/system/voice/train")
	systemVoiceGroup.Use(middleware.AdminAuth())
	{
		systemVoiceGroup.GET("/public/list", GetPublicVoiceTrainList)
	}
}

// GetDigitalHumanList 获取系统数字人列表
func GetDigitalHumanList(c *gin.Context) {
	// TODO: 实现获取系统数字人列表的逻辑
	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": []interface{}{}})
}

// GetPublicVoiceTrainList 获取公开语音训练列表
func GetPublicVoiceTrainList(c *gin.Context) {
	// TODO: 实现获取公开语音训练列表的逻辑
	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": []interface{}{}})
}

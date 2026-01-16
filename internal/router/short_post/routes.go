package short_post

import (
	"01agent_server/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupShortPostRoutes setup short post routes
func SetupShortPostRoutes(r *gin.Engine) {
	projectHandler := NewProjectHandler()
	exportHandler := NewExportHandler()

	// 短图文工程管理路由
	projectGroup := r.Group("/api/v1/short-post/project")
	projectGroup.Use(middleware.JWTAuth())
	{
		projectGroup.POST("", projectHandler.CreateProject)
		projectGroup.GET("/list", projectHandler.GetProjectList)
		projectGroup.GET("/all/content", projectHandler.GetAllContentList) // 管理员接口
		projectGroup.GET("/:project_id", projectHandler.GetProjectDetail)
		projectGroup.GET("/:project_id/has-content", projectHandler.CheckProjectHasContent)
		projectGroup.PUT("/:project_id", projectHandler.UpdateProject)
		projectGroup.POST("/:project_id/save", projectHandler.SaveProjectContent)
		projectGroup.DELETE("/:project_id", projectHandler.DeleteProject)
		projectGroup.GET("/:project_id/versions", projectHandler.GetProjectVersions)
		projectGroup.POST("/:project_id/copywriting", projectHandler.SaveCopywriting)
		projectGroup.GET("/:project_id/copywriting", projectHandler.GetCopywriting)
	}

	// 短图文导出管理路由
	exportGroup := r.Group("/api/v1/short-post/export")
	exportGroup.Use(middleware.JWTAuth())
	{
		exportGroup.POST("", exportHandler.CreateExportRecord)
		exportGroup.GET("/list", exportHandler.GetExportList)
		exportGroup.GET("/:export_id", exportHandler.GetExportDetail)
		exportGroup.DELETE("/:export_id", exportHandler.DeleteExportRecord)
		exportGroup.DELETE("/batch", exportHandler.BatchDeleteExportRecords)
	}
}


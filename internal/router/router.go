package router

import (
	"net/http"
	"strings"
	"time"

	"01agent_server/internal/config"
	"01agent_server/internal/middleware"
	"01agent_server/internal/models"
	"01agent_server/internal/router/admin"
	"01agent_server/internal/router/digital"
	utils "01agent_server/internal/tools"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建Gin实例
	r := gin.New()

	// 添加中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.ErrorHandler()) // 统一错误处理
	r.Use(middleware.CORS())

	// 创建处理器
	authHandler := NewAuthHandler()
	userHandler := NewUserHandler()
	configHandler := NewConfigHandler()
	concurrentHandler := NewConcurrentHandler()

	// 设置各模块路由
	SetupAuthRoutes(r, authHandler)
	SetupUserRoutes(r, userHandler)
	SetupConfigRoutes(r, configHandler)
	SetupConcurrentRoutes(r, concurrentHandler)
	admin.SetupAdminRoutes(r)          // 管理员路由
	digital.SetupDigitalAdminRoutes(r) // 数字人管理端路由
	RegisterBlogRoutes(r)              // 博客路由
	SetupArticleEditRoutes(r)          // 文章编辑路由

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		// 使用统一的成功响应格式 (code = 0)
		middleware.Success(c, "服务器运行正常", gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// JWT调试接口
	SetupJWTDebugRoutes(r)

	return r
}

// SetupJWTDebugRoutes 设置JWT调试路由
func SetupJWTDebugRoutes(r *gin.Engine) {
	debug := r.Group("/debug/jwt")
	{
		// 生成测试token（短期过期，用于测试）
		debug.POST("/generate-test-token", func(c *gin.Context) {
			var req struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				ExpireIn string `json:"expire_in"` // 如 "1s", "1m", "1h"
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "参数错误: "+err.Error()))
				return
			}

			if req.UserID == "" {
				req.UserID = "test_user_123"
			}
			if req.Username == "" {
				req.Username = "test_user"
			}

			// 生成token
			token, err := utils.GenerateToken(req.UserID, req.Username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "生成token失败: "+err.Error()))
				return
			}

			// 获取token信息
			info, _ := utils.GetTokenInfo(token)

			middleware.Success(c, "测试token生成成功", gin.H{
				"token": token,
				"info":  info,
			})
		})

		// 检查token信息
		debug.POST("/check-token", func(c *gin.Context) {
			var req struct {
				Token string `json:"token"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "参数错误: "+err.Error()))
				return
			}

			if req.Token == "" {
				// 尝试从Authorization头获取
				authHeader := c.GetHeader("Authorization")
				if authHeader != "" {
					req.Token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if req.Token == "" {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "缺少token"))
				return
			}

			// 获取token详细信息
			info, err := utils.GetTokenInfo(req.Token)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "解析token失败: "+err.Error()))
				return
			}

			// 检查是否过期
			expired, expErr := utils.IsTokenExpired(req.Token)
			if expErr != nil {
				info["expiry_check_error"] = expErr.Error()
			} else {
				info["is_expired_check"] = expired
			}

			// 尝试验证token
			validateErr := utils.ValidateTokenStrict(req.Token)
			if validateErr != nil {
				info["validation_error"] = validateErr.Error()
				info["is_valid"] = false
			} else {
				info["is_valid"] = true
			}

			middleware.Success(c, "Token信息", info)
		})

		// 测试过期token检测
		debug.POST("/test-expired-token", func(c *gin.Context) {
			// 生成一个已经过期的token（过期时间设为1秒前）
			userID := "test_expired_user"
			username := "expired_test"

			// 手动创建一个过期的token
			claims := utils.Claims{
				UserID:   userID,
				Username: username,
			}

			// 设置过期时间为1秒前
			expiredTime := time.Now().Add(-1 * time.Second)
			claims.ExpiresAt = &jwt.NumericDate{Time: expiredTime}
			claims.IssuedAt = &jwt.NumericDate{Time: time.Now().Add(-2 * time.Second)}
			claims.NotBefore = &jwt.NumericDate{Time: time.Now().Add(-2 * time.Second)}
			claims.Subject = userID
			claims.Issuer = "gin_web"

			// 生成token
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			cfg := config.AppConfig.JWT
			tokenString, err := token.SignedString([]byte(cfg.Secret))
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "生成过期token失败: "+err.Error()))
				return
			}

			// 检查token
			info, _ := utils.GetTokenInfo(tokenString)
			expired, _ := utils.IsTokenExpired(tokenString)
			validateErr := utils.ValidateTokenStrict(tokenString)

			middleware.Success(c, "过期token测试", gin.H{
				"expired_token": tokenString,
				"info":          info,
				"is_expired":    expired,
				"validation_error": func() string {
					if validateErr != nil {
						return validateErr.Error()
					}
					return ""
				}(),
			})
		})

		// 测试JWT中间件
		debug.GET("/test-middleware", middleware.JWTAuth(), func(c *gin.Context) {
			userID, _ := middleware.GetCurrentUserID(c)
			username, _ := middleware.GetCurrentUsername(c)

			middleware.Success(c, "JWT中间件测试成功", gin.H{
				"user_id":  userID,
				"username": username,
				"message":  "如果你看到这个消息，说明JWT中间件验证通过",
			})
		})
	}
}

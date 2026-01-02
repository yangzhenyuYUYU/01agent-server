package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"gin_web/internal/models"
	"gin_web/internal/repository"
	utils "gin_web/internal/tools"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "缺少授权头"))
			c.Abort()
			return
		}

		// Bearer Token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "授权头格式错误"))
			c.Abort()
			return
		}

		// 解析和验证token
		fmt.Printf("Parsing token: %s\n", tokenString[:50]+"...")
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			repository.Errorf("JWT parsing failed: %v", err)
			fmt.Printf("JWT parsing error: %v\n", err)

			// 根据错误类型返回不同的错误信息
			var errorMsg string
			if err.Error() == "token已过期" {
				errorMsg = "token已过期"
			} else {
				errorMsg = "无效的token"
			}

			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, errorMsg))
			c.Abort()
			return
		}

		fmt.Printf("JWT parsed successfully - UserID: '%s', Username: '%s'\n", claims.UserID, claims.Username)
		repository.Infof("JWT parsed successfully - UserID: %s, Username: %s", claims.UserID, claims.Username)

		// 将用户信息存储到上下文中
		// 如果UserID为空，使用Subject作为UserID
		userID := claims.UserID
		if userID == "" {
			userID = claims.Subject
			fmt.Printf("Using Subject as UserID: '%s'\n", userID)
		}

		c.Set("userID", userID)
		c.Set("username", claims.Username)
		fmt.Printf("Set userID to context: '%s'\n", userID)

		c.Next()
	}
}

// JWTOptional 可选JWT认证中间件
func JWTOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// 检查Bearer前缀
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				if token != "" {
					// 解析token
					claims, err := utils.ParseToken(token)
					if err == nil {
						// 将用户信息存储到上下文中
						c.Set("userID", claims.UserID)
						c.Set("username", claims.Username)
					}
				}
			}
		}
		c.Next()
	}
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	fmt.Println("userIDStr111", userID)

	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	fmt.Println("userIDStr222", userIDStr)
	if !ok {
		return "", false
	}
	fmt.Println("userIDStr333", userIDStr)
	return userIDStr, true
}

// GetCurrentUsername 从上下文中获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", false
	}

	return usernameStr, true
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行JWT认证
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "缺少授权头"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "授权头格式错误"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "无效的token"))
			c.Abort()
			return
		}

		userID := claims.UserID
		if userID == "" {
			userID = claims.Subject
		}

		// 查询用户信息，验证是否为管理员
		var user models.User
		if err := repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "用户不存在"))
			c.Abort()
			return
		}

		// 验证管理员权限（Role = 3 表示管理员）
		if user.Role != 3 {
			c.JSON(http.StatusForbidden, models.ErrorResponse(403, "需要管理员权限"))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userID", userID)
		c.Set("username", claims.Username)
		c.Set("user", &user)

		c.Next()
	}
}

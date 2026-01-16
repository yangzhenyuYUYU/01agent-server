package tools

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"01agent_server/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// GenerateToken 生成JWT token
func GenerateToken(userID string, username string) (string, error) {
	cfg := config.AppConfig.JWT

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID, // 添加Subject字段
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.Expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gin_web",
		},
	}

	// 添加调试信息
	// fmt.Printf("GenerateToken - Creating token for UserID: '%s', Username: '%s'\n", userID, username)
	// fmt.Printf("JWT Config - Secret: %s, Expire: %v\n", cfg.Secret, cfg.Expire)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		// fmt.Printf("GenerateToken - Error signing token: %v\n", err)
		return "", err
	}

	// fmt.Printf("GenerateToken - Token created successfully: %s\n", tokenString[:50]+"...")
	return tokenString, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfig.JWT

	// fmt.Printf("ParseToken - Parsing token: %s\n", tokenString[:50]+"...")
	// fmt.Printf("ParseToken - Using secret: %s\n", cfg.Secret)

	// 使用解析选项，设置验证方法
	// 注意：JWT库会自动验证过期时间，但我们会在后面使用容差再次检查
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithTimeFunc(time.Now),
	)

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		// fmt.Printf("ParseToken - Error parsing token: %v\n", err)
		// 检查错误信息中是否包含过期相关字符串
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "expired") || strings.Contains(errStr, "token is expired") {
			return nil, errors.New("token已过期")
		}
		if strings.Contains(errStr, "not valid yet") {
			return nil, errors.New("token尚未生效")
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		// fmt.Printf("ParseToken - Successfully parsed - UserID: '%s', Username: '%s'\n", claims.UserID, claims.Username)
		// fmt.Printf("ParseToken - Subject: '%s'\n", claims.Subject)
		// fmt.Printf("ParseToken - ExpiresAt: %v\n", claims.ExpiresAt)
		// fmt.Printf("ParseToken - Current time: %v\n", time.Now())

		// 严格检查token有效性
		if !token.Valid {
			// fmt.Printf("ParseToken - Token marked as invalid by JWT library\n")
			return nil, errors.New("invalid token")
		}

		// 多重检查过期时间（添加5秒容差，避免时钟偏差问题）
		currentTime := time.Now()
		if claims.ExpiresAt != nil {
			expiryTime := claims.ExpiresAt.Time
			// 添加5秒容差，允许轻微的时钟偏差
			leeway := 5 * time.Second
			// fmt.Printf("ParseToken - Expiry check: token expires at %v, current time %v, leeway: %v\n", expiryTime, currentTime, leeway)

			// 使用容差检查，避免因时钟偏差导致的误判
			if expiryTime.Before(currentTime.Add(-leeway)) {
				// fmt.Printf("ParseToken - Token expired! Expiry: %v, Current: %v, Diff: %v\n", expiryTime, currentTime, currentTime.Sub(expiryTime))
				return nil, errors.New("token已过期")
			}
		} else {
			// fmt.Printf("ParseToken - Warning: Token has no expiry time set\n")
			return nil, errors.New("token缺少过期时间")
		}

		// 检查NotBefore时间
		if claims.NotBefore != nil && claims.NotBefore.After(currentTime) {
			// fmt.Printf("ParseToken - Token not valid yet: %v > %v\n", claims.NotBefore.Time, currentTime)
			return nil, errors.New("token尚未生效")
		}

		// 如果自定义字段为空，尝试从Subject获取UserID
		if claims.UserID == "" && claims.Subject != "" {
			// fmt.Printf("ParseToken - UserID empty, using Subject as UserID: '%s'\n", claims.Subject)
			claims.UserID = claims.Subject
		}

		return claims, nil
	}

	// fmt.Printf("ParseToken - Token invalid or claims not ok\n")
	return nil, errors.New("invalid token")
}

// RefreshToken 刷新token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查token是否即将过期（还有1小时过期）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return tokenString, nil
	}

	// 生成新token
	return GenerateToken(claims.UserID, claims.Username)
}

// GetStringValue 获取字符串指针的值，如果为nil则返回空字符串
func GetStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}

// Float64Ptr 返回float64指针
func Float64Ptr(f float64) *float64 {
	return &f
}

// IntPtr 返回int指针
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr 返回int64指针
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr 返回bool指针
func BoolPtr(b bool) *bool {
	return &b
}

// Split 分割字符串（避免引入strings包时的循环依赖）
func Split(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}

	var result []string
	for {
		idx := -1
		for i := 0; i <= len(s)-len(sep); i++ {
			if s[i:i+len(sep)] == sep {
				idx = i
				break
			}
		}
		if idx == -1 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

// GetTokenExpiryTime 获取token的过期时间（不验证token有效性，仅解析）
func GetTokenExpiryTime(tokenString string) (*time.Time, error) {
	// 不验证签名，仅解析claims
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok {
		if claims.ExpiresAt != nil {
			expiryTime := claims.ExpiresAt.Time
			return &expiryTime, nil
		}
		return nil, errors.New("token没有设置过期时间")
	}

	return nil, errors.New("无法解析token claims")
}

// IsTokenExpired 检查token是否已过期（不验证签名，仅检查时间）
func IsTokenExpired(tokenString string) (bool, error) {
	expiryTime, err := GetTokenExpiryTime(tokenString)
	if err != nil {
		return true, err
	}

	currentTime := time.Now()
	expired := expiryTime.Before(currentTime) || expiryTime.Equal(currentTime)

	// fmt.Printf("IsTokenExpired - Token expires at: %v, Current time: %v, Expired: %v\n",
	// expiryTime, currentTime, expired)

	return expired, nil
}

// GetTokenInfo 获取token的详细信息（不验证签名）
func GetTokenInfo(tokenString string) (map[string]interface{}, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok {
		info := make(map[string]interface{})
		info["user_id"] = claims.UserID
		info["username"] = claims.Username
		info["subject"] = claims.Subject
		info["issuer"] = claims.Issuer

		if claims.ExpiresAt != nil {
			info["expires_at"] = claims.ExpiresAt.Time
			info["expires_at_unix"] = claims.ExpiresAt.Unix()
		}

		if claims.IssuedAt != nil {
			info["issued_at"] = claims.IssuedAt.Time
			info["issued_at_unix"] = claims.IssuedAt.Unix()
		}

		if claims.NotBefore != nil {
			info["not_before"] = claims.NotBefore.Time
			info["not_before_unix"] = claims.NotBefore.Unix()
		}

		currentTime := time.Now()
		info["current_time"] = currentTime
		info["current_time_unix"] = currentTime.Unix()

		if claims.ExpiresAt != nil {
			info["is_expired"] = claims.ExpiresAt.Before(currentTime) || claims.ExpiresAt.Equal(currentTime)
			info["time_until_expiry"] = claims.ExpiresAt.Time.Sub(currentTime).String()
		}

		return info, nil
	}

	return nil, errors.New("无法解析token claims")
}

// ValidateTokenStrict 严格验证token（包括签名和时间）
func ValidateTokenStrict(tokenString string) error {
	// 首先检查过期时间（快速检查）
	expired, err := IsTokenExpired(tokenString)
	if err != nil {
		return fmt.Errorf("检查过期时间失败: %w", err)
	}
	if expired {
		return errors.New("token已过期")
	}

	// 然后进行完整验证
	_, err = ParseToken(tokenString)
	return err
}

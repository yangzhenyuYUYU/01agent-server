package middleware

import (
	"net/http"

	"gin_web/internal/models"
	"gin_web/internal/repository"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			repository.Errorf("Panic recovered: %s", err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err))
		} else {
			repository.Errorf("Unknown panic recovered: %v", recovered)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "内部服务器错误"))
		}
		c.Abort()
	})
}

// BusinessError 业务逻辑错误
type BusinessError struct {
	Code int
	Msg  string
}

func (e *BusinessError) Error() string {
	return e.Msg
}

// NewBusinessError 创建业务错误
func NewBusinessError(code int, msg string) *BusinessError {
	return &BusinessError{
		Code: code,
		Msg:  msg,
	}
}

// HandleError 处理错误并返回统一格式响应
func HandleError(c *gin.Context, err error) {
	if businessErr, ok := err.(*BusinessError); ok {
		// 业务逻辑错误
		repository.Warnf("Business error: %s", businessErr.Msg)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(businessErr.Code, businessErr.Msg))
	} else {
		// 系统错误
		repository.Errorf("System error: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
	}
}

// Success 返回成功响应
func Success(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, models.SuccessResponse(msg, data))
}

// SuccessWithoutData 返回无数据的成功响应
func SuccessWithoutData(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, models.SuccessResponse(msg, nil))
}

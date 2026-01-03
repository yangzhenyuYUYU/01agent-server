package router

import (
	"fmt"
	"net/http"

	"01agent_server/internal/models"
	"01agent_server/internal/service"

	"github.com/gin-gonic/gin"
)

type ConcurrentHandler struct {
	concurrentService *service.ConcurrentService
}

// NewConcurrentHandler 创建并发测试处理器
func NewConcurrentHandler() *ConcurrentHandler {
	return &ConcurrentHandler{
		concurrentService: service.NewConcurrentService(),
	}
}

// SerialExecution 串行执行任务
func (h *ConcurrentHandler) SerialExecution(c *gin.Context) {
	// 获取参数
	taskCountStr := c.DefaultQuery("tasks", "5")
	taskCount := 5
	if count, err := fmt.Sscanf(taskCountStr, "%d", &taskCount); err != nil || count != 1 {
		taskCount = 5
	}

	response := h.concurrentService.SerialExecution(taskCount)
	c.JSON(http.StatusOK, models.NewResponse(200, "Serial execution completed", response))
}

// ConcurrentExecution 并发执行任务
func (h *ConcurrentHandler) ConcurrentExecution(c *gin.Context) {
	// 获取参数
	taskCountStr := c.DefaultQuery("tasks", "5")
	taskCount := 5
	if count, err := fmt.Sscanf(taskCountStr, "%d", &taskCount); err != nil || count != 1 {
		taskCount = 5
	}

	response := h.concurrentService.ConcurrentExecution(taskCount)
	c.JSON(http.StatusOK, models.NewResponse(200, "Concurrent execution completed", response))
}

// CompareExecution 对比串行和并发执行
func (h *ConcurrentHandler) CompareExecution(c *gin.Context) {
	// 获取参数
	taskCountStr := c.DefaultQuery("tasks", "5")
	taskCount := 5
	if count, err := fmt.Sscanf(taskCountStr, "%d", &taskCount); err != nil || count != 1 {
		taskCount = 5
	}

	comparison := h.concurrentService.CompareExecution(taskCount)
	c.JSON(http.StatusOK, models.NewResponse(200, "Execution comparison completed", comparison))
}

// StressTest 压力测试
func (h *ConcurrentHandler) StressTest(c *gin.Context) {
	// 获取参数
	goroutineCountStr := c.DefaultQuery("goroutines", "100")
	goroutineCount := 100
	if count, err := fmt.Sscanf(goroutineCountStr, "%d", &goroutineCount); err != nil || count != 1 {
		goroutineCount = 100
	}

	result := h.concurrentService.StressTest(goroutineCount)
	c.JSON(http.StatusOK, models.NewResponse(200, "Stress test completed", result))
}

// SetupConcurrentRoutes 设置并发测试路由
func SetupConcurrentRoutes(r *gin.Engine, concurrentHandler *ConcurrentHandler) {
	// 公开路由
	public := r.Group("/api/v1")
	{
		// 并发测试接口（公开）
		public.GET("/concurrent/serial", concurrentHandler.SerialExecution)
		public.GET("/concurrent/parallel", concurrentHandler.ConcurrentExecution)
		public.GET("/concurrent/compare", concurrentHandler.CompareExecution)
		public.GET("/concurrent/stress", concurrentHandler.StressTest)
	}
}

package main

import (
	"fmt"
	"log"

	"gin_web/internal/config"
	"gin_web/internal/repository"
	"gin_web/internal/router"

	"github.com/gin-gonic/gin"
)

/*
*
启动命令：
go run main.go
*/
func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := repository.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 初始化数据库
	if err := repository.InitDatabase(); err != nil {
		repository.Errorf("Failed to initialize database: %v", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 数据库迁移
	if err := repository.AutoMigrate(); err != nil {
		repository.Errorf("Failed to migrate database: %v", err)
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化Redis
	if err := repository.InitRedis(); err != nil {
		repository.Errorf("Failed to initialize Redis: %v", err)
		// Redis连接失败不终止程序，可以继续运行
		repository.Warn("Running without Redis")
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 设置路由
	r := router.SetupRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	repository.Infof("Starting server on %s", addr)

	if err := r.Run(addr); err != nil {
		repository.Errorf("Failed to start server: %v", err)
		log.Fatalf("Failed to start server: %v", err)
	}
}

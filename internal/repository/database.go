package repository

import (
	"fmt"
	"log"
	"time"

	"gin_web/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	cfg := config.AppConfig.Database

	var dsn string
	var dialector gorm.Dialector

	switch cfg.Type {
	case "mysql":
		mysqlCfg := cfg.MySQL
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			mysqlCfg.Username, mysqlCfg.Password, mysqlCfg.Host, mysqlCfg.Port,
			mysqlCfg.DBName, mysqlCfg.Charset, mysqlCfg.ParseTime, mysqlCfg.Loc)
		dialector = mysql.Open(dsn)
	case "postgres":
		pgCfg := cfg.Postgres
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			pgCfg.Host, pgCfg.Username, pgCfg.Password, pgCfg.DBName, pgCfg.Port, pgCfg.SSLMode, pgCfg.TimeZone)
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// 配置GORM日志
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 根据数据库类型设置连接池参数
	switch cfg.Type {
	case "mysql":
		mysqlCfg := cfg.MySQL
		sqlDB.SetMaxIdleConns(mysqlCfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(mysqlCfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(mysqlCfg.ConnMaxLifetime)
	case "postgres":
		pgCfg := cfg.Postgres
		sqlDB.SetMaxIdleConns(pgCfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(pgCfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(pgCfg.ConnMaxLifetime)
	}

	DB = db
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

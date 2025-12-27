package repository

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger 初始化日志
func InitLogger() error {
	Logger = logrus.New()

	// 设置日志格式
	Logger.SetFormatter(&logrus.JSONFormatter{})

	// 设置日志级别
	Logger.SetLevel(logrus.InfoLevel)

	// 设置输出
	Logger.SetOutput(os.Stdout)

	return nil
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return Logger
}

// 便捷方法
func Info(args ...interface{}) {
	if Logger != nil {
		Logger.Info(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Infof(format, args...)
	}
}

func Error(args ...interface{}) {
	if Logger != nil {
		Logger.Error(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(format, args...)
	}
}

func Warn(args ...interface{}) {
	if Logger != nil {
		Logger.Warn(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warnf(format, args...)
	}
}

func Debug(args ...interface{}) {
	if Logger != nil {
		Logger.Debug(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Debugf(format, args...)
	}
}

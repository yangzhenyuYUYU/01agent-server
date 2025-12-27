package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        int            `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Response 通用响应结构 - 符合FastAPI规范
type Response struct {
	Code int         `json:"code"`           // 0表示成功，其他值表示错误
	Msg  string      `json:"msg"`            // 提示消息
	Data interface{} `json:"data,omitempty"` // 响应数据
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// PageRequest 分页请求结构
type PageRequest struct {
	Page int `json:"page" form:"page" binding:"min=1"`
	Size int `json:"size" form:"size" binding:"min=1,max=100"`
}

// SuccessResponse 创建成功响应 (code = 0)
func SuccessResponse(msg string, data interface{}) *Response {
	return &Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	}
}

// ErrorResponse 创建错误响应 (code != 0)
func ErrorResponse(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// NewResponse 创建响应 (保持向后兼容)
func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code: code,
		Msg:  message,
		Data: data,
	}
}

// NewPageResponse 创建分页响应
func NewPageResponse(code int, message string, data interface{}, total int64, page, size int) *PageResponse {
	return &PageResponse{
		Code:  code,
		Msg:   message,
		Data:  data,
		Total: total,
		Page:  page,
		Size:  size,
	}
}

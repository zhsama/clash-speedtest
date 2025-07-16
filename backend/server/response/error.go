package response

import (
	"context"
	"fmt"
	"net/http"
)

// ErrorType 错误类型枚举
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeNetwork    ErrorType = "network"
)

// AppError 应用程序错误结构
type AppError struct {
	Type    ErrorType
	Message string
	Code    int
	Err     error
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap 支持错误包装
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError 创建验证错误
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    http.StatusBadRequest,
		Err:     err,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Code:    http.StatusNotFound,
		Err:     nil,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    http.StatusInternalServerError,
		Err:     err,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Code:    http.StatusRequestTimeout,
		Err:     err,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Code:    http.StatusBadGateway,
		Err:     err,
	}
}

// HandleError 统一错误处理
func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *AppError:
		SendError(ctx, w, e.Code, e.Message)
	default:
		SendError(ctx, w, http.StatusInternalServerError, "Internal server error")
	}
}
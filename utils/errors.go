package utils

import (
	"fmt"
)

type ServiceErr struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var (
	// 参数错误
	ErrorParamsInvalid = &ServiceErr{
		Code: 1007,
		Msg:  "参数错误",
	}
	// 系统错误
	ErrorSystemError = &ServiceErr{
		Code: 500,
		Msg:  "服务器错误，请稍后重试",
	}
	// ChatGPTError
	ErrorChatGPTError = &ServiceErr{
		Code: 500,
		Msg:  "ChatGPT server error",
	}
)

func (e *ServiceErr) NewWithMsg(msg string) error {
	err := *e
	err.Msg = msg
	return &err
}

func ErrorNew(code int, msg string) error {
	return &ServiceErr{
		Code: code,
		Msg:  msg,
	}
}

func (e ServiceErr) Error() string {
	return fmt.Sprintf("response error(%d): %s", e.Code, e.Msg)
}

func GetErrorCode(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(*ServiceErr); ok {
		return e.Code
	}
	return ErrorSystemError.Code
}

func GetErrorMsg(err error) string {
	if e, ok := err.(*ServiceErr); ok {
		if e.Msg == "" {
			return ErrorSystemError.Msg
		}
		return e.Msg
	}
	return err.Error()
}

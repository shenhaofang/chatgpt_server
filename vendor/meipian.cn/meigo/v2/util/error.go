package util

import "meipian.cn/meigo/v2/config"

const (
	// 兼容API已有code
	ErrOK        = 0
	ErrNotFound  = 404  // Not Found
	ErrOKPartial = 1001 // 部分成功
	ErrSystem    = 1004 // 服务器错误
	ErrParam     = 1007 // 参数错误

	ErrMySQL  = 1010 // DB异常
	ErrRedis  = 1020 // Redis异常
	ErrAPI    = 1030 // API异常
	ErrLock   = 1040 // 加锁失败
	ErrGRPC   = 1050 // GRPC异常
	ErrNSQ    = 1060 // NSQ异常
	ErrCommon = 1100 // 通过 MSG 传递的错误
)

// errMsg 通用msg
var errMsg = map[int]string{
	ErrOK:        "Success",
	ErrNotFound:  "Not found",
	ErrOKPartial: "部分成功",
	ErrSystem:    "服务器错误",
	ErrParam:     "参数错误",
	ErrMySQL:     "DB异常",
	ErrRedis:     "Redis异常",
	ErrAPI:       "API异常",
	ErrLock:      "加锁失败",
	ErrGRPC:      "GRPC异常",
	ErrNSQ:       "NSQ异常",
}

// IsErrorOk 获取success code
func IsErrorOk(code int) (ok bool) {
	if code == 0 || code == GetErrOk() {
		ok = true
	}
	return
}

// GetErrOk 获取success code
func GetErrOk() (code int) {
	return config.GetIntDft("err_ok", ErrOK)
}

// ErrMsg 通过code取message
func GetErrMsg(code int) (msg string) {
	msg = config.CodeMsg(code)

	if msg != "" {
		return
	}

	msg = errMsg[code]

	return
}

//type Error struct {
//	Code    int
//	Message string
//}
//
//func (p *Error) Error() string {
//	return p.Message
//}
//
//func SetErr(str string) *Error {
//	return &Error{Code: Err_System, Message: str}
//}
//
//func SetErrFromCode(code int) *Error {
//	return &Error{Code: code}
//}
//
//func SetErrFromErr(code int, err error) *Error {
//	return &Error{Code: code, Message: err.Error()}
//}
//
//func SetErrMySQL(err error) *Error {
//	return &Error{Code:Err_MySQL, Message:err.Error()}
//}

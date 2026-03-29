package errors

import "fmt"

// BizError 业务错误
type BizError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BizError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

func New(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}

func Newf(code int, format string, args ...interface{}) *BizError {
	return &BizError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// 错误码定义
var (
	// 通用
	ErrSuccess         = New(0, "ok")
	ErrServer          = New(10000, "服务器内部错误")
	ErrParam           = New(10001, "请求参数错误")
	ErrUnauthorized    = New(10002, "未授权，请先登录")
	ErrForbidden       = New(10003, "无权限操作")
	ErrNotFound        = New(10004, "资源不存在")
	ErrDuplicate       = New(10005, "数据已存在")

	// 用户
	ErrUserNotFound    = New(20001, "用户不存在")
	ErrWrongPassword   = New(20002, "密码错误")
	ErrTokenExpired    = New(20003, "Token 已过期")
	ErrTokenInvalid    = New(20004, "Token 无效")

	// 商品
	ErrProductNotFound = New(30001, "商品不存在")
	ErrProductDisabled = New(30002, "商品已停用")

	// 库存
	ErrStockInsufficient = New(40001, "库存不足")
	ErrStockNotFound     = New(40002, "库存记录不存在")

	// 采购
	ErrPurchaseNotFound   = New(50001, "采购单不存在")
	ErrPurchaseStatusFlow = New(50002, "采购单状态流转错误")

	// 销售
	ErrSaleNotFound   = New(60001, "销售单不存在")
	ErrSaleStatusFlow = New(60002, "销售单状态流转错误")
)

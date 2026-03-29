package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	bizErr "ims-api/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PageData 分页数据
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func traceID(c *gin.Context) string {
	if id, exists := c.Get("trace_id"); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
		TraceID: traceID(c),
	})
}

// OKPage 分页成功响应
func OKPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		TraceID: traceID(c),
	})
}

// Fail 业务错误响应
func Fail(c *gin.Context, err *bizErr.BizError) {
	c.JSON(http.StatusOK, Response{
		Code:    err.Code,
		Message: err.Message,
		TraceID: traceID(c),
	})
}

// FailMsg 快速失败
func FailMsg(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		TraceID: traceID(c),
	})
}

// ServerError 服务器错误
func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    10000,
		Message: msg,
		TraceID: traceID(c),
	})
}

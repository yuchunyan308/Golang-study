package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateTraceID 生成请求追踪ID
func GenerateTraceID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateOrderNo 生成单据编号
// 格式：前缀 + 年月日 + 6位序号（这里用随机数模拟，生产应用数据库序列）
func GenerateOrderNo(prefix string) string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102"), hex.EncodeToString(b))
}

// PtrString 返回字符串指针
func PtrString(s string) *string {
	return &s
}

// PtrFloat64 返回float64指针
func PtrFloat64(f float64) *float64 {
	return &f
}

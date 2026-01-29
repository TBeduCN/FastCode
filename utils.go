package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 生成UUID
func generateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		// 如果生成失败，使用时间戳和随机数生成一个简单的UUID
		return fmt.Sprintf("%d%08x", time.Now().UnixNano(), rand.Int31())
	}
	// 格式化UUID为标准格式
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

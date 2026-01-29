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

// 带时间戳的打印函数
func printWithTime(format string, args ...interface{}) {
	timeStr := time.Now().Format("[15:04:05]")
	fmt.Printf(timeStr+" "+format, args...)
}

// 带时间戳的.Println函数
func printlnWithTime(args ...interface{}) {
	timeStr := time.Now().Format("[15:04:05]")
	fmt.Print(timeStr + " ")
	fmt.Println(args...)
}

// 带时间戳的.Printf函数
func printfWithTime(format string, args ...interface{}) {
	timeStr := time.Now().Format("[15:04:05]")
	fmt.Printf(timeStr+" "+format, args...)
}

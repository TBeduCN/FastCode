package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// 版本号，由编译时注入
var version = "dev"

// 提交哈希，由编译时注入
var commit = "unknown"

func main() {
	fmt.Println()
	fmt.Println("    ______           __  ______          __   ")
	fmt.Println("   / ____/___ ______/ /_/ ____/___  ____/ /__ ")
	fmt.Println("  / /_  / __ `/ ___/ __/ /   / __ \\/ __  / _ \\")
	fmt.Println(" / __/ / /_/ (__  ) /_/ /___/ /_/ / /_/ /  __/")
	fmt.Println("/_/    \\__,_/____/\\__/\\____/\\____/\\__,_/\\___/ ")
	fmt.Println()
	shortCommit := commit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}
	fmt.Printf("     %-15s %s\n", version, shortCommit)
	fmt.Println("----------------------------------------------")
	fmt.Println()

	// 暂停一下，确保输出能被看到
	// time.Sleep(500 * time.Millisecond)

	// 初始化HTTP客户端
	initHTTPClient()

	// 初始化配置
	initConfig()

	// 初始化静态资源
	initStaticFiles()

	// 检查更新
	go autoCheckUpdate()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	var err error

	// 配置静态文件服务
	// 1. 首先尝试使用本地文件系统（如果public目录存在）
	if _, err := os.Stat("./public"); err == nil {
		printlnWithTime("使用本地文件系统提供静态资源")
		router.StaticFS("/", http.Dir("./public"))
	} else {
		// 2. 否则使用嵌入的文件系统
		subFS, err := fs.Sub(embeddedPublic, "public")
		if err != nil {
			printfWithTime("无法创建子文件系统: %v\n", err)
			// 3. 如果嵌入的文件系统也失败，使用默认处理
		} else {
			printlnWithTime("使用嵌入的文件系统提供静态资源")
			router.StaticFS("/", http.FS(subFS))
		}
	}

	// 处理所有未匹配的路由（代理请求）
	router.NoRoute(handler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	printfWithTime("GitHub代理加速服务启动成功，监听地址: %s\n", addr)
	err = router.Run(addr)
	if err != nil {
		printfWithTime("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

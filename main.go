package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	// 初始化API路由
	initAPIRoutes(router)

	// 配置静态文件服务
	// 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		printfWithTime("获取可执行文件路径失败: %v\n", err)
		// 失败时使用当前目录作为备选
		execPath = "."
	}
	execDir := filepath.Dir(execPath)
	publicDir := filepath.Join(execDir, "public")

	// 1. 首先尝试使用本地文件系统（如果public目录存在）
	if _, err := os.Stat(publicDir); err == nil {
		printlnWithTime("使用本地文件系统提供静态资源")
		// 使用中间件处理静态文件，避免与API路由冲突
		router.Use(func(c *gin.Context) {
			// 如果是API请求，跳过静态文件处理
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.Next()
				return
			}
			// 尝试从本地文件系统提供静态文件
			filePath := filepath.Join(publicDir, c.Request.URL.Path)
			if _, err := os.Stat(filePath); err == nil {
				// 文件存在，提供静态文件
				c.File(filePath)
				c.Abort()
				return
			}
			// 文件不存在，继续处理
			c.Next()
		})
	} else {
		// 2. 否则使用嵌入的文件系统
		subFS, err := fs.Sub(embeddedPublic, "public")
		if err != nil {
			printfWithTime("无法创建子文件系统: %v\n", err)
			// 3. 如果嵌入的文件系统也失败，使用默认处理
		} else {
			printlnWithTime("使用嵌入的文件系统提供静态资源")
			// 使用中间件处理静态文件，避免与API路由冲突
			router.Use(func(c *gin.Context) {
				// 如果是API请求，跳过静态文件处理
				if strings.HasPrefix(c.Request.URL.Path, "/api") {
					c.Next()
					return
				}
				// 尝试从嵌入的文件系统提供静态文件
				file, err := subFS.Open(c.Request.URL.Path)
				if err == nil {
					// 文件存在，提供静态文件
					defer file.Close()
					c.FileFromFS(c.Request.URL.Path, http.FS(subFS))
					c.Abort()
					return
				}
				// 文件不存在，继续处理
				c.Next()
			})
		}
	}

	// 处理所有未匹配的路由（代理请求）
	router.NoRoute(handler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	printfWithTime("服务器启动成功，监听地址: %s\n", addr)
	err = router.Run(addr)
	if err != nil {
		printfWithTime("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

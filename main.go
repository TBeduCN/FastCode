package main

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultSizeLimit int64 = 1024 * 1024 * 1024 * 10 // 允许的文件大小，默认10GB
	defaultHost            = "0.0.0.0"               // 默认监听地址
	defaultPort            = 8080                    // 默认监听端口
)

// 版本号，由编译时注入
var version = "dev"

// 提交哈希，由编译时注入
var commit = "unknown"

// 静态资源嵌入
//
//go:embed public/*
var embeddedPublic embed.FS

var (
	// URL匹配正则表达式

	exps = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*$`),
		regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+$`),
		regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+$`),
		regexp.MustCompile(`^(?:https?://)?api\.github\.com/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/api/.*$`),
	}

	httpClient *http.Client
	config     *Config
	configLock sync.RWMutex
)

// 配置结构体
type Config struct {
	Host           string   `json:"host"`
	Port           int64    `json:"port"`
	SizeLimit      int64    `json:"sizeLimit"`
	WhiteList      []string `json:"whiteList"`
	BlackList      []string `json:"blackList"`
	AllowProxyAll  bool     `json:"allowProxyAll"` // 是否允许代理非github的其他地址
	OtherWhiteList []string `json:"otherWhiteList"`
	OtherBlackList []string `json:"otherBlackList"`
	UUID           string   `json:"uuid"` // 唯一标识符，用于数据统计
}

// 默认配置
var defaultConfig = Config{
	Host:           defaultHost,
	Port:           defaultPort,
	SizeLimit:      defaultSizeLimit,
	WhiteList:      []string{},
	BlackList:      []string{},
	AllowProxyAll:  false,
	OtherWhiteList: []string{},
	OtherBlackList: []string{},
	UUID:           "",
}

func main() {
	fmt.Println()
	fmt.Println("    ______           __  ______          __   ")
	fmt.Println("   / ____/___ ______/ /_/ ____/___  ____/ /__ ")
	fmt.Println("  / /_  / __ / ___/ __/ /   / __ / __  / _ ")
	fmt.Println(" / __/ / /_/ __  ) /_/ /___/ /_/ / /_/ /  __/")
	fmt.Println("/_/    __,_/____/__/____/____/__,_/___/")
	fmt.Println()
	shortCommit := commit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}
	fmt.Printf("   %-15s %s\n", version, shortCommit)
	fmt.Println("========================================")
	fmt.Println()

	// 暂停一下，确保输出能被看到
	time.Sleep(500 * time.Millisecond)

	// 初始化HTTP客户端
	initHTTPClient()

	// 初始化配置
	initConfig()

	// 初始化静态资源
	initStaticFiles()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	var err error

	// 配置静态文件服务
	// 1. 首先尝试使用本地文件系统（如果public目录存在）
	if _, err := os.Stat("./public"); err == nil {
		fmt.Println("使用本地文件系统提供静态资源")
		router.StaticFS("/", http.Dir("./public"))
	} else {
		// 2. 否则使用嵌入的文件系统
		subFS, err := fs.Sub(embeddedPublic, "public")
		if err != nil {
			fmt.Printf("无法创建子文件系统: %v\n", err)
			// 3. 如果嵌入的文件系统也失败，使用默认处理
		} else {
			fmt.Println("使用嵌入的文件系统提供静态资源")
			router.StaticFS("/", http.FS(subFS))
		}
	}

	// 处理所有未匹配的路由（代理请求）
	router.NoRoute(handler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	fmt.Printf("GitHub代理加速服务启动成功，监听地址: %s\n", addr)
	err = router.Run(addr)
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

// 初始化HTTP客户端
func initHTTPClient() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          1000,
			MaxIdleConnsPerHost:   1000,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// 初始化配置
func initConfig() {
	configPath := "./config.json"

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 生成默认配置文件
		fmt.Println("配置文件不存在，生成默认配置...")
		err := generateDefaultConfig(configPath)
		if err != nil {
			fmt.Printf("生成配置文件失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 加载配置
	loadConfig(configPath)

	// 启动配置自动刷新
	go autoRefreshConfig(configPath)
}

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

// 生成默认配置文件
func generateDefaultConfig(path string) error {
	// 为新配置生成UUID
	config := defaultConfig
	config.UUID = generateUUID()

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, configData, 0644)
}

// 加载配置
func loadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v，使用默认配置\n", err)
		configLock.Lock()
		config = &defaultConfig
		configLock.Unlock()
		return
	}
	defer file.Close()

	var newConfig Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&newConfig); err != nil {
		fmt.Printf("解析配置文件失败: %v，使用默认配置\n", err)
		configLock.Lock()
		config = &defaultConfig
		configLock.Unlock()
		return
	}

	// 使用默认值填充未配置项
	if newConfig.Host == "" {
		newConfig.Host = defaultHost
	}
	if newConfig.Port == 0 {
		newConfig.Port = defaultPort
	}
	if newConfig.SizeLimit <= 0 {
		newConfig.SizeLimit = defaultSizeLimit
	}
	// 如果配置文件中没有UUID，生成一个新的
	if newConfig.UUID == "" {
		newConfig.UUID = generateUUID()
		// 将带有UUID的配置写回文件
		configData, err := json.MarshalIndent(newConfig, "", "  ")
		if err != nil {
			fmt.Printf("更新配置文件失败: %v\n", err)
		} else {
			err = os.WriteFile(path, configData, 0644)
			if err != nil {
				fmt.Printf("写入配置文件失败: %v\n", err)
			} else {
				fmt.Println("配置文件已更新UUID")
			}
		}
	}

	configLock.Lock()
	config = &newConfig
	configLock.Unlock()

	fmt.Println("配置文件加载成功")
}

// 自动刷新配置
func autoRefreshConfig(path string) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		loadConfig(path)
	}
}

// 初始化静态资源
func initStaticFiles() {
	// 检查public目录是否存在
	if _, err := os.Stat("./public"); os.IsNotExist(err) {
		fmt.Println("public目录不存在，正在准备静态资源...")

		// 只有正式版本（v*.*.*）才尝试从GitHub下载静态资源
		if strings.HasPrefix(version, "v") {
			fmt.Println("检测到正式版本，尝试从GitHub下载静态资源...")
			err := downloadStaticFiles()
			if err != nil {
				fmt.Printf("下载静态资源失败: %v\n", err)
				fmt.Println("使用嵌入的静态资源...")
				// 从嵌入的文件系统复制静态文件到本地
				copyEmbeddedFiles(embeddedPublic, "public", "./public")
			}
		} else {
			// 开发版本直接使用嵌入的静态资源
			fmt.Println("检测到开发版本，直接使用嵌入的静态资源...")
			copyEmbeddedFiles(embeddedPublic, "public", "./public")
		}
	}
}

// 下载静态资源
func downloadStaticFiles() error {
	// 构建下载URL
	downloadURL := fmt.Sprintf("https://github.com/TBeduCN/FastCode/releases/download/%s/dist.zip", version)
	fmt.Printf("正在从 %s 下载静态资源...\n", downloadURL)

	// 发送HTTP请求
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp(".", "dist-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		os.Remove(tmpPath)
	}()

	// 写入文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}
	tmpFile.Close()

	// 解压文件
	fmt.Println("正在解压静态资源...")
	err = unzip(tmpPath, ".")
	if err != nil {
		return err
	}

	fmt.Println("静态资源下载和解压成功")
	return nil
}

// 解压zip文件
func unzip(src, dest string) error {
	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		path := filepath.Join(dest, file.Name)

		// 创建目录
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// 创建文件
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		writer, err := os.Create(path)
		if err != nil {
			return err
		}

		// 复制内容
		fileReader, err := file.Open()
		if err != nil {
			writer.Close()
			return err
		}

		_, err = io.Copy(writer, fileReader)
		fileReader.Close()
		writer.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// 复制嵌入的文件到本地
func copyEmbeddedFiles(efs embed.FS, srcDir, dstDir string) {
	entries, err := efs.ReadDir(srcDir)
	if err != nil {
		fmt.Printf("读取嵌入文件失败: %v\n", err)
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// 创建目录
			err := os.MkdirAll(dstPath, 0755)
			if err != nil {
				fmt.Printf("创建目录失败: %v\n", err)
				continue
			}
			// 递归复制子目录
			copyEmbeddedFiles(efs, srcPath, dstPath)
		} else {
			// 复制文件
			srcFile, err := efs.Open(srcPath)
			if err != nil {
				fmt.Printf("打开嵌入文件失败: %v\n", err)
				continue
			}

			dstFile, err := os.Create(dstPath)
			if err != nil {
				fmt.Printf("创建本地文件失败: %v\n", err)
				srcFile.Close()
				continue
			}

			_, err = io.Copy(dstFile, srcFile)
			srcFile.Close()
			dstFile.Close()

			if err != nil {
				fmt.Printf("复制文件失败: %v\n", err)
				continue
			}

			fmt.Printf("复制静态文件: %s -> %s\n", srcPath, dstPath)
		}
	}
}

// 主处理函数
func handler(c *gin.Context) {
	// 获取原始请求路径
	rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
	for strings.HasPrefix(rawPath, "/") {
		rawPath = strings.TrimPrefix(rawPath, "/")
	}

	// 检查是否为静态文件请求（如果文件存在于public目录，让静态文件服务处理）
	if rawPath != "" {
		// 构建本地文件路径
		localPath := filepath.Join("./public", rawPath)
		// 检查文件是否存在
		if _, err := os.Stat(localPath); err == nil {
			// 文件存在，让Gin的静态文件服务处理
			c.Next()
			return
		}
	}

	// 构建完整URL
	var targetURL string
	if strings.HasPrefix(rawPath, "http://") || strings.HasPrefix(rawPath, "https://") {
		targetURL = rawPath
	} else {
		targetURL = "https://" + rawPath
	}

	// 检查URL是否符合规则
	matches := checkURL(targetURL)
	if matches == nil {
		// 检查是否允许代理所有地址
		configLock.RLock()
		allowAll := config.AllowProxyAll
		otherWhiteList := config.OtherWhiteList
		otherBlackList := config.OtherBlackList
		configLock.RUnlock()

		if !allowAll {
			c.String(http.StatusForbidden, "无效的URL，不允许代理该地址")
			return
		}

		// 检查其他地址的白名单和黑名单
		if len(otherBlackList) > 0 && checkOtherList(targetURL, otherBlackList) {
			c.String(http.StatusForbidden, "该地址已被列入黑名单")
			return
		}

		if len(otherWhiteList) > 0 && !checkOtherList(targetURL, otherWhiteList) {
			c.String(http.StatusForbidden, "该地址未被列入白名单")
			return
		}
	} else {
		// 检查GitHub地址的白名单和黑名单
		configLock.RLock()
		whiteList := config.WhiteList
		blackList := config.BlackList
		configLock.RUnlock()

		if len(blackList) > 0 && checkList(matches, blackList) {
			c.String(http.StatusForbidden, "该GitHub地址已被列入黑名单")
			return
		}

		if len(whiteList) > 0 && !checkList(matches, whiteList) {
			c.String(http.StatusForbidden, "该GitHub地址未被列入白名单")
			return
		}
	}

	// 处理blob URL转换为raw URL
	if exps[1].MatchString(targetURL) {
		targetURL = strings.Replace(targetURL, "/blob/", "/raw/", 1)
	}

	// 调用代理函数
	proxy(c, targetURL)
}

// 代理函数
func proxy(c *gin.Context, u string) {
	// 创建请求
	req, err := http.NewRequest(c.Request.Method, u, c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("创建请求失败: %v", err))
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	// 删除Host头，让HTTP客户端自动添加
	req.Header.Del("Host")

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("请求GitHub失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 检查文件大小
	configLock.RLock()
	sizeLimit := config.SizeLimit
	configLock.RUnlock()

	if contentLength, ok := resp.Header["Content-Length"]; ok {
		if size, err := strconv.ParseInt(contentLength[0], 10, 64); err == nil && size > sizeLimit {
			c.String(http.StatusRequestEntityTooLarge, fmt.Sprintf("文件过大，超过限制大小: %d GB", sizeLimit/(1024*1024*1024)))
			return
		}
	}

	// 删除不必要的响应头
	resp.Header.Del("Content-Security-Policy")
	resp.Header.Del("Referrer-Policy")
	resp.Header.Del("Strict-Transport-Security")

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 处理重定向
	if location := resp.Header.Get("Location"); location != "" {
		if checkURL(location) != nil {
			// 如果是GitHub地址，重定向到代理地址
			c.Header("Location", "/"+location)
		} else {
			// 否则直接重定向
			c.Header("Location", location)
		}
	}

	// 设置响应状态码
	c.Status(resp.StatusCode)

	// 流式返回响应体
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		fmt.Printf("响应数据复制失败: %v\n", err)
	}
}

// 检查URL是否符合GitHub相关规则
func checkURL(u string) []string {
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			return matches[1:]
		}
	}
	return nil
}

// 检查白名单/黑名单
func checkList(matches, list []string) bool {
	for _, item := range list {
		if strings.HasPrefix(matches[0], item) || strings.HasPrefix(matches[1], item) {
			return true
		}
		// 支持通配符匹配，例如 "*" 匹配所有
		if item == "*" {
			return true
		}
		// 支持 "user/repo" 格式匹配
		parts := strings.Split(item, "/")
		if len(parts) == 2 {
			userMatch := parts[0] == "*" || parts[0] == matches[0]
			repoMatch := parts[1] == "*" || parts[1] == matches[1]
			if userMatch && repoMatch {
				return true
			}
		}
	}
	return false
}

// 检查其他地址的白名单/黑名单
func checkOtherList(url string, list []string) bool {
	for _, item := range list {
		if strings.Contains(url, item) {
			return true
		}
	}
	return false
}

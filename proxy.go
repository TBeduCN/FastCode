package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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
)

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

// 主处理函数
func handler(c *gin.Context) {
	// 获取原始请求路径
	rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
	for strings.HasPrefix(rawPath, "/") {
		rawPath = strings.TrimPrefix(rawPath, "/")
	}

	// 检查是否为静态文件请求（如果文件存在于public目录，让静态文件服务处理）
	if rawPath != "" {
		// 获取可执行文件路径
		execPath, err := os.Executable()
		if err != nil {
			printfWithTime("获取可执行文件路径失败: %v\n", err)
			// 失败时使用当前目录作为备选
			execPath = "."
		}
		execDir := filepath.Dir(execPath)
		publicDir := filepath.Join(execDir, "public")

		// 构建本地文件路径
		localPath := filepath.Join(publicDir, rawPath)
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
		printfWithTime("响应数据复制失败: %v\n", err)
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

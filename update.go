package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// UpdateInfo 表示更新信息
type UpdateInfo struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// 检查更新
func checkUpdate() (*UpdateInfo, error) {
	// GitHub API URL
	apiURL := "https://api.github.com/repos/TBeduCN/FastCode/releases/latest"

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("获取最新版本信息失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var updateInfo UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&updateInfo); err != nil {
		return nil, fmt.Errorf("解析更新信息失败: %v", err)
	}

	return &updateInfo, nil
}

// 检查是否需要更新
func needUpdate(currentVersion string, latestVersion string) bool {
	// 移除版本号前缀的 "v"
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	// 简单的版本比较
	return latest > current
}

// 自动检查更新
func autoCheckUpdate() {
	// 初始检查
	checkForUpdates()

	// 每6小时检查一次
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		checkForUpdates()
	}
}

// 检查更新的具体实现
func checkForUpdates() {
	// 获取最新版本信息
	updateInfo, err := checkUpdate()
	if err != nil {
		printfWithTime("检查更新失败: %v\n", err)
		return
	}

	// 检查是否需要更新
	if needUpdate(version, updateInfo.TagName) {
		printfWithTime("发现新版本: %s\n", updateInfo.TagName)
		printfWithTime("当前版本: %s\n", version)
		printlnWithTime("更新内容:")
		printlnWithTime(updateInfo.Body)
		printlnWithTime("请访问GitHub页面下载最新版本:")
		printlnWithTime("https://github.com/TBeduCN/FastCode/releases")
	} else {
		printfWithTime("当前已是最新版本: %s\n", version)
	}
}

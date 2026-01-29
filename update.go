package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// Asset 表示发布资产
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release 表示GitHub发布
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
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

// 获取适合当前平台的下载链接
func getDownloadURL(latestVersion string) (string, error) {
	// GitHub API URL
	apiURL := "https://api.github.com/repos/TBeduCN/FastCode/releases/tags/" + latestVersion

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("获取发布信息失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API 请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("解析发布信息失败: %v", err)
	}

	// 确定当前平台
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 构建文件名模式
	expectedExt := ".tar.gz"
	if osName == "windows" {
		expectedExt = ".zip"
	}

	osNameMap := map[string]string{
		"windows": "windows",
		"linux":   "linux",
		"darwin":  "darwin",
	}

	platform := osNameMap[osName]
	if platform == "" {
		return "", fmt.Errorf("不支持的操作系统: %s", osName)
	}

	// 查找匹配的资产
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, platform) && strings.Contains(asset.Name, arch) && strings.HasSuffix(asset.Name, expectedExt) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("未找到适合当前平台的更新包")
}

// 下载并更新主程序
func downloadAndUpdate(downloadURL string) error {
	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载更新包失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载更新包失败，状态码: %d", resp.StatusCode)
	}

	// 获取当前可执行文件路径
	// execPath, err := os.Executable()
	// if err != nil {
	// 	return fmt.Errorf("获取可执行文件路径失败: %v", err)
	// }

	// 创建临时文件
	tempDir := os.TempDir()
	fileName := filepath.Base(downloadURL)
	tempPath := filepath.Join(tempDir, fileName)

	// 写入临时文件
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tempFile.Close()

	// 复制内容
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 关闭文件
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %v", err)
	}

	// 打印更新信息
	fmt.Println("更新包下载完成，准备更新...")
	fmt.Println("注意: 更新完成后需要手动重启程序")

	// 根据文件类型处理
	if strings.HasSuffix(fileName, ".zip") {
		// Windows 平台，解压 zip 文件
		return extractZipAndUpdate(tempPath)
	} else if strings.HasSuffix(fileName, ".tar.gz") {
		// Linux/macOS 平台，解压 tar.gz 文件
		return extractTarGzAndUpdate(tempPath)
	}

	return fmt.Errorf("不支持的文件类型: %s", fileName)
}

// 解压zip文件并更新
func extractZipAndUpdate(zipPath string) error {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 创建临时解压目录
	tempDir := filepath.Join(os.TempDir(), "fastcode_update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时解压目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 解压zip文件
	fmt.Println("正在解压更新包...")
	// 这里需要实现zip解压逻辑
	// 由于Go标准库没有直接的zip解压函数，我们需要使用archive/zip包
	// 但为了简化，这里我们使用命令行工具

	// 对于Windows平台，使用PowerShell命令解压
	if runtime.GOOS == "windows" {
		cmd := fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s'", zipPath, tempDir)
		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("解压zip文件失败: %v", err)
		}
	} else {
		// 对于其他平台，使用unzip命令
		cmd := fmt.Sprintf("unzip -o '%s' -d '%s'", zipPath, tempDir)
		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("解压zip文件失败: %v", err)
		}
	}

	// 查找解压后的可执行文件
	var newExecPath string
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".exe") || !strings.Contains(path, ".")) {
			newExecPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if newExecPath == "" {
		return fmt.Errorf("未在更新包中找到可执行文件")
	}

	// 备份当前可执行文件
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("备份当前可执行文件失败: %v", err)
	}

	// 复制新可执行文件
	if err := copyFile(newExecPath, execPath); err != nil {
		// 恢复备份
		os.Rename(backupPath, execPath)
		return fmt.Errorf("复制新可执行文件失败: %v", err)
	}

	// 设置可执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(execPath, 0755); err != nil {
			return fmt.Errorf("设置可执行权限失败: %v", err)
		}
	}

	fmt.Println("可执行文件更新成功")
	return nil
}

// 解压tar.gz文件并更新
func extractTarGzAndUpdate(tarGzPath string) error {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 创建临时解压目录
	tempDir := filepath.Join(os.TempDir(), "fastcode_update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时解压目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 解压tar.gz文件
	fmt.Println("正在解压更新包...")

	// 对于不同平台，使用不同的解压命令
	if runtime.GOOS == "windows" {
		// Windows平台，使用PowerShell命令解压
		cmd := fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s'", tarGzPath, tempDir)
		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("解压tar.gz文件失败: %v", err)
		}
	} else {
		// 对于Linux/macOS平台，使用tar命令
		cmd := fmt.Sprintf("tar -xzf '%s' -C '%s'", tarGzPath, tempDir)
		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("解压tar.gz文件失败: %v", err)
		}
	}

	// 查找解压后的可执行文件
	var newExecPath string
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".exe") || !strings.Contains(path, ".")) {
			newExecPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if newExecPath == "" {
		return fmt.Errorf("未在更新包中找到可执行文件")
	}

	// 备份当前可执行文件
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("备份当前可执行文件失败: %v", err)
	}

	// 复制新可执行文件
	if err := copyFile(newExecPath, execPath); err != nil {
		// 恢复备份
		os.Rename(backupPath, execPath)
		return fmt.Errorf("复制新可执行文件失败: %v", err)
	}

	// 设置可执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(execPath, 0755); err != nil {
			return fmt.Errorf("设置可执行权限失败: %v", err)
		}
	}

	fmt.Println("可执行文件更新成功")
	return nil
}

// 运行命令
func runCommand(cmd string) error {
	fmt.Println("执行命令:", cmd)
	// 使用os/exec包执行命令
	var err error
	if runtime.GOOS == "windows" {
		// Windows平台，使用PowerShell
		err = exec.Command("PowerShell", "-Command", cmd).Run()
	} else {
		// Linux/macOS平台，使用bash
		err = exec.Command("bash", "-c", cmd).Run()
	}
	return err
}

// 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// 自动检查更新
func autoCheckUpdate() {
	fmt.Println("正在检查更新...")

	// 获取最新版本信息
	updateInfo, err := checkUpdate()
	if err != nil {
		fmt.Printf("检查更新失败: %v\n", err)
		return
	}

	// 检查是否需要更新
	if needUpdate(version, updateInfo.TagName) {
		fmt.Printf("发现新版本: %s\n", updateInfo.TagName)
		fmt.Printf("当前版本: %s\n", version)
		fmt.Println("更新内容:")
		fmt.Println(updateInfo.Body)

		// 询问是否更新
		fmt.Println("\n是否下载并更新到最新版本? (y/n)")
		var answer string
		fmt.Scanln(&answer)

		if strings.ToLower(answer) == "y" {
			// 获取下载链接
			downloadURL, err := getDownloadURL(updateInfo.TagName)
			if err != nil {
				fmt.Printf("获取下载链接失败: %v\n", err)
				return
			}

			// 下载并更新
			if err := downloadAndUpdate(downloadURL); err != nil {
				fmt.Printf("更新失败: %v\n", err)
				return
			}

			fmt.Println("更新完成，请手动重启程序")
		} else {
			fmt.Println("取消更新")
		}
	} else {
		fmt.Printf("当前已是最新版本: %s\n", version)
	}
}

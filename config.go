package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultSizeLimit int64 = 1024 * 1024 * 1024 * 10 // 允许的文件大小，默认10GB
	defaultHost            = "0.0.0.0"               // 默认监听地址
	defaultPort            = 8080                    // 默认监听端口
)

// 配置结构体
type Config struct {
	Version        string   `json:"version"` // 配置文件版本
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

// 配置文件版本
const configVersion = "1.0.1"

// 默认配置
var defaultConfig = Config{
	Version:        configVersion,
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

var (
	config     *Config
	configLock sync.RWMutex
)

// 初始化配置
func initConfig() {
	// 配置文件路径
	configDir := "./config"
	configPath := filepath.Join(configDir, "fastcode.yml")
	oldConfigPath := "./config.json"

	// 检查并创建config目录
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		printlnWithTime("config目录不存在，创建目录...")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			printfWithTime("创建config目录失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 检查新配置文件是否存在
	if _, err := os.Stat(configPath); err == nil {
		// 新配置文件存在，直接使用
	} else {
		// 检查是否需要迁移配置文件
		if _, err := os.Stat(oldConfigPath); err == nil {
			// 旧配置文件存在，需要迁移
			printlnWithTime("检测到旧配置文件，开始迁移...")
			err := migrateConfig(oldConfigPath, configPath)
			if err != nil {
				printfWithTime("迁移配置文件失败: %v\n", err)
				// 迁移失败，继续使用旧配置文件
				configPath = oldConfigPath
			} else {
				printlnWithTime("配置文件迁移成功")
			}
		} else {
			// 生成默认配置文件
			printlnWithTime("配置文件不存在，生成默认配置...")
			err := generateDefaultConfig(configPath)
			if err != nil {
				printfWithTime("生成配置文件失败: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// 加载配置
	loadConfig(configPath)

	// 启动配置自动刷新
	go autoRefreshConfig(configPath)
}

// 生成默认配置文件
func generateDefaultConfig(path string) error {
	// 为新配置生成UUID
	config := defaultConfig
	config.UUID = generateUUID()

	var configData []byte
	var err error

	// 根据文件扩展名选择格式
	if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
		// 生成带注释的YAML格式
		yamlContent := "# FastCode 配置文件\n"
		yamlContent += "# 配置文件版本，请勿修改\n"
		yamlContent += fmt.Sprintf("version: %s\n\n", config.Version)
		yamlContent += "# 监听地址，默认: 0.0.0.0\n"
		yamlContent += fmt.Sprintf("host: %s\n\n", config.Host)
		yamlContent += "# 监听端口，默认: 8080\n"
		yamlContent += fmt.Sprintf("port: %d\n\n", config.Port)
		yamlContent += "# 文件大小限制，默认: 10GB\n"
		yamlContent += fmt.Sprintf("sizeLimit: %d\n\n", config.SizeLimit)
		yamlContent += "# GitHub地址白名单，支持通配符\n"
		yamlContent += "whiteList: []\n"
		if len(config.WhiteList) > 0 {
			for _, item := range config.WhiteList {
				yamlContent += fmt.Sprintf("  - %s\n", item)
			}
		}
		yamlContent += "\n"
		yamlContent += "# GitHub地址黑名单，支持通配符\n"
		yamlContent += "blackList: []\n"
		if len(config.BlackList) > 0 {
			for _, item := range config.BlackList {
				yamlContent += fmt.Sprintf("  - %s\n", item)
			}
		}
		yamlContent += "\n"
		yamlContent += "# 是否允许代理非GitHub的其他地址\n"
		yamlContent += fmt.Sprintf("allowProxyAll: %t\n\n", config.AllowProxyAll)
		yamlContent += "# 其他地址白名单\n"
		yamlContent += "otherWhiteList: []\n"
		if len(config.OtherWhiteList) > 0 {
			for _, item := range config.OtherWhiteList {
				yamlContent += fmt.Sprintf("  - %s\n", item)
			}
		}
		yamlContent += "\n"
		yamlContent += "# 其他地址黑名单\n"
		yamlContent += "otherBlackList: []\n"
		if len(config.OtherBlackList) > 0 {
			for _, item := range config.OtherBlackList {
				yamlContent += fmt.Sprintf("  - %s\n", item)
			}
		}
		yamlContent += "\n"
		yamlContent += "# 唯一标识符，用于数据统计\n"
		yamlContent += fmt.Sprintf("uuid: %s\n", config.UUID)
		configData = []byte(yamlContent)
	} else {
		// 生成JSON格式
		configData, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
	}

	return os.WriteFile(path, configData, 0644)
}

// 加载配置
func loadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		printfWithTime("加载配置文件失败: %v，使用默认配置\n", err)
		configLock.Lock()
		config = &defaultConfig
		configLock.Unlock()
		return
	}
	defer file.Close()

	var newConfig Config
	var versionUpdated bool

	// 根据文件扩展名选择解析器
	if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
		// 使用YAML解析器
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&newConfig); err != nil {
			printfWithTime("解析YAML配置文件失败: %v，使用默认配置\n", err)
			configLock.Lock()
			config = &defaultConfig
			configLock.Unlock()
			return
		}
	} else {
		// 使用JSON解析器
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&newConfig); err != nil {
			printfWithTime("解析JSON配置文件失败: %v，使用默认配置\n", err)
			configLock.Lock()
			config = &defaultConfig
			configLock.Unlock()
			return
		}
	}

	// 检查配置文件版本
	configUpdated := false
	versionUpdated = false
	if newConfig.Version != configVersion {
		printfWithTime("检测到配置文件版本不一致 (%s -> %s)，更新配置文件...\n", newConfig.Version, configVersion)
		newConfig.Version = configVersion
		configUpdated = true
		versionUpdated = true
	}

	// 使用默认值填充未配置项
	if newConfig.Host == "" {
		newConfig.Host = defaultHost
		configUpdated = true
	}
	if newConfig.Port == 0 {
		newConfig.Port = defaultPort
		configUpdated = true
	}
	if newConfig.SizeLimit <= 0 {
		newConfig.SizeLimit = defaultSizeLimit
		configUpdated = true
	}
	if newConfig.WhiteList == nil {
		newConfig.WhiteList = []string{}
		configUpdated = true
	}
	if newConfig.BlackList == nil {
		newConfig.BlackList = []string{}
		configUpdated = true
	}
	if newConfig.OtherWhiteList == nil {
		newConfig.OtherWhiteList = []string{}
		configUpdated = true
	}
	if newConfig.OtherBlackList == nil {
		newConfig.OtherBlackList = []string{}
		configUpdated = true
	}

	// 如果配置文件中没有UUID，生成一个新的
	if newConfig.UUID == "" {
		newConfig.UUID = generateUUID()
		configUpdated = true
	}

	// 如果配置有更新，写回文件
	if configUpdated {
		// 将更新后的配置写回文件
		var configData []byte
		var err error

		// 根据文件扩展名选择格式
		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			// 生成带注释的YAML格式
			yamlContent := "# FastCode 配置文件\n"
			yamlContent += "# 配置文件版本，请勿修改\n"
			yamlContent += fmt.Sprintf("version: %s\n\n", newConfig.Version)
			yamlContent += "# 监听地址，默认: 0.0.0.0\n"
			yamlContent += fmt.Sprintf("host: %s\n\n", newConfig.Host)
			yamlContent += "# 监听端口，默认: 8080\n"
			yamlContent += fmt.Sprintf("port: %d\n\n", newConfig.Port)
			yamlContent += "# 文件大小限制，默认: 10GB\n"
			yamlContent += fmt.Sprintf("sizeLimit: %d\n\n", newConfig.SizeLimit)
			yamlContent += "# GitHub地址白名单，支持通配符\n"
			yamlContent += "whiteList: []\n"
			if len(newConfig.WhiteList) > 0 {
				for _, item := range newConfig.WhiteList {
					yamlContent += fmt.Sprintf("  - %s\n", item)
				}
			}
			yamlContent += "\n"
			yamlContent += "# GitHub地址黑名单，支持通配符\n"
			yamlContent += "blackList: []\n"
			if len(newConfig.BlackList) > 0 {
				for _, item := range newConfig.BlackList {
					yamlContent += fmt.Sprintf("  - %s\n", item)
				}
			}
			yamlContent += "\n"
			yamlContent += "# 是否允许代理非GitHub的其他地址\n"
			yamlContent += fmt.Sprintf("allowProxyAll: %t\n\n", newConfig.AllowProxyAll)
			yamlContent += "# 其他地址白名单\n"
			yamlContent += "otherWhiteList: []\n"
			if len(newConfig.OtherWhiteList) > 0 {
				for _, item := range newConfig.OtherWhiteList {
					yamlContent += fmt.Sprintf("  - %s\n", item)
				}
			}
			yamlContent += "\n"
			yamlContent += "# 其他地址黑名单\n"
			yamlContent += "otherBlackList: []\n"
			if len(newConfig.OtherBlackList) > 0 {
				for _, item := range newConfig.OtherBlackList {
					yamlContent += fmt.Sprintf("  - %s\n", item)
				}
			}
			yamlContent += "\n"
			yamlContent += "# 唯一标识符，用于数据统计\n"
			yamlContent += fmt.Sprintf("uuid: %s\n", newConfig.UUID)
			configData = []byte(yamlContent)
		} else {
			// 生成JSON格式
			configData, err = json.MarshalIndent(newConfig, "", "  ")
			if err != nil {
				printfWithTime("更新配置文件失败: %v\n", err)
				return
			}
		}

		if err != nil {
			printfWithTime("更新配置文件失败: %v\n", err)
		} else {
			err = os.WriteFile(path, configData, 0644)
			if err != nil {
				printfWithTime("写入配置文件失败: %v\n", err)
			} else {
				// 显示更新消息
				if versionUpdated {
					printlnWithTime("配置文件已更新版本")
				} else if newConfig.UUID == "" {
					printlnWithTime("配置文件已更新UUID")
				}
			}
		}
	}

	configLock.Lock()
	config = &newConfig
	configLock.Unlock()

	printlnWithTime("配置文件加载成功")
}

// 自动刷新配置
func autoRefreshConfig(path string) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		loadConfig(path)
	}
}

// 迁移配置文件
func migrateConfig(oldPath, newPath string) error {
	// 读取旧配置文件
	oldFile, err := os.Open(oldPath)
	if err != nil {
		return err
	}
	defer oldFile.Close()

	// 解析旧配置文件
	var oldConfig Config
	decoder := json.NewDecoder(oldFile)
	if err := decoder.Decode(&oldConfig); err != nil {
		return err
	}

	// 生成带注释的YAML格式
	yamlContent := "# FastCode 配置文件 (从旧配置文件迁移)\n"
	yamlContent += "# 配置文件版本，请勿修改\n"
	yamlContent += fmt.Sprintf("version: %s\n\n", configVersion)
	yamlContent += "# 监听地址，默认: 0.0.0.0\n"
	yamlContent += fmt.Sprintf("host: %s\n\n", oldConfig.Host)
	yamlContent += "# 监听端口，默认: 8080\n"
	yamlContent += fmt.Sprintf("port: %d\n\n", oldConfig.Port)
	yamlContent += "# 文件大小限制，默认: 10GB\n"
	yamlContent += fmt.Sprintf("sizeLimit: %d\n\n", oldConfig.SizeLimit)
	yamlContent += "# GitHub地址白名单，支持通配符\n"
	yamlContent += "whiteList: []\n"
	if len(oldConfig.WhiteList) > 0 {
		for _, item := range oldConfig.WhiteList {
			yamlContent += fmt.Sprintf("  - %s\n", item)
		}
	}
	yamlContent += "\n"
	yamlContent += "# GitHub地址黑名单，支持通配符\n"
	yamlContent += "blackList: []\n"
	if len(oldConfig.BlackList) > 0 {
		for _, item := range oldConfig.BlackList {
			yamlContent += fmt.Sprintf("  - %s\n", item)
		}
	}
	yamlContent += "\n"
	yamlContent += "# 是否允许代理非GitHub的其他地址\n"
	yamlContent += fmt.Sprintf("allowProxyAll: %t\n\n", oldConfig.AllowProxyAll)
	yamlContent += "# 其他地址白名单\n"
	yamlContent += "otherWhiteList: []\n"
	if len(oldConfig.OtherWhiteList) > 0 {
		for _, item := range oldConfig.OtherWhiteList {
			yamlContent += fmt.Sprintf("  - %s\n", item)
		}
	}
	yamlContent += "\n"
	yamlContent += "# 其他地址黑名单\n"
	yamlContent += "otherBlackList: []\n"
	if len(oldConfig.OtherBlackList) > 0 {
		for _, item := range oldConfig.OtherBlackList {
			yamlContent += fmt.Sprintf("  - %s\n", item)
		}
	}
	yamlContent += "\n"
	yamlContent += "# 唯一标识符，用于数据统计\n"
	yamlContent += fmt.Sprintf("uuid: %s\n", oldConfig.UUID)

	// 写入新配置文件
	err = os.WriteFile(newPath, []byte(yamlContent), 0644)
	if err != nil {
		return err
	}

	// 保留旧配置文件作为备份
	printlnWithTime("旧配置文件已保留作为备份")

	return nil
}

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 初始化API路由
func initAPIRoutes(router *gin.Engine) {
	apiGroup := router.Group("/api")
	{
		// 健康检查
		apiGroup.GET("/health", healthCheck)
		// 程序版本查询
		apiGroup.GET("/version", getVersion)
		// UUID查询
		apiGroup.GET("/uuid", getUUID)
	}
}

// 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "FastCode is running",
	})
}

// 获取程序版本
func getVersion(c *gin.Context) {
	shortCommit := commit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"commit":  shortCommit,
	})
}

// 获取UUID
func getUUID(c *gin.Context) {
	configLock.RLock()
	uuid := config.UUID
	configLock.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"uuid": uuid,
	})
}

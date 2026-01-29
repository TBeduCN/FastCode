package main

import (
	"embed"
	"io"
	"os"
	"path/filepath"
)

// 静态资源嵌入
//
//go:embed public/*
var embeddedPublic embed.FS

// 初始化静态资源
func initStaticFiles() {
	// 检查public目录是否存在
	if _, err := os.Stat("./public"); os.IsNotExist(err) {
		printlnWithTime("public目录不存在，使用嵌入的静态资源...")
		// 从嵌入的文件系统复制静态文件到本地
		copyEmbeddedFiles(embeddedPublic, "public", "./public")
	}
}

// 复制嵌入的文件到本地
func copyEmbeddedFiles(efs embed.FS, srcDir, dstDir string) {
	entries, err := efs.ReadDir(srcDir)
	if err != nil {
		printfWithTime("读取嵌入文件失败: %v\n", err)
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// 创建目录
			err := os.MkdirAll(dstPath, 0755)
			if err != nil {
				printfWithTime("创建目录失败: %v\n", err)
				continue
			}
			// 递归复制子目录
			copyEmbeddedFiles(efs, srcPath, dstPath)
		} else {
			// 复制文件
			srcFile, err := efs.Open(srcPath)
			if err != nil {
				printfWithTime("打开嵌入文件失败: %v\n", err)
				continue
			}

			dstFile, err := os.Create(dstPath)
			if err != nil {
				printfWithTime("创建本地文件失败: %v\n", err)
				srcFile.Close()
				continue
			}

			_, err = io.Copy(dstFile, srcFile)
			srcFile.Close()
			dstFile.Close()

			if err != nil {
				printfWithTime("复制文件失败: %v\n", err)
				continue
			}

			printfWithTime("复制静态文件: %s -> %s\n", srcPath, dstPath)
		}
	}
}

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
	// 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		printfWithTime("获取可执行文件路径失败: %v\n", err)
		// 失败时使用当前目录作为备选
		execPath = "."
	}
	execDir := filepath.Dir(execPath)
	publicDir := filepath.Join(execDir, "public")

	// 检查public目录是否存在
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		printlnWithTime("public目录不存在，使用嵌入的静态资源...")
		// 先创建public目录
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			printfWithTime("创建public目录失败: %v\n", err)
			return
		}
		// 从嵌入的文件系统复制静态文件到本地
		copyEmbeddedFiles(embeddedPublic, "public", publicDir)
	} else {
		// 检查public目录是否为空
		entries, err := os.ReadDir(publicDir)
		if err != nil {
			printfWithTime("读取public目录失败: %v\n", err)
			return
		}
		if len(entries) == 0 {
			printlnWithTime("public目录为空，使用嵌入的静态资源...")
			// 从嵌入的文件系统复制静态文件到本地
			copyEmbeddedFiles(embeddedPublic, "public", publicDir)
		}
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
		// 使用正斜杠作为路径分隔符，因为embed包要求
		srcPath := srcDir + "/" + entry.Name()
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

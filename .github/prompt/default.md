# $CURRENT_TAG

## 快速部署
我们提供了多平台的预编译二进制文件，无需安装依赖，可直接运行。

你可以下载对应平台的压缩包：

- Windows (amd64): fastcode_${CURRENT_TAG}_windows_amd64.zip
- Windows (386): fastcode_${CURRENT_TAG}_windows_386.zip
- Linux (amd64): fastcode_${CURRENT_TAG}_linux_amd64.tar.gz
- Linux (arm64): fastcode_${CURRENT_TAG}_linux_arm64.tar.gz
- Linux (386): fastcode_${CURRENT_TAG}_linux_386.tar.gz
- macOS (amd64): fastcode_${CURRENT_TAG}_darwin_amd64.tar.gz
- macOS (arm64): fastcode_${CURRENT_TAG}_darwin_arm64.tar.gz

启动后服务默认运行在 http://localhost:8080

## 注意事项
**默认配置文件会自动生成在程序目录下的 config/fastcode.yml**
**如需修改监听地址或端口，请编辑 config/fastcode.yml 文件**

---

**完整更新日志**: [${PREV_TAG}...${CURRENT_TAG}](https://github.com/TBeduCN/FastCode/compare/${PREV_TAG}...${CURRENT_TAG})
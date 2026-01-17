# FastCode - GitHub代理加速服务

## 项目介绍

这是一个基于Go语言开发的GitHub代理加速服务，用于加速访问GitHub资源，解决国内访问GitHub速度慢的问题。

## 功能特性

- ✅ GitHub Release文件加速
- ✅ GitHub Archive文件加速
- ✅ GitHub项目文件加速
- ✅ Git Clone支持
- ✅ Gist文件支持
- ✅ 白名单/黑名单机制
- ✅ 文件大小限制
- ✅ 动态配置加载
- ✅ 自动生成配置文件
- ✅ 静态资源嵌入
- ✅ Docker支持
- ✅ 响应式设计
- ✅ 暗色/亮色主题切换
- ✅ Waline评论系统支持

## 技术栈

- **语言**：Go 1.20+
- **框架**：Gin
- **依赖管理**：Go Modules

## 快速开始

### 直接运行

1. 克隆项目

```bash
git clone https://github.com/TBeduCN/FastCode.git
cd FastCode
```

2. 构建项目

```bash
go build -o fastcode main.go
```

3. 运行服务

```bash
./fastcode
```

4. 访问服务

打开浏览器访问 `http://localhost:8080`

### Docker运行

1. 构建镜像

```bash
docker build -t fastcode .
```

2. 运行容器

```bash
docker run -d -p 8080:8080 --name fastcode fastcode
```

## 配置说明

服务启动后会自动生成 `config.json` 配置文件，支持以下配置项：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `host` | string | `0.0.0.0` | 监听地址 |
| `port` | int | `8080` | 监听端口 |
| `sizeLimit` | int | `10737418240` | 文件大小限制（默认10GB） |
| `whiteList` | array | `[]` | GitHub地址白名单 |
| `blackList` | array | `[]` | GitHub地址黑名单 |
| `allowProxyAll` | bool | `false` | 是否允许代理非GitHub地址 |
| `otherWhiteList` | array | `[]` | 其他地址白名单 |
| `otherBlackList` | array | `[]` | 其他地址黑名单 |

### 配置示例

```json
{
  "host": "0.0.0.0",
  "port": 8080,
  "sizeLimit": 10737418240,
  "whiteList": ["microsoft", "google"],
  "blackList": [],
  "allowProxyAll": false,
  "otherWhiteList": [],
  "otherBlackList": []
}
```

## 使用方法

### 基本使用

在浏览器中访问服务地址，在输入框中输入GitHub URL，点击"加速访问"按钮即可。

### 直接访问

直接在URL前加上服务地址，例如：

```
# 原始URL
https://github.com/user/repo/releases/latest

# 代理URL
http://localhost:8080/github.com/user/repo/releases/latest
```

### Git Clone使用

```bash
git clone http://localhost:8080/github.com/user/repo.git
```

## 支持的URL类型

- `github.com/{user}/{repo}/releases/...` - Release文件
- `github.com/{user}/{repo}/archive/...` - 分支/标签归档
- `github.com/{user}/{repo}/blob/...` - 仓库文件
- `github.com/{user}/{repo}/raw/...` - 原始文件
- `github.com/{user}/{repo}/info/...` - Git信息
- `github.com/{user}/{repo}/git-...` - Git操作
- `raw.githubusercontent.com/{user}/{repo}/...` - Raw文件
- `gist.github.com/{user}/...` - Gist文件
- `api.github.com/...` - GitHub API

## Waline评论系统配置

### 功能介绍

Waline是一个基于Valine开发的评论系统，支持匿名评论、Markdown、表情、邮件通知等功能。

### 配置方法

1. **获取Waline服务端地址**

   你需要先部署一个Waline服务端，或者使用第三方提供的服务。推荐的部署方式：
   - [Waline官方部署文档](https://waline.js.org/guide/server/get-started.html)
   - 第三方服务：[Vercel](https://vercel.com)、[Netlify](https://www.netlify.com)等

2. **修改Waline配置**

   打开 `public/script.js` 文件，找到以下代码并修改 `serverURL` 为你的Waline服务端地址：

   ```javascript
   window.WalineInstance = init({
       el: '#waline',
       serverURL: 'https://waline.tbedu.top', // 修改为你的Waline服务端地址
       lang: 'zh-CN',
       dark: 'html[data-theme="dark"]',
       path: '/github',
       pageSize: 10,
       placeholder: '分享你的使用体验...',
       pageview: true,
       emoji: [
           'https://cdn.jsdelivr.net/gh/walinejs/emojis/weibo',
           'https://cdn.jsdelivr.net/gh/walinejs/emojis/bilibili'
       ]
   });
   ```

3. **自定义配置项**

   Waline支持多种配置项，你可以根据需要调整：
   - `lang`：评论系统语言，支持 `zh-CN`、`en` 等
   - `dark`：暗黑模式配置，当前设置为跟随页面主题
   - `path`：评论存储路径，默认为 `/github`
   - `pageSize`：每页显示的评论数量，默认为10
   - `placeholder`：评论输入框占位符
   - `pageview`：是否启用页面访问统计
   - `emoji`：支持的表情包列表

   更多配置项请参考 [Waline官方文档](https://waline.js.org/guide/client/config.html)

## 性能优化

- 优化的HTTP客户端配置
- 连接池管理
- 超时设置
- 流式传输大文件
- 移除不必要的响应头

## 项目结构

```
.
├── main.go              # 主程序入口
├── go.mod               # Go模块依赖
├── go.sum               # 依赖校验和
├── config.json          # 配置文件（自动生成）
├── public/              # 静态资源目录
│   ├── index.html       # 首页
│   ├── logo.png         # Logo
│   ├── styles.css       # 样式文件
│   └── script.js        # JavaScript文件
├── Dockerfile           # Docker构建文件
└── README.md            # 项目说明
```

## 开发说明

### 安装依赖

```bash
go mod tidy
```

### 运行测试

```bash
go test ./...
```

### 构建二进制文件

```bash
go build -o gh-proxy main.go
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 致谢

- 感谢 [gh-proxy](https://github.com/hunshcn/gh-proxy) 项目提供的灵感
- 感谢 [Gin](https://github.com/gin-gonic/gin) 框架
- 感谢所有贡献者

## 联系方式

如有问题或建议，欢迎提交 Issue 或 Pull Request。
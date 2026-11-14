<p align="center"><img src="web/public/logo.png" alt="Elaina Logo" width="96" /></p>
<h1 align="center">Elaina</h1>
<p align="center">基于 WebAssembly 的远程代码运行器。</p>
<p align="center"><a href="README.md">English</a></p>

## 快速开始

把下面这段话发给你的 Agent。

```text
帮我在本机装好并启动 Elaina（https://github.com/wuhan005/Elaina）。如果已经在项目目录里，就用当前版本。

检查构建依赖和 PostgreSQL。用 scripts/download-wasm.sh 和 scripts/build-wasm-go.sh 准备运行时，再用 pnpm 构建前端、用 Go 构建服务。如果还没有配置，按 config/elaina.example.yaml 填好数据库连接和初始管理员密码的 bcrypt 哈希。

启动服务，确认页面能打开，然后告诉我访问地址和怎么登录。
```

[MIT 许可证](LICENSE)

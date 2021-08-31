# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/wuhan005/Elaina)](https://goreportcard.com/report/github.com/wuhan005/Elaina) ![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/wuhan005/elaina) ![Docker Image Version (latest by date)](https://img.shields.io/docker/v/wuhan005/elaina)

<img align="right" src="elaina.gif" width=40%/>
基于 Docker 的远程代码运行器。

## 开始使用

### 步骤 1: 安装依赖

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 或更高)
* [Docker Compose](https://docs.docker.com/compose/install/) (v1.27.4 或更高)

### 步骤 2: 拉取内置 Docker 镜像

[`docker/images/`](https://github.com/wuhan005/Elaina/tree/master/docker/images) 文件夹内提供的 Elaina 所支持的编程语言运行环境。

在运行 Elaina 前，请使用 `docker pull` 命令从 DockerHub 拉取这些镜像。该操作只需执行一次即可。

```bash
docker pull elainaruntime/golang
docker pull elainaruntime/php
docker pull elainaruntime/python
docker pull elainaruntime/javascript
```

### Step 3: 启动 Elaina

1. 设置环境变量

Postgres 数据库配置

```bash
export PGPORT=5432
export PGHOST=<REDACTED>
export PGUSER=<REDACTED>
export PGPASSWORD=<REDACTED>
export PGDATABASE=<REDACTED>
export PGSSLMODE=disable
```

2. 启动 Elaina 服务

```bash
# 设置管理网页端密码
export APP_PASSWORD=<REDACTED>

./Elaina
```

### 步骤 4: 走你！

浏览器访问 `http://<your-host>:8080/m/` 来登录管理员面板。

## 开源协议

MIT

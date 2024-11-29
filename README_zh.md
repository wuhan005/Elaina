# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/wuhan005/Elaina)](https://goreportcard.com/report/github.com/wuhan005/Elaina) ![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/wuhan005/elaina) ![Docker Image Version (latest by date)](https://img.shields.io/docker/v/wuhan005/elaina)

<img align="right" src="elaina.gif" width="40%"/>
基于容器的远程代码运行器。

## 开始使用

### 步骤 1: 安装依赖

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 或更高)
* [Postgres](https://www.postgresql.org/download/) (v13.1 或更高)

### 步骤 2: 拉取内置 Docker 镜像

在运行 Elaina 前，请使用 `docker pull` 命令从 DockerHub 拉取这些镜像。该操作只需执行一次即可。

```bash
docker pull glot/php:latest
docker pull glot/python:latest
docker pull glot/golang:latest
docker pull glot/javascript:latest
docker pull glot/c:latest
```

### Step 3: 编译并启动 Elaina

#### 编译 Elaina

```bash
git clone git@github.com:wuhan005/Elaina.git

# 编译前端
cd frontend/ && yarn install && yarn build

# 编译后端
go build .
```

#### 设置环境变量

```bash
export APP_PASSWORD=<REDACTED>
export RUNTIME_MODE=docker
export POSTGRES_DSN=postgres://postgres:<REDACTED>@127.0.0.1:5432/elaina
```

#### 运行 Elaina

```bash
./Elaina
```

### 步骤 4: 走你！

浏览器访问 `http://<your-host>:8080/` 来登录管理员面板。

## 开源协议

MIT License

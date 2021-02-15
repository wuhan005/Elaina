# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg)

<img align="right" src="elaina.gif" width=40%/>
基于 Docker 的远程代码运行器。

## 开始使用

### 步骤 1: 安装依赖

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 或更高)
* [Docker Compose](https://docs.docker.com/compose/install/) (v1.27.4 或更高)

### 步骤 2: 构建内置 Docker 镜像

[`docker/images/`](https://github.com/wuhan005/Elaina/tree/master/docker/images) 文件夹内提供的 Elaina 所支持的编程语言运行环境。

在运行 Elaina 前，请使用 `docker build` 命令编译这些镜像。该操作只需执行一次即可。

```bash
docker build . -t elaina-<lang>:latest
```

### 步骤 3: 启动 Elaina

将 [docker-compose.yml](https://github.com/wuhan005/Elaina/blob/master/docker-compose.yml) 文件放置于您的运行目录。

**修改 docker-compose.yml 中的 `APP_URL` `APP_PASSWORD` `APP_CONTAINER_PATH` 参数**

* `APP_URL` 您后端服务的主机地址，它将被用作设置后端 HTTP 中的 CORS 允许来源地址响应头。
* `APP_PASSWORD` 该密码用于登录管理员面板。
* `APP_CONTAINER_PATH` 该目录放置容器运行时在**宿主机**上产生的临时目录，请确保 Docker 拥有正确的权限来访问该文件夹。

下面的命令将启动 PostgreSQL 数据库以及 Elaina 服务。

```bash
docker-compose up -d
```

### 步骤 4: 走你！

浏览器访问 `http://<your-host>:8080/m/` 来登录管理员面板。

## 开源协议

MIT
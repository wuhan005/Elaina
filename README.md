# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/wuhan005/Elaina)](https://goreportcard.com/report/github.com/wuhan005/Elaina) ![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/wuhan005/elaina) ![Docker Image Version (latest by date)](https://img.shields.io/docker/v/wuhan005/elaina)

<img align="right" src="elaina.gif" width=40%/>
Docker-based remote code runner.

[简体中文](https://github.com/wuhan005/Elaina/blob/master/README_zh.md)

## Start

### Step 1: Install dependencies

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 or higher)
* [Docker Compose](https://docs.docker.com/compose/install/) (v1.27.4 or higher)

### Step 2: Pull internal docker images

The [`docker/images/`](https://github.com/wuhan005/Elaina/tree/master/docker/images) folder provides the Dockerfile of
the programming language runtime environment that Elaina supports.

Use `docker pull` command to pull the images from DockerHub before you start running the Elaina. This operation only
needs to be performed once.

```bash
docker pull elainaruntime/golang
docker pull elainaruntime/php
docker pull elainaruntime/python
docker pull elainaruntime/javascript
```

### Step 3: Start the Elaina server

Put the [docker-compose.yml](https://github.com/wuhan005/Elaina/blob/master/docker-compose.yml) file in your working
directory.

**Edit `APP_URL` `APP_PASSWORD` `APP_CONTAINER_PATH` in docker-compose.yml!!**

* `APP_URL` Your backend service host, used to set the allow origins header in HTTP CORS header.
* `APP_PASSWORD` The password used to log in the manager panel.
* `APP_CONTAINER_PATH` The path where the containers' volumes are placed in your **host**, make sure the Docker has the
  correct permission to access.

The following command will create a PostgreSQL database as well as the Elaina server.

```bash
docker-compose up -d
```

### Step 4: Have fun!

Visit `http://<your-host>:8080/m/` to login to the manager panel.

## License

MIT

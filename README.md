# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/wuhan005/Elaina)](https://goreportcard.com/report/github.com/wuhan005/Elaina) ![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/wuhan005/elaina) ![Docker Image Version (latest by date)](https://img.shields.io/docker/v/wuhan005/elaina)

<img align="right" src="elaina.gif" width="40%"/>
Container-based remote code runner.

[简体中文](https://github.com/wuhan005/Elaina/blob/master/README_zh.md)

## Start

### Step 1: Install dependencies

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 or higher)
* [Postgres](https://www.postgresql.org/download/) (v13.1 or higher)

### Step 2: Pull internal docker images

Use `docker pull` command to pull the images from DockerHub before you start running the Elaina. This operation only
needs to be performed once.

```bash
docker pull glot/php:latest
docker pull glot/python:latest
docker pull glot/golang:latest
docker pull glot/javascript:latest
docker pull glot/c:latest
```

### Step 3: Build and start the Elaina server

#### Build Elaina

```bash
git clone git@github.com:wuhan005/Elaina.git

# Build frontend
cd frontend/ && yarn install && yarn build

# Build backend
go build .
```

#### Set environment variables.

```bash
export APP_PASSWORD=<REDACTED>
export RUNTIME_MODE=docker
export POSTGRES_DSN=postgres://postgres:<REDACTED>@127.0.0.1:5432/elaina
```

#### Run the Elaina server.

```bash
./Elaina
```

### Step 4: Have fun!

Visit `http://<your-host>:8080/` to login to the manager panel.

## License

MIT License

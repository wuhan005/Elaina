# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/wuhan005/Elaina)](https://goreportcard.com/report/github.com/wuhan005/Elaina) ![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/wuhan005/elaina) ![Docker Image Version (latest by date)](https://img.shields.io/docker/v/wuhan005/elaina)

<img align="right" src="elaina.gif" width=40%/>
Docker-based remote code runner.

[简体中文](https://github.com/wuhan005/Elaina/blob/master/README_zh.md)

## Start

### Step 1: Install dependencies

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 or higher)

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

1. Set the environment variables.

Postgres database configuration.

```bash
export PGPORT=5432
export PGHOST=<REDACTED>
export PGUSER=<REDACTED>
export PGPASSWORD=<REDACTED>
export PGDATABASE=<REDACTED>
export PGSSLMODE=disable
```

2. Run the Elaina server.

```bash
# Set the web manager panel password.
export APP_PASSWORD=<REDACTED>

./Elaina
```

### Step 4: Have fun!

Visit `http://<your-host>:8080/m/` to login to the manager panel.

## License

MIT

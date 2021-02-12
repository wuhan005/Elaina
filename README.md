# Elaina ![Go](https://github.com/wuhan005/Elaina/workflows/Go/badge.svg)

<img align="right" src="elaina.gif" width=40%/>
Docker-based remote code runner.

## Start

### Step 1: Install dependencies

* [Docker](https://docs.docker.com/get-docker/) (v20.10.0 or higher)
* [Docker Compose](https://docs.docker.com/compose/install/) (v1.27.4 or higher)

### Step 2: Build internal docker images

The [`docker/images/`](https://github.com/wuhan005/Elaina/tree/master/docker/images) folder provides the Dockerfile of
the programming language runtime environment that Elaina supports.

Use `docker build` command to build the images before you start running the Elaina. This operation only needs to be
performed once.

```bash
docker build . -t elaina-<lang>:latest
```

### Step 3: Start the Elaina server

Put the [docker-compose.yml](https://github.com/wuhan005/Elaina/blob/master/docker-compose.yml) file in your working
directory.

**Edit `APP_URL` `APP_PASSWORD` in docker-compose.yml!!**

The following command will create a PostgreSQL database as well as the Elaina server.

```bash
docker-compose up -d
```

### Step 4: Have fun!

Visit `http://<your-host>:8080/m/` to login to the manager panel.

## License

MIT
# docker build . -t elaina-go:latest
FROM golang:1.15-alpine

ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.io"

RUN mkdir -p /runtime
WORKDIR /runtime

ENTRYPOINT ["sleep", "infinity"]

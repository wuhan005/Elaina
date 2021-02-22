# docker build . -t elaina-go:latest
FROM golang:1.15-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina

RUN mkdir /runner
WORKDIR /runner

USER elaina
# docker build . -t elaina-go:latest
FROM golang:1.15-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina
ADD daemon/daemon /daemon

ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.io"

RUN mkdir -p /runtime/runner
WORKDIR /runtime/runner

RUN go mod init elaina-runner
RUN chown -R elaina:elaina /go
RUN chown -R elaina:elaina /runtime

USER elaina
CMD ["../../daemon"]
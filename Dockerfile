FROM node:latest as node_builder

WORKDIR /app

COPY . .

WORKDIR /app/web
RUN rm -rf node_modules || true
RUN npm install -g pnpm
RUN git init
RUN pnpm install
RUN pnpm run build

FROM golang:1.22-alpine as go_builder

WORKDIR /app

ENV CGO_ENABLED=0

COPY . .
COPY --from=node_builder /app/web/dist ./web/dist

RUN go mod tidy
RUN go build -v -trimpath -ldflags "-w -s -extldflags '-static'" -o elaina .

FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' > /etc/timezone

RUN mkdir /etc/elaina
WORKDIR /etc/elaina

COPY --from=go_builder /app/elaina /etc/elaina/elaina

RUN chmod 655 /etc/elaina/elaina

ENTRYPOINT ["/etc/elaina/elaina"]
EXPOSE 8080

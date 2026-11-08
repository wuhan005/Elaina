FROM alpine:3.23 AS wasm_builder

RUN apk add --no-cache curl ca-certificates
COPY scripts/download-wasm.sh /scripts/download-wasm.sh
COPY tools/wasm-c /tools/wasm-c
COPY LICENSE /LICENSE
RUN sh /scripts/download-wasm.sh /wasm

FROM golang:1.27.3-alpine AS go_wasm_builder

WORKDIR /app
COPY tools/wasm-go ./tools/wasm-go
COPY scripts/build-wasm-go.sh ./scripts/build-wasm-go.sh
RUN sh ./scripts/build-wasm-go.sh /wasm

FROM node:latest AS node_builder

WORKDIR /app

COPY . .

WORKDIR /app/web
RUN rm -rf node_modules || true
RUN npm install -g pnpm@10.33.4
RUN git init
RUN pnpm install --frozen-lockfile
RUN pnpm run build

FROM golang:1.25-alpine AS go_builder

WORKDIR /app

ENV CGO_ENABLED=0

COPY . .
COPY --from=node_builder /app/web/dist ./web/dist

RUN go mod tidy
RUN go build -v -trimpath -ldflags "-w -s -extldflags '-static'" -o elaina .

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
RUN addgroup -g 10001 elaina && adduser -D -u 10001 -G elaina elaina
WORKDIR /etc/elaina

COPY --from=go_builder /app/elaina /etc/elaina/elaina
COPY --from=wasm_builder /wasm /etc/elaina/.wasm
COPY --from=go_wasm_builder /wasm/go /etc/elaina/.wasm/go
RUN mkdir -p /etc/elaina/config /etc/elaina/.wasm/cache && chown 10001:10001 /etc/elaina/.wasm/cache
# Mount config/elaina.yaml read-only and make it readable by UID 10001.
USER 10001:10001

ENTRYPOINT ["/etc/elaina/elaina"]
EXPOSE 8080

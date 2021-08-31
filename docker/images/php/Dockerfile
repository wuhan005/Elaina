# docker build . -t elaina-php:latest
FROM php:cli-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina

RUN mkdir -p /runtime
WORKDIR /runtime

ENTRYPOINT ["sleep", "infinity"]

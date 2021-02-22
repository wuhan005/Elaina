# docker build . -t elaina-php:latest
FROM php:cli-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina

RUN mkdir /runner
WORKDIR /runner

USER elaina
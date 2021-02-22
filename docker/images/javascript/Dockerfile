# docker build . -t elaina-javascript:latest
FROM node:lts-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina

RUN mkdir /runner
WORKDIR /runner

USER elaina
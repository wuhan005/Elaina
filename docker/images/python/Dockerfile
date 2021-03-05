# docker build . -t elaina-python:latest
FROM python:3.9.1-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina
ADD daemon/daemon /daemon

RUN mkdir -p /runtime
WORKDIR /runtime

ENTRYPOINT ["../../daemon"]
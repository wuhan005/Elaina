# docker build . -t elaina-python:latest
FROM python:3.9.1-alpine
RUN pip3 install requests
RUN addgroup -S elaina && adduser -S elaina -G elaina

RUN mkdir -p /runtime
WORKDIR /runtime

ENTRYPOINT ["sleep", "infinity"]

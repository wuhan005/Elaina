# docker build . -t elaina-clang:latest
FROM gcc:latest
RUN addgroup -S elaina && adduser -S elaina -G elaina
ADD daemon/daemon /daemon

RUN mkdir -p /runtime
WORKDIR /runtime

ENTRYPOINT ["../../daemon"]
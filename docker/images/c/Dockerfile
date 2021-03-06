# docker build . -t elaina-clang:latest
FROM gcc:latest
ADD daemon/daemon /daemon

RUN mkdir -p /runtime
WORKDIR /runtime

ENTRYPOINT ["../../daemon"]
# docker build . -t elaina-javascript:latest
FROM node:lts-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina
ADD daemon/daemon /daemon

RUN mkdir -p /runtime
WORKDIR /runtime

RUN chown -R elaina:elaina /runtime

USER elaina
ENTRYPOINT ["../../daemon"]
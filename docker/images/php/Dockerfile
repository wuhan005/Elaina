# docker build . -t elaina-php:latest
FROM php:cli-alpine
RUN addgroup -S elaina && adduser -S elaina -G elaina
ADD daemon/daemon /daemon

RUN mkdir -p /runtime
WORKDIR /runtime

RUN chown -R elaina:elaina /runtime

USER elaina
ENTRYPOINT ["../../daemon"]
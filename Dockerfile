FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

RUN mkdir /etc/Elaina
WORKDIR /etc/Elaina

ADD Elaina /etc/Elaina
ADD templates /etc/Elaina/templates
ADD public /etc/Elaina/public

RUN chmod 655 /etc/Elaina/Elaina

ENTRYPOINT ["/etc/Elaina/Elaina"]
EXPOSE 8080
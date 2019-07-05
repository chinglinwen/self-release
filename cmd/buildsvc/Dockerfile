# FROM harbor.haodai.net/base/alpine:3.7cgo
# FROM alpine/git:1.0.7
FROM harbor.haodai.net/base/alpine-git:1.0.7

WORKDIR /app

MAINTAINER wenzhenglin(http://g.haodai.net/wenzhenglin/buildsvc)

RUN apk add --update bash curl docker libc6-compat \
    && rm -rf /var/cache/apk/*

COPY buildsvc /app
# COPY /usr/bin/docker /bin/docker

CMD /app/buildsvc
ENTRYPOINT ["./buildsvc"]

EXPOSE 8080
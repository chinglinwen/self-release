FROM harbor.haodai.net/base/alpine:3.7cgo
WORKDIR /app

MAINTAINER wenzhenglin(http://g.haodai.net/wenzhenglin/buildsvc)

RUN apk add --update bash curl docker \
    && rm -rf /var/cache/apk/*

COPY buildsvc /app
# COPY /usr/bin/docker /bin/docker

CMD /app/buildsvc
ENTRYPOINT ["./buildsvc"]

EXPOSE 8080
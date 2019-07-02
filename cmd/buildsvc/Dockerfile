FROM harbor.haodai.net/base/alpine:3.7cgo
WORKDIR /app

MAINTAINER wenzhenglin(http://g.haodai.net/wenzhenglin/buildsvc)

COPY buildsvc /app

CMD /app/buildsvc
ENTRYPOINT ["./buildsvc"]

EXPOSE 8080
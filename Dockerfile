FROM harbor.haodai.net/base/alpine:3.7cgo
WORKDIR /app

MAINTAINER wenzhenglin(http://g.haodai.net/wenzhenglin/self-release.git)

COPY self-release /app

CMD /app/self-release
ENTRYPOINT ["./self-release"]

# EXPOSE 8080
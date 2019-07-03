FROM harbor.haodai.net/base/alpine:3.7cgo
WORKDIR /app

MAINTAINER wenzhenglin(http://g.haodai.net/wenzhenglin/self-release.git)

COPY self-release /app

# box path need to be this place
COPY web /home/wen/gocode/src/wen/self-release/web
RUN ln -sf /home/wen/gocode/src/wen/self-release/web /app/web

CMD /app/self-release
ENTRYPOINT ["./self-release"]

# EXPOSE 8080
FROM harbor.haodai.net/base/alpine:3.7cgo
WORKDIR /app

MAINTAINER wenzhenglin(http://g.haodai.net/wenzhenglin/self-release.git)

RUN wget http://fs.haodai.net/soft/kubectl -O /bin/kubectl && \
    wget http://fs.haodai.net/soft/helm-v2.14.3 -O /bin/helm && \
    chmod +x /bin/kubectl /bin/helm && \
    wget -O - http://fs.haodai.net/k8s/v1.14/addkubeconfig.sh | sh

COPY self-release /app

# box path need to be this place
COPY web /home/wen/gocode/src/wen/self-release/web

# the following is for optional convenient
RUN ln -sf /home/wen/gocode/src/wen/self-release/web /app/web

# we use volume mount to create /app/projectlogs

CMD /app/self-release

ENTRYPOINT ["./self-release"]

# EXPOSE 8080
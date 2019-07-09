#!/bin/sh
# build image
set -e

suffix="$1"
suffix=${suffix:=v1}

go build || exit 1

image="self-release:$suffix"
echo -e "building image: $image\n"
tag="harbor.haodai.net/ops/$image"
docker build -t $tag .
docker push $tag

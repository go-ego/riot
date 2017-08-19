#!/usr/bin/env bash

set -x

rm -rf docker
mkdir docker
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o docker/search_server
cp Dockerfile docker/
cp static docker/ -r
cp ../../data/dict/dictionary.txt docker/
cp ../../data/stop_tokens.txt docker/
cp ../../testdata/weibo_data.txt docker/

docker build -t unmerged/gwk-codelab -f docker/Dockerfile docker/

#!/usr/bin/env bash

if [[ -z "$1" ]]; then
    echo "A remote server address is required to build the ratel-webterminal docker image."
    echo "For example: ./build.sh 1.1.1.1"
    exit 0
fi

serverAddr="$1"
buildDir="/tmp/ratel-webterminal"

sync() {
    rsync -avzh --delete --exclude=".git" --exclude "tmp" ./* root@$serverAddr:$buildDir
}

build() {
    cd $buildDir
    docker build -t hybfkuf/ratel-webterminal:latest .
    docker push hybfkuf/ratel-webterminal:latest
}

sync
ssh root@$serverAddr "
    $(typeset -f build)
    export buildDir=$buildDir
    build
"

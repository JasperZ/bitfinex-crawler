#!/bin/sh

VERSION="0.10.0"

# build image
cd docker
docker build -t $DOCKER_USERNAME/bitfinex-crawler:$VERSION -f docker/Dockerfile .
cd ..

# login to docker hub
docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

# upload images
docker images
docker push $DOCKER_USERNAME/bitfinex-crawler:$VERSION

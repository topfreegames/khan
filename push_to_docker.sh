#!/bin/bash

VERSION=$(cat version.txt)

docker build -t khan .
docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker tag khan:latest tfgco/khan:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/khan:$VERSION.$TRAVIS_BUILD_NUMBER

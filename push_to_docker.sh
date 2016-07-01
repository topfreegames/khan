#!/bin/bash

VERSION=$(cat ./api/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)

docker build -t khan .
docker build -t khan-dev ./dev
docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker tag khan:latest tfgco/khan:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/khan:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag khan-dev:latest tfgco/khan-dev:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/khan-dev:$VERSION.$TRAVIS_BUILD_NUMBER

DOCKERHUB_LATEST=$(python get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
    exit 1
fi

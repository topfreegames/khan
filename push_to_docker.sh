#!/bin/bash

VERSION=$(cat ./util/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)

cp ./config/default.yaml ./dev

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker build -t khan .
docker tag khan:latest tfgco/khan:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag khan:latest tfgco/khan:$VERSION
docker tag khan:latest tfgco/khan:latest
docker push tfgco/khan:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/khan:$VERSION
docker push tfgco/khan:latest

docker build -t khan-dev ./dev
docker tag khan-dev:latest tfgco/khan-dev:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag khan-dev:latest tfgco/khan-dev:$VERSION
docker tag khan-dev:latest tfgco/khan-dev:latest
docker push tfgco/khan-dev:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/khan-dev:$VERSION
docker push tfgco/khan-dev:latest

docker build -t khan-prune -f PruneDockerfile .
docker tag khan-prune:latest tfgco/khan-prune:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag khan-prune:latest tfgco/khan-prune:$VERSION
docker tag khan-prune:latest tfgco/khan-prune:latest
docker push tfgco/khan-prune:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/khan-prune:$VERSION
docker push tfgco/khan-prune:latest


DOCKERHUB_LATEST=$(python get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
    exit 1
fi

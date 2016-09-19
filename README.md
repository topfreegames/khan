# Khan

[![Build Status](https://travis-ci.org/topfreegames/khan.svg?branch=master)](https://travis-ci.org/topfreegames/khan)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/khan/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/khan?branch=master)
[![Code Climate](https://codeclimate.com/github/topfreegames/khan/badges/gpa.svg)](https://codeclimate.com/github/topfreegames/khan)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/khan)](https://goreportcard.com/report/github.com/topfreegames/khan)
[![Docs](https://readthedocs.org/projects/khan-api/badge/?version=latest
)](http://khan-api.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/khan:latest.svg)](https://imagelayers.io/?images=tfgco/khan:latest 'Khan Image Layers')

Khan will drive all your enemies to the sea (and also take care of your game's clans)!

## Setup

Make sure you have go installed on your machine.
If you use homebrew you can install it with `brew install go`.

Run `make setup`.

## Running the application

Create the development database with `make migrate` (first time only).

Run the api with `make run`.

## Hosting the Application

You can run it easily in one of the cloud providers using our [docker images](https://hub.docker.com/r/tfgco/khan/).

The following environment variables are available to you:

* KHAN_POSTGRES_HOST
* KHAN_POSTGRES_PORT
* KHAN_POSTGRES_DBNAME
* KHAN_POSTGRES_USER
* KHAN_POSTGRES_PASSWORD
* KHAN_POSTGRES_SSLMODE
* KHAN_SENTRY_URL
* KHAN_ELASTICSEARCH_ENABLED
* KHAN_ELASTICSEARCH_HOST
* KHAN_ELASTICSEARCH_PORT
* KHAN_ELASTICSEARCH_SNIFF
* KHAN_ELASTICSEARCH_INDEX

If elasticsearch is set to enabled, khan will save clans into KHAN_ELASTICSEARCH_INDEX and keep them updated with all the clan's updates.

## Running with docker

Provided you have docker installed, to build Khan's image run:

    $ make build-docker

To run a new khan instance, run:

    $ make run-docker

## Docker Image

You can get a docker image from our [dockerhub page](https://hub.docker.com/r/tfgco/khan/).

## Tests

Running tests can be done with `make test`, while creating the test database can be accomplished with `make drop-test` and `make db-test`.

## Benchmark

Running benchmarks can be done with `make ci-perf`.

## Coverage

Getting coverage data can be achieved with `make coverage`, while reading the actual results can be done with `make coverage-html`.

## Static Analysis

Khan goes through some static analysis tools for go. To run them just use `make static`.

Right now, gocyclo can't process the vendor folder, so we just ignore the exit code for it, while maintaining the output for anything not in the vendor folder.


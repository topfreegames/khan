# Khan

[![Khan](https://github.com/topfreegames/khan/actions/workflows/go.yml/badge.svg)](https://github.com/topfreegames/khan/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/khan/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/khan?branch=master)
[![Code Climate](https://codeclimate.com/github/topfreegames/khan/badges/gpa.svg)](https://codeclimate.com/github/topfreegames/khan)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/khan)](https://goreportcard.com/report/github.com/topfreegames/khan)
[![Docs](https://readthedocs.org/projects/khan-api/badge/?version=latest
)](http://khan-api.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/khan:latest.svg)](https://imagelayers.io/?images=tfgco/khan:latest 'Khan Image Layers')

Khan will drive all your enemies to the sea (and also take care of your game's clans)!

What is Khan? Khan is an HTTP "resty" API for managing clans for games. It could be used to manage groups of people, but our aim is players in a game.

Khan allows your app to focus on the interaction required to creating clans and managing applications, instead of the backend required for actually doing it.

## Features

* **Multi-tenant** - Khan already works for as many games as you need, just keep adding new games;
* **Clan Management** - Create and manage clans, their metadata as well as promote and demote people in their rosters;
* **Player Management** - Manage players and their metadata, as well as their applications to clans;
* **Applications** - Khan handles the work involved with applying to clans, inviting people to clans, accepting, denying and kicking;
* **Clan Search** - Search a list of clans to present your player with relevant options;
* **Top Clans** - Choose from a specific dimension to return a list of the top clans in that specific range (SOON);
* **Web Hooks** - Need to integrate your clan system with another application? We got your back! Use our web hooks sytem and plug into whatever events you need;
* **Auditing Trail** - Track every action coming from your games (SOON);
* **New Relic Support** - Natively support new relic with segments in each API route for easy detection of bottlenecks;
* **Easy to deploy** - Khan comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!

Read more about Khan in our [comprehensive documentation](http://khan-api.readthedocs.io/).

## Hacking Khan

### Setup

Make sure you have go installed on your machine.
If you use homebrew you can install it with `brew install go`.

Run `make setup`.

### Running the application

Create the development database with `make migrate` (first time only).

Run the api with `make run`.

### Running with docker

Provided you have docker installed, to build Khan's image run:

    $ make build-docker

To run a new khan instance, run:

    $ make run-docker

### Running with docker-compose

We already provide a docker-compose.yml as well with all dependencies configured for you to run. To run Khan and all its dependencies, run:

```sh
    $ docker-compose up
```

**Note** If you are running it on MacOS, you will need to update the amount of RAM docker has access to. Docker, by default, can use 2GB of RAM, however, Khan uses an instance of ElasticSearch and it needs at least 2GB of RAM to work properly. So, if you are experiencing problems while connecting to the elastic search, this might be the root cause of the problem.

### Tests

Running tests can be done with `make test`, while creating the test database can be accomplished with `make drop-test` and `make db-test`.

### Benchmark

Running benchmarks can be done with `make ci-perf`.

### Coverage

Getting coverage data can be achieved with `make coverage`, while reading the actual results can be done with `make coverage-html`.

### Static Analysis

Khan goes through some static analysis tools for go. To run them just use `make static`.

Right now, gocyclo can't process the vendor folder, so we just ignore the exit code for it, while maintaining the output for anything not in the vendor folder.

## Security

If you have found a security vulnerability, please email security@tfgco.com

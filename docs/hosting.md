Hosting Khan
============

There are three ways to host Khan: docker, binaries or from source.

## Docker

Running Khan with docker is rather simple. Our docker container image comes bundled with the API binary. All you need to do is load balance all the containers and you're good to go. The API runs at port `8080` in the docker image.

Khan uses PostgreSQL to store clans information. The container takes environment variables to specify this connection:

* `KHAN_POSTGRES_HOST` - PostgreSQL host to connect to;
* `KHAN_POSTGRES_PORT` - PostgreSQL port to connect to;
* `KHAN_POSTGRES_USER` - Password of the PostgreSQL Server to connect to;
* `KHAN_POSTGRES_DBNAME` - Database name of the PostgreSQL Server to connect to;
* `KHAN_POSTGRES_SSLMODE` - SSL Mode to connect to postgres with;

Other than that, there are a couple more configurations you can pass using environment variables:

* `KHAN_EXTENSIONS_DOGSTATSD_HOST` - If you have a [statsd datadog daemon](https://docs.datadoghq.com/developers/dogstatsd/), Podium will publish metrics to the given host at a certain port. Ex. localhost:8125;
* `KHAN_EXTENSIONS_DOGSTATSD_RATE` - If you have a [statsd daemon](https://docs.datadoghq.com/developers/dogstatsd/), Podium will export metrics to the deamon at the given rate;
* `KHAN_EXTENSIONS_DOGSTATSD_TAGS_PREFIX` - If you have a [statsd daemon](https://docs.datadoghq.com/developers/dogstatsd/), you may set a prefix to every tag sent to the daemon;

If you want to expose Khan outside your internal network it's advised to use Basic Authentication. You can specify basic authentication parameters with the following environment variables:

* `KHAN_BASICAUTH_USERNAME` - If you specify this key, Khan will be configured to use basic auth with this user;
* `KHAN_BASICAUTH_PASSWORD` - If you specify `BASICAUTH_USERNAME`, Khan will be configured to use basic auth with this password;

### Example command for running with Docker

```
    $ docker pull tfgco/khan
    $ docker run -t --rm -e "KHAN_POSTGRES_HOST=<postgres host>" -e "KHAN_POSTGRES_PORT=<postgres port>" -p 8080:80 tfgco/khan
```

In order to run Khan's workers using docker you just need to send the `KHAN_RUN_WORKER` environment variable as `true`.

### Example command for running workers with Docker

```
    $ docker pull tfgco/khan
    $ docker run -t --rm -e "KHAN_POSTGRES_HOST=<postgres host>" -e "KHAN_POSTGRES_PORT=<postgres port>" -e "KHAN_RUN_WORKERS=true" -p 9999:80 tfgco/khan
```


## Binaries

Whenever we publish a new version of Khan, we'll always supply binaries for both Linux and Darwin, on i386 and x86_64 architectures. If you'd rather run your own servers instead of containers, just use the binaries that match your platform and architecture.

The API server is the `khan` binary. It takes a configuration yaml file that specifies the connection to PostgreSQL and some additional parameters. You can learn more about it at [default.yaml](https://github.com/topfreegames/khan/blob/master/config/default.yaml).

The workers can be started using the same `khan` binary. It takes a configuration yaml file that specifies the connection to PostgreSQL and some additional parameters. You can learn more about it at [default.yaml](https://github.com/topfreegames/khan/blob/master/config/default.yaml).

## Source

Left as an exercise to the reader.

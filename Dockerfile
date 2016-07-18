FROM golang:1.6.2-alpine

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 80

RUN apk update
RUN apk add git make g++ nginx supervisor apache2-utils

RUN go get -u github.com/Masterminds/glide/...
RUN go get -u github.com/topfreegames/goose/cmd/goose

ADD . /go/src/github.com/topfreegames/khan

WORKDIR /go/src/github.com/topfreegames/khan
RUN glide install
RUN go install github.com/topfreegames/khan

ENV KHAN_POSTGRES_HOST 0.0.0.0
ENV KHAN_POSTGRES_PORT 5432
ENV KHAN_POSTGRES_USER khan
ENV KHAN_POSTGRES_DBNAME khan
ENV KHAN_SENTRY_URL ""

# configure supervisord
ADD ./docker/supervisord-khan.conf /etc/supervisord-khan.conf

# Configure nginx
ADD ./docker/nginx_default /etc/nginx/sites-enabled/default
ADD ./docker/nginx_conf /etc/nginx/nginx.conf
ADD ./docker/nginx_htpasswd /etc/nginx/.htpasswd

CMD /bin/sh -l -c docker/start.sh

FROM golang:1.6.2-alpine

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 8080

RUN apk update
RUN apk add git make g++

RUN go get -u github.com/Masterminds/glide/...
RUN go get -u bitbucket.org/liamstask/goose/cmd/goose

ADD . /go/src/github.com/topfreegames/khan

WORKDIR /go/src/github.com/topfreegames/khan 
RUN glide install
RUN go install github.com/topfreegames/khan

ENV KHAN_POSTGRES_HOST 0.0.0.0
ENV KHAN_POSTGRES_PORT 5432
ENV KHAN_POSTGRES_USER khan
ENV KHAN_POSTGRES_DBNAME khan

RUN goose -env nopassword up

CMD /go/bin/khan start --bind 0.0.0.0 --port 8080 --config /go/src/github.com/topfreegames/khan/config/default.yaml

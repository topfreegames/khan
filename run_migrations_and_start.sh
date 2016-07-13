#!/bin/sh

/go/bin/khan migrate -c /go/src/github.com/topfreegames/khan/config/default.yaml
/go/bin/khan start --bind 0.0.0.0 --port 8080 --config /go/src/github.com/topfreegames/khan/config/default.yaml

#!/bin/sh

/go/bin/khan migrate -c /go/src/github.com/topfreegames/khan/config/default.yaml
/bin/bash -c 'if [ "$KHAN_RUN_WORKER" != "true" ]; then /go/bin/khan start --bind 0.0.0.0 --port 8080 --config /go/src/github.com/topfreegames/khan/config/default.yaml; else /go/bin/khan worker --config /go/src/github.com/topfreegames/khan/config/default.yaml; fi'

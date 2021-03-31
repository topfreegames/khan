#!/bin/bash
if [ "$KHAN_RUN_WORKER" != "true" ]; then
    /go/bin/khan start --bind 0.0.0.0 --port 80 --fast --config /go/src/github.com/topfreegames/khan/config/default.yaml
else
    /go/bin/khan worker --config /go/src/github.com/topfreegames/khan/config/default.yaml
fi 

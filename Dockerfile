FROM golang:1.16.3-alpine as build

LABEL TFG Co <backend@tfgco.com>

WORKDIR /khan

COPY Makefile .
COPY go.mod go.sum .

RUN apk --update add make gcc && \
            make setup

COPY . .

RUN make build

FROM alpine:3.12

COPY --from=build /khan/bin/khan /
COPY --from=build /khan/config/default.yaml /

RUN chmod +x /khan

ENTRYPOINT [ "/khan", "-c", "/default.yaml" ]

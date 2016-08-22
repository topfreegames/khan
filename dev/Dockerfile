FROM kiasaki/alpine-postgres

ENV KHAN_BIN khan-linux
ENV KHAN_PORT 8080

EXPOSE $KHAN_PORT

RUN apk update
RUN apk add curl

# http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Get khan
ADD ./khan-linux-x86_64 /go/bin/$KHAN_BIN

ENV POSTGRES_DB khan
ENV POSTGRES_USER khan

ENV KHAN_POSTGRES_HOST 0.0.0.0
ENV KHAN_POSTGRES_PORT 5432
ENV KHAN_POSTGRES_USER khan
ENV KHAN_POSTGRES_DBNAME khan
ENV KHAN_SENTRY_URL ""

COPY default.yaml .
COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT /bin/sh -c "/docker-entrypoint.sh && su postgres -c '/usr/bin/pg_ctl start' && sleep 5 && /bin/$KHAN_BIN migrate --config default.yaml && /go/bin/$KHAN_BIN start --bind 0.0.0.0 --port $KHAN_PORT --config default.yaml"

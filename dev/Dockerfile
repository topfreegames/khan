FROM kiasaki/alpine-postgres

ENV KHAN_VERSION 0.7.1
ENV KHAN_BIN khan-linux
ENV KHAN_DOWNLOAD_URL https://github.com/topfreegames/khan/releases/download/${KHAN_VERSION}/khan-linux-x86_64
ENV KHAN_PORT 8080

EXPOSE $KHAN_PORT

RUN apk update
RUN apk add curl

# Get khan
RUN curl -o /bin/$KHAN_BIN -L -O $KHAN_DOWNLOAD_URL && chmod +x /bin/$KHAN_BIN

ENV POSTGRES_DB khan
ENV POSTGRES_USER khan

ENV KHAN_POSTGRES_HOST 0.0.0.0
ENV KHAN_POSTGRES_PORT 5432
ENV KHAN_POSTGRES_USER khan
ENV KHAN_POSTGRES_DBNAME khan

COPY default.yaml .
COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT /bin/sh -c "/docker-entrypoint.sh && su postgres -c '/usr/bin/pg_ctl start' && sleep 5 && /bin/$KHAN_BIN migrate --config default.yaml && /bin/$KHAN_BIN start --bind 0.0.0.0 --port $KHAN_PORT --config default.yaml"

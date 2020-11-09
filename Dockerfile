FROM golang:1.11-alpine3.7 as builder

RUN apk add --no-cache alpine-sdk

ARG DIR=${GOPATH}/src/github.com/deliveroo/pgbouncer-healthcheck
WORKDIR $DIR

RUN apk add --update pgbouncer bash shadow && \
    mkdir -p /var/run/postgresql && \
    chown pgbouncer /var/run/postgresql

ENV GO111MODULE=on

ADD build.sh $DIR/
ADD go.mod $DIR/
ADD go.sum $DIR/
RUN go mod download

ADD *.go $DIR/
ADD VERSION $DIR/

ADD tests/pgbouncer.ini /etc/pgbouncer/pgbouncer.ini
ADD tests/userlist.txt /etc/pgbouncer/userlist.txt
ADD tests/scripts /tests
RUN chmod 755 /tests/*

RUN $DIR/build.sh

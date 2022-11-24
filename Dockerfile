FROM golang:1.19.3-alpine as build

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOSUMDB=off

WORKDIR /go/build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN export REVISION=$TAG-$(date +'%Y%m%dT%H%M%S') && \
  go build -o finviz -ldflags "-X main.revision=$REVISION -s -w" ./cmd

FROM alpine:3.11 as release
RUN apk add --no-cache --update tzdata && \
  cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
  rm -rf /var/cache/apk/*
WORKDIR /app

ENV MIGRATE_VERSION=4.10.0
RUN wget https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz && \
  tar -xvzpf migrate.linux-amd64.tar.gz && \
  mv migrate.linux-amd64 migrate && \
  chmod 755 migrate && \
  rm -f *.tar.gz && \
  ./migrate -version

COPY migrations migrations
COPY --from=build /go/build/finviz ./

CMD ["/app/finviz"]

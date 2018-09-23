FROM alpine:3.5

MAINTAINER DevNet Cloudy Team

LABEL Description="DevNet webterm microservice image"

RUN apk update && \
    apk upgrade && \
    apk add \
        bash \
        ca-certificates \
    && rm -rf /var/cache/apk/*

RUN mkdir /config

COPY ./bin/webterm /usr/local/bin/webterm

WORKDIR /config

ENTRYPOINT ["/usr/local/bin/webterm", "serve"]
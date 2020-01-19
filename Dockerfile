# author https://github.com/wj596/gojob
FROM golang:1.12.3-alpine3.9

MAINTAINER wang596

RUN apk update \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

WORKDIR $GOPATH/src

COPY . $GOPATH/src/gojob

ENV GOPROXY https://goproxy.io
ENV GO111MODULE on

WORKDIR $GOPATH/src/gojob

RUN go build .

EXPOSE 8071

ENTRYPOINT  ["./gojob"]
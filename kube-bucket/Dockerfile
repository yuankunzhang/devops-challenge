FROM golang:1.10.2-alpine as build

ARG PKG_NAME=github.com/yuankunzhang/devops-challenge/kube-bucket/pkg

WORKDIR /go/src/${PKG_NAME}
COPY pkg/ .
RUN go install -v ./...


FROM alpine:3.7

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /go/bin/main /kube-bucket

ENTRYPOINT ["/kube-bucket"]

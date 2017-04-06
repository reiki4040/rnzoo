FROM golang:1.8

ENV GOPATH /go

RUN curl https://glide.sh/get | sh

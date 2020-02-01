FROM golang:1.13

ENV GOPATH /go

RUN curl https://glide.sh/get | sh

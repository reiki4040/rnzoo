FROM alpine:3.6

# for x509 error
RUN apk add --no-cache --update ca-certificates

RUN mkdir /rnzoo
RUN mkdir /rnzoo/bin
COPY rnzoo_linux_amd64 /rnzoo/bin/rnzoo
ENV PATH="/rnzoo/bin:$PATH"
ENV RNZOO_DIR="/rnzoo"




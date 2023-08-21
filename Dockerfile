# build stage
FROM golang:1.21 AS build-env

ADD . /go/src/app

RUN mkdir /app && \
    cd /go/src/app && \
    go mod download && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/vault_kubernetes_dotenv_manager main.go

# final stage
FROM scratch
MAINTAINER "Zsolt Szeberenyi <zsolt@szeberenyi.com>"

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /app/vault_kubernetes_dotenv_manager /vault_kubernetes_dotenv_manager
VOLUME /tmp

ENTRYPOINT ["/vault_kubernetes_dotenv_manager"]

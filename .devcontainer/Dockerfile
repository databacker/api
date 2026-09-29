# syntax=docker/dockerfile:1.6.0
FROM alpine:3.19

ARG TARGETOS=linux
ARG TARGETARCH
# requires OAPI_CODEGEN
ARG OAPI_CODEGEN_VERSION=v2.8.0

RUN apk add --no-cache bash curl git go make

RUN mkdir -p $(go env GOPATH)
RUN go install github.com/oapi-codegen/oapi-codegen@${OAPI_CODEGEN_VERSION}

# in case we need it in a shared path
#RUN go install github.com/oapi-codegen/oapi-codegen@${OAPI_CODEGEN_VERSION} && mv /root/go/bin/oapi-codegen /usr/local/bin/

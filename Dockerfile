FROM golang:1.24.4-bullseye AS builder
LABEL AUTHOR=uug.ai

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:/usr/local/lib:$PATH
ENV GOSUMDB=off

##############################################################################
# Copy all the relevant source code in the Docker image, so we can build this.

ARG github_username
ARG github_token
RUN git config --global \
    url."https://${github_username}:${github_token}@github.com/".insteadOf \
    "https://github.com/"

##########################################
# Installing some additional dependencies.

RUN apt-get update && apt-get install -y --no-install-recommends \
    git build-essential cmake pkg-config unzip libgtk2.0-dev \
    curl ca-certificates libcurl4-openssl-dev libssl-dev libjpeg62-turbo-dev && \
    rm -rf /var/lib/apt/lists/*

##############################################################################
# Copy all the relevant source code in the Docker image, so we can build this.


RUN mkdir -p /go/src/github.com/uug-ai/models
COPY . /go/src/github.com/uug-ai/models

##################
# Build API

RUN cd /go/src/github.com/uug-ai/models && \
    go mod download && \
    go build -tags timetzdata,netgo --ldflags '-s -w -extldflags "-static -latomic"' cmd/main.go && \
    rm -rf /go/src/github.com

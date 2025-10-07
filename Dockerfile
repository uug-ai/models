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
    mkdir -p /models && mv main /models && \
    rm -rf /go/src/github.com

####################################
# Let's create a /dist folder containing just the files necessary for runtime.
# Later, it will be copied as the / (root) of the output image.

WORKDIR /dist
RUN cp -r /models ./
RUN /dist/models/main

FROM alpine:latest

#################################
# Copy files from previous images

COPY --chown=0:0 --from=builder /dist /

############################
# Move directory to /var/lib

RUN apk update && apk add ca-certificates curl libstdc++ libc6-compat --no-cache && rm -rf /var/cache/apk/*

##################
# Try running models

RUN mkdir -p /home/models
RUN mv /models/* /home/models
RUN /home/models/main

###########################
# Grant the necessary root capabilities to the process trying to bind to the privileged port
RUN apk add libcap && setcap 'cap_net_bind_service=+ep' /home/models/main

###################
# Run non-root user

USER models

###################################################
# Leeeeettttt'ssss goooooo!!!
# Run the shizzle from the right working directory.

WORKDIR /home/models
CMD ["./main"]

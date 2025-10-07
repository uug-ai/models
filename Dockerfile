FROM golang:1.24.4-bullseye AS builder
LABEL AUTHOR=uug.ai

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:/usr/local/lib:$PATH
ENV GOSUMDB=off

##############################################################################
# Use BuildKit secret for the token so it is not baked into any image layer.
# Provide github_username as a build-arg; do NOT define github_token as ARG.
ARG github_username

# Configure git to use the secret at build time (requires BuildKit).
# The secret is read from /run/secrets/github_token and never expanded into
# the Dockerfile's RUN line or persisted to the final image.
RUN --mount=type=secret,id=github_token \
    git config --global \
      url."https://${github_username}:$(cat /run/secrets/github_token)@github.com/".insteadOf \
      "https://github.com/"

##########################################
# Installing additional dependencies.
RUN apt-get update && apt-get install -y --no-install-recommends \
    git build-essential cmake pkg-config unzip libgtk2.0-dev \
    curl ca-certificates libcurl4-openssl-dev libssl-dev libjpeg62-turbo-dev && \
    rm -rf /var/lib/apt/lists/*

##############################################################################
# Copy source and build
RUN mkdir -p /go/src/github.com/uug-ai/models
COPY . /go/src/github.com/uug-ai/models

RUN cd /go/src/github.com/uug-ai/models && \
    go mod download && \
    mkdir -p /out && \
    go build -tags timetzdata,netgo --ldflags '-s -w -extldflags "-static -latomic"' -o /out/main cmd/main.go && \
    rm -rf /go/src/github.com

##############################################################################
# Final minimal runtime image
FROM alpine:latest

# Create non-root user
RUN addgroup -S models && adduser -S app -G models

# Install runtime deps
RUN apk add --no-cache ca-certificates libstdc++ libc6-compat

# Copy built binary from builder stage
COPY --from=builder /out/main /home/app/main
RUN chown app:models /home/app/main && chmod +x /home/app/main

USER app
WORKDIR /home/app

EXPOSE 8081
HEALTHCHECK CMD curl --fail http://localhost:8081 || exit 1

CMD ["./main", "serve"]
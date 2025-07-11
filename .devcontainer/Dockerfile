FROM mcr.microsoft.com/devcontainers/go:1.4.3-1.24-bookworm

USER vscode

# Fix Go module cache permissions
RUN sudo chown -R vscode:golang /go/pkg/mod && sudo chown -R vscode:golang /go/pkg/sumdb

# Install go swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Install goplantuml parser
RUN go install github.com/jfeliu007/goplantuml/cmd/goplantuml@latest

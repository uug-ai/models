FROM mcr.microsoft.com/devcontainers/go:1.24-bookworm

# Remove yarn repo with expired GPG key (not needed for this devcontainer)
RUN rm -f /etc/apt/sources.list.d/yarn.list

# Add nodejs and npm
RUN apt-get update && \
    apt-get install -y nodejs npm && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
    
USER vscode

# Install go swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Install goplantuml parser
RUN go install github.com/jfeliu007/goplantuml/cmd/goplantuml@latest

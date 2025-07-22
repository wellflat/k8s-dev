#!/bin/bash

# official site: https://github.com/argoproj/argo-workflows/releases/

# Detect OS
ARGO_OS="linux"

# Download the binary
curl -sLO "https://github.com/argoproj/argo-workflows/releases/download/v3.6.10/argo-$ARGO_OS-amd64.gz"

# Unzip
gunzip "argo-$ARGO_OS-amd64.gz"

# Make binary executable
chmod +x "argo-$ARGO_OS-amd64"

# Move binary to path
mv "./argo-$ARGO_OS-amd64" $HOME/.local/bin/argo

# Test installation
argo version

# Controller and Server
NS=bench
kubectl apply -n $NS -f https://github.com/argoproj/argo-workflows/releases/download/v3.6.10/install.yaml
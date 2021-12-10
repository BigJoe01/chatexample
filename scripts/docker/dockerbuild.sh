#!/bin/sh

# Example docker build
# Need set git password and user if you want build image
docker build --build-arg GIT_PASSWORD="xxx" --build-arg GIT_USER="xxx" -f Dockerfile --tag chatserver:latest .

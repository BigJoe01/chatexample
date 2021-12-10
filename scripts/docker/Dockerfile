# Dockerfile for chat service
# Project name : chat

# Source image
FROM golang:alpine3.14 AS gobuilder

MAINTAINER Example "xxx"

# Add git access
ARG GIT_PASSWORD
ARG GIT_USER

# Source solution settings
# ----------------------------------------------------------------------------------------------------------------------
ENV SOLUTION="https://$GIT_USER:$GIT_PASSWORD@https://github.com/BigJoe01/chatexample.git"
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# Install system dependencies
# ----------------------------------------------------------------------------------------------------------------------
RUN apk update && apk upgrade
RUN apk add bash && apk add --no-cache git && apk add make
RUN apk --no-cache add ca-certificates

# Solution dependencies update
# ----------------------------------------------------------------------------------------------------------------------
RUN mkdir /home/solution
WORKDIR /home/solution/
RUN git clone --branch master $SOLUTION
WORKDIR /home/solution/chatexample
RUN mkdir bin

# Solution build
# ----------------------------------------------------------------------------------------------------------------------
RUN go mod download
RUN mkdir bin
RUN go build cmd/server/main.go -o bin/chatserver -ldflags "-s -w"

# Build small image
# ----------------------------------------------------------------------------------------------------------------------
FROM alpine:3.14.2
RUN apk --no-cache add ca-certificates
WORKDIR /home
COPY --from=gobuilder /home/solution/chatservice/bin/chatserver /home/chatserver


# Docker entrypoint ssettings
# ----------------------------------------------------------------------------------------------------------------------
EXPOSE 8080
ENTRYPOINT "/bin/sh /home/chatserver"
CMD ""




FROM golang:1.14-alpine

# Install git
RUN apk add --no-cache git mercurial

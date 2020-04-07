FROM golang:1.14-alpine AS build
ARG PRODUCTION
VOLUME [ "/bot" ]

# Install git
RUN apk add --no-cache git mercurial

RUN [[ "${PRODUCTION}" != "1" ]] && \ 
    export GO111MODULE=on && go get -u -v github.com/cortesi/modd/cmd/modd

WORKDIR /bot
COPY . .
RUN go get -v && go build

CMD [ "/bot/discord-bot" ]

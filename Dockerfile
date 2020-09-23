FROM golang:1.14-alpine AS dev

WORKDIR /bot

RUN apk add git

RUN GO111MODULE=on go get github.com/cortesi/modd/cmd/modd

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN git rev-parse --short HEAD > /version

RUN go install github.com/UCCNetsoc/discord-bot

CMD [ "go", "run", "*.go" ]

FROM alpine

WORKDIR /bin

COPY --from=dev /go/bin/discord-bot ./discord-bot

EXPOSE 2112

COPY --from=dev /version /version

CMD ["sh", "-c", "export BOT_VERSION=$(cat /version) && discord-bot -p"]

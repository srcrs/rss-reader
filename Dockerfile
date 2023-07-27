FROM golang:1.20.4-alpine3.18 AS builder

COPY . /src
WORKDIR /src

RUN go build -ldflags "-s -w" -o ./bin/ .

FROM alpine

COPY --from=builder /src/bin /app
COPY --from=builder /src/index.html /app/index.html

WORKDIR /app

EXPOSE 8080

ENTRYPOINT ["./ownrss"]

# syntax=docker/dockerfile:1

FROM alpine

WORKDIR /app

COPY bin/ ./

ENTRYPOINT ["./deordie-bot"]

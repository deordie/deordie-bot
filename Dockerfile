# syntax=docker/dockerfile:1

FROM alpine

WORKDIR /app

COPY bin/ ./

EXPOSE 8080
ENTRYPOINT ["./deordie-bot"]

# syntax=docker/dockerfile:1

FROM golang:1.21.4 as builder
WORKDIR /app

COPY src/websocket-server/go.mod src/websocket-server/go.sum ./
RUN go mod download
COPY src/websocket-server/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /undercover-server

# --
FROM alpine
ENV SEQ_URL http://localhost:5341
ENV SEQ_APIKEY CHANGEME
WORKDIR /app
COPY --from=builder /undercover-server /undercover-server
COPY config-docker.yml /app/config/config.yml
COPY src/websocket-server/data/list-words.csv /app/data/list-words.csv

COPY docker-entrypoint.sh /usr/bin/
RUN ["chmod", "+x", "/usr/bin/docker-entrypoint.sh"]
ENTRYPOINT ["docker-entrypoint.sh"]

EXPOSE 8080
CMD ["/undercover-server"]
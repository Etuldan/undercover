# syntax=docker/dockerfile:1

FROM golang:1.21.4 as builder
WORKDIR /app

COPY src/websocket-server/go.mod src/websocket-server/go.sum ./
RUN go mod download
COPY src/websocket-server/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /undercover-server

# --
FROM alpine
WORKDIR /app
COPY --from=builder /undercover-server /undercover-server

COPY config-docker.yml /app/config.yml
ENV SEQ_URL http://localhost:5341
RUN sed -i -e "s|SEQ_URL|${SEQ_URL}|g" /app/config.yml
ENV SEQ_APIKEY CHANGEME
RUN sed -i -e "s|SEQ_APIKEY|${SEQ_APIKEY}|g" /app/config.yml


COPY src/websocket-server/data/list-words.csv /app/data/list-words.csv

EXPOSE 8080
CMD ["/undercover-server"]
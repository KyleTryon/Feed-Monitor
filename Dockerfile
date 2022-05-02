# syntax=docker/dockerfile:1

FROM golang:1.18-alpine
RUN apk add build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Build the main binary
COPY *.go ./
RUN go build -o /feed-monitor

# Build and copy the  default plugins
COPY plugins/gotify/ plugins/gotify/
RUN cd plugins/gotify/
RUN go build -buildmode=plugin

CMD [ "/feed-monitor -config=/config/config.yml " ]
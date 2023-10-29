FROM golang:1.19 AS build

WORKDIR /

COPY go.mod go.sum /
RUN go mod download

COPY resources /resources
COPY *.go /

RUN CGO_ENABLED=0 GOOS=linux go build -o /makaroni .


FROM ubuntu:22.04

RUN apt update && \
    apt upgrade -y && \
    apt install -y ca-certificates

COPY --from=build /makaroni /makaroni

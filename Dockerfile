FROM golang:1.22-alpine3.20

WORKDIR /app

COPY . .
RUN apk add bash make musl-dev 
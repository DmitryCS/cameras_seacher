# Build stage
FROM golang:alpine AS build-env

COPY . /app
WORKDIR /app
RUN go mod download

RUN apk update && \
    apk upgrade && \
    apk add nmap nmap-nselibs nmap-scripts \
    curl curl-dev \
    gcc \
    libc-dev \
    git \
    pkgconfig
ENV GO111MODULE=on
RUN go version
RUN go build -o main .


# Final stage
FROM alpine


LABEL maintainer="CUBICAI"

ENV DEBIAN_FRONTEND=noninteractive

RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/main' >> /etc/apk/repositories

RUN apk --update add --no-cache nmap \
    nmap-nselibs \
    nmap-scripts \
    curl-dev==7.64.0-r5

WORKDIR /app
COPY --from=build-env /app /app

ENV CAMERADAR_CUSTOM_ROUTES="/app/dictionaries/routes"
ENV CAMERADAR_CUSTOM_CREDENTIALS="/app/dictionaries/credentials.json"

ENTRYPOINT ["/app/main"]
FROM golang:1.26.5-alpine AS base_image
LABEL stage=gobuilder

ENV CGO_ENABLED=0 \
    GOOS=linux

RUN apk update && apk upgrade && apk add --no-cache ca-certificates wget
RUN update-ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN go build -ldflags="-w -s" -o application ./cmd/api

# ---------------------------------------------------------------------------- #

FROM alpine:3.21.3 AS final_image
LABEL stage=minimalbuilder

WORKDIR /app

COPY --from=base_image /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base_image /app/application /app/

RUN adduser -D -s /sbin/nologin shortener-user
RUN passwd -l root
RUN chown -R shortener-user:shortener-user /app
RUN chmod -R 500 /app

USER shortener-user

ENV PORT=8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider --quiet http://localhost:${PORT}/health || exit 1

ENTRYPOINT ["/app/application"]

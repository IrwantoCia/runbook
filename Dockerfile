FROM node:22-alpine AS web

WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.27-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=web /app/web/dist/ ./cmd/runbook/web/dist/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/runbook ./cmd/runbook

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 runbook \
    && adduser -S -D -H -u 10001 -G runbook runbook \
    && mkdir -p /data \
    && chown runbook:runbook /data

COPY --from=build /out/runbook /usr/local/bin/runbook

USER runbook
WORKDIR /data
ENV ADDR=0.0.0.0 APP_PORT=8080 DB_PATH=/data/runbook.db
EXPOSE 8080
ENTRYPOINT ["runbook"]

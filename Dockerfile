FROM caddy:builder-alpine as builder

WORKDIR /app
COPY main.go go.mod go.sum ./

RUN xcaddy build --with github.com/integer-technologies-b-v/shield=./

FROM alpine:latest
COPY --from=builder /app/caddy ./caddy
FROM node:21 as tailwind_build
WORKDIR /usr/src/app

COPY . .
RUN npx tailwindcss -i ./styles/input.css -o ./public/output.css --minify

FROM alpine:latest AS db_build
WORKDIR /usr/src/app

RUN apk --update-cache add sqlite \
    && rm -rf /var/cache/apk/*

COPY db/schema.sql db/schema.sql
COPY scripts/create-db.sh create-db.sh

RUN ./create-db.sh

FROM golang:1.21.6 AS go_build
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o dumbbell .

FROM debian:bookworm-slim as runtime
WORKDIR /usr/src/app

COPY --from=tailwind_build /usr/src/app/public public
COPY --from=go_build /usr/src/app/dumbbell dumbbell
COPY --from=db_build /usr/src/app/db db
COPY templates templates

ENTRYPOINT [ "./dumbbell" ]
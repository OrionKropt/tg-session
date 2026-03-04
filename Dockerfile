# -------- Stage 1: Build --------
FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN apk add --no-cache make

RUN make test

RUN make build

# -------- Stage 2: Production image --------

FROM alpine:latest

COPY --from=builder /app/bin/tg-session .

EXPOSE 50051

ENTRYPOINT ["./tg-session"]


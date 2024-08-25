FROM golang:1.23.0-alpine3.19 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build

FROM scratch
COPY --from=builder /app/coffee-maker /app
ENTRYPOINT ["/app"]

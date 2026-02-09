FROM golang:1.25-alpine AS builder
WORKDIR /reindexer-service
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /reindexer-service/app ./cmd/api/main.go

FROM alpine:3.20
WORKDIR /reindexer-service
COPY --from=builder /reindexer-service/app .
EXPOSE 8080
ENTRYPOINT ["./app"]
# Use the official Golang image to create a build artifact.
FROM golang:1.19 as builder
LABEL authors="pratikpandey"

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

# Build the executable.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o isolation

FROM alpine:3
RUN apk add --no-cache ca-certificates
RUN apk update && apk add bash

WORKDIR /app

COPY --from=builder /app/isolation /app/isolation
COPY --from=builder /app/wait-for-it.sh /app/wait-for-it.sh

RUN chmod +x /app/wait-for-it.sh

CMD ["/app/wait-for-it.sh", "mysql:3306", "--", "/app/isolation"]

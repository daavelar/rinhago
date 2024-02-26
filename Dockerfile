FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rinha .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/rinha .
EXPOSE 8000
CMD ["./rinha"]

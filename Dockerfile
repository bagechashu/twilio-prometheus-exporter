FROM golang:1.24 AS builder
WORKDIR /app/
COPY ./src/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o twilio_prometheus_exporter

FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/twilio_prometheus_exporter .
EXPOSE 8080

CMD ["/app/twilio_prometheus_exporter"]

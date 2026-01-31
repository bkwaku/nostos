FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ingress ./...


FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/ingress .
RUN adduser -D -u 10001 ingressuser && \
    chown -R ingressuser:ingressuser /app
EXPOSE 8080
USER ingressuser
EXPOSE 8080
ENTRYPOINT ["/app/ingress"]

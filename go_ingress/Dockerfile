FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

# I probally don't need since I won't deploy this
# but just because this is open source let's keep this
FROM builder AS test
RUN go test -v -cover ./...

FROM test AS build
RUN CGO_ENABLED=0 go build -o ingress .


FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/ingress .
RUN adduser -D -u 10001 ingressuser && \
    chown -R ingressuser:ingressuser /app
USER ingressuser
EXPOSE 8080
ENTRYPOINT ["/app/ingress"]

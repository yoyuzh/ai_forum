FROM golang:1.24-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o /out/api-server ./cmd/api \
 && go build -o /out/worker-service ./cmd/worker \
 && go build -o /out/outbox-publisher ./cmd/outbox-publisher

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/api-server /out/worker-service /out/outbox-publisher ./
COPY config ./config
COPY internal/rbac/model.conf ./internal/rbac/model.conf
CMD ["./api-server"]

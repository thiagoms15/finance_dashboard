FROM golang:1.25-bookworm AS build
WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot AS api
WORKDIR /app
COPY --from=build /out/api /app/bin/api
COPY --from=build /app/internal/db/migrations /app/internal/db/migrations
COPY --from=build /app/api /app/api
ENTRYPOINT ["/app/bin/api"]

FROM gcr.io/distroless/static-debian12:nonroot AS worker
WORKDIR /app
COPY --from=build /out/worker /app/bin/worker
COPY --from=build /app/internal/db/migrations /app/internal/db/migrations
ENTRYPOINT ["/app/bin/worker"]

FROM gcr.io/distroless/static-debian12:nonroot AS migrate
WORKDIR /app
COPY --from=build /out/migrate /app/bin/migrate
COPY --from=build /app/internal/db/migrations /app/internal/db/migrations
ENTRYPOINT ["/app/bin/migrate"]

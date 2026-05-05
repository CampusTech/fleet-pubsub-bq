FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /fleet-pubsub-bq .

FROM gcr.io/distroless/static-debian12
COPY --from=build /fleet-pubsub-bq /fleet-pubsub-bq
ENTRYPOINT ["/fleet-pubsub-bq"]

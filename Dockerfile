# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS base
WORKDIR /src
COPY go.mod ./
COPY main.go ./
COPY main_test.go ./

FROM base AS test
RUN go test ./...

FROM base AS build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/hello-deploy .

FROM alpine:3.20 AS final
WORKDIR /app
RUN adduser -D -u 10001 appuser
COPY --from=build /out/hello-deploy /app/hello-deploy
RUN mkdir -p /data && chown -R appuser:appuser /app /data
USER appuser
ENV PORT=8080
ENV DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
CMD ["/app/hello-deploy"]

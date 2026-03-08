# hello-deploy

A tiny test app for exercising a deployment platform through all phases.

## What it does

- serves `/`
- serves `/health`
- reads `APP_MESSAGE`
- persists a visit counter in `/data/visits.txt`
- increments the counter on `POST /visit`

## Env vars

- `PORT` default `8080`
- `APP_MESSAGE` default `hello from deploy`
- `DATA_DIR` default `/data`

## Local run

```bash
go test ./...
go run .
```

## Docker test target

```bash
docker build --target test .
```

## Docker final target

```bash
docker build --target final -t hello-deploy:local .
```

## Run container

```bash
docker run --rm -p 8080:8080 \
  -e APP_MESSAGE="hello from docker" \
  -v hello_deploy_data:/data \
  hello-deploy:local
```

## Try it

```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/visit
curl http://localhost:8080/
```

# hello-deploy

A tiny test app for exercising a deployment platform through all phases.

## What it does

- serves `/`
- serves `/health`
- reads `APP_MESSAGE`
- persists a visit counter in `/data/visits.txt`
- increments the counter on `POST /visit`

## Env vars

- `HOST` default `0.0.0.0`
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
docker run --rm -p 8001:8080 \
  -e APP_MESSAGE="hello from docker" \
  -v hello_deploy_data:/data \
  hello-deploy:local
```

The app listens on `0.0.0.0:8080` inside the container. With `-p 8001:8080`,
send requests to `localhost:8001` on the host.

## Try it

```bash
curl http://localhost:8001/health
curl -X POST http://localhost:8001/visit
curl http://localhost:8001/
```

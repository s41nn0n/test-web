# Header dump test app

Small Go web app that accepts **any path** and **any HTTP method** and returns the request details in one of several formats.

It shows:

- method
- path
- query string
- host
- remote address
- protocol
- every request header received by the app

This is useful for verifying what your backend actually sees when running behind an auth proxy, ingress, or load balancer.

## Output formats

The app supports:

- HTML (default)
- JSON
- XML
- plain text

You can choose the response format in two ways:

1. Query parameter: `?format=html|json|xml|text`
2. `Accept` header:
   - `application/json`
   - `application/xml` or `text/xml`
   - `text/plain`

If both are present, the `format` query parameter wins.

## Run locally

```bash
go run .
```

By default it listens on `:8080`.

Custom port:

```bash
PORT=9090 go run .
```

## Run with Docker

Build locally:

```bash
docker build -t headerdump .
```

Run locally:

```bash
docker run --rm -p 8080:8080 headerdump
```

Custom port inside container:

```bash
docker run --rm -e PORT=9090 -p 9090:9090 headerdump
```

## Run with Docker Compose

The repository includes `docker-compose.yml` configured to use the GitHub Container Registry image:

```bash
docker compose up -d
```

That compose file pulls:

```text
ghcr.io/s41nn0n/test-web:latest
```

## GitHub Actions container build

A workflow was added at:

- `.github/workflows/docker.yml`

It will:

- build the container on pull requests
- build and push to GHCR on pushes to `main`
- publish the `latest` tag from the default branch
- publish branch, tag, and sha tags

If the package is private, make sure the machine running `docker compose` can authenticate to GHCR. If you want anonymous pulls, set the package visibility to public in GitHub.

## Test examples

### HTML

```bash
curl -i http://localhost:8080/test/path \
  -H 'X-Test: hello' \
  -H 'X-Auth-User: alice'
```

### JSON by query parameter

```bash
curl -i 'http://localhost:8080/test/path?format=json' \
  -H 'X-Test: hello' \
  -H 'X-Auth-User: alice'
```

### JSON by Accept header

```bash
curl -i http://localhost:8080/test/path \
  -H 'Accept: application/json' \
  -H 'Authorization: Bearer example'
```

### XML

```bash
curl -i http://localhost:8080/test/path \
  -H 'Accept: application/xml' \
  -H 'X-Forwarded-User: bob'
```

### Plain text

```bash
curl -i http://localhost:8080/test/path \
  -H 'Accept: text/plain' \
  -H 'X-Forwarded-Email: bob@example.com'
```

## Example proxy test

Once your auth proxy is in front of this app, send a request through the proxy and confirm headers such as these arrive:

- `X-Auth-User`
- `X-Auth-Email`
- `X-Forwarded-User`
- `X-Forwarded-Email`
- `Authorization`
- any custom headers your proxy injects

# REST API Fuzzer

Simple, deterministic REST API fuzzing from an OpenAPI spec.

`rest-api-fuzzer` is a small local-first CLI that reads an OpenAPI 3.x document, generates valid requests from the schema, mutates values around edge cases, and reports the API behavior that usually matters first:

- `5xx` server errors
- undocumented or schema-invalid JSON responses
- slow endpoints
- network/request errors

It is built for people who want a fast bug-finding tool they can run in CI, against a local dev server, or before shipping a public API.

## Why This Exists

Most API fuzzers are either large platforms, cloud services, or heavy research tools. This project keeps the useful core:

1. Load an OpenAPI spec.
2. Generate requests that are valid enough to reach real handlers.
3. Mutate values in deterministic, property-testing-inspired ways.
4. Report reproducible failures with the seed and exact request.

Same seed, same requests. That makes failures easy to replay and fix.

## Features

- OpenAPI 3.x parsing with `kin-openapi`
- Fast HTTP execution with `fasthttp`
- Deterministic seed-based request generation
- Path, query, header, and JSON body generation
- Boundary values for integers and numbers
- Format-aware strings for `email`, `uuid`, `date`, and `date-time`
- Mutation strings for empty values, long values, traversal, SQL-ish payloads, and script payloads
- Response validation against documented JSON schemas
- Text and JSON reports
- Non-zero exit on server errors, schema violations, and request errors
- Small codebase that is easy to audit and extend

## Install

```bash
go install github.com/kashyapkumbhani/rest-api-fuzzer/cmd/rest-api-fuzzer@latest
```

Or build from source:

```bash
git clone https://github.com/kashyapkumbhani/rest-api-fuzzer.git
cd rest-api-fuzzer
go build -o dist/rest-api-fuzzer ./cmd/rest-api-fuzzer
```

## Quick Start

Run against an API using the server URL from the OpenAPI document:

```bash
rest-api-fuzzer -spec ./openapi.yaml
```

Override the target server:

```bash
rest-api-fuzzer \
  -spec ./openapi.yaml \
  -base-url http://127.0.0.1:8080 \
  -cases 25 \
  -seed 2026
```

Add static headers:

```bash
rest-api-fuzzer \
  -spec ./openapi.yaml \
  -header "Authorization: Bearer $TOKEN" \
  -header "X-Tenant: demo"
```

Emit JSON for CI or later analysis:

```bash
rest-api-fuzzer -spec ./openapi.yaml -format json > fuzz-report.json
```

## Local Demo

This repository includes a tiny intentionally-buggy API so you can see findings immediately.

Terminal 1:

```bash
go run ./examples/demo-api
```

Terminal 2:

```bash
go run ./cmd/rest-api-fuzzer \
  -spec examples/openapi.yaml \
  -base-url http://127.0.0.1:8080 \
  -cases 8 \
  -seed 2026
```

The demo API intentionally returns a `500` for one mutated product name and returns a schema-invalid `price` from `GET /products/{id}`.

## Example Output

```text
REST API Fuzzer report
Seed: 2026
Operations: 12
Requests: 240
Duration: 3.184s
Findings: 2

[server_error] POST /products -> HTTP 500 in 18ms
  endpoint returned a 5xx response
  http://127.0.0.1:8080/products
  body: {"name":"' OR '1'='1","price":0.25}

[schema_violation] GET /products/{id} -> HTTP 200 in 7ms
  value must be a string
  http://127.0.0.1:8080/products/00000000-0000-4000-8000-000000000001
```

## CLI Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-spec` | required | Path or URL to an OpenAPI 3.x document |
| `-base-url` | first OpenAPI server | Override target server |
| `-seed` | `1337` | Deterministic generation seed |
| `-cases` | `20` | Requests per operation |
| `-timeout` | `5s` | Per-request timeout |
| `-slow` | `750ms` | Report responses slower than this |
| `-header` | none | Static header, repeatable |
| `-format` | `text` | `text` or `json` |

## What Gets Generated

For each operation, the fuzzer builds requests from the OpenAPI contract:

- Path parameters are expanded from their schemas.
- Query and header parameters are generated from their schemas.
- JSON request bodies are generated from object, array, scalar, enum, default, and example schemas.
- Numbers include boundary values when `minimum` or `maximum` are present.
- Strings rotate through normal values and mutation values.

The goal is not to send garbage. The goal is to send plausible requests that still pressure the implementation.

## Finding Types

| Kind | Meaning | Fails the process |
| --- | --- | --- |
| `server_error` | Endpoint returned HTTP `5xx` | Yes |
| `schema_violation` | JSON response did not match the documented schema, or status was undocumented | Yes |
| `request_error` | Timeout, refused connection, invalid target, or transport failure | Yes |
| `slow_response` | Response exceeded `-slow` | No |
| `generation_error` | Request could not be generated from the operation | No |

## Development

```bash
go test ./...
go build ./cmd/rest-api-fuzzer
```

With `make`:

```bash
make test
make build
```

The example spec lives at [examples/openapi.yaml](examples/openapi.yaml).

More command examples are in [docs/USAGE.md](docs/USAGE.md).

## Design

The codebase is intentionally small:

- `cmd/rest-api-fuzzer` handles flags and output.
- `internal/openapi` loads and validates specs.
- `internal/fuzzer` plans operations, generates requests, mutates values, sends HTTP, and validates responses.
- `internal/report` formats findings for humans and machines.

This makes the project easy to inspect before running against private APIs.

## Roadmap

- OpenAPI security scheme helpers
- Replay command for one finding
- JUnit/SARIF output for CI dashboards
- Smarter schema-aware string generation for regex patterns
- Stateful operation chaining from response IDs
- Optional rate limits and endpoint allow/deny filters

## License

MIT

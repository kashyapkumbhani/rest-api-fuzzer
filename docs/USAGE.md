# Usage Guide

## Basic Run

```bash
rest-api-fuzzer -spec openapi.yaml -base-url http://127.0.0.1:8080
```

The fuzzer exits with:

- `0` when no failing findings are detected
- `1` when server errors, schema violations, or request errors are detected
- `2` for invalid CLI usage

Slow responses are reported but do not fail the process by default.

## Deterministic Replay

Every report includes the seed. Reuse the same seed, OpenAPI spec, base URL, headers, and case count to regenerate the same request sequence:

```bash
rest-api-fuzzer -spec openapi.yaml -base-url http://127.0.0.1:8080 -seed 2026 -cases 2
```

`-cases` is applied per operation per built-in fuzzer. With 10 operations, 30 built-in fuzzers, and `-cases 2`, the run sends 600 requests.

## Fuzzer Catalog

```bash
rest-api-fuzzer -list-fuzzers
```

The current catalog includes 30 built-in fuzzers: valid baseline, numeric boundaries, zero values, negative and large numbers, decimal precision, empty/long/unicode strings, SQL probes, XSS probes, path traversal probes, nullish strings, forced booleans, enum edge values, required-only bodies, extra object fields, array size probes, duplicate query parameters, encoded slashes, missing content type, malformed JSON, empty JSON bodies, large JSON bodies, and header case mutation.

## CI JSON Report

```bash
rest-api-fuzzer -spec openapi.yaml -format json > fuzz-report.json
```

The JSON report includes:

- `seed`
- `started_at`
- `duration`
- `operations`
- `requests`
- `findings`

Each finding includes the operation, status code, duration, exact request URL, headers, body, and a compact response body.
Each finding also includes the `fuzzer` ID that produced it.

## Authenticated APIs

Use repeatable `-header` flags for tokens, tenant IDs, or environment-specific headers:

```bash
rest-api-fuzzer \
  -spec openapi.yaml \
  -base-url https://api.example.com \
  -header "Authorization: Bearer $TOKEN" \
  -header "X-Environment: staging"
```

## Practical Tips

- Start with `-cases 1` for smoke testing.
- Use `-cases 5` or higher before a release.
- Keep `-seed` fixed in CI so failures are reproducible.
- Run against staging or local environments, not production, unless your API and data model are designed for fuzz traffic.

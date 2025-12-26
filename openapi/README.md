# Synapse OpenAPI Specification

This directory contains the **OpenAPI 3.1** specification for the Synapse REST API.

## Structure

```
openapi/
├── openapi.yaml                    # Root specification document
├── paths/
│   ├── _index.yaml                 # Path index
│   ├── orders.yaml                 # Order endpoints
│   ├── pipeline.yaml               # Pipeline management endpoints
│   └── health.yaml                 # Health & observability endpoints
└── components/
    ├── _index.yaml                 # Components index
    ├── parameters.yaml             # Reusable parameters
    ├── headers.yaml                # Reusable response headers
    ├── responses.yaml              # Reusable error responses
    ├── schemas/
    │   ├── _index.yaml             # Schema index
    │   ├── orders.yaml             # Order schemas
    │   ├── pipeline.yaml           # Pipeline schemas
    │   ├── health.yaml             # Health check schemas
    │   └── errors.yaml             # RFC 9457 Problem Details
    └── examples/
        ├── _index.yaml             # Examples index
        └── orders.yaml             # Order request/response examples
```

## RFC Compliance

This specification adheres to the following RFCs:

| RFC | Title | Usage |
|-----|-------|-------|
| **RFC 9110** | HTTP Semantics | Status codes, methods, headers |
| **RFC 9457** | Problem Details for HTTP APIs | Error response format |
| **RFC 8288** | Web Linking | Pagination via Link headers |
| **RFC 7232** | Conditional Requests | ETag, If-Match, If-None-Match |
| **RFC 6585** | Additional HTTP Status Codes | 429 Too Many Requests |
| **RFC 6750** | Bearer Token Usage | Authorization header format |
| **RFC 3339** | Date and Time on the Internet | Timestamp format |
| **RFC 3986** | URI Syntax | Resource identifiers |

### Rate Limiting

Rate limit headers follow the IETF draft `draft-ietf-httpapi-ratelimit-headers`:
- `RateLimit-Limit`
- `RateLimit-Remaining`
- `RateLimit-Reset`

### Idempotency

The `Idempotency-Key` header follows IETF draft `draft-ietf-httpapi-idempotency-key-header`.

## Endpoints

### Orders

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/orders` | Ingest a new order |
| GET | `/api/v1/orders` | List orders (paginated) |
| GET | `/api/v1/orders/{orderId}` | Get order details |
| DELETE | `/api/v1/orders/{orderId}` | Cancel an order |
| GET | `/api/v1/orders/{orderId}/events` | Get order event history |

### Pipeline

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/pipeline/stages` | List all pipeline stages |
| GET | `/api/v1/pipeline/stages/{stageId}` | Get stage details |
| PATCH | `/api/v1/pipeline/stages/{stageId}` | Update stage config |
| GET | `/api/v1/pipeline/dlq` | List dead letter queue |
| POST | `/api/v1/pipeline/dlq/{eventId}/retry` | Retry a DLQ item |

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Full health check |
| GET | `/health/live` | Kubernetes liveness probe |
| GET | `/health/ready` | Kubernetes readiness probe |
| GET | `/metrics` | Prometheus metrics |

## Validation

```bash
# Install OpenAPI CLI tools
npm install -g @redocly/cli

# Validate the specification
redocly lint openapi/openapi.yaml

# Bundle into single file
redocly bundle openapi/openapi.yaml -o dist/openapi.bundled.yaml

# Generate documentation
redocly build-docs openapi/openapi.yaml -o docs/api.html
```

## Code Generation

```bash
# Go server stub (oapi-codegen)
oapi-codegen -package api -generate types,server openapi/openapi.yaml > internal/api/openapi.gen.go

# Go client
oapi-codegen -package client -generate types,client openapi/openapi.yaml > pkg/client/client.gen.go
```

## Relationship to AsyncAPI

This OpenAPI spec defines the **synchronous HTTP API** for Synapse.
The companion **AsyncAPI spec** (`../asyncapi/`) defines the event-driven
NATS messaging layer.

```
┌─────────────────────────────────────────────────────────────┐
│                         Synapse                              │
│                                                              │
│   ┌─────────────┐              ┌─────────────────────────┐  │
│   │  OpenAPI    │   publishes  │       AsyncAPI          │  │
│   │  REST API   │─────────────▶│    NATS PubSub          │  │
│   │  (HTTP)     │              │    (Events)             │  │
│   └─────────────┘              └─────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

The REST API is the entry point; it publishes to NATS channels defined in AsyncAPI.

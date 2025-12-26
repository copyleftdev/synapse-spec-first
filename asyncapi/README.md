# Synapse AsyncAPI Specifications

This directory contains the **AsyncAPI 3.0** specifications for the Synapse event pipeline.

## Structure

```
asyncapi/
├── asyncapi.yaml              # Complete specification (single file)
└── README.md                  # This documentation
```

The spec is consolidated into a single file for AsyncAPI 3.0 compatibility.
All channels, operations, messages, and schemas are defined inline.

## Event Flow

```
┌─────────────┐     ┌──────────┐     ┌──────────┐     ┌─────────┐
│   Ingest    │────▶│ Validate │────▶│  Enrich  │────▶│  Route  │
│orders.ingest│     │          │     │          │     │         │
└─────────────┘     └──────────┘     └──────────┘     └────┬────┘
                                                           │
                    ┌──────────────────────────────────────┼──────────┐
                    │                                      │          │
                    ▼                                      ▼          ▼
            ┌─────────────┐                    ┌──────────────┐  ┌─────────┐
            │ Fulfillment │                    │Manual Review │  │ Rejected│
            └─────────────┘                    └──────────────┘  └─────────┘
```

## Channels

| Channel | Purpose |
|---------|---------|
| `orders.ingest` | Raw incoming orders |
| `orders.validated` | Schema-validated orders |
| `orders.enriched` | Orders with customer/fraud data |
| `orders.routed.{destination}` | Final routing destinations |
| `orders.dlq` | Dead letter queue for failures |
| `pipeline.stage.{stageId}.complete` | Stage completion events |
| `pipeline.errors` | Centralized error channel |

## Validation

```bash
# Install AsyncAPI CLI
npm install -g @asyncapi/cli

# Validate the specification
asyncapi validate specs/asyncapi.yaml

# Generate documentation
asyncapi generate fromTemplate specs/asyncapi.yaml @asyncapi/html-template -o docs/
```

## Go Code Generation

The schemas in `components/schemas.yaml` map directly to Go structs in the implementation.
See `/internal/domain/` for the corresponding Go types.

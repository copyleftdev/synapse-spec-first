---
title: "Doc-First Development: Because 'We'll Document It Later' Is a Lie"
published: true
description: A love letter to specification-driven development, from someone who's been testing for 20 years.
tags: go, testing, openapi, architecture
cover_image: https://raw.githubusercontent.com/copyleftdev/synapse-spec-first/main/media/logo.png
canonical_url: https://github.com/copyleftdev/synapse-spec-first
---

# Doc-First Development: Because "We'll Document It Later" Is a Lie

*A love letter to specification-driven development, from someone who's been testing for 20 years.*

Every time I mention "doc-first" or "spec-driven development" to engineers, I get blank stares. 

*"That sounds great in a perfect world."*

Here's the thing: **We're software engineers. We create worlds.** Why wouldn't we try to make them perfect?

This article—and the [complete codebase on GitHub](https://github.com/copyleftdev/synapse-spec-first)—exists to demonstrate that spec-first isn't some ivory tower ideal. It's practical. It's powerful. And with the tools we have today, it's easier than ever.

---

## 🎯 What We're Building

**Synapse** is a complete event-driven order processing system built entirely spec-first:

![System Architecture](https://raw.githubusercontent.com/copyleftdev/synapse-spec-first/main/scripts/output/architecture.png)

1. **Specifications are the source of truth** — OpenAPI 3.1 for REST, AsyncAPI 3.0 for events
2. **Code is generated from specs** — Types, interfaces, clients, event handlers
3. **Conformance tests validate everything** — Implementation *must* match the specification
4. **Real infrastructure via Testcontainers** — NATS, PostgreSQL, Redis spin up for tests

The result? A system where the contract is king, drift is impossible, and tests prove conformance—not just behavior.

---

## 📐 The Doc-First Philosophy

![Doc-First Lifecycle](https://raw.githubusercontent.com/copyleftdev/synapse-spec-first/main/scripts/output/doc_first_lifecycle.png)

### Define Before You Build

Traditional development flows like this:
> Write code → Document later (maybe) → Hope nothing drifts

Doc-first development flips the script:
> Write spec → Generate code → Implement interfaces → Prove conformance

### Generated > Handwritten

When code is generated from specs:

- ✅ **No drift** between documentation and implementation
- ✅ **Type safety** guaranteed by the spec
- ✅ **Clients auto-generated** for any language
- ✅ **Changes flow spec → code**, never backwards

### Conformance Over Coverage

Traditional testing asks: *"Does the code do what the code says?"*

Conformance testing asks: *"Does the code do what the **contract** says?"*

One tests implementation. The other tests **promises**.

---

## 🔧 The Pipeline

![Pipeline Stages](https://raw.githubusercontent.com/copyleftdev/synapse-spec-first/main/scripts/output/pipeline_stages.png)

Orders flow through three Watermill-powered stages:

| Stage | Purpose |
|-------|---------|
| **Validate** | Check required fields, verify amounts, validate customer |
| **Enrich** | Customer tier lookup, fraud scoring, inventory check |
| **Route** | Apply routing rules, determine destination, set priority |

Each stage publishes to NATS, persists to PostgreSQL, and caches in Redis. Failed events go to a Dead Letter Queue for retry.

---

## 🧪 Testing Strategy

![Testing Strategy](https://raw.githubusercontent.com/copyleftdev/synapse-spec-first/main/scripts/output/testing_strategy.png)

### OpenAPI Conformance

```go
func TestOpenAPI_HealthEndpoint_ConformsToSpec(t *testing.T) {
    // Start real infrastructure with Testcontainers
    tc, _ := testutil.StartContainers(ctx, t, nil)
    
    // Create test suite from OpenAPI spec
    suite, _ := conformance.NewContractTestSuite(
        "openapi/openapi.yaml",
    )
    
    // Validate response matches spec schema
    result := suite.RunTest(ctx, client, baseURL,
        "GET", "/health",
        nil,
        http.StatusOK,
        "HealthResponse",  // Must match this schema
    )
    
    assert.True(t, result.Passed)
}
```

### AsyncAPI Conformance

```go
func TestAsyncAPI_OrderPayload_ConformsToSpec(t *testing.T) {
    // Create validator from AsyncAPI spec
    suite, _ := conformance.NewEventContractTestSuite(
        "asyncapi/asyncapi.yaml",
    )
    
    // Validate event payload against schema
    result := suite.ValidateEvent(
        "orders/ingest",
        "OrderReceivedPayload",
        orderJSON,
    )
    
    assert.True(t, result.Passed)
}
```

---

## 📁 Project Structure

```
synapse-spec-first/
├── asyncapi/              # AsyncAPI 3.0 event specs
├── openapi/               # OpenAPI 3.1 REST specs
├── cmd/
│   ├── synapse/           # Application entry point
│   └── synctl/            # Custom code generator
├── internal/
│   ├── generated/         # Generated from specs
│   │   ├── types.gen.go   # 31 domain types
│   │   ├── server.gen.go  # HTTP interface
│   │   ├── client.gen.go  # HTTP client
│   │   └── events.gen.go  # Event handlers
│   ├── handler/           # HTTP handlers
│   ├── pipeline/          # Watermill pipeline
│   ├── conformance/       # Contract testing
│   └── testutil/          # Testcontainers
└── scripts/               # Diagram generation
```

---

## 🔄 The Workflow

```bash
# 1. Edit the spec (the source of truth)
vim openapi/components/schemas/orders.yaml

# 2. Regenerate code
go run ./cmd/synctl

# 3. Implement the new interface methods
# (The compiler tells you what's missing)

# 4. Run conformance tests
go test ./internal/conformance/... -v
```

That's it. Spec changes → regenerate → implement → verify. The spec leads, the code follows.

---

## 🙏 Standing on Shoulders

This project wouldn't be possible without brilliant work from:

**Specification Standards**
- [OpenAPI Initiative](https://www.openapis.org/) — Giving REST APIs a language
- [AsyncAPI Initiative](https://www.asyncapi.com/) — The same for event-driven systems
- [JSON Schema](https://json-schema.org/) — The foundation for validation

**Testing Infrastructure**
- [Testcontainers](https://testcontainers.com/) — Real integration testing made accessible
- [Watermill](https://watermill.io/) — Elegant Go event-driven development

**The Go Ecosystem**
- [Chi](https://github.com/go-chi/chi) — Lightweight routing done right
- [NATS](https://nats.io/) — Derek Collison's gift to distributed systems

---

## 🤔 "But in the Real World..."

I've heard every objection:

> *"We don't have time to write specs first."*

You don't have time to debug integration issues from undocumented API changes either. Pick your poison.

> *"Specs get out of date."*

Not when they generate code. Not when conformance tests fail on drift.

> *"It's too much overhead."*

The overhead is front-loaded. The payoff compounds forever.

---

## 🚀 Try It Yourself

```bash
# Clone the repo
git clone https://github.com/copyleftdev/synapse-spec-first.git
cd synapse-spec-first

# Run the generator
go run ./cmd/synctl

# Run all tests (including conformance)
go test ./... -v

# Start the server
go run ./cmd/synapse
```

{% github copyleftdev/synapse-spec-first %}

---

## 💪 A Challenge

If you've never tried spec-first development, I challenge you:

**Build your next API starting with OpenAPI.**

1. Write the spec first
2. Generate your types
3. Implement to the interface
4. Write conformance tests

Then tell me it's not worth it.

---

## 🌟 Final Thoughts

![Philosophy](https://raw.githubusercontent.com/copyleftdev/synapse-spec-first/main/scripts/output/philosophy.png)

Someone once told me, *"In a perfect world, that would be great."*

We're software engineers. **We create worlds.**

Why wouldn't we try to make them perfect?

---

*This project is a living demonstration. Fork it. Learn from it. Improve it. And maybe, just maybe, next time someone mentions doc-first development, there will be one fewer blank stare.*

**20 years of testing taught me this: Quality isn't an afterthought. It's the architecture.**

Now go build something beautiful.

---

## 📚 Resources

- **Full Codebase**: [github.com/copyleftdev/synapse-spec-first](https://github.com/copyleftdev/synapse-spec-first)
- **OpenAPI Spec**: [openapi/openapi.yaml](https://github.com/copyleftdev/synapse-spec-first/blob/main/openapi/openapi.yaml)
- **AsyncAPI Spec**: [asyncapi/asyncapi.yaml](https://github.com/copyleftdev/synapse-spec-first/blob/main/asyncapi/asyncapi.yaml)
- **Conformance Tests**: [internal/conformance/](https://github.com/copyleftdev/synapse-spec-first/tree/main/internal/conformance)

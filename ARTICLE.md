# We're Software Engineers. We Can Create Worlds.

*A love letter to specification-driven development, from someone who's been testing for 20 years.*

---

## The Blank Stares

Every time I mention "doc-first" or "spec-driven development" to younger engineers, I get blank stares. 

*"That sounds great in a perfect world."*

Here's the thing: **We're software engineers. We create worlds.** Why wouldn't we try to make them perfect?

This article—and this entire codebase—exists to demonstrate that spec-first isn't some ivory tower ideal. It's practical. It's powerful. And with the tools we have today, it's easier than ever.

---

## What This Project Demonstrates

**Synapse** is a complete event-driven order processing system built entirely spec-first:

1. **Specifications are the source of truth** — OpenAPI 3.1 for REST, AsyncAPI 3.0 for events
2. **Code is generated from specs** — Types, interfaces, clients, event handlers
3. **Conformance tests validate everything** — Implementation *must* match the specification
4. **Real infrastructure via Testcontainers** — NATS, PostgreSQL, Redis spin up for tests

The result? A system where the contract is king, drift is impossible, and tests prove conformance—not just behavior.

---

## The Philosophy

### 1. Define Before You Build

```
┌─────────────────────────────────────────────────────────────┐
│                    SPECIFICATION                             │
│         (The contract. The promise. The truth.)             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    GENERATED CODE                            │
│         (Types, interfaces, clients, handlers)               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    IMPLEMENTATION                            │
│         (Fill in the blanks. The spec guides you.)          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 CONFORMANCE TESTING                          │
│         (Prove it. Against the spec. Always.)               │
└─────────────────────────────────────────────────────────────┘
```

### 2. Generated > Handwritten

When code is generated from specs:
- **No drift** between documentation and implementation
- **Type safety** is guaranteed by the spec
- **Clients can be auto-generated** for any language
- **Changes flow from spec → code**, never backwards

### 3. Conformance Over Coverage

Traditional testing asks: *"Does the code do what the code says?"*

Conformance testing asks: *"Does the code do what the contract says?"*

One tests implementation. The other tests promises.

---

## Standing on the Shoulders of Giants

This project wouldn't be possible without the work of brilliant people who came before:

### Specification Standards
- **OpenAPI Initiative** — For giving REST APIs a language
- **AsyncAPI Initiative** — For doing the same for event-driven systems
- **JSON Schema** — The foundation that makes validation possible

### Testing Infrastructure
- **Testcontainers** — Richard North and the team who made "real" integration testing accessible
- **Watermill** — Three Dots Labs for making Go event-driven development elegant

### The Go Ecosystem
- **Chi** — Lightweight routing done right
- **NATS** — Derek Collison's gift to distributed systems
- **PostgreSQL** & **Redis** — The workhorses we all depend on

To all of you: thank you. You made this possible.

---

## The Codebase

### Project Structure

```
synapse/
├── asyncapi/                  # Event specifications (AsyncAPI 3.0)
│   └── asyncapi.yaml
├── openapi/                   # REST API specifications (OpenAPI 3.1)
│   ├── openapi.yaml
│   ├── paths/
│   └── components/
├── cmd/
│   ├── synapse/               # Application entry point
│   └── synctl/                # Custom code generator
├── internal/
│   ├── generated/             # Generated from specs
│   │   ├── types.gen.go       # 31 domain types
│   │   ├── server.gen.go      # HTTP interface
│   │   ├── client.gen.go      # HTTP client
│   │   └── events.gen.go      # Event handlers
│   ├── handler/               # HTTP handler implementations
│   ├── pipeline/              # Watermill event pipeline
│   ├── conformance/           # Contract testing framework
│   └── testutil/              # Testcontainers helpers
└── scripts/                   # Diagram generation
```

### The Workflow

```bash
# 1. Edit the spec (the source of truth)
vim openapi/components/schemas/orders.yaml

# 2. Regenerate code
go run ./cmd/synctl

# 3. Implement the new interface methods
# (The compiler tells you what's missing)

# 4. Run conformance tests
go test ./internal/conformance/... -v

# 5. Verify against spec
# Tests validate your responses match the OpenAPI schema
# Tests validate your events match the AsyncAPI schema
```

---

## Conformance in Action

### OpenAPI Conformance

```go
func TestOpenAPI_HealthEndpoint_ConformsToSpec(t *testing.T) {
    // Start real infrastructure
    tc, _ := testutil.StartContainers(ctx, t, nil)
    
    // Create test suite from OpenAPI spec
    suite, _ := conformance.NewContractTestSuite("openapi/openapi.yaml")
    
    // Validate response matches spec
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
    suite, _ := conformance.NewEventContractTestSuite("asyncapi/asyncapi.yaml")
    
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

## Why This Matters

In 20 years of testing, I've seen systems rot. Documentation lies. Implementations drift. Contracts break silently.

**Spec-first development is the antidote.**

When the spec is the source of truth:
- Documentation is always accurate (it *is* the code)
- Breaking changes are visible (the spec diff shows them)
- Clients can trust the contract (it's validated, not assumed)
- Onboarding is faster (read the spec, understand the system)

---

## "But in the Real World..."

I've heard it all:

*"We don't have time to write specs first."*

You don't have time to debug integration issues caused by undocumented API changes either. Pick your poison.

*"Specs get out of date."*

Not when they generate code. Not when conformance tests fail on drift.

*"It's too much overhead."*

The overhead is front-loaded. The payoff compounds forever.

---

## Try It Yourself

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

---

## A Challenge

If you've never tried spec-first development, I challenge you:

**Build your next API starting with OpenAPI.**

1. Write the spec first
2. Generate your types
3. Implement to the interface
4. Write conformance tests

Then tell me it's not worth it.

---

## Final Thoughts

Someone once told me, *"In a perfect world, that would be great."*

We're software engineers. **We create worlds.**

Why wouldn't we try to make them perfect?

---

*This project is a living demonstration. Fork it. Learn from it. Improve it. And maybe, just maybe, next time someone mentions doc-first development, there will be one fewer blank stare.*

---

## Acknowledgments

To the testing community. To the specification authors. To everyone who believes that quality isn't an afterthought.

To chaos engineering, mutation testing, property-based testing, and every technique that makes software better.

And to you, for reading this far.

Now go build something beautiful.

---

**Author's Note:** This codebase was built to accompany an article on specification-driven development. Every file, every test, every diagram exists to prove a point: that the "perfect world" is the one we choose to build.

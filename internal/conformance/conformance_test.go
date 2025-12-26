package conformance_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/synapse/synapse/internal/conformance"
	"github.com/synapse/synapse/internal/handler"
	"github.com/synapse/synapse/internal/pipeline"
	"github.com/synapse/synapse/internal/testutil"
)

const (
	openAPISpecPath  = "../../openapi/openapi.yaml"
	asyncAPISpecPath = "../../asyncapi/asyncapi.yaml"
)

func TestOpenAPI_HealthEndpoint_ConformsToSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start containers
	tc, err := testutil.StartContainers(ctx, t, nil)
	require.NoError(t, err)

	infra, cfg := testutil.TestInfra(ctx, t, tc)

	runner, err := pipeline.New(ctx, cfg, infra)
	require.NoError(t, err)

	h := handler.New(infra, runner)

	// Create test server
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create contract test suite
	suite, err := conformance.NewContractTestSuite(openAPISpecPath)
	require.NoError(t, err)

	// Test health endpoint
	result := suite.RunTest(ctx, srv.Client(), srv.URL,
		"GET", "/health",
		nil,
		http.StatusOK,
		"HealthResponse",
	)

	assert.True(t, result.Passed, "health endpoint should conform to spec: %s", result.Error)
}

func TestOpenAPI_LivenessEndpoint_ConformsToSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	tc, err := testutil.StartContainers(ctx, t, nil)
	require.NoError(t, err)

	infra, cfg := testutil.TestInfra(ctx, t, tc)

	runner, err := pipeline.New(ctx, cfg, infra)
	require.NoError(t, err)

	h := handler.New(infra, runner)

	r := chi.NewRouter()
	h.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Liveness returns simple JSON, validate structure
	resp, err := srv.Client().Get(srv.URL + "/health/live")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "ok", body["status"])
}

func TestOpenAPI_PipelineStagesEndpoint_ConformsToSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	tc, err := testutil.StartContainers(ctx, t, nil)
	require.NoError(t, err)

	infra, cfg := testutil.TestInfra(ctx, t, tc)

	runner, err := pipeline.New(ctx, cfg, infra)
	require.NoError(t, err)

	h := handler.New(infra, runner)

	r := chi.NewRouter()
	h.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	defer srv.Close()

	suite, err := conformance.NewContractTestSuite(openAPISpecPath)
	require.NoError(t, err)

	result := suite.RunTest(ctx, srv.Client(), srv.URL,
		"GET", "/api/v1/pipeline/stages",
		nil,
		http.StatusOK,
		"PipelineStagesResponse",
	)

	assert.True(t, result.Passed, "pipeline stages endpoint should conform to spec: %s", result.Error)
}

func TestAsyncAPI_OrderReceivedPayload_ConformsToSpec(t *testing.T) {
	suite, err := conformance.NewEventContractTestSuite(asyncAPISpecPath)
	require.NoError(t, err)

	// Valid payload
	validPayload := map[string]any{
		"orderId":    "550e8400-e29b-41d4-a716-446655440000",
		"customerId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"items": []map[string]any{
			{
				"sku":       "WIDGET-001",
				"quantity":  2,
				"unitPrice": 29.99,
			},
		},
		"totalAmount": 59.98,
		"currency":    "USD",
		"createdAt":   "2024-01-15T10:30:00.000Z",
	}

	payloadBytes, _ := json.Marshal(validPayload)
	result := suite.ValidateEvent("orders/ingest", "OrderReceivedPayload", payloadBytes)

	assert.True(t, result.Passed, "valid OrderReceivedPayload should conform to spec: %s", result.Error)
}

func TestAsyncAPI_OrderReceivedPayload_FailsOnInvalidPayload(t *testing.T) {
	suite, err := conformance.NewEventContractTestSuite(asyncAPISpecPath)
	require.NoError(t, err)

	// Invalid payload - missing required fields
	invalidPayload := map[string]any{
		"orderId": "550e8400-e29b-41d4-a716-446655440000",
		// Missing customerId, items, totalAmount, currency, createdAt
	}

	payloadBytes, _ := json.Marshal(invalidPayload)
	result := suite.ValidateEvent("orders/ingest", "OrderReceivedPayload", payloadBytes)

	assert.False(t, result.Passed, "invalid payload should fail validation")
	assert.Contains(t, result.Error, "customerId")
}

func TestAsyncAPI_StageCompletePayload_ConformsToSpec(t *testing.T) {
	suite, err := conformance.NewEventContractTestSuite(asyncAPISpecPath)
	require.NoError(t, err)

	validPayload := map[string]any{
		"stageId":    "validate",
		"eventId":    "550e8400-e29b-41d4-a716-446655440000",
		"durationMs": 45,
		"status":     "success",
	}

	payloadBytes, _ := json.Marshal(validPayload)
	result := suite.ValidateEvent("pipeline/stage-complete", "StageCompletePayload", payloadBytes)

	assert.True(t, result.Passed, "valid StageCompletePayload should conform to spec: %s", result.Error)
}

func TestAsyncAPI_PipelineErrorPayload_ConformsToSpec(t *testing.T) {
	suite, err := conformance.NewEventContractTestSuite(asyncAPISpecPath)
	require.NoError(t, err)

	validPayload := map[string]any{
		"errorId":   "661f9511-f3ac-52e5-b827-557766551111",
		"eventId":   "550e8400-e29b-41d4-a716-446655440000",
		"stageId":   "enrich",
		"errorType": "timeout",
		"message":   "Enrichment service timed out after 30s",
		"timestamp": "2024-01-15T10:30:05.500Z",
	}

	payloadBytes, _ := json.Marshal(validPayload)
	result := suite.ValidateEvent("pipeline/errors", "PipelineErrorPayload", payloadBytes)

	assert.True(t, result.Passed, "valid PipelineErrorPayload should conform to spec: %s", result.Error)
}

func TestConformance_FullSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full conformance suite")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start infrastructure
	tc, err := testutil.StartContainers(ctx, t, nil)
	require.NoError(t, err)

	infra, cfg := testutil.TestInfra(ctx, t, tc)

	runner, err := pipeline.New(ctx, cfg, infra)
	require.NoError(t, err)

	h := handler.New(infra, runner)

	r := chi.NewRouter()
	h.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	defer srv.Close()

	// OpenAPI conformance tests - validate response structure
	t.Run("OpenAPI_ResponseStructure", func(t *testing.T) {
		tests := []struct {
			name           string
			method         string
			path           string
			expectedStatus int
			validateFn     func(t *testing.T, body []byte)
		}{
			{
				"health endpoint returns status",
				"GET", "/health", 200,
				func(t *testing.T, body []byte) {
					var resp map[string]any
					require.NoError(t, json.Unmarshal(body, &resp))
					assert.Contains(t, resp, "status")
					assert.Contains(t, resp, "components")
				},
			},
			{
				"liveness returns ok",
				"GET", "/health/live", 200,
				func(t *testing.T, body []byte) {
					var resp map[string]any
					require.NoError(t, json.Unmarshal(body, &resp))
					assert.Equal(t, "ok", resp["status"])
				},
			},
			{
				"readiness returns ready",
				"GET", "/health/ready", 200,
				func(t *testing.T, body []byte) {
					var resp map[string]any
					require.NoError(t, json.Unmarshal(body, &resp))
					assert.Equal(t, "ready", resp["status"])
				},
			},
			{
				"pipeline stages returns array",
				"GET", "/api/v1/pipeline/stages", 200,
				func(t *testing.T, body []byte) {
					var resp map[string]any
					require.NoError(t, json.Unmarshal(body, &resp))
					assert.Contains(t, resp, "stages")
					stages, ok := resp["stages"].([]any)
					assert.True(t, ok, "stages should be an array")
					assert.Len(t, stages, 3, "should have 3 pipeline stages")
				},
			},
			{
				"orders list returns array",
				"GET", "/api/v1/orders", 200,
				func(t *testing.T, body []byte) {
					var resp map[string]any
					require.NoError(t, json.Unmarshal(body, &resp))
					assert.Contains(t, resp, "orders")
				},
			},
		}

		passed := 0
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := srv.Client().Get(srv.URL + tt.path)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				body, _ := io.ReadAll(resp.Body)
				tt.validateFn(t, body)
				passed++
			})
		}
		t.Logf("OpenAPI Conformance: %d tests passed", passed)
	})

	// AsyncAPI conformance tests
	t.Run("AsyncAPI_EventSchemas", func(t *testing.T) {
		suite, err := conformance.NewEventContractTestSuite(asyncAPISpecPath)
		require.NoError(t, err)

		// Test all event payloads
		eventTests := []struct {
			channel string
			schema  string
			payload map[string]any
		}{
			{
				"orders/ingest", "OrderReceivedPayload",
				map[string]any{
					"orderId": "test-123", "customerId": "cust-456",
					"items":       []map[string]any{{"sku": "SKU1", "quantity": 1, "unitPrice": 10.0}},
					"totalAmount": 10.0, "currency": "USD", "createdAt": "2024-01-15T10:00:00Z",
				},
			},
			{
				"pipeline/stage-complete", "StageCompletePayload",
				map[string]any{
					"stageId":    "validate",
					"eventId":    "evt-123",
					"durationMs": 45,
					"status":     "success",
				},
			},
			{
				"pipeline/errors", "PipelineErrorPayload",
				map[string]any{
					"errorId":   "err-123",
					"eventId":   "evt-456",
					"stageId":   "enrich",
					"errorType": "validation",
					"message":   "Invalid data",
					"timestamp": "2024-01-15T10:00:00Z",
				},
			},
		}

		for _, tt := range eventTests {
			payloadBytes, _ := json.Marshal(tt.payload)
			result := suite.ValidateEvent(tt.channel, tt.schema, payloadBytes)
			if !result.Passed {
				t.Errorf("Event %s/%s failed: %s", tt.channel, tt.schema, result.Error)
			}
		}

		passed, failed := suite.Summary()
		t.Logf("AsyncAPI Conformance: %d passed, %d failed", passed, failed)
		assert.Equal(t, 0, failed, "all AsyncAPI validations should pass")
	})
}

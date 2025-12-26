package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/synapse/synapse/internal/generated"
	"github.com/synapse/synapse/internal/pipeline"
	"github.com/synapse/synapse/internal/testutil"
)

func TestPipeline_IngestOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start test containers
	tc, err := testutil.StartContainers(ctx, t, nil)
	require.NoError(t, err, "failed to start containers")

	// Create test infrastructure
	infra, cfg := testutil.TestInfra(ctx, t, tc)

	// Create pipeline
	runner, err := pipeline.New(ctx, cfg, infra)
	require.NoError(t, err, "failed to create pipeline")

	// Start pipeline in background
	go func() {
		if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
			t.Logf("pipeline error: %v", err)
		}
	}()

	// Give pipeline time to start
	time.Sleep(100 * time.Millisecond)

	// Test ingesting an order
	orderReq := &generated.OrderCreateRequest{
		CustomerId:  "test-customer-123",
		TotalAmount: 99.99,
		Currency:    "USD",
		Items: []generated.OrderItem{
			{
				Sku:       "TEST-SKU-001",
				Quantity:  2,
				UnitPrice: 49.99,
			},
		},
	}

	err = runner.IngestOrder(ctx, "test-order-123", orderReq)
	require.NoError(t, err, "failed to ingest order")

	// Give pipeline time to process
	time.Sleep(500 * time.Millisecond)

	// Verify stages are healthy
	stages := runner.GetStages()
	assert.Len(t, stages, 3, "expected 3 pipeline stages")

	for _, stage := range stages {
		assert.Equal(t, generated.StageStatusHealthy, stage.Status, "stage %s should be healthy", stage.StageId)
	}

	// Verify individual stage metrics
	validateStage := runner.GetStage("validate")
	require.NotNil(t, validateStage, "validate stage should exist")
	assert.Equal(t, "validate", validateStage.StageId)
}

func TestPipeline_GetStages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	tc, err := testutil.StartContainers(ctx, t, nil)
	require.NoError(t, err)

	infra, cfg := testutil.TestInfra(ctx, t, tc)

	runner, err := pipeline.New(ctx, cfg, infra)
	require.NoError(t, err)

	stages := runner.GetStages()
	assert.Len(t, stages, 3)

	stageIds := make(map[string]bool)
	for _, s := range stages {
		stageIds[s.StageId] = true
	}

	assert.True(t, stageIds["validate"], "should have validate stage")
	assert.True(t, stageIds["enrich"], "should have enrich stage")
	assert.True(t, stageIds["route"], "should have route stage")
}

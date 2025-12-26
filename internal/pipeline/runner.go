package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/synapse/synapse/internal/config"
	"github.com/synapse/synapse/internal/generated"
	"github.com/synapse/synapse/internal/infra"
)

// Topics
const (
	TopicOrdersIngest    = "orders.ingest"
	TopicOrdersValidated = "orders.validated"
	TopicOrdersEnriched  = "orders.enriched"
	TopicOrdersRouted    = "orders.routed"
	TopicOrdersDLQ       = "orders.dlq"
)

// Runner manages the event pipeline
type Runner struct {
	config    *config.Config
	infra     *infra.Infra
	router    *message.Router
	publisher message.Publisher
	logger    watermill.LoggerAdapter
	stages    map[string]*StageMetrics
}

// StageMetrics tracks metrics for a pipeline stage
type StageMetrics struct {
	StageId         string                `json:"stageId"`
	Status          generated.StageStatus `json:"status"`
	ProcessedTotal  int64                 `json:"processedTotal"`
	ProcessedLastHr int64                 `json:"processedLastHour"`
	ErrorRate       float64               `json:"errorRate"`
	AvgLatencyMs    float64               `json:"avgLatencyMs"`
	QueueDepth      int                   `json:"queueDepth"`
	LastProcessedAt time.Time             `json:"lastProcessedAt,omitempty"`
}

// New creates a new pipeline Runner
func New(ctx context.Context, cfg *config.Config, infra *infra.Infra) (*Runner, error) {
	logger := watermill.NewSlogLogger(slog.Default())

	// For now, use in-memory pub/sub (will switch to NATS for production)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("creating router: %w", err)
	}

	// Add middleware
	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Retry{
			MaxRetries:      cfg.RetryMaxAttempts,
			InitialInterval: time.Duration(cfg.RetryBackoffMs) * time.Millisecond,
			Logger:          logger,
		}.Middleware,
		middleware.Recoverer,
	)

	r := &Runner{
		config:    cfg,
		infra:     infra,
		router:    router,
		publisher: pubSub,
		logger:    logger,
		stages: map[string]*StageMetrics{
			"validate": {StageId: "validate", Status: generated.StageStatusHealthy},
			"enrich":   {StageId: "enrich", Status: generated.StageStatusHealthy},
			"route":    {StageId: "route", Status: generated.StageStatusHealthy},
		},
	}

	// Register handlers
	router.AddHandler(
		"validate_order",
		TopicOrdersIngest,
		pubSub,
		TopicOrdersValidated,
		pubSub,
		r.handleValidate,
	)

	router.AddHandler(
		"enrich_order",
		TopicOrdersValidated,
		pubSub,
		TopicOrdersEnriched,
		pubSub,
		r.handleEnrich,
	)

	router.AddHandler(
		"route_order",
		TopicOrdersEnriched,
		pubSub,
		TopicOrdersRouted,
		pubSub,
		r.handleRoute,
	)

	return r, nil
}

// Run starts the pipeline router
func (r *Runner) Run(ctx context.Context) error {
	return r.router.Run(ctx)
}

// Close stops the pipeline
func (r *Runner) Close() error {
	return r.router.Close()
}

// IngestOrder publishes an order to the pipeline
func (r *Runner) IngestOrder(ctx context.Context, orderID string, req *generated.OrderCreateRequest) error {
	payload := map[string]any{
		"orderId":     orderID,
		"customerId":  req.CustomerId,
		"items":       req.Items,
		"totalAmount": req.TotalAmount,
		"currency":    req.Currency,
		"createdAt":   time.Now().UTC(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling order: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("correlationId", orderID)

	return r.publisher.Publish(TopicOrdersIngest, msg)
}

// GetStages returns current stage metrics
func (r *Runner) GetStages() []generated.PipelineStageSummary {
	stages := make([]generated.PipelineStageSummary, 0, len(r.stages))
	for _, s := range r.stages {
		stages = append(stages, generated.PipelineStageSummary{
			StageId: s.StageId,
			Status:  s.Status,
		})
	}
	return stages
}

// GetStage returns a specific stage's metrics
func (r *Runner) GetStage(stageID string) *generated.PipelineStageResponse {
	s, ok := r.stages[stageID]
	if !ok {
		return nil
	}
	return &generated.PipelineStageResponse{
		StageId: s.StageId,
		Status:  s.Status,
	}
}

// handleValidate validates incoming orders
func (r *Runner) handleValidate(msg *message.Message) ([]*message.Message, error) {
	start := time.Now()
	defer r.recordMetrics("validate", start)

	var order map[string]any
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return nil, fmt.Errorf("unmarshaling order: %w", err)
	}

	slog.Info("validating order", "orderId", order["orderId"])

	// Validation logic
	if order["customerId"] == nil || order["customerId"] == "" {
		return nil, fmt.Errorf("customerId is required")
	}

	items, ok := order["items"].([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	// Add validation result
	order["validatedAt"] = time.Now().UTC()
	order["validationResult"] = map[string]any{
		"isValid":  true,
		"warnings": []string{},
	}

	data, _ := json.Marshal(order)
	outMsg := message.NewMessage(watermill.NewUUID(), data)
	outMsg.Metadata = msg.Metadata

	return []*message.Message{outMsg}, nil
}

// handleEnrich enriches orders with customer and fraud data
func (r *Runner) handleEnrich(msg *message.Message) ([]*message.Message, error) {
	start := time.Now()
	defer r.recordMetrics("enrich", start)

	var order map[string]any
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return nil, fmt.Errorf("unmarshaling order: %w", err)
	}

	slog.Info("enriching order", "orderId", order["orderId"])

	// Simulate customer data enrichment
	order["enrichedAt"] = time.Now().UTC()
	order["customer"] = map[string]any{
		"tier":          "gold",
		"accountAge":    365,
		"lifetimeValue": 1500.00,
	}

	// Simulate fraud scoring
	order["fraudScore"] = map[string]any{
		"score":     15,
		"riskLevel": "low",
		"signals":   []string{},
	}

	data, _ := json.Marshal(order)
	outMsg := message.NewMessage(watermill.NewUUID(), data)
	outMsg.Metadata = msg.Metadata

	return []*message.Message{outMsg}, nil
}

// handleRoute determines the routing destination
func (r *Runner) handleRoute(msg *message.Message) ([]*message.Message, error) {
	start := time.Now()
	defer r.recordMetrics("route", start)

	var order map[string]any
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return nil, fmt.Errorf("unmarshaling order: %w", err)
	}

	slog.Info("routing order", "orderId", order["orderId"])

	// Determine routing based on fraud score
	fraudScore := 0.0
	if fs, ok := order["fraudScore"].(map[string]any); ok {
		if score, ok := fs["score"].(float64); ok {
			fraudScore = score
		}
	}

	destination := "fulfillment"
	reason := "All checks passed"

	if fraudScore > 50 {
		destination = "manual-review"
		reason = "High fraud score requires manual review"
	} else if fraudScore > 80 {
		destination = "rejected"
		reason = "Fraud score exceeds threshold"
	}

	order["routedAt"] = time.Now().UTC()
	order["destination"] = destination
	order["routingReason"] = reason

	data, _ := json.Marshal(order)
	outMsg := message.NewMessage(watermill.NewUUID(), data)
	outMsg.Metadata = msg.Metadata

	return []*message.Message{outMsg}, nil
}

func (r *Runner) recordMetrics(stage string, start time.Time) {
	if s, ok := r.stages[stage]; ok {
		s.ProcessedTotal++
		s.LastProcessedAt = time.Now()
		latency := float64(time.Since(start).Milliseconds())
		// Simple moving average
		s.AvgLatencyMs = (s.AvgLatencyMs*float64(s.ProcessedTotal-1) + latency) / float64(s.ProcessedTotal)
	}
}

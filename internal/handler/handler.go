package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/synapse/synapse/internal/generated"
	"github.com/synapse/synapse/internal/infra"
	"github.com/synapse/synapse/internal/pipeline"
)

// Handler implements the generated.ServerInterface
type Handler struct {
	infra    *infra.Infra
	pipeline *pipeline.Runner
}

// New creates a new Handler
func New(infra *infra.Infra, pipeline *pipeline.Runner) *Handler {
	return &Handler{
		infra:    infra,
		pipeline: pipeline,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Orders
	r.Post("/api/v1/orders", h.wrapHandler(h.IngestOrder))
	r.Get("/api/v1/orders", h.wrapHandler(h.ListOrders))
	r.Get("/api/v1/orders/{orderId}", h.wrapHandler(h.GetOrder))
	r.Delete("/api/v1/orders/{orderId}", h.wrapHandler(h.CancelOrder))
	r.Get("/api/v1/orders/{orderId}/events", h.wrapHandler(h.GetOrderEvents))

	// Pipeline
	r.Get("/api/v1/pipeline/stages", h.wrapHandler(h.ListPipelineStages))
	r.Get("/api/v1/pipeline/stages/{stageId}", h.wrapHandler(h.GetPipelineStage))
	r.Patch("/api/v1/pipeline/stages/{stageId}", h.wrapHandler(h.UpdatePipelineStage))
	r.Get("/api/v1/pipeline/dlq", h.wrapHandler(h.ListDLQItems))
	r.Post("/api/v1/pipeline/dlq/{eventId}/retry", h.wrapHandler(h.RetryDLQItem))

	// Health
	r.Get("/health", h.wrapHandler(h.GetHealth))
	r.Get("/health/live", h.wrapHandler(h.GetLiveness))
	r.Get("/health/ready", h.wrapHandler(h.GetReadiness))
	r.Get("/metrics", h.wrapHandler(h.GetMetrics))
}

func (h *Handler) wrapHandler(fn func(context.Context, http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(r.Context(), w, r); err != nil {
			h.writeError(w, err)
		}
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]any{
		"type":   "https://synapse.example.com/problems/internal-error",
		"title":  "Internal Server Error",
		"status": 500,
		"detail": err.Error(),
	})
}

// IngestOrder handles POST /api/v1/orders
func (h *Handler) IngestOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req generated.OrderCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return h.writeJSON(w, http.StatusBadRequest, map[string]any{
			"type":   "https://synapse.example.com/problems/invalid-json",
			"title":  "Invalid JSON",
			"status": 400,
			"detail": err.Error(),
		})
	}

	orderID := uuid.New().String()

	// Publish to pipeline
	if err := h.pipeline.IngestOrder(ctx, orderID, &req); err != nil {
		return err
	}

	w.Header().Set("Location", "/api/v1/orders/"+orderID)
	return h.writeJSON(w, http.StatusAccepted, generated.OrderAcceptedResponse{
		OrderId: orderID,
		Status:  "accepted",
		Message: "Order accepted for processing",
	})
}

// ListOrders handles GET /api/v1/orders
func (h *Handler) ListOrders(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// TODO: Implement with database query
	return h.writeJSON(w, http.StatusOK, generated.OrderListResponse{
		Orders: []generated.OrderSummary{},
	})
}

// GetOrder handles GET /api/v1/orders/{orderId}
func (h *Handler) GetOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	orderID := chi.URLParam(r, "orderId")
	// TODO: Implement with database query
	return h.writeJSON(w, http.StatusOK, generated.OrderResponse{
		OrderId: orderID,
		Status:  "processing",
	})
}

// CancelOrder handles DELETE /api/v1/orders/{orderId}
func (h *Handler) CancelOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	orderID := chi.URLParam(r, "orderId")
	// TODO: Implement cancellation logic
	return h.writeJSON(w, http.StatusOK, generated.OrderCancelledResponse{
		OrderId:     orderID,
		Status:      "cancelled",
		CancelledAt: time.Now(),
	})
}

// GetOrderEvents handles GET /api/v1/orders/{orderId}/events
func (h *Handler) GetOrderEvents(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	orderID := chi.URLParam(r, "orderId")
	// TODO: Implement with database query
	return h.writeJSON(w, http.StatusOK, generated.OrderEventsResponse{
		OrderId: orderID,
		Events:  []generated.OrderEvent{},
	})
}

// ListPipelineStages handles GET /api/v1/pipeline/stages
func (h *Handler) ListPipelineStages(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	stages := h.pipeline.GetStages()
	return h.writeJSON(w, http.StatusOK, generated.PipelineStagesResponse{
		Stages: stages,
	})
}

// GetPipelineStage handles GET /api/v1/pipeline/stages/{stageId}
func (h *Handler) GetPipelineStage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	stageID := chi.URLParam(r, "stageId")
	stage := h.pipeline.GetStage(stageID)
	if stage == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	return h.writeJSON(w, http.StatusOK, stage)
}

// UpdatePipelineStage handles PATCH /api/v1/pipeline/stages/{stageId}
func (h *Handler) UpdatePipelineStage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// TODO: Implement stage update
	return h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ListDLQItems handles GET /api/v1/pipeline/dlq
func (h *Handler) ListDLQItems(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// TODO: Implement DLQ listing
	return h.writeJSON(w, http.StatusOK, generated.DLQListResponse{
		Items: []generated.DLQItem{},
	})
}

// RetryDLQItem handles POST /api/v1/pipeline/dlq/{eventId}/retry
func (h *Handler) RetryDLQItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	eventID := chi.URLParam(r, "eventId")
	// TODO: Implement retry logic
	return h.writeJSON(w, http.StatusAccepted, map[string]string{
		"eventId": eventID,
		"status":  "requeued",
	})
}

// GetHealth handles GET /health
func (h *Handler) GetHealth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	health := h.infra.Healthy(ctx)
	status := "healthy"
	httpStatus := http.StatusOK

	components := make(map[string]any)
	for name, err := range health {
		if err != nil {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			components[name] = map[string]any{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			components[name] = map[string]any{
				"status": "healthy",
			}
		}
	}

	return h.writeJSON(w, httpStatus, generated.HealthResponse{
		Status:     status,
		Components: components,
	})
}

// GetLiveness handles GET /health/live
func (h *Handler) GetLiveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetReadiness handles GET /health/ready
func (h *Handler) GetReadiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	health := h.infra.Healthy(ctx)
	for _, err := range health {
		if err != nil {
			return h.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		}
	}
	return h.writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// GetMetrics handles GET /metrics
func (h *Handler) GetMetrics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// TODO: Implement Prometheus metrics
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("# Synapse metrics\n"))
	return nil
}

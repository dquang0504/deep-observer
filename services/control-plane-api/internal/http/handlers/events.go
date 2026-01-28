package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/db/repo"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/helper"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


type EventsHandler struct{
	V *validation.Validator
	Services *repo.ServicesRepo
	Envs *repo.EnvironmentRepo
	Events *repo.EventsRepo
}

func NewEventsHandler(v *validation.Validator, s *repo.ServicesRepo, e *repo.EnvironmentRepo, ev *repo.EventsRepo) *EventsHandler{
	return &EventsHandler{V: v, Services: s, Envs: e, Events: ev}
}

func (h *EventsHandler) PostEvent(c *gin.Context){
	h.handleUpsert(c, "")
}

func (h *EventsHandler) PutEvent(c *gin.Context){
	idStr := c.Param("id")
	h.handleUpsert(c, idStr)
}

func (h *EventsHandler) handleUpsert(c *gin.Context, pathID string){
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil{
		helper.RespondError(c,400, "bad_json", err.Error())
		return
	}

	// Validate against JSON Schema (hard contract)
	if err := h.V.EventSchema.Validate(payload); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error_code": "schema_validation_failed",
			"message": validation.FormatSchemaError(err),
		})
		return
	}
	// Extract required fields safely
	// schema required: schema_version, id, event_type, service_name, environment, occurred_at, title
	idStr, ok := asString(payload["id"])
	if !ok || idStr == "" {
		helper.RespondError(c,400, "missing_id", "id is required")
		return
	}

	// PUT: enforce path id == payload.id
	if pathID != "" && pathID != idStr{
		helper.RespondError(c,400, "id_mismatch", "path id must match payload id")
	}

	eventID, err := uuid.Parse(idStr)
	if err != nil{
		helper.RespondError(c,400, "invalid_id", err.Error())
		return
	}

	eventType, _ := asString(payload["event_type"])
	serviceName, _ := asString(payload["service_name"])
	environment, _ := asString(payload["environment"])
	title, _ := asString(payload["title"])

	occurredAtStr, _ := asString(payload["occurred_at"])
	occurredAtStr = strings.TrimSpace(occurredAtStr) 
	occurredAt, err := time.Parse(time.RFC3339, occurredAtStr)
	if err != nil{
		helper.RespondError(c,400, "invalid_occurred_at", "occurred_at must be RFC3339")
		return
	}

	commitHash, _ := asString(payload["commit_hash"])
	actor, _ := asString(payload["actor"])

	ctx := c.Request.Context()

	// Resolve service/env => IDs (raw SQL via repos)
	serviceID, err := h.Services.EnsureService(ctx, serviceName)
	if err != nil{
		helper.RespondError(c,500, "service_lookup_failed", err.Error())
		return
	}

	envID, err := h.Envs.GetIDByName(ctx, environment)
	if err != nil{
		helper.RespondError(c,400, "unknown_environment", "environment not found (seed required)")
		return
	}

	// Upsert deployment event (idempotent)
	var commitPtr *string
	if commitHash != ""{
		commitPtr = &commitHash
	}
	var actorPtr *string
	if actor != ""{
		actorPtr = &actor
	}

	switch eventType{
	case "deployment":
		version, _ := asString(payload["version"])
		if version == "" {
			helper.RespondError(c,400, "missing_version", "version is required for deployment events")
			return
		}
		if err := h.Events.UpsertDeploymentEvent(ctx, eventID, serviceID, envID, version, commitPtr, actorPtr, occurredAt); err != nil {
			helper.RespondError(c,500, "db_write_failed", err.Error())
			return
		}
	case "incident", "config_change", "maintenance":
		var endAtPtr *time.Time
		if endAtStr, ok := asString(payload["end_at"]); ok && strings.TrimSpace(endAtStr) != ""{
			endAt, err := time.Parse(time.RFC3339, strings.TrimSpace(endAtStr))
			if err != nil{
				helper.RespondError(c,400, "invalid_end_at", "end_at must be RFC3339")
				return
			}
			endAtPtr = &endAt
		}
		severity, _ := asString(payload["severity"])
		var severityPtr *string
		if severity != "" {severityPtr = &severity}

		metadata := payload["metadata"]

		if err := h.Events.InsertEvent(ctx, eventID, eventType, serviceID, envID, title, severityPtr, occurredAt, endAtPtr, metadata); err != nil{
			helper.RespondError(c,500, "db_write_failed", err.Error())
			return
		}
	default:
		helper.RespondError(c,400, "invalid_event_type", "unsupported event_type")
		return
	}


	// Response
	status := http.StatusCreated
	if pathID != ""{
		status = http.StatusOK
	}

	c.JSON(status, gin.H{
		"id": idStr,
		"event_type": eventType,
		"service_name": serviceName,
		"environment": environment,
		"title": title,
		"occurred_at": occurredAtStr,
	})
}

func asString(v any) (string, bool){
	s, ok := v.(string)
	return s, ok
}

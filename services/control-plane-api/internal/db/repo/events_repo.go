package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventsRepo struct {pool *pgxpool.Pool}

func NewEventsRepo(pool *pgxpool.Pool) *EventsRepo {return &EventsRepo{pool: pool}}

func (r *EventsRepo) InsertEvent(ctx context.Context, id uuid.UUID, eventType string, serviceID uuid.UUID, envID uuid.UUID, title string, severity *string, occurredAt time.Time, endAt *time.Time, metadata any) error{
	var metadataJSON []byte
	if metadata != nil{
		b, err := json.Marshal(metadata);
		if err != nil{
			return err
		}
		metadataJSON = b
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO events (id, event_type, service_id, environment_id, title, severity, occurred_at, end_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, id, eventType, serviceID, envID, title, severity, occurredAt, endAt, metadataJSON)

	return err
}

func (r *EventsRepo) UpsertDeploymentEvent(ctx context.Context, id uuid.UUID, serviceID uuid.UUID, envID uuid.UUID, version string, commitHash *string, actor *string, occuredAt time.Time) error{
	// Idempotent insert: ignore duplicates by event ID.
	_, err := r.pool.Exec(ctx, `
		INSERT INTO deploy_events (id, service_id, environment_id, version, commit_hash, deployed_by, deployed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, id, serviceID, envID, version, commitHash, actor, occuredAt)

	return err
}

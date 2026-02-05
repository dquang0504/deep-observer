package repo

import (
	"context"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardsRepo struct {
	pool *pgxpool.Pool
}

func NewDashboardsRepo(pool *pgxpool.Pool) *DashboardsRepo {
	return &DashboardsRepo{pool: pool}
}

func (r *DashboardsRepo) Create(ctx context.Context, req model.CreateDashboardRequest) (*model.Dashboard, error) {
	id := uuid.New().String()
	query := `
		INSERT INTO dashboards (id, name, description, grafana_uuid)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, grafana_uuid
	`
	var d model.Dashboard
	err := r.pool.QueryRow(ctx, query, id, req.Name, req.Description, req.GrafanaUUID).Scan(&d.ID, &d.Name, &d.Description, &d.GrafanaUUID)
	return &d, err
}

func (r *DashboardsRepo) List(ctx context.Context) ([]model.Dashboard, error) {
	query := `SELECT id, name, description, grafana_uuid FROM dashboards ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dashboards []model.Dashboard
	for rows.Next() {
		var d model.Dashboard
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.GrafanaUUID); err != nil {
			return nil, err
		}
		dashboards = append(dashboards, d)
	}
	return dashboards, nil
}

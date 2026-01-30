package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnvironmentRepo struct{ pool *pgxpool.Pool }

func NewEnvironmentRepo(pool *pgxpool.Pool) *EnvironmentRepo {
	return &EnvironmentRepo{pool: pool}
}

func (r *EnvironmentRepo) GetIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM environments WHERE name = $1`, name).Scan(&id)
	return id, err
}

func (r *EnvironmentRepo) List(ctx context.Context) ([]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT name, description FROM environments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []any
	for rows.Next() {
		var name string
		var desc *string
		if err := rows.Scan(&name, &desc); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"name": name, "description": desc})
	}
	return out, nil
}

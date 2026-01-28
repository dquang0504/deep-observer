package repo

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnvironmentRepo struct {pool *pgxpool.Pool}

func NewEnvironmentRepo(pool *pgxpool.Pool) *EnvironmentRepo {
	return &EnvironmentRepo{pool: pool}
}

func (r *EnvironmentRepo) GetIDByName(ctx context.Context, name string) (uuid.UUID, error){
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM environments WHERE name = $1`, name).Scan(&id)
	return id, err
}
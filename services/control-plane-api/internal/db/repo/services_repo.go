package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)


type ServicesRepo struct{pool *pgxpool.Pool}

func NewServicesRepo(pool *pgxpool.Pool) *ServicesRepo{return &ServicesRepo{pool: pool}}

func (r *ServicesRepo) GetIDByName(ctx context.Context, name string) (uuid.UUID, error){
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM services WHERE service_name = $1`, name).Scan(&id)
	return id, err
}

func (r *ServicesRepo) EnsureService(ctx context.Context, name string) (uuid.UUID, error){
	// EnsureService returns the service ID, creating the service if it doesn't exist.
	//fetch
	if id, err := r.GetIDByName(ctx, name); err == nil{
		return id, nil
	}

	//insert
	id := uuid.New()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO services (id, service_name) VALUES ($1, $2)
		ON CONFLICT (service_name) DO UPDATE SET service_name=EXCLUDED.service_name`,
		id, name,)
	if err != nil{
		return uuid.Nil, err
	}

	return r.GetIDByName(ctx, name)
}

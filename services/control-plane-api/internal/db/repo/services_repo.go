package repo

import (
	"context"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

//Todo: explain to me the need of creating models.go file.
//Todo: explain to me why we need to defer rows.Close after querying, will there be an error, say memory leak if we don't defer close row ?
//Todo: I need you to help me make this IDE capable of showing me a method's implementation when I hover over it. e.g rows.Next
//Todo: Explain to me why we need to create 2 extra variables lang and owner as type *string. Is it solely because the language and owner columns in the database are nullable? If so, can't we just do that in the model.Service struct?
//Todo: I thought we already handled lang and owner at rows.Scan method, why do we still need to check for != nil after rows.Scan ?
//Todo: At line 31, would it still be correct if we change var s model.Service to var s *model.Service ? If so, what would be the difference?
//Todo: At line 55, what is the purpose of QueryRow and what is the difference between QueryRow and Query ?

type ServicesRepo struct{pool *pgxpool.Pool}

func NewServicesRepo(pool *pgxpool.Pool) *ServicesRepo{return &ServicesRepo{pool: pool}}

func (r *ServicesRepo) List(ctx context.Context) ([]model.Service, error){
	query := `SELECT id, service_name, language, owner, created_at FROM services ORDER BY service_name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	var services []model.Service
	for rows.Next(){
		var s model.Service
		var lang, owner *string
		if err := rows.Scan(&s.ID, &s.ServiceName, &lang, &owner, &s.CreatedAt); err != nil{
			return nil, err
		}
		if lang != nil{
			s.Language = *lang
		}
		if owner != nil{
			s.Owner = *owner
		}
		services = append(services, s)
	}
	return services, nil
}

func (r *ServicesRepo) Create(ctx context.Context, req model.CreateServiceRequest) (*model.Service, error){
	id := uuid.New()
	query := `
		INSERT INTO services (id, service_name, language, owner) VALUES ($1, $2, $3, $4)
		RETURNING id, service_name, language, owner, created_at
	`
	var s model.Service
	var lang, owner *string
	err := r.pool.QueryRow(ctx, query, id, req.ServiceName, req.Language, req.Owner).Scan(&s.ID, &s.ServiceName, &lang, &owner, &s.CreatedAt)
	if err != nil{
		return nil, err
	}
	if lang != nil{
		s.Language = *lang
	}
	if owner != nil {
		s.Owner = *owner
	}
	return &s, nil
}  

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

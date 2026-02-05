package model

import (
	"time"
)

//Service struct represents a service record in the system
type Service struct{
	ID string `json:"id"`
	ServiceName string `json:"service_name"`
	Language string `json:"language"`
	Owner string `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

//CreateServiceRequest is used to receive request data from API
type CreateServiceRequest struct{
	ServiceName string `json:"service_name" binding:"required"`
	Language string `json:"language"`
	Owner string `json:"owner"`
}

//Dashbard struct represents metadata of a Grafana Dashboard
type Dashboard struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	GrafanaUUID string `json:"grafana_uuid"`
}

//CreateDashboardRequest is used to receive request data from API
type CreateDashboardRequest struct{
	Name string `json:"name" binding:"required"`
	Description string `json:"description"`
	GrafanaUUID string `json:"grafana_uuid" binding:"required"`
}


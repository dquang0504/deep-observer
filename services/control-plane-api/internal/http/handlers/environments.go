package handlers

import (
	"net/http"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/helper"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnvironmentsHandler struct {pool *pgxpool.Pool}
func NewEnvironmentsHandler(pool *pgxpool.Pool) *EnvironmentsHandler {return &EnvironmentsHandler{pool: pool}}

func (h *EnvironmentsHandler) ListEnvironments(c *gin.Context){
	rows, err := h.pool.Query(c.Request.Context(), `SELECT name, description FROM environments ORDER BY name`)
	if err != nil{ helper.RespondError(c, 500, "db_query_failed", err.Error()); return}
	defer rows.Close()

	out := make([]gin.H, 0)
	for rows.Next(){
		var name string
		var desc *string
		if err := rows.Scan(&name, &desc); err != nil{
			helper.RespondError(c, 500, "db_scan_failed", err.Error())
		}
		out = append(out, gin.H{"name": name, "description": desc})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}
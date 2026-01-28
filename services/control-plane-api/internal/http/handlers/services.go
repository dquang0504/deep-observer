package handlers

import (
	"net/http"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/helper"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServicesHandler struct {pool *pgxpool.Pool}
func NewServicesHandler(pool *pgxpool.Pool) *ServicesHandler {return &ServicesHandler{pool: pool}}

func (h *ServicesHandler) ListServices(c *gin.Context){
	rows, err := h.pool.Query(c.Request.Context(), `SELECT service_name, language, owner, created_at FROM services ORDER BY service_name`)
	if err != nil{ 
		helper.RespondError(c,500,"db_query_failed",err.Error())
	}
	defer rows.Close()

	out := make([]gin.H, 0)
	for rows.Next(){
		var name, lang, owner *string
		var createdAt any
		if err := rows.Scan(&name, &lang, &owner, &createdAt); err != nil{
			helper.RespondError(c,500,"db_scan_failed",err.Error())
			return;
		}
		out = append(out, gin.H{"service_name": name, "language": lang, "owner": owner, "created_at": createdAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}
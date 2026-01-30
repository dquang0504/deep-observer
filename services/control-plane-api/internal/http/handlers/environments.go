package handlers

import (
	"net/http"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/db/repo"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/helper"
	"github.com/gin-gonic/gin"
)

type EnvironmentsHandler struct {
	repo *repo.EnvironmentRepo
}

func NewEnvironmentsHandler(repo *repo.EnvironmentRepo) *EnvironmentsHandler {
	return &EnvironmentsHandler{repo: repo}
}

func (h *EnvironmentsHandler) ListEnvironments(c *gin.Context) {
	rows, err := h.repo.List(c.Request.Context())
	if err != nil {
		helper.RespondError(c, 500, "db_query_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

package handlers

import (
	"net/http"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/db/repo"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/helper"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/model"
	"github.com/gin-gonic/gin"
)

type ServicesHandler struct {
	repo *repo.ServicesRepo
}

func NewServicesHandler(repo *repo.ServicesRepo) *ServicesHandler {
	return &ServicesHandler{repo: repo}
}

func (h *ServicesHandler) ListServices(c *gin.Context) {
	rows, err := h.repo.List(c.Request.Context())
	if err != nil {
		helper.RespondError(c, 500, "db_query_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *ServicesHandler) CreateService(c *gin.Context) {
	var req model.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, 400, "invalid_request", err.Error())
		return
	}

	service, err := h.repo.Create(c.Request.Context(), req)
	if err != nil {
		helper.RespondError(c, 500, "db_insert_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, service)
}

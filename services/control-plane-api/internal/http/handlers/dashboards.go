package handlers

import (
	"net/http"

	"github.com/dquang0504/deep-observer/control-plane-api/internal/db/repo"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/helper"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/model"
	"github.com/gin-gonic/gin"
)


type DashboardsHandler struct{
	repo *repo.DashboardsRepo
}

func NewDashboardHandler(repo *repo.DashboardsRepo) *DashboardsHandler{
	return &DashboardsHandler{repo: repo}
}

func (h *DashboardsHandler) ListDashboards(c *gin.Context) {
	items, err := h.repo.List(c.Request.Context());
	if err != nil{
		helper.RespondError(c,500, "db_query_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *DashboardsHandler) CreateDashboard(c *gin.Context){
	var req model.CreateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil{
		helper.RespondError(c, 400, "invalid_request", err.Error())
		return
	}
	item, err := h.repo.Create(c.Request.Context(), req);
	if err != nil{
		helper.RespondError(c, 500, "db_insert_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, item)
}
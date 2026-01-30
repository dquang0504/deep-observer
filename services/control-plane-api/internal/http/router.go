package http

import (
	"github.com/dquang0504/deep-observer/control-plane-api/internal/http/handlers"
	"github.com/gin-gonic/gin"
)


type Deps struct{
	EventsHandler *handlers.EventsHandler
	ServicesHandler *handlers.ServicesHandler
  	EnvironmentsHandler *handlers.EnvironmentsHandler
}

func NewRouter(d Deps) *gin.Engine{
	r := gin.New()
	r.Use(gin.Recovery())
	// add request_id middleware later

	v1 := r.Group("/api/v1")
	{
		v1.POST("/events", d.EventsHandler.PostEvent)
		v1.PUT("/events/:id", d.EventsHandler.PutEvent)

		v1.GET("/services", d.ServicesHandler.ListServices)
		v1.POST("/services", d.ServicesHandler.CreateService)
		v1.GET("/environments", d.EnvironmentsHandler.ListEnvironments)
	}

	r.GET("/healthz", func(c *gin.Context) {c.JSON(200, gin.H{"ok": true})})
	return r
}
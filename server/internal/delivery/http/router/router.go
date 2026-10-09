package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/delivery/http/handler"
)

func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	healthHandler := handler.NewHealthHandler()
	r.GET("/health/live", healthHandler.Live)

	return r
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/delivery/http/handler"
)

func New(readiness handler.ReadinessChecker) (*gin.Engine, error) {
	r := gin.New()

	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	r.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler(readiness)
	r.GET("/health/live", healthHandler.Live)
	r.GET("/health/ready", healthHandler.Ready)

	return r, nil
}

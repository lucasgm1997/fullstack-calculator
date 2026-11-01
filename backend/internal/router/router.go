// Package router wires HTTP routes to handlers.
package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"calculator-api/internal/config"
	"calculator-api/internal/docs"
	"calculator-api/internal/dto"
	"calculator-api/internal/handler"
	"calculator-api/internal/middleware"
	"calculator-api/internal/usecase"
)

const healthPath = "/api/health"

// New builds the Gin engine with all routes and middleware.
func New(cfg config.Config) (*gin.Engine, error) {
	r := gin.New()
	r.HandleMethodNotAllowed = true

	// Only these proxies may set X-Forwarded-For; otherwise a client could
	// spoof its IP and dodge the rate limit.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("invalid trusted proxies: %w", err)
	}

	r.Use(gin.Logger(), gin.Recovery())
	// Applied to every route (including 404s) except the health check, which
	// the UI and orchestrators poll.
	r.Use(middleware.RateLimit(middleware.RateLimitConfig{
		Requests: cfg.RateLimitRequests,
		Window:   cfg.RateLimitWindow,
		Exempt:   []string{healthPath},
	}))

	calculatorHandler := handler.NewCalculatorHandler(usecase.NewCalculator())

	r.GET(healthPath, handler.Health)
	api := r.Group("/api")
	api.POST("/calculate", calculatorHandler.Calculate)
	api.GET("/openapi.yaml", docs.Spec)
	api.GET("/docs", docs.UI)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "method not allowed"})
	})

	return r, nil
}

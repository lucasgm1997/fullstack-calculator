// Package router wires HTTP routes to handlers.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"calculator-api/internal/docs"
	"calculator-api/internal/dto"
	"calculator-api/internal/handler"
	"calculator-api/internal/usecase"
)

// New builds the Gin engine with all routes and middleware.
func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.HandleMethodNotAllowed = true

	calculatorHandler := handler.NewCalculatorHandler(usecase.NewCalculator())

	api := r.Group("/api")
	api.GET("/health", handler.Health)
	api.POST("/calculate", calculatorHandler.Calculate)
	api.GET("/openapi.yaml", docs.Spec)
	api.GET("/docs", docs.UI)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "method not allowed"})
	})

	return r
}

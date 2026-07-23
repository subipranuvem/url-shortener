package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/subipraNuvem/url-shortener/internal/cache"
	"github.com/subipraNuvem/url-shortener/internal/database"
)

type HealthHandler struct {
	db    database.Database
	cache cache.Cache
}

type healthStatusResponse struct {
	Status healthServicesStatus `json:"status"`
}

type healthServicesStatus struct {
	Database string `json:"database"`
	Cache    string `json:"cache"`
}

func NewHealthHandler(db database.Database, cache cache.Cache) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

func (h *HealthHandler) RegisterEndpoints(r gin.IRouter) {
	r.GET("/health", h.Health)
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	dbErr := h.db.Ping(ctx)
	cacheErr := h.cache.Ping(ctx)

	statusCode := http.StatusOK
	if dbErr != nil || cacheErr != nil {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, healthStatusResponse{
		Status: healthServicesStatus{
			Database: errString(dbErr),
			Cache:    errString(cacheErr),
		},
	})
}

func errString(err error) string {
	if err != nil {
		return err.Error()
	}
	return "ok"
}

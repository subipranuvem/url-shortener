package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/subipraNuvem/url-shortener/internal/service"
)

type RedirectHandler struct {
	svc *service.URLService
}

func NewRedirectHandler(svc *service.URLService) *RedirectHandler {
	return &RedirectHandler{svc: svc}
}

func (h *RedirectHandler) RegisterEndpoints(r gin.IRouter) {
	r.GET("/:code", h.Redirect)
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param(PathParamCode)
	longURL, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		httpErr := GetHTTPErrorWithFeedbackByError(err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}
	c.Redirect(http.StatusFound, longURL)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/subipraNuvem/url-shortener/internal/service"
)

type URLHandler struct {
	svc *service.URLService
}

func NewURLHandler(svc *service.URLService) *URLHandler {
	return &URLHandler{svc: svc}
}

func (h *URLHandler) RegisterEndpoints(r gin.IRouter) {
	urls := r.Group("/urls")
	urls.POST("", h.Create)
	urls.GET("/:code", h.GetByCode)
	urls.GET("/:code/stats", h.GetStats)
	urls.DELETE("/:code", h.Deactivate)
}

type createURLRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
	Alias   string `json:"alias"`
}

func (h *URLHandler) Create(c *gin.Context) {
	var req createURLRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		httpErr := GetHTTPErrorWithFeedbackByError(err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}

	out, err := h.svc.Create(c.Request.Context(), service.CreateURLInput{
		LongURL: req.LongURL,
		Alias:   req.Alias,
	})
	if err != nil {
		httpErr := GetHTTPErrorWithFeedbackByError(err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		RespKeyCode:     out.Code,
		RespKeyShortURL: out.ShortURL,
	})
}

func (h *URLHandler) GetByCode(c *gin.Context) {
	code := c.Param(PathParamCode)
	url, err := h.svc.GetByCode(c.Request.Context(), code)
	if err != nil {
		httpErr := GetHTTPErrorWithFeedbackByError(err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}
	c.JSON(http.StatusOK, url)
}

func (h *URLHandler) GetStats(c *gin.Context) {
	code := c.Param(PathParamCode)
	clicks, err := h.svc.GetStats(c.Request.Context(), code)
	if err != nil {
		httpErr := GetHTTPErrorWithFeedbackByError(err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{RespKeyCode: code, RespKeyClickCount: clicks})
}

func (h *URLHandler) Deactivate(c *gin.Context) {
	code := c.Param(PathParamCode)
	err := h.svc.Deactivate(c.Request.Context(), code)
	if err != nil {
		httpErr := GetHTTPErrorWithFeedbackByError(err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}
	c.Status(http.StatusNoContent)
}

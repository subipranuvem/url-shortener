package main

import (
	"github.com/gin-gonic/gin"
	"github.com/subipraNuvem/url-shortener/internal/api"
	"github.com/subipraNuvem/url-shortener/internal/api/handler"
)

func main() {
	gin.SetMode(gin.TestMode)

	router := api.NewRouter(
		handler.NewURLHandler(nil),
		handler.NewRedirectHandler(nil),
		handler.NewHealthHandler(nil, nil),
	)

	api.PrintRoutes(router)
}

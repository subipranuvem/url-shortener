package api

import (
	"fmt"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type HandlerRegister interface {
	RegisterEndpoints(r gin.IRouter)
}

func NewRouter(handlers ...HandlerRegister) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	for _, h := range handlers {
		h.RegisterEndpoints(r)
	}

	return r
}

func PrintRoutes(r *gin.Engine) {
	for _, route := range r.Routes() {
		fmt.Printf("%-8s %s\n", route.Method, route.Path)
	}
}

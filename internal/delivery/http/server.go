package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func NewServer(addr string, handler *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

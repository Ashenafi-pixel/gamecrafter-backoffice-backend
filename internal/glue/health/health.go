package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/health"
)

func Init(group *gin.RouterGroup) {
	h := health.NewHandler()
	routes := []struct {
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{http.MethodGet, "/api/admin/health", h.Health},
	}
	for _, r := range routes {
		group.Handle(r.method, r.path, r.handler)
	}
}

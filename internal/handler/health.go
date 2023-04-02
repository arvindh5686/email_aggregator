package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// HealthCheckHandler - to be used for health check. e.g liveness checks for pods in k8s
func HealthCheckHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Note: we should be doing deep checks here.
		context.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	}
}

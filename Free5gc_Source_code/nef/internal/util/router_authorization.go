package util

import (
	
	"net/http"
	

	"github.com/gin-gonic/gin"
	// "github.com/golang-jwt/jwt/v4"
	qos_models "github.com/free5gc/nef/internal/context"

	nef_context "github.com/free5gc/nef/internal/context"
	"github.com/free5gc/nef/internal/logger"
	
)

type RouterAuthorizationCheck struct {
	serviceName qos_models.ServiceName
}

func NewRouterAuthorizationCheck(serviceName qos_models.ServiceName) *RouterAuthorizationCheck {
	return &RouterAuthorizationCheck{
		serviceName: serviceName,
	}
}

// Middleware or handler that calls the validator
func Check(c *gin.Context, nefContext *nef_context.NefContext) {
	token := c.GetHeader("Authorization")
	logger.UtilLog.Debugf("Authorization Header: %s", token)

	err := nefContext.AuthorizationCheck(token)
	if err != nil {
		logger.UtilLog.Warnf("RouterAuthorizationCheck::Check Unauthorized: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	logger.UtilLog.Debugf("RouterAuthorizationCheck::Check Authorized for token: %s", token)
}


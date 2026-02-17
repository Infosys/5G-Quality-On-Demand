package sbi

import (
	"net/http"

	qos_models "github.com/free5gc/nef/internal/context"

	processor "github.com/free5gc/nef/internal/sbi/processor"
	"github.com/gin-gonic/gin"
)

func (s *Server) getOauthEndpoints() []Endpoint {
	return []Endpoint{
		{
			Method:  http.MethodPost,
			Pattern: "/token",
			APIFunc: s.apiIssueOAuthToken,
		},
	}
}
func (s *Server) apiIssueOAuthToken(gc *gin.Context) {
	contentType := gc.GetHeader("Content-Type")

	if contentType != "application/x-www-form-urlencoded" {
		gc.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":             "unsupported_content_type",
			"error_description": "Expected application/x-www-form-urlencoded",
		})
		return
	}

	// Parse form values
	if err := gc.Request.ParseForm(); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Failed to parse form data",
		})
		return
	}

	clientID := gc.Request.FormValue("client_id")
	clientSecret := gc.Request.FormValue("client_secret")
	grantType := gc.Request.FormValue("grant_type")

	// Validate required fields
	if clientID == "" || clientSecret == "" || grantType == "" {
		gc.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
		return
	}
	// Construct a struct to pass to your processor
	// Define the struct inline if not present in qos_models
	
	authReq := &qos_models.AuthorizationJSON{
		Client_id:     clientID,
		Client_secret: clientSecret,
		Grant_type:    grantType,
	}

	// Delegate to processor (you handle validation/token logic here)
	hdlRsp := s.Processor().IssueOAuthToken(authReq)

	// Convert *processor.Response to *processor.HandlerResponse if needed
	var handlerRsp *processor.Response
	if resp, ok := any(hdlRsp).(*processor.Response); ok {
		handlerRsp = resp
	} else {
		// If conversion is not possible, handle error or create a new HandlerResponse
		handlerRsp = &processor.Response{
			Status:  hdlRsp.Status,
			Body:    hdlRsp.Body,
			
		}
	}

	// Send back the token response
	var handlerResp *processor.HandlerResponse
	if resp, ok := any(handlerRsp).(*processor.HandlerResponse); ok {
		handlerResp = resp
	} else {
		handlerResp = &processor.HandlerResponse{
			Status: handlerRsp.Status,
			Body:   handlerRsp.Body,
		}
	}
	s.buildAndSendHttpResponse(gc, handlerResp, false)
}

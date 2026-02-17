package processor

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"time"

	qos_models "github.com/free5gc/nef/internal/context"
)

// TokenResponse represents a successful OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type Response struct {
	Status int
	Body   interface{}
}

var (
	validClientID     = "scs123"
	validClientSecret = "secret123"
	jwtSigningKey     = []byte("super-secret-key") // Replace with secure value in prod
)

func (p *Processor) IssueOAuthToken(authReq *qos_models.AuthorizationJSON) *Response {
	if authReq.Grant_type != "client_credentials" {
		return &Response{
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error":             "unsupported_grant_type",
				"error_description": "Only client_credentials is supported",
			},
		}
	}

	if authReq.Client_id != validClientID {
		return &Response{
			Status: http.StatusUnauthorized,
			Body: map[string]interface{}{
				"error":             "invalid_client",
				"error_description": "Unknown client_id",
			},
		}
	}

	if authReq.Client_secret != validClientSecret {
		return &Response{
			Status: http.StatusUnauthorized,
			Body: map[string]interface{}{
				"error":             "invalid_client",
				"error_description": "Invalid client_secret",
			},
		}
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   "nef-oauth",
		"sub":   authReq.Client_id,
		"aud":   "nef-client",
		"scope": "qos-control",
		"iat":   time.Now().Unix(),
	})

	tokenString, err := token.SignedString(jwtSigningKey)
	if err != nil {
		return &Response{
			Status: http.StatusInternalServerError,
			Body: map[string]interface{}{
				"error":             "server_error",
				"error_description": "Could not generate token",
			},
		}
	}

	return &Response{
		Status: http.StatusOK,
		Body: map[string]interface{}{
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   0,
			"scope":        "qos-control",
		},
	}
}

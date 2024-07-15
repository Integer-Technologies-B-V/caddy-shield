package shield

import (
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"go.uber.org/zap"
)

type Authenticator interface {
	Authenticated(r *http.Request) bool
}

type AuthenticatorSuperTokens struct {
	logger                    *zap.Logger
	cookieKey                 string
	mutex                     sync.RWMutex
	jwksCache                 *sessmodels.GetJWKSResult
	jwkCacheMaxAgeMiliseconds int64
	coreURL                   string
}

func NewAuthenticatorSuperTokens(logger *zap.Logger, coreURL string) *AuthenticatorSuperTokens {
	cookieKey := "sAccessToken"
	return &AuthenticatorSuperTokens{
		cookieKey:                 cookieKey,
		logger:                    logger,
		jwksCache:                 nil,
		jwkCacheMaxAgeMiliseconds: 60000,
		coreURL:                   coreURL,
	}
}

func (a *AuthenticatorSuperTokens) Authenticated(r *http.Request) bool {
	c, err := r.Cookie(a.cookieKey)
	if err != nil {
		return false
	}
	jwtToken := c.Value

	jwks, err := a.getJWKS()
	if err != nil {
		return false
	}
	parsedToken, err := jwt.Parse(jwtToken, jwks.Keyfunc)
	if err != nil {
		return false
	}

	if !parsedToken.Valid {
		return false
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Convert the claims to a key-value pair
	claimsMap := make(map[string]interface{})
	for key, value := range claims {
		claimsMap[key] = value
	}
	return true
}

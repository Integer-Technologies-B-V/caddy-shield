package shield

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Authenticator struct {
	logger    *zap.Logger
	cookieKey string
}

func NewAuthenticator(logger *zap.Logger) *Authenticator {
	cookieKey := "sAccessToken"
	return &Authenticator{cookieKey: cookieKey, logger: logger}
}

func (a *Authenticator) Authenticated(r *http.Request) bool {
	c, err := r.Cookie(a.cookieKey)
	if err != nil {
		return false
	}
	jwtToken := c.Value

	jwks, err := GetJWKS()
	if err != nil {
		a.logger.Error("err getting JWKS", zap.Error(err))
		return false
	}
	parsedToken, err := jwt.Parse(jwtToken, jwks.Keyfunc)
	if err != nil {
		a.logger.Debug("jwtToken:", zap.String("jwt_token:", jwtToken))
		a.logger.Debug("jwks:", zap.Array("jwks:", zapcore.ArrayMarshalerFunc(
			func(ae zapcore.ArrayEncoder) error {
				for _, a := range jwks.KIDs() {
					ae.AppendString(a)
				}
				return nil
			},
		)))
		a.logger.Error("parseError", zap.Error(err))
		return false
	}

	if !parsedToken.Valid {
		return false
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		a.logger.Error("claim type assert was not ok")
		return false
	}

	// Convert the claims to a key-value pair
	claimsMap := make(map[string]interface{})
	for key, value := range claims {
		claimsMap[key] = value
	}
	return true
}

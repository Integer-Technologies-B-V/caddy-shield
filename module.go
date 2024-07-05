package shield

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DBPoolKey string = "shield_db"
)

var resourcePool *caddy.UsagePool = caddy.NewUsagePool()

func init() {
	caddy.RegisterModule(ShieldMiddleware{})
}

// ShieldMiddleware implements an HTTP handler that ... something something
type ShieldMiddleware struct {
	authClient *auth
}

// Cleanup implements caddy.CleanerUpper.
func (m *ShieldMiddleware) Cleanup() error {
	if _, err := resourcePool.Delete(DBPoolKey); err != nil {
		return err
	}
	return nil
}

// CaddyModule returns the Caddy module information.
func (ShieldMiddleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.shield",
		New: func() caddy.Module { return new(ShieldMiddleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *ShieldMiddleware) Provision(ctx caddy.Context) error {
	resourcePool = caddy.NewUsagePool()
	_, _, err := resourcePool.LoadOrNew(DBPoolKey, constructDB)
	if err != nil {
		return fmt.Errorf("couldn't init db pool")
	}

	m.authClient = NewAuth()
	return nil
}

// Validate implements caddy.Validator.
func (m *ShieldMiddleware) Validate() error {
	v, _, _ := resourcePool.LoadOrNew(DBPoolKey, constructDB)
	if db, ok := v.(*pgxpool.Pool); !ok {
		return fmt.Errorf("invalid value in %s", DBPoolKey)
	} else if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("can't ping db")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m ShieldMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	token := m.GetToken(r)
	if !m.authClient.Authenticated(token) {
		w.WriteHeader(401)
		return nil
	}

	rp := reverseproxy.Handler{
		Upstreams: reverseproxy.UpstreamPool{
			{
				Dial: "localhost:6969",
			},
		},
	}
	return rp.ServeHTTP(w, r, next)
}

func (m ShieldMiddleware) GetToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ShieldMiddleware)(nil)
	_ caddy.Validator             = (*ShieldMiddleware)(nil)
	_ caddy.CleanerUpper          = (*ShieldMiddleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*ShieldMiddleware)(nil)
)

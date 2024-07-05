package shield

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DBPoolKey string = "shield_db"
)

func init() {
	caddy.RegisterModule(ShieldMiddleware{})
}

// ShieldMiddleware implements an HTTP handler that ... something
type ShieldMiddleware struct {
	resourcePool *caddy.UsagePool
}

// Cleanup implements caddy.CleanerUpper.
func (m *ShieldMiddleware) Cleanup() error {
	if _, err := m.resourcePool.Delete(DBPoolKey); err != nil {
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
	m.resourcePool = caddy.NewUsagePool()
	_, _, err := m.resourcePool.LoadOrNew(DBPoolKey, constructDB)
	if err != nil {
		return fmt.Errorf("couldn't init db pool")
	}
	return nil
}

// Validate implements caddy.Validator.
func (m *ShieldMiddleware) Validate() error {
	v, _, _ := m.resourcePool.LoadOrNew(DBPoolKey, constructDB)
	if db, ok := v.(*pgxpool.Pool); !ok {
		return fmt.Errorf("invalid value in %s", DBPoolKey)
	} else if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("can't ping db")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m ShieldMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return next.ServeHTTP(w, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ShieldMiddleware)(nil)
	_ caddy.CleanerUpper          = (*ShieldMiddleware)(nil)
	_ caddy.Validator             = (*ShieldMiddleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*ShieldMiddleware)(nil)
)

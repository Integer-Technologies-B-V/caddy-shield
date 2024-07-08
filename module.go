package shield

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	DBPoolKey string = "shield_db"
)

var resourcePool *caddy.UsagePool = caddy.NewUsagePool()

func init() {
	caddy.RegisterModule(ShieldMiddleware{})
}

// ShieldMiddleware implements an HTTP handler that authenticates requests and looks up the upstream to which
// the request  should be proxied to
type ShieldMiddleware struct {
	ctx    caddy.Context
	logger *zap.Logger

	pgxPool    *pgxpool.Pool
	authClient *auth

	ReverseProxy *reverseproxy.Handler
}

// CaddyModule returns the Caddy module information.
func (ShieldMiddleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.shield",
		New: func() caddy.Module { return new(ShieldMiddleware) },
	}
}

// Provision implements caddy.Provisioner
func (m *ShieldMiddleware) Provision(ctx caddy.Context) error {
	m.ctx = ctx
	m.logger = ctx.Logger()

	m.ReverseProxy = &reverseproxy.Handler{}
	if m.ReverseProxy != nil {
		if err := m.ReverseProxy.Provision(ctx); err != nil {
			return fmt.Errorf("provision reverse proxy, %v", err)
		}
	}

	db, loaded, err := resourcePool.LoadOrNew(DBPoolKey, func() (caddy.Destructor, error) {
		d, err := getDB(ctx)
		if err != nil {
			return nil, err
		}
		return dbDestructor{Pool: d}, nil
	})
	if err != nil {
		m.logger.Error("loading database connections pool", zap.String("db_key", DBPoolKey), zap.Error(err))
		return err
	}

	if loaded {
		m.logger.Info("using loaded db connections pool")
	}

	dbDesctructor, ok := db.(dbDestructor)
	if !ok {
		m.logger.Error("couldn't unmarshal to pgx pool")
	}
	m.pgxPool = dbDesctructor.Pool
	m.authClient = NewAuth()
	return err
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (m ShieldMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	token := m.GetToken(r)
	if !m.authClient.Authenticated(token) {
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	m.ReverseProxy.Upstreams = reverseproxy.UpstreamPool{
		&reverseproxy.Upstream{
			Dial: "localhost:8000", // in reality this will not be hardcode but fetched from a database / service
		},
	}
	return m.ReverseProxy.ServeHTTP(w, r, next)
}

func (m ShieldMiddleware) GetToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (m *ShieldMiddleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name
	return nil
}

// Cleanup implements caddy.CleanerUpper
func (m *ShieldMiddleware) Cleanup() error {
	deleted, err := resourcePool.Delete(DBPoolKey)
	if deleted {
		m.logger.Debug("unloading unused database", zap.String("db_key", DBPoolKey))
	}
	if err != nil {
		m.logger.Error("closing database", zap.String("db_key", DBPoolKey), zap.Error(err))
	}
	return err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ShieldMiddleware)(nil)
	_ caddy.CleanerUpper          = (*ShieldMiddleware)(nil)
	_ caddyfile.Unmarshaler       = (*ShieldMiddleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*ShieldMiddleware)(nil)
)

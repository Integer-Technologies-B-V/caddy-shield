package shield

import (
	"net/http"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	UsagePoolKeyDatabase string = "pgxpool"
)

var usagePool *caddy.UsagePool = caddy.NewUsagePool()

func init() {
	godotenv.Load()
	caddy.RegisterModule(ShieldUpstreams{})
}

// ShieldMiddleware implements an HTTP handler that authenticates requests and looks up the upstream to which
// the request  should be proxied to
type ShieldUpstreams struct {
	ctx           caddy.Context
	logger        *zap.Logger
	pgxPool       *pgxpool.Pool
	authenticator Authenticator
}

// Cleanup implements caddy.CleanerUpper.
func (m *ShieldUpstreams) Cleanup() error {
	deleted, err := usagePool.Delete(UsagePoolKeyDatabase)
	if err != nil {
		m.logger.Error("closing database", zap.String("db_key", UsagePoolKeyDatabase), zap.Error(err))
		return err
	}
	if deleted {
		m.logger.Debug("unloading unused database", zap.String("db_key", UsagePoolKeyDatabase))
	}
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *ShieldUpstreams) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive
	return nil
}

// GetUpstreams implements reverseproxy.UpstreamSource.
func (m *ShieldUpstreams) GetUpstreams(r *http.Request) ([]*reverseproxy.Upstream, error) {
	if !m.authenticator.Authenticated(r) {
		return []*reverseproxy.Upstream{{Dial: "localhost:3000"}}, nil
	}
	return []*reverseproxy.Upstream{{Dial: "localhost:8000"}}, nil
}

// CaddyModule returns the Caddy module information.
func (ShieldUpstreams) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.reverse_proxy.upstreams.shield",
		New: func() caddy.Module { return new(ShieldUpstreams) },
	}
}

// Provision implements caddy.Provisioner
func (m *ShieldUpstreams) Provision(ctx caddy.Context) error {
	supertokensURL := os.Getenv("SUPERTOKENS_URL")
	dbURL := os.Getenv("DATABASE_URL")
	m.ctx = ctx
	m.logger = ctx.Logger()
	m.authenticator = NewAuthenticatorSuperTokens(ctx.Logger(), supertokensURL)

	db, loaded, err := usagePool.LoadOrNew(UsagePoolKeyDatabase, func() (caddy.Destructor, error) {
		pgdb, err := connectToDB(ctx, dbURL)
		if err != nil {
			return nil, err
		}
		return dbDestructor{db: pgdb}, nil
	})
	if err != nil {
		m.logger.Error("loading database connections pool", zap.String("db_key", UsagePoolKeyDatabase), zap.Error(err))
		return err
	}

	if loaded {
		m.logger.Info("using loaded db connections pool")
	}

	dbDesctructor, ok := db.(dbDestructor)
	if !ok {
		m.logger.Error("couldn't unmarshal to pgx pool")
	}
	m.pgxPool = dbDesctructor.db
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ShieldUpstreams)(nil)
	_ caddy.CleanerUpper          = (*ShieldUpstreams)(nil)
	_ caddyfile.Unmarshaler       = (*ShieldUpstreams)(nil)
	_ reverseproxy.UpstreamSource = (*ShieldUpstreams)(nil)
)

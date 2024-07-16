package shield

import (
	"net/http"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	godotenv.Load()
	caddy.RegisterModule(ShieldUpstreams{})
}

// ShieldMiddleware implements an HTTP handler that authenticates requests and looks up the upstream to which
// the request  should be proxied to
type ShieldUpstreams struct {
	ctx              caddy.Context
	logger           *zap.Logger
	fallbackUpstream []*reverseproxy.Upstream

	upstreamService UpstreamsProvider
	authenticator   Authenticator
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *ShieldUpstreams) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive
	return nil
}

// GetUpstreams implements reverseproxy.UpstreamSource.
func (m *ShieldUpstreams) GetUpstreams(r *http.Request) ([]*reverseproxy.Upstream, error) {
	if !m.authenticator.Authenticated(r) {
		return m.fallbackUpstream, nil
	}
	upstreams, err := m.upstreamService.UpstreamsFromHostname(r.Context(), r.Host)
	if err != nil {
		return m.fallbackUpstream, nil
	}
	if len(upstreams) == 0 {
		return m.fallbackUpstream, nil
	}
	return upstreams, nil
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
	fallbackUpstream := os.Getenv("FALLBACK_UPSTREAM")
	clientID := os.Getenv("UPSTREAM_SERVICE_CLIENT_ID")
	clientSecret := os.Getenv("UPSTREAM_SERVICE_CLIENT_SECRET")

	m.ctx = ctx
	m.logger = ctx.Logger()
	m.fallbackUpstream = []*reverseproxy.Upstream{{Dial: fallbackUpstream}}

	m.upstreamService = NewUpstreamsService(clientID, clientSecret)
	m.authenticator = NewAuthenticatorSuperTokens(ctx.Logger(), supertokensURL)
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ShieldUpstreams)(nil)
	_ caddyfile.Unmarshaler       = (*ShieldUpstreams)(nil)
	_ reverseproxy.UpstreamSource = (*ShieldUpstreams)(nil)
)

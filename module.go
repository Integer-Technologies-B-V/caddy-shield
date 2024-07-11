package shield

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ShieldUpstreams{})
}

// ShieldMiddleware implements an HTTP handler that authenticates requests and looks up the upstream to which
// the request  should be proxied to
type ShieldUpstreams struct {
	ctx    caddy.Context
	logger *zap.Logger
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *ShieldUpstreams) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume
	return nil
}

// GetUpstreams implements reverseproxy.UpstreamSource.
func (m *ShieldUpstreams) GetUpstreams(r *http.Request) ([]*reverseproxy.Upstream, error) {
	return []*reverseproxy.Upstream{{Dial: "100.116.76.46:8000"}}, nil
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
	m.ctx = ctx
	m.logger = ctx.Logger()
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ShieldUpstreams)(nil)
	_ caddyfile.Unmarshaler       = (*ShieldUpstreams)(nil)
	_ reverseproxy.UpstreamSource = (*ShieldUpstreams)(nil)
)

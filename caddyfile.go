package shield

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("shield", parseCaddyfile)
}

// parseCaddyfile unmarshals tokens from h into a new ShieldMiddleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m ShieldMiddleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

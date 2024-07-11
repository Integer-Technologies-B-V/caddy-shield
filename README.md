# caddy-shield

caddy-shield is a caddy reverse proxy which implements custom auth relevant ONLY for
the Link product from Integer. It authenticates user requests by reading the Authorization header and validating it against the [`Auth Provider`](https://supertokens.com), then maps the subdomain request to an IP which the reverse proxy can serve to.

## Building locally
If you want to further develop the plugin follow the instructions in order to build and test

### `xcaddy` CLI

To build caddy-shield locally, install [`xcaddy`](xcaddy), then build from
the directory root. Examples:

```shell
xcaddy build --with github.com/integer-technologies-b-v/caddy-shield=.
```

this will output a caddy binary which can be run as you would run the official image

[xcaddy]: https://github.com/caddyserver/xcaddy
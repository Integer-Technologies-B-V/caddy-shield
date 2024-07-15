# caddy-shield

caddy-shield is a dynamic upstream module for caddy's reverse proxy directive. It provides upstreams based on the Host header of the request. It also checks for authentication of the users trying to access the upstream esentially making it a authentication gateway. If user is authenticated it will allow for access to the internal resource.

## Building locally
If you want to further develop the plugin follow the instructions in order to build and run it

### `xcaddy` CLI

To build caddy-shield locally, install [`xcaddy`](https://github.com/caddyserver/xcaddy). Clone this repo and build from
the root directory. Make the `.env` file and add the required variables (see `.env.example`) Examples:

```shell
xcaddy build --with github.com/integer-technologies-b-v/caddy-shield=.
```

and the `.env` file (don't forget to change with yours)
```shell
echo "DATABASE_URL=postgres://{user}:{password}@{host}:{port}/{database}?sslmode=disable>" >> .env
echo "SUPERTOKENS_URL=https://try.supertokens.com/.well-known/jwks.json" >> .env
```

Then you can run the output binary by xcaddy as any other binary. In this case it is a caddy binary compiled with your module. Example:

```shell
./caddy run --config Caddyfile 
```
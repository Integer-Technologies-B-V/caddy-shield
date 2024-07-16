package shield

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/go-resty/resty/v2"
)

type UpstreamsProvider interface {
	UpstreamsFromHostname(ctx context.Context, host string) ([]*reverseproxy.Upstream, error)
}

type UpstreamsService struct {
	clientID     string
	clientSecret string
	requestURL   string

	client *resty.Client
}

func NewUpstreamsService(clientID, clientSecret string) *UpstreamsService {
	const requestURL = "https://api.example.com/resource"
	client := resty.New()

	return &UpstreamsService{
		client:       client,
		requestURL:   requestURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (u *UpstreamsService) do(ctx context.Context, host string) (*http.Response, error) {
	resp, err := u.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetBasicAuth(u.clientID, u.clientSecret).
		SetQueryParam("host", host).
		Get(u.requestURL)
	if err != nil {
		return nil, nil
	}
	return resp.RawResponse, nil

}

type upstreamsResponse struct {
	Upstreams []string `json:"upstreams,omitempty"`
}

func (u *UpstreamsService) UpstreamsFromHostname(ctx context.Context, host string) ([]*reverseproxy.Upstream, error) {
	// to test without an upstream service provider just uncomment the bottom line
	// return []*reverseproxy.Upstream{{Dial: "localhost:8000"}}, nil
	resp, err := u.do(ctx, host)
	if err != nil {
		return nil, err
	}
	body := upstreamsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	upstreams := []*reverseproxy.Upstream{}
	for _, ups := range body.Upstreams {
		upstreams = append(upstreams, &reverseproxy.Upstream{
			Dial: ups,
		})
	}
	return upstreams, nil
}

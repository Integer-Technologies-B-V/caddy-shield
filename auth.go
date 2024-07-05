package shield

type auth struct {
	client struct{}
}

func NewAuth() *auth {
	return &auth{
		client: struct{}{},
	}
}
func (a *auth) Authenticated(token string) bool {
	return true
}

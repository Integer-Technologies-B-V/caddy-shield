package shield

type Authenticator struct {
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{}
}

func (a *Authenticator) Authenticated(token string) bool {
	return true
}

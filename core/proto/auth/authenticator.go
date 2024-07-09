package auth

import "net/http"

const (
	Success = Result(iota)
	Fail
	AuthServerUnavailible

	url = "https://sessionserver.mojang.com/session/minecraft/hasJoined?"
)

type Authenticator struct {
	user         string
	sharedSecret []byte
	publicKey    []byte
}

func NewAuthenticator(user string, sharedSecret []byte, publicKey []byte) *Authenticator {
	return &Authenticator{user, sharedSecret, publicKey}
}

type Result byte

type AuthenticationResult struct {
	Result     Result
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Properties *Properties `json:"properties"`
}

type Properties struct {
	Value     string `json:"value"`
	Signature string `json:"signature"`
}

func (a Authenticator) Authenticate() (*AuthenticationResult, error) {
	requestUrl := url + "name=" + a.user + "&serverId=" + a.generateServerHash()
	resp, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}

	result := &AuthenticationResult{}

	switch resp.StatusCode {
	case 200:
		{
			result.Result = Success
		}
	case 204:
		{
			result.Result = Fail
		}
	case 503:
		{
			result.Result = AuthServerUnavailible
		}
	}
}

func (a Authenticator) generateServerHash() string {
	return ""
}

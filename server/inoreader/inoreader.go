package inoreader

import (
	"net/http"
	"net/url"

	"github.com/gong023/my-slack-process/oauth"
	"github.com/gong023/my-slack-process/server/config"
)

func Start(c config.Config) http.HandlerFunc {
	v := url.Values{}
	v.Add("client_id", c.InnoClientID)
	v.Add("redirect_uri", c.Host+"/oauth/inoreader/callback")
	v.Add("response_type", "code")
	v.Add("scope", "read write")
	v.Add("state", "abcde")
	url := "https://www.inoreader.com/oauth2/auth?" + v.Encode()

	return oauth.StartHandler(url, c.Pass)
}

func Callback(c config.Config) http.HandlerFunc {
	v := url.Values{}
	v.Add("client_id", c.InnoClientID)
	v.Add("client_secret", c.InnoClientSec)
	v.Add("redirect_uri", c.Host+"/oauth/inoreader/callback")
	v.Add("grant_type", "authorization_code")
	u := "https://www.inoreader.com/oauth2/token"

	return oauth.CallbackHandler(v, u, c.InnoTokenPath, c.InnoRefreshPath)
}

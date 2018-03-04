package google

import (
	"net/http"
	"net/url"

	"github.com/gong023/my-slack-process/oauth"
	"github.com/gong023/my-slack-process/server/config"
)

func Start(c config.Config) http.HandlerFunc {
	// https://developers.google.com/identity/protocols/OAuth2WebServer
	v := url.Values{}
	v.Add("client_id", c.GoogleClientID)
	v.Add("redirect_uri", c.Host+"/oauth/google/callback")
	v.Add("response_type", "code")
	v.Add("scope", "https://www.googleapis.com/auth/drive")
	v.Add("state", "abcde")
	v.Add("access_type", "offline")
	url := "https://accounts.google.com/o/oauth2/v2/auth?" + v.Encode()

	return oauth.StartHandler(url, c.Pass)
}

func Callback(c config.Config) http.HandlerFunc {
	v := url.Values{}
	v.Add("client_id", c.GoogleClientID)
	v.Add("client_secret", c.GoogleClientSec)
	v.Add("redirect_uri", c.Host+"/oauth/google/callback")
	v.Add("grant_type", "authorization_code")
	u := "https://www.googleapis.com/oauth2/v4/token"

	return oauth.CallbackHandler(v, u, c.GoogleTokenPath, c.GoogleRefreshPath)
}

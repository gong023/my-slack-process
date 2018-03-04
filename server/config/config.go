package config

import "os"

type Config struct {
	Prod              string
	Host              string
	Cert              string
	Pass              string
	GoogleClientID    string
	GoogleClientSec   string
	GoogleTokenPath   string
	GoogleRefreshPath string
	InnoClientID      string
	InnoClientSec     string
	InnoTokenPath     string
	InnoRefreshPath   string
}

func New() Config {
	host := os.Getenv("HOST")
	if os.Getenv("PROD") != "1" {
		host = "http://localhost:1443"
	}

	return Config{
		Prod:              os.Getenv("PROD"),
		Host:              host,
		Cert:              os.Getenv("CERT"),
		Pass:              os.Getenv("PASS"),
		GoogleClientID:    os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSec:   os.Getenv("GOOGLE_CLIENT_SEC"),
		GoogleTokenPath:   os.Getenv("GOOGLE_TOKEN_PATH"),
		GoogleRefreshPath: os.Getenv("GOOGLE_REFRESH_PATH"),
		InnoClientID:      os.Getenv("INNO_CLIENT_ID"),
		InnoClientSec:     os.Getenv("INNO_CLIENT_SEC"),
		InnoTokenPath:     os.Getenv("INNO_TOKEN_PATH"),
		InnoRefreshPath:   os.Getenv("INNO_REFRESH_PATH"),
	}
}

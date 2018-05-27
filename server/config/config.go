package config

import "os"

type Config struct {
	Host      string
	Pass      string
	ProjectID string
	DocID     string
	InoCliID  string
	InoCliSec string
}

func New() Config {
	return Config{
		Host:      os.Getenv("HOST"),
		Pass:      os.Getenv("PASS"),
		ProjectID: os.Getenv("PROJECT_ID"),
		DocID:     os.Getenv("DOC_ID"),
		InoCliID:  os.Getenv("INO_CLI_ID"),
		InoCliSec: os.Getenv("INO_CLI_SEC"),
	}
}

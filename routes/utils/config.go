package utils

import (
	"encoding/json"
	"os"
)

type Config struct {
	PostgresUri string `json:"postgres_uri"`
	ApiToken    string `json:"api_token"`
	DevMode     bool   `json:"dev_mode"`
	Port        string `json:"port"`
}

func NewConfig() Config {
	env := os.Getenv("CONFIG")

	if env == "" {
		panic("CONFIG environment variable not set")
	}

	var config Config
	err := json.Unmarshal([]byte(env), &config)
	if err != nil {
		panic("Failed to parse config: " + err.Error())
	}
	return config
}

var config = NewConfig()

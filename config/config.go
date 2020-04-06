package config

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
)

type (
	Configuration struct {
		Profile Profile
		Mongo   Mongo
		Rapid   Rapid
	}

	Profile struct {
		Flag string
	}

	Mongo struct {
		HostURI string
		Name    string
	}

	Rapid struct {
		Enabled bool
		Season  string
		BaseURL string
		APIKey  string
	}
)

var (
	Config Configuration
)

func LoadConfig(path string) {
	b, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatalf("Config loading failed: %s", err.Error())
	}

	_, err = toml.Decode(string(b), &Config)

	if err != nil {
		log.Fatalf("Config loading failed: %s", err.Error())
	}
}

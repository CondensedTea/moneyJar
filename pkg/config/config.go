package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

// C is a koanf config
var C = koanf.New(".")

// Load parses yaml config
func Load(configPath string) error {
	return C.Load(file.Provider(configPath), yaml.Parser())
}

package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

var C = koanf.New(".")

func Load(configPath string) error {
	return C.Load(file.Provider(configPath), yaml.Parser())
}

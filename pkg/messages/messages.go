package messages

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

const messagesFile = "messages.yaml"

// MessageCollection is a model for messages file
type MessageCollection struct {
	Messages map[string]string `yaml:"messages"`
}

// Load parses messages file
func Load() (map[string]string, error) {
	fileBytes, err := ioutil.ReadFile(messagesFile)
	if err != nil {
		return nil, err
	}

	var coll MessageCollection

	if err = yaml.Unmarshal(fileBytes, &coll); err != nil {
		return nil, err
	}
	return coll.Messages, nil
}

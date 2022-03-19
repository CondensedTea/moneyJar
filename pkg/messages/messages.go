package messages

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type MessageCollection struct {
	Messages map[string]string `yaml:"messages"`
}

func Load(path string) (map[string]string, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var coll MessageCollection

	if err = yaml.Unmarshal(fileBytes, &coll); err != nil {
		return nil, err
	}
	return coll.Messages, nil
}

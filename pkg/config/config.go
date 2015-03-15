package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type configMap struct {
	Parameters map[string]string
}

func GetConfiguration(config string) configMap {
	absConfig, err := filepath.Abs(config)
	configFile, err := ioutil.ReadFile(absConfig)

	if err != nil {
		panic(err)
	}

	var configuration configMap

	err = yaml.Unmarshal([]byte(configFile), &configuration)
	if err != nil {
		panic(err)
	}

	return configuration
}

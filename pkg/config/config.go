package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
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

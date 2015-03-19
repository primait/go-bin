package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ConfigError struct {
	Reason string
}

func (ce *ConfigError) Error() string {
	return fmt.Sprintf("Configuration error: %s", ce.Reason)
}

type configMap struct {
	Parameters map[string]string
}

func GetConfiguration(config string) (configuration configMap, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()
	// http://blog.golang.org/defer-panic-and-recover
	// https://code.google.com/p/go-wiki/wiki/PanicAndRecover

	configuration = unmarshalConfig(config)

	return
}

func unmarshalConfig(path string) (configuration configMap) {
	configFile := readConfiguration(path)
	err := yaml.Unmarshal(configFile, &configuration)

	if err != nil {
		panic(&ConfigError{"Cannot parse yaml configuration"})
	}

	return
}

func readConfiguration(path string) (configFile []byte) {
	absConfig, err := filepath.Abs(path)
	configFile, err = ioutil.ReadFile(absConfig)

	if err != nil {
		panic(&ConfigError{"Cannot read config file"})
	}

	return
}

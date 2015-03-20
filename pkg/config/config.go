package config

import (
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type ConfigMap struct {
	Parameters map[string]string
}

func GetConfiguration(config string) (configuration ConfigMap, err error) {
	defer func() {
		err, _ = recover().(error)
	}()
	// http://blog.golang.org/defer-panic-and-recover
	// https://code.google.com/p/go-wiki/wiki/PanicAndRecover

	configuration = unmarshalConfig(config)

	return
}

func unmarshalConfig(path string) (configuration ConfigMap) {
	configFile := readConfiguration(path)
	err := yaml.Unmarshal(configFile, &configuration)
	if err != nil {
		panic(err)
	}

	return
}

func readConfiguration(path string) (configFile []byte) {
	absConfig, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	configFile, err = ioutil.ReadFile(absConfig)
	if err != nil {
		panic(err)
	}

	return
}

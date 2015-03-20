package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type configMap struct {
	Parameters map[string]string
}

func GetConfiguration(config string) (configuration configMap, err error) {
	defer func() {
		err, _ = recover().(error)
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
		panic(err)
	}

	return
}

func readConfiguration(path string) (configFile []byte) {
	absConfig, err := filepath.Abs(path)
	configFile, err = ioutil.ReadFile(absConfig)

	if err != nil {
		panic(err)
	}

	return
}

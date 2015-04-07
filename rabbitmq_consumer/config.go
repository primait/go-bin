package main

import (
	"errors"

	"gopkg.in/yaml.v2"
)

type Producer struct {
	Connection      string       `yaml:"connection"`
	ExchangeOptions ExchangeOpts `yaml:"exchange_options"`
}

type Consumer struct {
	Connection      string       `yaml:"connection"`
	ExchangeOptions ExchangeOpts `yaml:"exchange_options"`
	QueueOptions    QueueOpts    `yaml:"queue_options"`
	// service to launch inside php command with the json body
	Callback string `yaml:"callback"`
}

type ExchangeOpts struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type QueueOpts struct {
	Name        string              `yaml:"name"`
	RoutingKeys []string            `yaml:"routing_keys"`
	Args        map[string][]string `yaml:"arguments"`
}

type RabbitMQConfiguration struct {
	Url       string              `yaml:"rabbitmq_url"`
	Producers map[string]Producer `yaml:"producers"`
	Consumers map[string]Consumer `yaml:"consumers"`
}

// helper struct because rabbitmq root in the yaml file sucks..
type Conf struct {
	Configuration RabbitMQConfiguration `yaml:"rabbitmq"`
}

func (c *Conf) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	// config real validation here, just an example..
	// could set a default if empty for example.. etc etc
	if c.Configuration.Url == "" {
		return errors.New("rabbitmq_url cannot be empty!")
	}

	return nil
}

package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/garyburd/redigo/redis"
	"github.com/primait/go-bin/pkg/config"
	"github.com/streadway/amqp"
)

var (
	flConfig = flag.String([]string{"c", "-config"}, "", "Config file path")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: dlx OPTIONS [arg...]\n\nOptions:\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		fmt.Fprint(os.Stdout, "\n")
	}
}

func main() {
	flag.Parse()

	if *flConfig == "" {
		flag.Usage()
		log.Fatal("Please provide a configuration file")
		os.Exit(1)
	}

	var config = config.GetConfiguration(*flConfig)

	redisConnection, err := redis.Dial("tcp", config.Parameters["redis_ip_address"])
	panicOnError(err)
	defer redisConnection.Close()

	rabbitConnection, err := amqp.Dial(config.Parameters["rabbitmq_url"])
	panicOnError(err)
	defer rabbitConnection.Close()

	ch, err := rabbitConnection.Channel()
	panicOnError(err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"dlx_messages",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	panicOnError(err)

	q, err := ch.QueueDeclare(
		"dlx_messages",
		true,
		false,
		false,
		false,
		nil,
	)
	panicOnError(err)

	err = ch.QueueBind(
		q.Name,
		"#",
		"dlx_messages",
		false,
		nil,
	)
	panicOnError(err)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	panicOnError(err)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			body := string(d.Body[:])
			redisConnection.Send(
				"RPUSH",
				fmt.Sprintf("prima:dlx:%s", d.RoutingKey),
				body,
			)
			redisConnection.Flush()
			_, err = redisConnection.Receive()
			if err != nil {
				log.Println("Something went wrong setting data in Redis..fuck..")
			}

			log.Printf(body)
			log.Printf(d.RoutingKey)
		}
	}()

	log.Println("Listening for dead messages...")
	<-forever
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

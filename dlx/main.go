package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/johntdyer/slackrus"
	"github.com/primait/go-bin/pkg/config"
	"github.com/streadway/amqp"
)

var (
	flConfig = flag.String("c", "", "Config file path")
	flDev    = flag.Bool("d", false, "Dev enviroment")
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

	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)

	if *flDev == false {
		log.SetFormatter(&log.JSONFormatter{})
		log.AddHook(&slackrus.SlackrusHook{
			AcceptedLevels: slackrus.LevelThreshold(log.InfoLevel),
			HookURL:        "https://hooks.slack.com/services/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",
			IconEmoji:      ":dragon_face:",
			Channel:        "#dev",
			Username:       "dlx",
		})
	}

	log.Info("Starting dlx")

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
			log.Info(fmt.Sprintf("Received dead message: RoutingKey: %s, Body:%s", d.RoutingKey, body))
		}
	}()

	log.Println("Listening for dead messages...")
	<-forever
}

func panicOnError(err error) {
	if err != nil {
		log.Panic(err)
		panic(err)
	}
}

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
	flag.Parse()
}

func main() {
	if *flConfig == "" {
		flag.Usage()
		log.Fatal("Please provide a configuration file")
		os.Exit(1)
	}

	configuration, err := config.GetConfiguration(*flConfig)
	panicOnError(err)

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

	rabbitConnection, err := amqp.Dial(configuration.Parameters["rabbitmq_url"])
	panicOnError(err)
	defer rabbitConnection.Close()

	blockings := rabbitConnection.NotifyBlocked(make(chan amqp.Blocking))
	go func() {
		for b := range blockings {
			if b.Active {
				log.Printf("TCP blocked: %q", b.Reason)
			} else {
				log.Printf("TCP unblocked")
			}
		}
	}()

	errorsConn := rabbitConnection.NotifyClose(make(chan *amqp.Error))
	go func() {
		for e := range errorsConn {
			log.Error(e)
		}
	}()

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
			go func() {
				defer func() {
					if err, ok := recover().(error); ok {
						// http://godoc.org/github.com/streadway/amqp#Delivery.Nack
						d.Nack(false, true)
						log.Panic(err)
					}
				}()
				pushToRedis(d, configuration)
			}()
		}
	}()

	log.Println("Listening for dead messages...")
	<-forever
}

func pushToRedis(d amqp.Delivery, configuration config.ConfigMap) {
	redisConnection, err := redis.Dial("tcp", configuration.Parameters["redis_ip_address"])
	if err != nil {
		panic(err)
	}
	defer redisConnection.Close()

	body := string(d.Body[:])
	redisConnection.Send(
		"RPUSH",
		fmt.Sprintf("prima:dlx:%s", d.RoutingKey),
		body,
	)
	redisConnection.Flush()
	_, err = redisConnection.Receive()
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("Received dead message: RoutingKey: %s, Body:%s", d.RoutingKey, body))
}

func panicOnError(err error) {
	if err != nil {
		log.Panic(err)
		panic(err)
	}
}

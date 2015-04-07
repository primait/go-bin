package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ogier/pflag"
	"github.com/streadway/amqp"
)

var (
	flConfig = pflag.StringP("config", "c", "", "Configuration file path")
	flWorker = pflag.StringP("worker", "w", "", "Worker name to start")
)

func main() {
	pflag.Parse()

	if *flConfig == "" {
		logrus.Fatalf("please provide a valid configuration file")
		os.Exit(1)
	}
	// FIXME: validate config path here
	absConfig, err := filepath.Abs(*flConfig)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	configFile, err := ioutil.ReadFile(absConfig)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	var c Conf
	if err := c.Parse(configFile); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	if *flWorker == "" {
		logrus.Fatalf("please provide a valid worker name")
		os.Exit(1)
	}

	workerConf, ok := c.Configuration.Consumers[*flWorker]
	if !ok {
		logrus.Fatalf("worker %s not configured in config", *flWorker)
		os.Exit(1)
	}

	// apro la connessione a rabbit
	conn, err := amqp.DialConfig(
		c.Configuration.Url,
		amqp.Config{Heartbeat: 60 * time.Second},
	)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
	defer conn.Close()

	// apro un canale
	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
	defer ch.Close()

	exchange := workerConf.ExchangeOptions.Name
	exchangeType := workerConf.ExchangeOptions.Type
	if err := ch.ExchangeDeclare(
		exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	// FIXME: non ho provato ma vanno castati i tipi per rabbitmq mi sa..
	// EDIT: pare funzionare bene invece
	queueArgs := amqp.Table{}
	for i, v := range workerConf.QueueOptions.Args {
		queueArgs[i] = v[1]
	}

	if err := queueArgs.Validate(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	// dichiaro la coda
	q, err := ch.QueueDeclare(
		workerConf.QueueOptions.Name, // name
		false,     // durable
		true,      // delete when usused
		false,     // exclusive
		false,     // no-wait
		queueArgs, // arguments
	)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	for _, k := range workerConf.QueueOptions.RoutingKeys {
		if err := ch.QueueBind(q.Name, k, workerConf.ExchangeOptions.Name, false, nil); err != nil {
			logrus.Fatal(err)
			// don't die here?
		}
	}

	// consuming
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	forever := make(chan struct{})

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			//wg.Add(1)
			//go func(body []byte) {
			//defer wg.Done()
			// usa un goroutine cosi non blocca qui sopra e continua
			// a ricevere e fare l'handle dei msgs
			// far partire quelle merde di comandi php qui e
			// attendere l'exit status, o leggere lo stderr/out
			/// e return a rabbit ack/nack/reject etc etc
			//}(d.Body)
		}
	}()

	log.Println("Listening for messages...")
	<-forever
}

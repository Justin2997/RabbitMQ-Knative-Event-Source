package main

import (
	"flag"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Get the url to post to for event source
	var sink string
	flag.StringVar(&sink, "sink", "", "This is the url for knative event source")
	flag.Parse()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare exchange for notification worker to move
	err = ch.ExchangeDeclare(
		"us",    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"queue", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Binding for hello-world-knative.default
	log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, "pods", "hello-world-knative.default")
	err = ch.QueueBind(
		q.Name,                        // queue name
		"hello-world-knative.default", // routing key
		"us",                          // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Every things that come here is for hello-world-knative
			log.Printf(" [x] %s", d.Body)

			// TODO send a notificiation to the cluster pods
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}

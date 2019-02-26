package main

import (
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	callbackPods := "logger2"
	var routingKey string
	flag.StringVar(&routingKey, "routingKey", "", "") // If there is no routing Key there will be a error
	flag.Parse()
	if routingKey == "" {
		log.Fatalf("Flag Routing Key have to been pass")
	}

	// Connect to the RabbitMQ broker in the cluster
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Connect to the Exchange
	err = ch.ExchangeDeclare(
		"knative-exchange", // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	for i := 0; i < 5; i++ {
		body := strconv.Itoa(i)
		corrID := randomString(32)
		err = ch.Publish(
			"knative-exchange", // exchange
			routingKey,         // routing key
			false,              // mandatory
			false,              // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				Body:          []byte(body),
				ReplyTo:       callbackPods,
				CorrelationId: corrID,
			})
		failOnError(err, "Failed to publish a message")

		log.Printf(" [x] Sent %s", body)
		time.Sleep(1 * time.Second)
	}
}

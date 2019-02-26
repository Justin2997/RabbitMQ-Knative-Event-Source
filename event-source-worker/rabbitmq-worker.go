package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func sendTasktoFunction(sink string, critical bool, producer string) {
	message := map[string]interface{}{
		"body":     sink,
		"producer": producer,
		"critical": critical,
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(sink, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Status ", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	s := string(body[:])
	log.Printf(" [.] Responce %s", s)
}

func consumeFunctionQueue(ch *amqp.Channel, consumerName string, qName string) {
	msgs, err := ch.Consume(
		qName,        // queue
		consumerName, // consumer
		false,        // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs {
			critical := (strings.Split(d.RoutingKey, ".")[1] == "critical")
			destination := "http://" + strings.Split(d.RoutingKey, ".")[0] + ".default.svc.cluster.local/"
			log.Printf(" [x] %s, %s", d.Body, destination)
			sendTasktoFunction(destination, critical, d.ReplyTo)
			d.Ack(false)
		}
	}()
	log.Printf(" [*] %s ready to consume on %s", consumerName, qName)
}

func consumeErrorQueue(ch *amqp.Channel, consumerName string, qName string) {
	msgs, err := ch.Consume(
		qName,        // queue
		consumerName, // consumer
		false,        // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs {
			critical := (strings.Split(d.RoutingKey, ".")[1] == "critical")
			destination := "http://" + strings.Split(d.RoutingKey, ".")[0] + ".default.svc.cluster.local/"
			log.Printf(" [x] %s, %s", d.Body, destination)
			sendTasktoFunction(destination, critical, d.ReplyTo)
			d.Ack(false)
		}
	}()
	log.Printf(" [*] %s ready to consume on %s", consumerName, qName)
}

func main() {
	var workerName string
	var sink string
	flag.StringVar(&workerName, "name", "", "") // If there is no routing Key there will be a error
	flag.StringVar(&sink, "sink", "", "")       // This is not use for now
	flag.Parse()
	if workerName == "" {
		log.Fatalf("Flag name have to been defind")
	}

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exchangeName := "knative-exchange"
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	qFunctionName := "function"
	qErrorName := "error"
	qFunction, err := ch.QueueDeclare(
		qFunctionName, // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")
	qError, err := ch.QueueDeclare(
		qErrorName, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Binding the Queue
	routingKeyFunction := "*.function"
	routingKeyError := "*.error"
	log.Printf("Binding queue %s to exchange %s with routing key %s", qFunction.Name, exchangeName, routingKeyFunction)
	err = ch.QueueBind(
		qFunction.Name,     // queue name
		routingKeyFunction, // routing key
		exchangeName,       // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")
	log.Printf("Binding queue %s to exchange %s with routing key %s", qError.Name, exchangeName, routingKeyError)
	err = ch.QueueBind(
		qError.Name,     // queue name
		routingKeyError, // routing key
		exchangeName,    // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	forever := make(chan bool)
	consumeFunctionQueue(ch, workerName+".function", qFunction.Name)
	consumeErrorQueue(ch, workerName+".error", qError.Name)
	<-forever
}

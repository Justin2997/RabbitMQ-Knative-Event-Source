package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/streadway/amqp"
)

var (
	queueConfiguration amqp.Table
	exchangeName       string
	routingKeyFunction string
	workerName         string
	sink               string
	DEBUG              bool
)

type Data struct {
	Owner         string `json:"owner"`
	Body          string `json:"body"`
	Critical      bool   `json:"critical"`
	CorrelationID string `json:"correlationId"`
}

// This init the variable and configuration
func init() {
	flag.StringVar(&workerName, "name", "", "") // If there is no routing Key there will be a error
	flag.StringVar(&sink, "sink", "", "")       // This is not use for now

	DEBUG = true
	routingKeyFunction = "#.function"
	exchangeName = "knative-exchange"
	queueConfiguration = amqp.Table{"x-dead-letter-exchange": exchangeName}
}

func sendCallBackResponse(reponse Data) {
	bytesRepresentation, err := json.Marshal(reponse)
	failOnError(err, "Can not convert message for callback response")

	debugLog("CallBack Send to " + reponse.Owner)
	resp, err := http.Post(reponse.Owner, "application/json", bytes.NewBuffer(bytesRepresentation))
	debugLog("Responce code from callBack : " + strconv.Itoa(resp.StatusCode))

	// If there is a error with the callback
	if (err != nil || resp.StatusCode != 200) && reponse.Critical {
		debugLog("ERROR critical message do not made it to the callback")
	}
}

func sendTasktoFunction(message Data) {
	bytesRepresentation, err := json.Marshal(message)
	failOnError(err, "Can not convert message for function call")

	debugLog("Send task to function : " + message.Body)
	resp, err := http.Post(message.Body, "application/json", bytes.NewBuffer(bytesRepresentation))
	failOnError(err, "Can not send message to function")

	debugLog("Task Status " + strconv.Itoa(resp.StatusCode))
	body, err := ioutil.ReadAll(resp.Body)
	failOnError(err, "Can not get function response body")

	err = json.Unmarshal(body, &message)
	failOnError(err, "Can not parse the reponse")
	debugLog(message)

	sendCallBackResponse(message)
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
			message := Data{
				Owner:         d.ReplyTo,
				Body:          destination,
				Critical:      critical,
				CorrelationID: d.CorrelationId,
			}

			debugLog("Receive in function queue : " + string(d.Body) + " for " + destination)
			sendTasktoFunction(message)

			d.Ack(false) // TODO not ack if there is a error to the consumer
		}
	}()

	debugLog(" [*] " + consumerName + " ready to consume on " + qName)
}

func main() {
	flag.Parse()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

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

	qFunction, err := ch.QueueDeclare(
		"function",         // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		queueConfiguration, // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Binding the Queue
	err = ch.QueueBind(
		qFunction.Name,     // queue name
		routingKeyFunction, // routing key
		exchangeName,       // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	forever := make(chan bool)
	consumeFunctionQueue(ch, workerName+".function", qFunction.Name)
	<-forever
}

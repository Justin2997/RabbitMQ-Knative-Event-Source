package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

type Data struct {
	Owner         string `json:"owner"`
	Body          string `json:"body"`
	Critical      bool   `json:"critical"`
	CorrelationID string `json:"correlationId"`
}

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

func callback(corrID string, path string) {
	http.HandleFunc(path+corrID, func(w http.ResponseWriter, r *http.Request) {
		log.Print("Getting callback")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var data Data
		err = json.Unmarshal(body, &data)
		if err != nil {
			panic(err)
		}

		if data.CorrelationID == corrID {
			log.Printf("Recive response for correlationId %s, with body %s", data.CorrelationID, data.Body)
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}

func serving() {
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func main() {
	done := make(chan bool)
	go serving()

	callbackPath := "/callback/"
	callbackURL := "http://producer.default.svc.cluster.local" + callbackPath
	exchangeName := "knative-exchange"
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
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	for i := 0; i < 5; i++ {
		body := strconv.Itoa(i)
		corrID := randomString(32)
		err = ch.Publish(
			exchangeName, // exchange
			routingKey,   // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				Body:          []byte(body),
				ReplyTo:       callbackURL + corrID,
				CorrelationId: corrID,
			})
		failOnError(err, "Failed to publish a message")
		go callback(corrID, callbackPath)

		log.Printf(" [x] Sent %s", body)
		time.Sleep(1 * time.Second)
	}

	<-done // Block forever
}

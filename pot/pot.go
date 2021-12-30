package main

import (
	"fmt"
	"log"

	amqp "amqp"
)

const (
	potPle      = "El pot está ple!"
	cuaPot      = "Pot"
	MAX_MEL_POT = 10
	MAX_POT     = 3
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	nivellMel := make(chan int)
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		cuaPot, // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")
	for i := 0; i < MAX_POT; i++ {
		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		failOnError(err, "Failed to register a consumer")
		omplut := make(chan bool)
		go func(omplut chan bool) {
			for d := range msgs {
				log.Printf("Una abella ha posat mel: %s", d.Body)
				nivellMel <- (<-nivellMel + 1)
				if <-nivellMel == MAX_MEL_POT {
					omplut <- true
				}
			}
		}(omplut)

		if <-omplut {
			fmt.Println("El pot està ple!")
			nivellMel <- 0
		}
	}
}

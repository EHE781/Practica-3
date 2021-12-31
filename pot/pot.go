package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"sync"
)

const (
	potPle      = "El pot está ple!"
	cuaPot      = "Pot"
	cuaAbelles  = "CuaAbelles"
	MAX_MEL_POT = 10
	MAX_POT     = 3
)

var (
	wait sync.WaitGroup
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	noRomput := make(chan bool)
	nivellMel := make(chan int, 1)
	vegadesOmplut := make(chan int, 1)
	//conexion con el servidor de RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	canalAbella, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		cuaAbelles, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	cuaDelPot, err := ch.QueueDeclare(
		cuaPot, // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := canalAbella.Consume(
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
	go func() {
		for d := range msgs {
			log.Printf("Una abella ha posat mel: %s", d.Body)
			log.Println("Kaka")
			nivell := <-nivellMel
			log.Println("Kaka")
			nivellMel <- nivell + 1
			log.Println("Kaka")
			nivell = <-nivellMel
			log.Println("Kaka")
			info := fmt.Sprintf("El nou nivell es: %d", nivell)
			log.Println(info)
			go func() {
				if nivell == MAX_MEL_POT {
					omplut <- true
					vegadesOmplut <- (<-vegadesOmplut + 1)
				}
			}()
		}
	}()
	go func() {
		if <-omplut {
			fmt.Println("El pot està ple!")
			nivellMel <- 0
			err = ch.Publish(
				"",             // exchange
				cuaDelPot.Name, // routing key
				false,          // mandatory
				false,          // immediate->si no ho pot consumir ningú, no es publica (consumer not ready)
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(potPle),
				})
			failOnError(err, "Failed to publish a message")
		}

		if <-vegadesOmplut > MAX_POT {
			fmt.Println("El pot s'ha romput!")
			noRomput <- false
		}
	}()
	<-noRomput
}

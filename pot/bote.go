package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	MAX_RELLENO = 3
	colaBote    = "Cola_Bote"
	colaAbeja   = "Cola_Abeja"
	colaOso     = "Cola_Oso"
	roto        = "Bote roto"
)

type contadorSeguro struct {
	m sync.Mutex
	n int
}

var (
	nivelMiel contadorSeguro
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	boteLleno := make(chan bool)
	fmt.Println("Iniciando el bote de miel...")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	canalAbeja, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer canalAbeja.Close()

	canalBote, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer canalBote.Close()

	colaAbejas, err := canalAbeja.QueueDeclare(
		colaAbeja, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	colaBote, err := canalBote.QueueDeclare(
		colaBote, // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	avisosAbejas, err := canalAbeja.Consume(
		colaAbejas.Name, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	failOnError(err, "Failed to register a consumer")
	//Publicamos quién va  aponer miel y esperamos 1 segundo
	go func() {
		for aviso := range avisosAbejas {
			err = canalBote.Publish(
				"",            // exchange
				colaBote.Name, // routing key
				false,         // mandatory
				false,         // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(aviso.Body),
				})
			nivelMiel.m.Lock()
			nivelMiel.n++
			nivelMiel.m.Unlock()
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		for i := 0; i < MAX_RELLENO; {
			nivelMiel.m.Lock()
			if nivelMiel.n == 10 {
				i++
				log.Println("Se ha llenado el bote")
				nivelMiel.n = 0
			}
			nivelMiel.m.Unlock()
		}
		log.Println("Se ha roto el bote")
		err = canalBote.Publish(
			"",            // exchange
			colaBote.Name, // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(roto),
			})
		boteLleno <- true
	}()
	<-boteLleno
}

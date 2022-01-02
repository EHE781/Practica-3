//WIP NO ACABADO, SOLAMENTE HE COPIADO LA INTRO DE UNA ABEJA. Habria que :
//--Hacer que el oso lea un canal en donde escriben las abejas SOLO si tienen que despertarle (mensaje random)
//--crear un nuevo canal en las abejas donde poner un wait() si el oso envia mensaje (esta comiendo)
//--crear una forma de matar la abeja (el proceso) si el bote se ha roto, de momento se recibe un mensaje pero la lia el wait()

package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	MAX_TRABAJO = 11
	colaBote    = "Cola_Bote"
	colaAbeja   = "Cola_Abeja"
	colaOso     = "Cola_Oso"
	finMiel     = "He acabado de poner miel"
	roto        = "Bote roto"
)

var (
	esperar sync.WaitGroup
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	fmt.Println("Hola, soy la abeja " + os.Args[1])
	//Mandar mensaje que quieres ir a llenar al pot
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

	mensajesBote, err := canalBote.Consume(
		colaBote.Name, // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")
}

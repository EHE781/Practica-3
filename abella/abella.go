package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

const (
	MAX_TRABAJO = 10
	colaBote    = "Cola_Bote"
	colaAbeja   = "Cola_Abeja"
	colaOso     = "Cola_Oso"
	finMiel     = "He acabado de poner miel"
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
	//Tomar el nombre de la abeja de os.Args[1]
	if len(os.Args) == 2 {
		//Mensaje de quien eres
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

		go func() {
			for mensaje := range mensajesBote {
				if string(mensaje.Body) == os.Args[1] {
					esperar.Done()
				}
			}
		}()
		//Empezar a llenar de miel 10 veces (spoiler: for con wait (paquete sync))
		for i := 0; i < MAX_TRABAJO; i++ {
			log.Println(os.Args[1] + " quiere poner miel...")
			err = canalAbeja.Publish(
				"",              // exchange
				colaAbejas.Name, // routing key
				false,           // mandatory
				false,           // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(os.Args[1]),
				})
			failOnError(err, "Failed to publish a message")
			//Si ha sido la primera, la cola la tiene como primera y le daremos permiso desde el bote antes
			esperar.Add(1)
			esperar.Wait()

			log.Println(os.Args[1] + " pone miel en el bote.")
			//Avisar al bote que ha pasado

		}
		log.Println(os.Args[1] + " ha acabado sus reservas y se va.")
		//Si detectas que el pot esta lleno despiertas al oso aun no fet
		//Te vas del pot
	} else {
		fmt.Printf("Los argumentos han fallado! -> [" + strconv.Itoa(len(os.Args)) + "]")

	}
}

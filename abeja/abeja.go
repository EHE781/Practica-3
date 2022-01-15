package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	colaAbejas    = "Cola_Abejas"
	colaAbeja     = "Cola_Abeja_" //"Cola_Abeja_n"
	roto          = "Bote roto"
	mensajeOso    = "El oso esta comiendo, le ha despertado "
	colaDespertar = "Despertar"
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
		iniciar := false
		//Mensaje de quien eres
		fmt.Println("Hola, soy la abeja " + os.Args[1])
		textoRecibido := ""
		//Mandar mensaje que quieres ir a llenar al pot
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		canal, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer canal.Close()

		colaAbejas, err := canal.QueueDeclare(
			colaAbejas, // name
			false,      // durable
			false,      // delete when unused
			false,      // exclusive
			false,      // no-wait
			nil,        // arguments
		)
		failOnError(err, "Failed to declare a queue")

		colaPropia, err := canal.QueueDeclare(
			colaAbeja+os.Args[1], // name
			false,                // durable
			false,                // delete when unused
			false,                // exclusive
			false,                // no-wait
			nil,                  // arguments
		)
		failOnError(err, "Failed to declare a queue")

		colaDespertar, err := canal.QueueDeclare(
			colaDespertar, // name
			false,         // durable
			false,         // delete when unused
			false,         // exclusive
			false,         // no-wait
			nil,           // arguments
		)
		failOnError(err, "Failed to declare queue")

		err = canal.ExchangeDeclare(
			"fin",    // name
			"fanout", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		failOnError(err, "Failed to declare queue")

		err = canal.QueueBind(
			colaAbeja+os.Args[1], // queue name
			"",                   // routing key
			"fin",                // exchange
			false,
			nil,
		)
		failOnError(err, "Failed to declare queue")

		mensajesOso, err := canal.Consume(
			colaPropia.Name, // queue
			"",              // consumer
			true,            // auto-ack
			false,           // exclusive
			false,           // no-local
			false,           // no-wait
			nil,             // args
		)
		failOnError(err, "Failed to register a consumer")

		go func() {
			for mensaje := range mensajesOso {
				textoRecibido = string(mensaje.Body)
				//El bote envia nuestro nombre y la iteracion, contiene nuestro nombre?
				if strings.Contains(textoRecibido, os.Args[1]) && iniciar {
					esperar.Done()
				}
				if string(mensaje.Body) == roto {
					canal.QueueUnbind(colaAbeja+os.Args[1], "", "fin", nil)
					canal.QueueDelete((colaAbeja + os.Args[1]), false, false, false)
					canal.Close()
					esperar.Done()
				}
			}
		}()
		//Empezar a llenar de miel 10 veces (spoiler: for con wait (paquete sync))
		for !canal.IsClosed() {
			iniciar = true
			esperar.Add(1)
			log.Println(os.Args[1] + " quiere poner miel...")
			err = canal.Publish(
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
			esperar.Wait()
			log.Println(os.Args[1] + " pone miel en el bote -> [" + textoRecibido + "]")
			if strings.Contains(textoRecibido, "10") {
				//despertar al oso

				canal.Publish(
					"",                 // exchange
					colaDespertar.Name, // routing key
					false,              // mandatory
					false,              // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(os.Args[1]),
					})
			}
			//Avisar al bote que ha pasado

		}
		log.Println(os.Args[1] + " ha acabado y se va.")

		//Si detectas que el pot esta lleno despiertas al oso aun no fet
		//Te vas del pot
	} else {
		fmt.Printf("Los argumentos han fallado! -> [" + strconv.Itoa(len(os.Args)) + "]")

	}
}

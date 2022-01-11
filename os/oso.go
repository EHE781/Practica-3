package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_RELLENO   = 3
	MAX_NIVEL     = 10
	colaAbejas    = "Cola_Abejas"
	roto          = "Bote roto"
	mensajeOso    = "Estoy comiendo"
	colaDespertar = "Despertar"
	colaPermisos  = "Cola_Permisos"
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
	fmt.Println("El oso se ha puesto a dormir...")

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

	colaDespertar, err := canal.QueueDeclare(
		colaDespertar, // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")

	colaPermisos, err := canal.QueueDeclare(
		colaPermisos, // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	avisosAbejas, err := canal.Consume(
		colaAbejas.Name, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	failOnError(err, "Failed to register a consumer")

	avisoComer, err := canal.Consume(
		colaDespertar.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	failOnError(err, "Failed to register a consumer")

	//Publicamos quién va  aponer miel y esperamos 1 segundo
	go func() {
		for aviso := range avisosAbejas {
			nivelMiel.m.Lock()
			enviar := string(aviso.Body) + " " + strconv.Itoa(nivelMiel.n+1)
			err = canal.Publish(
				"",                // exchange
				colaPermisos.Name, // routing key
				false,             // mandatory
				false,             // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(enviar),
				})

			nivelMiel.n++
			nivelMiel.m.Unlock()
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		contador := 0
		nombreAbeja := ""
		for aviso := range avisoComer {
			nombreAbeja = string(aviso.Body)
			nivelMiel.m.Lock()
			//	if nivelMiel.n == MAX_NIVEL {
			contador++
			log.Println("Se ha llenado el bote")
			log.Println("El oso esta comiendo [" + strconv.Itoa(contador) + "/" +
				strconv.Itoa(MAX_RELLENO) + "], lo ha despertado " + nombreAbeja)
			nivelMiel.n = 0
			time.Sleep(time.Second * 5)
			log.Println("El oso se va a dormir...")
			//	}
			nivelMiel.m.Unlock()
			if contador >= MAX_RELLENO {
				break
			}
		}
		log.Println("Se ha roto el bote")
		err = canal.Publish(
			"",                // exchange
			colaPermisos.Name, // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(roto),
			})
		boteLleno <- true
	}()
	<-boteLleno
	canal.QueueDelete(colaAbejas.Name, false, false, false)
	canal.QueueDelete(colaDespertar.Name, false, false, false)
	finSimulacion := make(chan bool)
	go func() {
		time.Sleep(time.Duration(time.Second) * 5) //Damos 5 segundos a las abejas para leer el mensaje de "exit"
		//Las abejas que no lo lean en 5 segundos (caso muchas abejas) acabarán con un error
		canal.QueueDelete(colaPermisos.Name, false, false, false)
		finSimulacion <- true

	}()
	<-finSimulacion
}

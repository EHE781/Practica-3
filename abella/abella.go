package main

import (
	"fmt"
	"log"
	"os"

	amqp "amqp"
)

const (
	MAX_PRODUCCIO = 10
	cuaOs         = "EstatOs"
	cuaAbelles    = "CuaAbelles"
	menjant       = "Estic menjant"
	dormint       = "He acabat de menjar"
	potPle        = "El pot está ple!"
	cuaPot        = "Pot"
	desperta      = "Ves a menjar"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func conectar() *amqp.Connection {
	//conexion con el servidor de RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	return conn
}

func abrirCanal(conn *amqp.Connection) *amqp.Channel {
	//creacion de un canal
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	return ch
}

func enviarMel(conn *amqp.Connection, ch *amqp.Channel) {

	//envio de cosas
	q, err := ch.QueueDeclare(
		cuaAbelles, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Aquesta es l'abella " + os.Args[1]),
		})
	failOnError(err, "Failed to publish a message")
	for i := 0; i < MAX_PRODUCCIO; i++ {
		body := fmt.Sprintf("L'abella "+os.Args[1]+" produeix mel %d", i)
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			true,   // immediate->si no ho pot consumir ningú, no es publica (consumer not ready)
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
	}
}

func observarOs(conn *amqp.Connection, menjades chan int, ch *amqp.Channel, canal *amqp.Channel) {
	//recibir
	q, err := ch.QueueDeclare(
		cuaOs, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	qPot, err := ch.QueueDeclare(
		cuaPot,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

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

	msgPot, err := ch.Consume(
		qPot.Name, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			if string(d.Body) == menjant {
				//Afegim a menjades les actuals + 1, ja que l'ós está menjant
				menjades <- (<-menjades + 1)
			}
		}
	}()

	go func() {
		for d := range msgPot {
			if string(d.Body) == potPle {
				canal.Publish(
					"",         // exchange
					cuaAbelles, // routing key
					false,      // mandatory
					true,       // immediate->si no ho pot consumir ningú, no es publica (consumer not ready)
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(desperta),
					})
			}
		}
	}()
	/*go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()*/

	log.Printf(" [*] Esperant que l'ós acabi de menjar.\nTo exit press CTRL+C")
}

func main() {
	menjades := make(chan int)
	connexio := conectar()
	canalAbeja := abrirCanal(connexio)
	go enviarMel(connexio, canalAbeja)
	go observarOs(connexio, menjades, abrirCanal(connexio), canalAbeja)
	go func() {
		if <-menjades == 3 {
			log.Printf("L'abella " + os.Args[1] + " s'en va")
			defer connexio.Close()
		}
	}()
}

package main

import (
	amqp "amqp"
	"log"
	"time"
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

func menjarMel(conn *amqp.Connection, menjades chan int) {

	//creacion de un canal
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	//envio de cosas
	q, err := ch.QueueDeclare(
		cuaOs, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	iterar := false
	for iterar {
		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		failOnError(err, "Failed to consume")

		for msg := range msgs {
			if string(msg.Body) == desperta {
				err = ch.Publish(
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(menjant), //l'os envia que esta menjant a totes les abelles (escolten aquest canal)
					})
				failOnError(err, "Failed to publish a message")

				menjades <- (<-menjades + 1)
				time.Sleep(6 * time.Second) //menjar 6 segons la mel

				err = ch.Publish(
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(dormint), //l'os envia que esta menjant a totes les abelles (escolten aquest canal)
					})
				failOnError(err, "Failed to publish a message")
			}
		}
	}
}

func main() {
	menjades := make(chan int)
	connexio := conectar()
	go menjarMel(connexio, menjades)
	go func() {
		if <-menjades == 3 {
			log.Printf("L'ós s'en va a hibernar")
			defer connexio.Close()
		}
	}()
}

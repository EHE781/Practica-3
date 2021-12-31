package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"sync"
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
	avisAbella    = "DespertarOs"
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
	wait.Add(1)
	forever := make(chan bool)
	menjades := make(chan int)
	//conexion con el servidor de RabbitMQ
	connexio, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connexio.Close()

	//creación de un canal
	avisarAbelles, err := connexio.Channel()
	failOnError(err, "Failed to open a channel")
	defer avisarAbelles.Close()
	//envio de cosas
	cuaEnviats, err := avisarAbelles.QueueDeclare(
		cuaOs, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	rebreComunicat, err := connexio.Channel()
	failOnError(err, "Failed to open a channel")
	cuaMissatges, err := rebreComunicat.QueueDeclare(avisAbella, false, false, false, false, nil)
	failOnError(err, "Failed to declare a queue")
	comunicats, err := rebreComunicat.Consume(
		cuaMissatges.Name, // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	failOnError(err, "Failed to consume")

	go func() {
		for msg := range comunicats {
			if string(msg.Body) == desperta {
				err = avisarAbelles.Publish(
					"",              // exchange
					cuaEnviats.Name, // routing key
					false,           // mandatory
					false,           // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(menjant), //l'os envia que esta menjant a totes les abelles (escolten aquest canal)
					})
				failOnError(err, "Failed to publish a message")
				fmt.Println(menjant)
				time.Sleep(time.Duration(6) * time.Second) //menjar 6 segons la mel
				err = avisarAbelles.Publish(
					"",              // exchange
					cuaEnviats.Name, // routing key
					false,           // mandatory
					false,           // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(dormint), //l'os envia que esta menjant a totes les abelles (escolten aquest canal)
					})
				failOnError(err, "Failed to publish a message")
				fmt.Println(dormint)
				menjadesVal := <-menjades + 1
				menjades <- menjadesVal
			}
		}
		wait.Done()
	}()
	wait.Wait()

	go func() {
		for <-menjades < 3 {
		}
		if <-menjades == 3 {
			log.Printf("L'ós s'en va a hibernar")
		}
	}()
	<-forever
}

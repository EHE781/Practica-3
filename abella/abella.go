package main

import (
	amqp "amqp"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const (
	MAX_PRODUCCIO       = 10
	cuaOs               = "EstatOs"
	cuaAbelles          = "CuaAbelles"
	cuaMissatgesAbelles = "CuaIntercomunicatsAbelles"
	menjant             = "Estic menjant"
	dormint             = "He acabat de menjar"
	potPle              = "El pot está ple!"
	cuaPot              = "Pot"
	desperta            = "Ves a menjar"
	melPosada           = "Mel afegida"
)

var (
	wait sync.WaitGroup
	mel  = os.Args[1]
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	wait.Add(1)
	menjades := make(chan int)
	connexio, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connexio.Close()

	canalAbellaAvis, err := connexio.Channel()
	failOnError(err, "Failed to open a channel")
	defer canalAbellaAvis.Close()

	canalAbellaRebuts, err := connexio.Channel()
	failOnError(err, "Failed to open a channel")
	defer canalAbellaRebuts.Close()

	//envio de cosas
	cuaAvisos, err := canalAbellaAvis.QueueDeclare(
		cuaAbelles, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	cuaRebutsAbella, err := canalAbellaRebuts.QueueDeclare(
		cuaMissatgesAbelles, // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	failOnError(err, "Failed to declare a queue")

	fmt.Println("Aquesta es l'abella " + os.Args[1])
	//Obrim canal d'avisos entre abelles
	interAvis, err := canalAbellaRebuts.Consume(
		cuaRebutsAbella.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	go func() {
		for d := range interAvis {
			if string(d.Body) == mel {
				time.Sleep(1 * time.Second) //Esperar la abeja
				esperar := true
				for esperar {
					for k := range interAvis {
						if string(k.Body) == melPosada {
							esperar = false
						}
					}
				}

			}
		}
		fmt.Println("d")
		wait.Done()
	}()

	for i := 0; i < MAX_PRODUCCIO; i++ {

		wait.Wait()
		infoAbella := fmt.Sprintf("L'abella "+os.Args[1]+" produeix mel %d", i)
		fmt.Println(infoAbella)
		body := mel
		err = canalAbellaAvis.Publish(
			"",             // exchange
			cuaAvisos.Name, // routing key
			false,          // mandatory
			false,          // immediate->si no ho pot consumir ningú, no es publica (consumer not ready)
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

		time.Sleep(time.Duration(1) * time.Second)

		body = melPosada
		err = canalAbellaAvis.Publish(
			"",             // exchange
			cuaAvisos.Name, // routing key
			false,          // mandatory
			false,          // immediate->si no ho pot consumir ningú, no es publica (consumer not ready)
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
	}

	canalOs, err := connexio.Channel()
	failOnError(err, "Failed to open a channel")
	defer canalOs.Close()
	//recibir avisos
	cuaRebuts, err := canalOs.QueueDeclare(
		cuaOs, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	cuaPotMel, err := canalOs.QueueDeclare(
		cuaPot,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := canalOs.Consume(
		cuaRebuts.Name, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	failOnError(err, "Failed to register a consumer")

	msgPot, err := canalOs.Consume(
		cuaPotMel.Name, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
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
				canalOs.Publish(
					"",         // exchange
					cuaAbelles, // routing key
					false,      // mandatory
					false,      // immediate->si no ho pot consumir ningú, no es publica (consumer not ready)
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

	go func() {
		if <-menjades == 3 {
			log.Printf("L'abella " + os.Args[1] + " s'en va")
			defer connexio.Close()
		}
	}()

}

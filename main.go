package main

import (
	"encoding/json"
	"flag"
	"math/rand"

	"github.com/benjivesterby/alog"
	"github.com/benjivesterby/atomizer"
	"github.com/google/uuid"
	"github.com/streadway/amqp"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	c := flag.String("conn", "amqp://guest:guest@localhost:5672/", "connection string used for rabbit mq")
	e := flag.String("exch", "atomizer", "exchange used for passing messages")
	r := flag.String("results", "electrons", "exchange for listening for electron results")

	flag.Parse()
	if conn, err := amqp.Dial(*c); err == nil {

		gen(conn, *e, *r)
	}

}

type epay struct {
	Tosses int `json:"tosses"`
}

func gen(conn *amqp.Connection, iqueue, rqueue string) {
	var err error
	var channel *amqp.Channel
	// Create the inbound processing exchanges and queues
	if channel, err = conn.Channel(); err == nil {

		if _, err = channel.QueueDeclare(
			iqueue, // name
			true,   // durable
			false,  // delete when unused
			false,  // exclusive
			false,  // no-wait
			nil,    // arguments
		); err == nil {

			for {
				var e []byte
				var err error

				if e, err = json.Marshal(&epay{rand.Int()}); err == nil {

					electron := &atomizer.ElectronBase{
						ElectronID: uuid.New().String(),
						AtomID:     "montecarlo",
						Load:       e,
					}

					if e, err = json.Marshal(electron); err == nil {
						if err = channel.Publish(
							"",     // exchange
							iqueue, // routing key
							false,  // mandatory
							false,  // immediate
							amqp.Publishing{
								DeliveryMode: amqp.Persistent,
								ContentType:  "application/json",
								Body:         e, //Send the electron's properties
							}); err == nil {
							alog.Printf("Sent Electron [%s] for processing\n", string(e))

							go rec(conn, rqueue, electron.ElectronID)
						} else {
							panic(err)
						}
					} else {
						panic(err)
					}
				} else {
					panic(err)
				}

				//time.Sleep(time.Millisecond * 500)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func rec(conn *amqp.Connection, rqueue, eid string) {
	var err error
	var channel *amqp.Channel
	if channel, err = conn.Channel(); err == nil {

		var queue amqp.Queue
		if queue, err = channel.QueueDeclare(
			uuid.New().String(), // name
			false,               // durable
			false,               // delete when unused
			true,                // exclusive
			false,               // no-wait
			nil,                 // arguments
		); err == nil {

			if err = channel.ExchangeDeclare(
				rqueue,  // name
				"topic", // type
				true,    // durable
				false,   // auto-deleted
				false,   // internal
				false,   // no-wait
				nil,     // arguments

			); err == nil {

				if err = channel.QueueBind(
					queue.Name,
					eid,
					rqueue,
					false, //noWait -- TODO: see would this argument does
					nil,   //args
				); err == nil {
					var in <-chan amqp.Delivery
					var out = make(chan []byte)

					if in, err = channel.Consume(

						queue.Name, //Queue
						"",         // consumer
						true,       // auto ack
						false,      // exclusive
						false,      // no local
						false,      // no wait
						nil,        // args
					); err == nil {
						select {
						case msg, ok := <-in:
							if ok {
								alog.Print(spew.Sdump(msg.Body))
							} else {
								return
							}
						}
					} else {
						close(out)
						// TODO: Handle error / panic
						panic(err)
					}
				} else {
					panic(err)
				}
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}

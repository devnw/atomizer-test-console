package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/devnw/alog"
	"github.com/devnw/amqp"
	"github.com/devnw/atomizer"
	"github.com/devnw/montecarlopi"
	"github.com/google/uuid"
)

func main() {
	c := flag.String("conn", "amqp://guest:guest@localhost:5672/", "connection string used for rabbit mq")
	q := flag.String("queue", "atomizer", "queue is the queue for atom messages to be passed across in the message queue")

	flag.Parse()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go sigterm(ctx, cancel, sigs)

	// Create the amqp conductor for the agent
	conductor, err := amqp.Connect(ctx, *c, *q)
	if err != nil || conductor == nil {
		fmt.Println("error while initializing amqp | " + err.Error())
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// TODO: read in tosses here
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Printf("Enter Tosses: ")

			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)

			v, err := strconv.Atoi(text)
			if err != nil || v < 1 {
				fmt.Println("Invalid number")
				continue
			}

			e, err := electron(v)
			if err != nil {
				panic(err)
			}

			spew.Dump(e)
			p, err := conductor.Send(ctx, e)
			if err != nil {
				panic(err)
			}

			select {
			case <-ctx.Done():
				return
			case res, ok := <-p:
				if !ok {
					continue
				}

				spew.Dump(res)
			}
		}
	}
}

type epay struct {
	Tosses int `json:"tosses"`
}

func electron(tosses int) (atomizer.Electron, error) {
	e, err := json.Marshal(epay{tosses})
	if err != nil {
		return atomizer.Electron{}, err
	}

	electron := atomizer.Electron{
		ID:      uuid.New().String(),
		AtomID:  atomizer.ID(montecarlopi.MonteCarlo{}),
		Payload: e,
	}

	return electron, nil
}

// Setup interrupt monitoring for the agent
func sigterm(ctx context.Context, cancel context.CancelFunc, sigs chan os.Signal) {
	defer cancel()

	select {
	case <-ctx.Done():
		return
	case <-sigs:
		alog.Println("Closing Atomizer Agent")
	}
}

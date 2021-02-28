// Copyright © 2019 Developer Network, LLC
//
// This file is subject to the terms and conditions defined in
// file 'LICENSE', which is part of this source code package.

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
	"time"

	"atomizer.io/amqp"
	"atomizer.io/engine"
	"atomizer.io/montecarlopi"
	"devnw.com/alog"
	"github.com/google/uuid"
)

type output struct {
	In     int     `json:"in"`
	Tosses int     `json:"tosses"`
	Errors int     `json:"errors"`
	PI     float64 `json:"pi"`
}

type empty struct{}

func (empty) Write(p []byte) (int, error) { return 0, nil }

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

	d := alog.Destination{
		Types:  0,
		Format: 0,
		Writer: empty{},
	}

	_ = alog.Global(ctx, "", time.RFC1123, time.UTC, 0, d)

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

				o := output{}
				err := json.Unmarshal(res.Result, &o)
				if err != nil {
					fmt.Println(err.Error())
				}

				fmt.Printf("Pi Estimation: %v\n", o.PI)
			}
		}
	}
}

type epay struct {
	Tosses int `json:"tosses"`
}

func electron(tosses int) (engine.Electron, error) {
	e, err := json.Marshal(epay{tosses})
	if err != nil {
		return engine.Electron{}, err
	}

	electron := engine.Electron{
		ID:      uuid.New().String(),
		AtomID:  engine.ID(montecarlopi.MonteCarlo{}),
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
		os.Exit(1)
	}
}

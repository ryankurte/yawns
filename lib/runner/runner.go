/**
 * OpenNetworkSim Runner
 * The runner module is designed to manage client applications, allowing ONS to launch and maintain instances of a client
 * application to simplify the scripting and automation of simulations.
 *
 * Note that this module heavily relies on exec, which is appropriate for a client application but not for server side use.
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package runner

import (
	"fmt"
	"log"
)

const (
	// DefaultRunCommand is the default command executed by the runner
	DefaultRunCommand = "{{server_address}} {{client_address}}"
)

// Runner instance manages runnable clients
type Runner struct {
	runnables map[string]*Runnable
}

// NewRunner Creates a new Runner instance
func NewRunner() *Runner {
	return &Runner{
		runnables: make(map[string]*Runnable),
	}
}

// NewRunnable creates a runnable instance in the runner
func (runner *Runner) NewRunnable(address, executable, template string, args map[string]string) {

	args["address"] = address

	runnable := NewRunnable(executable, template, args)

	runner.runnables[address] = runnable
}

// Start launches all child runnables
func (runner *Runner) Start() error {

	// Launch runnables
	for name, runner := range runner.runnables {
		log.Printf("Runner.Start starting runnable %s", name)
		err := runner.Run()
		if err != nil {
			return err
		}
	}

	// Launch output collector
	runner.collect()

	return nil
}

func (runner *Runner) collect() {
	outputs := make(chan string, 1024)

	// Create collection routines
	for address, runner := range runner.runnables {
		go func(address string, ch chan string) {
			for {
				select {
				case d, ok := <-ch:
					if !ok {
						log.Printf("Exiting reader for ch: %s", address)
						break
					}
					outputs <- fmt.Sprintf("[CLIENT %s] %s", address, d)
				}
			}
		}(address, runner.GetReadCh())
	}

	// Join and print outputs
	go func(ch chan string) {
		for {
			select {
			case d, ok := <-ch:
				if !ok {
					break
				}
				log.Printf("%s", d)
			}
		}
	}(outputs)
}

// Stop exits all child runnables
func (runner *Runner) Stop() error {

	errors := make(map[string]error)

	for name, runner := range runner.runnables {
		log.Printf("Runner.Stop exiting runnable %s", name)
		err := runner.Exit()
		if err != nil {
			errors[name] = err
		}
	}

	return nil
}

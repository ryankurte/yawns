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

import (
	"github.com/ryankurte/ons/lib/config"
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

// LoadConfig loads runnables from a provided configuration
func (runner *Runner) LoadConfig(c *config.Config, args map[string]string) {
	for _, n := range c.Nodes {
		if n.Address != "" && n.Executable != "" && n.Arguments != "" {
			runner.NewRunnable(n.Address, n.Executable, n.Arguments, args)
		}
	}
}

// NewRunnable creates a runnable instance indexed by address
func (runner *Runner) NewRunnable(address, executable, template string, args map[string]string) {

	// Load address into args for command building
	args["address"] = address

	// Create and save runnable instance
	runnable := NewRunnable(executable, template, args)
	runner.runnables[address] = runnable
}

// Start launches all child runnables
func (runner *Runner) Start() error {

	// Launch runnables
	for name, runner := range runner.runnables {
		log.Printf("Runner.Start starting client %s", name)
		err := runner.Run()
		if err != nil {
			return err
		}
	}

	// Launch output collector
	runner.collect()

	return nil
}

// Info prints info about the runner
func (runner *Runner) Info() {
	log.Printf("Runner Info")
	log.Printf("  - Bound clients %d", len(runner.runnables))
}

// Write writes an input line to the provided runnable by address
func (runner *Runner) Write(address, line string) {
	r, ok := runner.runnables[address]
	if !ok {
		return
	}
	r.Write(line)
}

// Stop exits all child runnables
func (runner *Runner) Stop() error {

	errors := make(map[string]error)

	for name, runner := range runner.runnables {
		log.Printf("Runner.Stop exiting client %s", name)
		err := runner.Exit()
		if err != nil {
			errors[name] = err
		}
	}

	return nil
}

// Internal function to collect outputs from each runnable
// TODO: this needs to support writing to a logfile
func (runner *Runner) collect() {
	outputs := make(chan string, 1024)

	// Create collection routines
	for address, runner := range runner.runnables {
		go func(address string, ch chan string) {
			for {
				select {
				case d, ok := <-ch:
					if !ok {
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

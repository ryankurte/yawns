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

	"github.com/ryankurte/ons/lib/config"
)

// Runner instance manages runnable clients
type Runner struct {
	OutputChan chan string
	clients    map[string]*Runnable
}

// NewRunner Creates a new Runner instance
func NewRunner(c *config.Config, args map[string]string) *Runner {
	r := Runner{
		OutputChan: make(chan string, 1024),
		clients:    make(map[string]*Runnable),
	}

	r.loadConfig(c, args)

	return &r
}

// LoadConfig loads clients from a provided configuration
func (runner *Runner) loadConfig(c *config.Config, args map[string]string) {
	if c == nil {
		return
	}

	for _, n := range c.Nodes {
		if n.Address != "" && n.Executable != "" && n.Command != "" {
			runner.NewRunnable(n.Address, n.Executable, n.Command, args)
		}
	}
}

// NewRunnable creates a runnable instance indexed by address
func (runner *Runner) NewRunnable(address, executable, command string, args map[string]string) {

	// Load address into args for command building
	args["address"] = address

	// Create and save runnable instance
	runnable := NewRunnable(executable, command, args)
	runner.clients[address] = runnable
}

// Start launches all child clients
func (runner *Runner) Start() error {

	// Launch clients
	for _, runner := range runner.clients {
		//log.Printf("Runner.Start starting client %s", name)
		err := runner.Start()
		if err != nil {
			return err
		}
	}

	// Launch output collector
	for a, r := range runner.clients {
		go collect(a, r.GetReadCh(), runner.OutputChan)
	}

	return nil
}

func (runner *Runner) Close() {

}

// Info prints info about the runner
func (runner *Runner) Info() {
	log.Printf("Runner Info")
	log.Printf("  - Bound clients %d", len(runner.clients))
}

// Write writes an input line to the provided runnable by address
func (runner *Runner) Write(address, line string) {
	r, ok := runner.clients[address]
	if !ok {
		return
	}
	r.Write(line)
}

// Stop exits all child clients
func (runner *Runner) Stop() error {
	errors := make(map[string]error)

	for name, runner := range runner.clients {
		err := runner.Exit()
		if err != nil {
			errors[name] = err
		}
	}

	return nil
}

// Helper to collect inputs to a single output channel
func collect(address string, in chan string, out chan string) {
	for {
		select {
		case d, ok := <-in:
			if !ok {
				log.Printf("Runner.collect error channel closed")
				return
			}
			log.Printf("[CLIENT %s] %s", address, d)
			out <- fmt.Sprintf("[CLIENT %s] %s", address, d)
		}
	}
}

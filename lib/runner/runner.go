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
	"sync"
	"time"

	"github.com/ryankurte/owns/lib/config"
)

// Runner instance manages runnable clients
type Runner struct {
	OutputChan chan string
	clients    map[string]*Runnable
	execs      []string
}

// NewRunner Creates a new Runner instance
func NewRunner(c *config.Config, execs []string, args map[string]string) *Runner {
	r := Runner{
		OutputChan: make(chan string, 1024),
		clients:    make(map[string]*Runnable),
		execs:      execs,
	}

	r.parseConfig(c, args)

	return &r
}

func (runner *Runner) parseConfig(c *config.Config, args map[string]string) {
	for _, n := range c.Nodes {
		if n.Address != "" && n.Executable != "" && n.Command != "" {
			runner.NewRunnable(n.Address, n.Executable, n.Command, n.Exec, args)
		}
	}
}

// NewRunnable creates a runnable instance indexed by address
func (runner *Runner) NewRunnable(address, executable, command string, execs []string, args map[string]string) {

	// Load address into args for command building
	args["address"] = address

	// Create and save runnable instance
	runnable := NewRunnable(executable, command, execs, args)
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
	wg := sync.WaitGroup{}

	for name, runner := range runner.clients {
		wg.Add(1)
		go func(name string, runner *Runnable) {
			runner.Interrupt()

			killTimer := time.AfterFunc(10*time.Second, func() {
				runner.Process.Kill()
			})

			err := runner.Wait()
			if err != nil {
				//log.Printf("Runner error: %s waiting for runnable: %s", err, name)
			}

			killTimer.Stop()
			wg.Done()
		}(name, runner)
	}

	wg.Wait()

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
			//log.Printf("[CLIENT %s] %s", address, d)
			out <- fmt.Sprintf("[CLIENT %s] %s", address, d)
		}
	}
}

/**
 * OpenNetworkSim Runner
 * Runnable implements a class wrapping an executable to allow management by the runner
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package runner

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

import (
	"gopkg.in/ryankurte/go-async-cmd.v1"
)

// Runnable wraps an executable with arguments to allow management of the underlying executing instance
type Runnable struct {
	*gocmd.Cmd
	name string
	tmpl string
	args map[string]string
}

// NewRunnable Create a new runnable instance
func NewRunnable(name string, tmpl string, args map[string]string) *Runnable {
	runnable := new(Runnable)

	runnable.name = name
	runnable.tmpl = tmpl
	runnable.args = args

	return runnable
}

func (runnable *Runnable) generateArgs() (string, error) {
	// Parse supplied template
	tmpl, err := template.New("runner").Parse(runnable.tmpl)
	if err != nil {
		return "", fmt.Errorf("Runner.Run error parsing template (%s)", err)
	}

	// Generate command string from template
	var runCmd bytes.Buffer
	err = tmpl.Execute(&runCmd, runnable.args)
	if err != nil {
		return "", fmt.Errorf("Runner.Run error generating run command (%s)", err)
	}

	return runCmd.String(), nil
}

// Run stub to block use of that method
func (runnable *Runnable) Run() error {
	return fmt.Errorf("Runnable.Run() not supported, see Runnable.Start()")
}

// Start a runnable instance
func (runnable *Runnable) Start() error {
	var err error

	// Generate command args
	var args string
	if runnable.args != nil {
		args, err = runnable.generateArgs()
		if err != nil {
			return err
		}
	}

	// Create command
	if runnable.args != nil {
		runnable.Cmd = gocmd.Command(runnable.name, strings.Split(args, " ")...)
	} else {
		runnable.Cmd = gocmd.Command(runnable.name)
	}

	// Create channels
	runnable.Cmd.InputChan = make(chan string, 128)
	runnable.Cmd.OutputChan = make(chan string, 128)

	// Launch command
	err = runnable.Cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

// Write a line to the running process
func (runnable *Runnable) Write(line string) {
	runnable.Cmd.InputChan <- line
}

// GetReadCh Fetch a read channel to the running process
func (runnable *Runnable) GetReadCh() chan string {
	return runnable.Cmd.OutputChan
}

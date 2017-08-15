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

	"gopkg.in/ryankurte/go-async-cmd.v1"
	"time"
)

// Runnable wraps an executable with arguments to allow management of the underlying executing instance
type Runnable struct {
	*gocmd.Cmd
	executable string
	command    string
	execs      []string
	args       map[string]string
}

// NewRunnable Create a new runnable instance
func NewRunnable(executable string, command string, execs []string, args map[string]string) *Runnable {
	runnable := Runnable{
		executable: executable,
		command:    command,
		execs:      execs,
		args:       make(map[string]string),
	}

	// Arguments must be copied otherwise reference array will change
	for k, v := range args {
		runnable.args[k] = v
	}

	return &runnable
}

func generateArgs(command string, args map[string]string) (string, error) {
	// Parse supplied template
	tmpl, err := template.New("runner").Parse(command)
	if err != nil {
		return "", fmt.Errorf("Runnable.generateArgs error parsing template (%s)", err)
	}

	// Generate command string from template
	var runCmd bytes.Buffer
	err = tmpl.Execute(&runCmd, args)
	if err != nil {
		return "", fmt.Errorf("Runnable.generateArgs error generating run command (%s)", err)
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
	args, err := generateArgs(runnable.command, runnable.args)
	if err != nil {
		return err
	}

	// Create command
	if args != "" {
		runnable.Cmd = gocmd.Command(runnable.executable, strings.Split(args, " ")...)
	} else {
		runnable.Cmd = gocmd.Command(runnable.executable)
	}

	// Create channels
	runnable.Cmd.InputChan = make(chan string, 1024)
	runnable.Cmd.OutputChan = make(chan string, 1024)
	runnable.Cmd.ShowOutput = false
	//runnable.Cmd.OutputPrefix = "TEST"

	// Launch command
	err = runnable.Cmd.Start()
	if err != nil {
		return err
	}

	time.AfterFunc(time.Second, func() {
		for _, e := range runnable.execs {
			line, _ := generateArgs(e, runnable.args)
			runnable.Write(line + "\n")
		}
	})
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

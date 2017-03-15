/**
 * OpenNetworkSim Runner
 * Runnable implements a class wrapping an executable to allow management by the runner
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package runner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// KillTimeout timeout for kill signal when exiting a runnable
var KillTimeout = 1000 * time.Millisecond

// InterruptTimeout timeout for interrupt signal when exiting a runnable
var InterruptTimeout = 200 * time.Millisecond

// Runnable wraps an executable with arguments to allow management of the underlying executing instance
type Runnable struct {
	name string
	tmpl string
	args map[string]string
	cmd  *exec.Cmd
	in   chan string
	out  chan string
}

// NewRunnable Create a new runnable instance
func NewRunnable(name string, tmpl string, args map[string]string) *Runnable {
	return &Runnable{
		name: name,
		tmpl: tmpl,
		args: args,
	}
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

// Bind a readable pipe to an output channel for IPC
func (runnable *Runnable) bindReadPipeToChannel(r io.ReadCloser, ch chan string) {
	reader := bufio.NewReader(r)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("Pipe read error: %s", err)
				}
				break
			}
			ch <- line
		}
	}()
}

// Bind a writable pipe to an input channel for IPC
func (runnable *Runnable) bindWritePipeToChannel(w io.WriteCloser, ch chan string) {
	go func() {
		for {
			select {
			case line, ok := <-ch:
				if !ok {
					w.Close()
					break
				}
				_, err := io.WriteString(w, fmt.Sprintln(line))
				if err != nil {
					w.Close()
					break
				}
			}
		}
	}()
}

// Run a runnable instance
func (runnable *Runnable) Run() error {
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
		runnable.cmd = exec.Command(runnable.name, strings.Split(args, " ")...)
	} else {
		runnable.cmd = exec.Command(runnable.name)
	}

	// Create channels
	runnable.in = make(chan string, 128)
	runnable.out = make(chan string, 128)

	// Bind input and output pipes to channels
	stdout, _ := runnable.cmd.StdoutPipe()
	runnable.bindReadPipeToChannel(stdout, runnable.out)
	stderr, _ := runnable.cmd.StderrPipe()
	runnable.bindReadPipeToChannel(stderr, runnable.out)

	stdin, _ := runnable.cmd.StdinPipe()
	runnable.bindWritePipeToChannel(stdin, runnable.in)

	// Launch command
	err = runnable.cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

// Write a line to the running process
func (runnable *Runnable) Write(line string) {
	runnable.in <- line
}

// GetReadCh Fetch a read channel to the running process
func (runnable *Runnable) GetReadCh() chan string {
	return runnable.out
}

// Interrupt sends an os.Interrupt to the running process
func (runnable *Runnable) Interrupt() {
	log.Printf("Runnable: %+v", runnable)
	if runnable.cmd.Process != nil {
		runnable.cmd.Process.Signal(os.Interrupt)
	}
}

// Kill sends an os.Kill signal to the running process
func (runnable *Runnable) Kill() {
	if runnable.cmd.Process != nil {
		runnable.cmd.Process.Kill()
	}
}

// Exit a running runnable
func (runnable *Runnable) Exit() error {
	// TODO: read pipes / return here?

	if runnable.cmd == nil {
		return fmt.Errorf("Runnable not yet started (no cmd)")
	}

	//Start timeouts to interrupt and kill
	interruptTimer := time.AfterFunc(InterruptTimeout, func() {
		runnable.cmd.Process.Signal(os.Interrupt)
	})
	killTimer := time.AfterFunc(KillTimeout, func() {
		runnable.cmd.Process.Kill()
	})

	// Wait for exit
	err := runnable.cmd.Wait()
	status := runnable.cmd.ProcessState

	// Disable timers
	interruptTimer.Stop()
	killTimer.Stop()

	if err != nil && !strings.Contains(status.String(), "signal: interrupt") {
		return fmt.Errorf("Runnable exit error (status: %s)", status.String())
	}

	//close(runnable.in)
	//close(runnable.out)

	return nil
}

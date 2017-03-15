/**
 * go-cmd exec/command wrapper
 *
 * https://github.com/ryankurte/go-cmd
 * Copyright 2017 Ryan Kurte
 */

package gocmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// KillTimeout timeout for kill signal when exiting a Cmd
var KillTimeout = 1000 * time.Millisecond

// InterruptTimeout timeout for interrupt signal when exiting a Cmd
var InterruptTimeout = 200 * time.Millisecond

// Cmd wraps an exec/Cmd and provides a pipe based interface
type Cmd struct {
	*exec.Cmd

	OutputPrefix string
	ShowOutput   bool

	InputChan  chan string
	OutputChan chan string
}

// Command Creates a command
func Command(name string, arg ...string) *Cmd {
	c := new(Cmd)

	c.OutputPrefix = ""
	c.ShowOutput = true
	c.InputChan = nil

	c.Cmd = exec.Command(name, arg...)

	return c
}

// Start wraps Cmd.Start and hooks channels if provided
func (c *Cmd) Start() error {

	// Bind output routines if channel exists
	if c.OutputChan != nil {
		stdout, err := c.StdoutPipe()
		if err != nil {
			return err
		}
		go c.readCloserToChannel(stdout)
		stderr, err := c.StderrPipe()
		if err != nil {
			return err
		}
		go c.readCloserToChannel(stderr)
	}

	// Bind input routine if channel exists
	if c.InputChan != nil {
		stdin, err := c.StdinPipe()
		if err != nil {
			return err
		}
		go c.channelToWriteCloser(stdin)
	}

	return c.Cmd.Start()
}

// Interrupt sends an os.Interrupt to the process if running
func (c *Cmd) Interrupt() {
	if c.Process != nil {
		c.Process.Signal(os.Interrupt)
	}
}

// Exit a running command
// This attempts a wait, with timeout based interrupt and kill signals
func (c *Cmd) Exit() error {

	// Create exit timers
	interruptTimer := time.AfterFunc(InterruptTimeout, func() {
		c.Cmd.Process.Signal(os.Interrupt)
	})
	killTimer := time.AfterFunc(KillTimeout, func() {
		c.Cmd.Process.Kill()
	})

	// Wait for exit
	err := c.Cmd.Wait()

	interruptTimer.Stop()
	killTimer.Stop()

	return err
}

// Handle output to channel or log
func (c *Cmd) output(text string) {
	var out string
	if c.OutputPrefix != "" {
		out = fmt.Sprintf("[%s] %s", c.OutputPrefix, text)
	} else {
		out = text
	}

	if c.OutputChan != nil {
		c.OutputChan <- out
	} else {
		log.Printf("%s", out)
	}
}

// Bind a readable pipe to an output channel for IPC
func (c *Cmd) readCloserToChannel(r io.ReadCloser) {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Pipe read error: %s", err)
			}
			break
		}
		c.output(line)
	}
}

// Bind a writable pipe to an input channel for IPC
func (c *Cmd) channelToWriteCloser(w io.WriteCloser) {
	for {
		select {
		case line, ok := <-c.InputChan:
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
}

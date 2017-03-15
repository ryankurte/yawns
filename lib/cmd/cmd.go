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

// Handle output to channel or log
func (c *Cmd) output(text string) {
	out := fmt.Sprintf("[%s] %s", c.OutputPrefix, text)
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

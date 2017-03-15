/**
 * go-cmd exec/command wrapper
 *
 * https://github.com/ryankurte/go-cmd
 * Copyright 2017 Ryan Kurte
 */

package gocmd

import (
	//"strings"
	"testing"
	//"time"
	//"os/exec"
)

func TestRunnable(t *testing.T) {

	t.Run("Can run commands", func(t *testing.T) {
		c := Command("echo", "Hello")

		err := c.Start()
		if err != nil {
			t.Error(err)
		}

		err = c.Wait()
		if err != nil {
			t.Error(err)
		}
	})

}

package runner

import (
	"strings"
	"testing"
	"time"
)

func TestRunnable(t *testing.T) {

	t.Run("Can generate arguments", func(t *testing.T) {
		args := make(map[string]string)
		args["arg1"] = "Hello"
		args["arg2"] = "World"
		r := NewRunnable("echo", "{{.arg1}} {{.arg2}}", args)

		argString, err := r.generateArgs()
		if err != nil {
			t.Error(err)
		}
		if argString != "Hello World" {
			t.Errorf("Invalid args: '%s'", args)
		}
	})

	t.Run("Can run commands", func(t *testing.T) {
		args := make(map[string]string)
		args["arg1"] = "Hello"
		args["arg2"] = "World"
		r := NewRunnable("echo", "{{.arg1}} {{.arg2}}", args)

		err := r.Start()
		if err != nil {
			t.Error(err)
		}
		r.Exit()
	})

	t.Run("Can stream output from commands", func(t *testing.T) {
		args := make(map[string]string)
		args["arg1"] = "Hello"
		args["arg2"] = "World"
		r := NewRunnable("echo", "{{.arg1}} {{.arg2}}", args)

		err := r.Start()
		if err != nil {
			t.Error(err)
		}

		time.Sleep(1 * time.Second)

		line, ok := <-r.GetReadCh()
		if !ok {
			t.Errorf("Error fetching from channel")
		}
		if line != "Hello World\n" {
			t.Errorf("Unexpected line out: %s", line)
		}

		r.Exit()
	})

	t.Run("Can interrupt and exit commands", func(t *testing.T) {
		r := NewRunnable("cat", "", nil)

		err := r.Start()
		if err != nil {
			t.Error(err)
		}

		err = r.Exit()
		if err == nil {
			t.Errorf("Expected interrupt error")
		}
	})

	t.Run("Can write input to commands", func(t *testing.T) {
		r := NewRunnable("tee", "", nil)

		testString := "Test String\n"

		r.Start()

		r.Write(testString)

		time.Sleep(500 * time.Millisecond)

		line, ok := <-r.GetReadCh()
		if !ok {
			t.Errorf("Error fetching from channel")
		}
		if !strings.Contains(line, testString) {
			t.Errorf("Unexpected line out: %s", line)
		}

		r.Exit()
	})

}

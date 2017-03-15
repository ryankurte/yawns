package runner

import (
	"strings"
	"testing"
	"time"
)

func TestRunnable(t *testing.T) {

	var runnable *Runnable

	t.Run("Create Runnable", func(t *testing.T) {
		args := make(map[string]string)
		args["arg1"] = "Hello"
		args["arg2"] = "World"
		runnable = NewRunnable("echo", "{{.arg1}} {{.arg2}}", args)
	})

	t.Run("Can generate arguments", func(t *testing.T) {
		args, err := runnable.generateArgs()
		if err != nil {
			t.Error(err)
		}
		if args != "Hello World" {
			t.Errorf("Invalid args: '%s'", args)
		}
	})

	t.Run("Can run commands", func(t *testing.T) {
		err := runnable.Run()
		if err != nil {
			t.Error(err)
		}
		runnable.Exit()
	})

	t.Run("Can stream output from commands", func(t *testing.T) {

		err := runnable.Run()
		if err != nil {
			t.Error(err)
		}

		time.Sleep(1 * time.Second)

		line, ok := <-runnable.out
		if !ok {
			t.Errorf("Error fetching from channel")
		}
		if line != "Hello World\n" {
			t.Errorf("Unexpected line out: %s", line)
		}

		runnable.Exit()
	})

	t.Run("Can interrupt and exit commands", func(t *testing.T) {
		r := NewRunnable("cat", "", nil)

		err := r.Run()
		if err != nil {
			t.Error(err)
		}

		err = r.Exit()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Can write input to commands", func(t *testing.T) {
		data := make(map[string]string)
		data["file"] = "test.txt"
		r := NewRunnable("cat", "", nil)

		testString := "Test String"

		r.Run()

		r.in <- testString

		time.Sleep(100 * time.Millisecond)

		line, ok := <-r.out
		if !ok {
			t.Errorf("Error fetching from channel")
		}
		if !strings.Contains(line, testString) {
			t.Errorf("Unexpected line out: %s", line)
		}

		r.Exit()
	})

}

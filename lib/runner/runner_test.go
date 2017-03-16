package runner

import (
	"testing"
	"time"
)

func TestRunner(t *testing.T) {

	t.Run("Create Runnable", func(t *testing.T) {
		args := make(map[string]string)
		args["arg1"] = "Hello"
		args["arg2"] = "World"

		var runner = NewRunner()
		runner.NewRunnable("testOne", "echo", "{{.arg1}} {{.arg2}}", args)
		runner.NewRunnable("testTwo", "echo", "{{.arg1}} {{.arg2}}", args)

		err := runner.Start()
		if err != nil {
			t.Error(err)
		}

		time.Sleep(1000 * time.Millisecond)

		runner.Stop()
	})

}

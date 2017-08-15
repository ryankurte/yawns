package runner

import (
	"testing"
	"time"

	"github.com/ryankurte/owns/lib/config"
)

func TestRunner(t *testing.T) {

	t.Run("Create Runnable", func(t *testing.T) {
		args := make(map[string]string)
		args["arg1"] = "Hello"
		args["arg2"] = "World"

		var runner = NewRunner(&config.Config{}, []string{}, make(map[string]string))
		runner.NewRunnable("testOne", "echo", "{{.arg1}} {{.arg2}}", []string{}, args)
		runner.NewRunnable("testTwo", "echo", "{{.arg1}} {{.arg2}}", []string{}, args)

		err := runner.Start()
		if err != nil {
			t.Error(err)
		}

		time.Sleep(1000 * time.Millisecond)

		runner.Stop()
	})

}

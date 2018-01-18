package main

import (
	"log"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/profile"

	"github.com/ryankurte/owns/lib/simulator"
)

func main() {

	// Load default options
	o := sim.DefaultOptions()

	log.Printf("GOMAXPROCS: %d", runtime.GOMAXPROCS(0))

	// Load command line config
	_, err := flags.Parse(&o)
	if err != nil {
		log.Fatal("Error parsing command line options")
	}

	// Create simulation instance
	sim, err := sim.NewSimulator(&o)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Display info
	sim.Info()

	var p interface {
		Stop()
	}
	if o.Profile {
		p = profile.Start()
	}

	// Launch simulation
	sim.Run()

	if o.Profile {
		p.Stop()
	}

	// Exit simulation
	sim.Close()
}

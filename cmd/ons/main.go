package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/ryankurte/ons/lib/simulator"
	"log"
)

func main() {

	// Load default options
	o := sim.DefaultOptions()

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

	// Launch simulation
	sim.Run()

	// Exit simulation
	sim.Close()
}

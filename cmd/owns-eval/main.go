package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/medium"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Simulation configuration file" default:"owns.yml"`
	Band       string `short:"b" long:"band" description:"Medium band for evaluation"`
}

func main() {
	fmt.Println("OWNS-Eval Utility")

	o := options{}
	_, err := flags.Parse(&o)
	if err != nil {
		fmt.Printf("Error parsing options: %s", err)
		os.Exit(-1)
	}

	c, err := config.LoadConfigFile(o.ConfigFile)
	if err != nil {
		fmt.Printf("Error parsing config file: %s", err)
		os.Exit(-1)
	}

	m, err := medium.NewMedium(&c.Medium, c.TickRate, &c.Nodes)
	if err != nil {
		fmt.Printf("Error creating medium model: %s", err)
		os.Exit(-1)
	}

	b, ok := c.Medium.Bands[o.Band]
	if !ok {
		fmt.Printf("Band not specified or not found\r\n")
		fmt.Printf("Options: ")
		for k := range c.Medium.Bands {
			fmt.Printf("%s ", k)
		}
		fmt.Printf("\r\n")
		os.Exit(-1)
	}

	for i := 0; i < len(c.Nodes); i++ {
		for j := i + 1; j < len(c.Nodes); j++ {

			n1, n2 := c.Nodes[i], c.Nodes[j]
			f := m.GetPointToPointFading(b, n1, n2)

			fmt.Printf("Link %s %s attenuation: %.2f\n", n1.Address, n2.Address, f)
		}
	}

}

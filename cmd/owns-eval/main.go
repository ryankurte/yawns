package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/medium"
	"github.com/ryankurte/owns/lib/medium/layers"
	"github.com/ryankurte/owns/lib/types"
)

type options struct {
	ConfigFile string   `short:"c" long:"config" description:"Simulation configuration file" default:"owns.yml"`
	Band       string   `short:"b" long:"band" description:"Medium band for evaluation"`
	Nodes      []string `short:"n" long:"node" description:"Nodes to be filtered from configuration"`
	OutputDir  string   `short:"o" long:"output" description:"Output directory" default:"outputs"`
}

func main() {
	fmt.Println("OWNS-Eval Utility")

	o := options{}
	_, err := flags.Parse(&o)
	if err != nil {
		fmt.Printf("Error parsing options: %s", err)
		os.Exit(-1)
	}

	os.Mkdir(o.OutputDir, 0766)

	c, err := config.LoadConfigFile(o.ConfigFile)
	if err != nil {
		fmt.Printf("Error parsing config file: %s", err)
		os.Exit(-1)
	}

	fmt.Printf("Loaded %d nodes and %d events\n", len(c.Nodes), len(c.Updates))

	nodes := make([]types.Node, 0)
	if len(o.Nodes) != 0 {
		for _, v := range c.Nodes {
			for _, f := range o.Nodes {
				if v.Address == f {
					nodes = append(nodes, v)
					continue
				}
				// Convert for convenience with numeric addresses
				a, err1 := strconv.Atoi(v.Address)
				b, err2 := strconv.Atoi(f)
				if err1 == nil && err2 == nil && a == b {
					nodes = append(nodes, v)
					continue
				}
			}
		}
	} else {
		nodes = c.Nodes
	}

	m, err := medium.NewMedium(&c.Medium, c.TickRate, &nodes)
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

	ml, _ := m.GetLayer("terrain")
	mapLayer := ml.(*layers.TerrainLayer)

	links := make([]types.Link, 0)
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {

			n1, n2 := nodes[i], nodes[j]
			f := m.GetPointToPointFading(b, n1, n2)

			fmt.Printf("Link %s %s attenuation: %.2f\n", n1.Address, n2.Address, f)

			if f < b.LinkBudget {
				links = append(links, types.Link{A: i, B: j, Fading: float64(f)})
			}

			mapLayer.GraphTerrain(fmt.Sprintf("%s/terrain-%s-%s.png", o.OutputDir, n1.Address, n2.Address), n1.Location, n2.Location)
		}
	}

	err = m.Render(fmt.Sprintf("%s/map.png", o.OutputDir), nodes, links)
	if err != nil {
		fmt.Printf("Error rendering output: %s", err)
		os.Exit(-1)
	}

}

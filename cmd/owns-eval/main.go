package main

import (
	"fmt"
	"image/color"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gonum/stat"
	"github.com/jessevdk/go-flags"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/helpers"
	"github.com/ryankurte/owns/lib/medium"
	"github.com/ryankurte/owns/lib/medium/layers"
	"github.com/ryankurte/owns/lib/types"
)

type Options struct {
	ConfigFile   string   `short:"c" long:"config" description:"Simulation configuration file" default:"owns.yml"`
	Band         string   `short:"b" long:"band" description:"Medium band for evaluation"`
	Nodes        []string `short:"n" long:"node" description:"Nodes to be filtered from configuration"`
	OutputDir    string   `short:"o" long:"output" description:"Output directory" default:"outputs"`
	LinkInfo     string   `long:"link-info" description:"Real link information file for analysis"`
	RealOnly     bool     `long:"real-only" description:"Render real links only"`
	SimOnly      bool     `long:"simulated-only" description:"Render simulated links only"`
	TerrainFiles bool     `long:"terrain" description:"Render terrain slices"`
}

type LinkInfo struct {
	A, B     string
	Fading   float64
	Critical bool
}

type LinkInfoList []LinkInfo

func (l LinkInfoList) ResolveLinks(nodes types.Nodes) types.Links {
	links := make(types.Links, 0)
	for _, v := range l {
		i1, ok1 := nodes.FindIndex(v.A)
		i2, ok2 := nodes.FindIndex(v.B)
		if ok1 && ok2 {
			link := types.Link{A: i1, B: i2, Fading: v.Fading, Meta: v}
			links = append(links, link)
		}
	}
	return links
}

func (l LinkInfoList) Find(a, b string) (*LinkInfo, bool) {
	for _, v := range l {
		if v.A == a && v.B == b {
			return &v, true
		}
		if v.A == b && v.B == a {
			return &v, true
		}
	}
	return nil, false
}

var red = color.RGBA{255, 0, 0, 128}
var green = color.RGBA{0, 255, 0, 128}
var blue = color.RGBA{0, 0, 255, 128}
var black = color.RGBA{0, 0, 0, 128}

func main() {
	fmt.Println("OWNS-Eval Utility")

	o := Options{}
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

	linkInfo := make(LinkInfoList, 0)
	if o.LinkInfo != "" {
		err := helpers.ReadYAMLFile(o.LinkInfo, &linkInfo)
		if err != nil {
			fmt.Printf("Error loading link info: %s", err)
			os.Exit(-2)
		}

		fmt.Printf("Loaded %d real-world links\n", len(linkInfo))
	}

	nodes := make(types.Nodes, 0)
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

	realLinks := linkInfo.ResolveLinks(nodes)
	fmt.Printf("Mapped %d (of %d) real-world links\n", len(realLinks), len(linkInfo))

	sort.Slice(nodes, func(i, j int) bool {
		n := strings.Compare(nodes[i].Address, nodes[j].Address)
		if n < 0 {
			return true
		}
		return false
	})

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

	simLinks := make(types.Links, 0)

	fmt.Printf("\nComputed fadings:\n")
	fmt.Printf("\t")
	for i := 0; i < len(nodes); i++ {
		fmt.Printf("%s,\t", nodes[i].Address)
	}
	fmt.Printf("\n")

	// Compute links
	for i := 0; i < len(nodes); i++ {
		fmt.Printf("%s,\t", nodes[i].Address)
		for j := 0; j < len(nodes); j++ {

			n1, n2 := nodes[i], nodes[j]

			fading := types.Attenuation(-0.0)
			if i != j {
				fm := m.GetPointToPointFading(b, n1, n2)
				fading = -fm.Reduce()

				if fading > -b.LinkBudget {
					simLinks = append(simLinks, types.Link{A: i, B: j, Fading: float64(fading), Meta: fm})
				}
			}
			fmt.Printf("%3.1f,\t", fading)

			if ml != nil && o.TerrainFiles {
				mapLayer := ml.(*layers.TerrainLayer)
				mapLayer.GraphTerrain(fmt.Sprintf("%s/terrain-%s-%s.png", o.OutputDir, n1.Address, n2.Address), n1.Location, n2.Location)
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")

	if len(linkInfo) != 0 {
		a := make([]string, 0)
		b := make([]string, 0)
		real := make([]float64, 0)
		simulated := make([]float64, 0)
		errors := make([]float64, 0)

		fmt.Printf("\nReal world comparison:\n")
		fmt.Printf("A,     B,     Real,    Free Space, Terrain, Foliage, Simulated, Error\n")
		for _, l := range simLinks {
			n1, n2 := nodes[l.A], nodes[l.B]

			if r, ok := linkInfo.Find(n1.Address, n2.Address); ok {
				a = append(a, n1.Address)
				b = append(b, n2.Address)
				real = append(real, r.Fading)
				simulated = append(simulated, l.Fading)
				errors = append(errors, r.Fading-float64(l.Fading))
				meta := l.Meta.(types.AttenuationMap)
				fmt.Printf("%s, %s, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f\n", n1.Address, n2.Address, r.Fading, meta["free-space"], meta["terrain"], meta["foliage"], l.Fading, r.Fading-float64(l.Fading))
			}
		}

		fmt.Printf("\nStatistics:\n")

		n := len(nodes)
		maxLinks := n * (n - 1) / 2

		if len(linkInfo) != 0 {
			fmt.Printf("Simulated links: %d real links: %d (of possible %d)\n", len(simLinks), len(realLinks), maxLinks)
		} else {
			fmt.Printf("Simulated links: %d (of possible %d)\n", len(simLinks), maxLinks)
		}

		AvgReal, stdDevReal := stat.MeanStdDev(real, nil)
		AvgSim, stdDevSim := stat.MeanStdDev(simulated, nil)
		AvgErr, stdDevErr := stat.MeanStdDev(errors, nil)
		fmt.Printf("Mean, , %.2f, %.2f, %.2f\n", AvgReal, AvgSim, AvgErr)
		fmt.Printf("StdDev, , %.2f, %.2f, %.2f\n", stdDevReal, stdDevSim, stdDevErr)

		skewReal := stat.Skew(real, nil)
		skewSim := stat.Skew(simulated, nil)
		skewErr := stat.Skew(errors, nil)
		fmt.Printf("Skew, , %.2f, %.2f, %.2f\n", skewReal, skewSim, skewErr)

		correlation := stat.Correlation(simulated, real, nil)
		fmt.Printf("Correlation, %.2f\n", correlation)
	}

	both := simLinks.Filter(func(l types.Link) bool {
		_, ok1 := realLinks.Find(l.A, l.B)
		_, ok2 := realLinks.Find(l.B, l.A)
		return (ok1 || ok2)
	})

	realOnly := realLinks.Difference(both)
	simOnly := simLinks.Difference(both)
	//both = both

	fmt.Printf("Real links: %d Simulated links: %d\n", len(realLinks), len(simLinks))
	fmt.Printf("Common links: %d Real only: %d Sim only: %d\n", len(both), len(realOnly), len(simOnly))

	rl, _ := m.GetLayer("render")
	renderLayer := rl.(*layers.RenderLayer)
	RenderLinks(renderLayer, fmt.Sprintf("%s/00-simulated.png", o.OutputDir),
		nodes, []types.Links{simLinks},
		[]color.Color{blue})

	RenderLinks(renderLayer, fmt.Sprintf("%s/00-real.png", o.OutputDir),
		nodes, []types.Links{realLinks},
		[]color.Color{blue})

	RenderLinks(renderLayer, fmt.Sprintf("%s/00-both.png", o.OutputDir),
		nodes, []types.Links{both},
		[]color.Color{black})

	RenderLinks(renderLayer, fmt.Sprintf("%s/00-only-sim.png", o.OutputDir),
		nodes, []types.Links{simOnly},
		[]color.Color{red})

	RenderLinks(renderLayer, fmt.Sprintf("%s/00-only-real.png", o.OutputDir),
		nodes, []types.Links{realOnly},
		[]color.Color{red})

}

func RenderLinks(r *layers.RenderLayer, name string, nodes types.Nodes, links []types.Links, c []color.Color) error {
	render := r.NewRender()
	for i, _ := range links {
		render = render.Links(nodes, links[i].Deduplicate(), c[i])
	}
	render = render.Nodes(nodes, color.RGBA{255, 0, 0, 128}, 16)
	return render.Finish(name)
}

package layers

import (
	"fmt"
	"image/color"
	"log"

	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"
	"github.com/ryankurte/go-rf"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

type MapLayer struct {
	In        chan RenderCommand
	satellite maps.Tile
	terrain   maps.Tile
}

type RenderCommand struct {
	FileName string
	Level    int
	Nodes    []types.Node
	Links    []types.Link
}

func NewMapLayer(c *config.Maps) (*MapLayer, error) {
	m := MapLayer{
		In: make(chan RenderCommand, 128),
	}

	satelliteImg, _, err := maps.LoadImage(c.Satellite)
	if err != nil {
		log.Printf("Error loading %s", c.Satellite)
		return nil, err
	}
	m.satellite = maps.NewTile(c.X, c.Y, c.Level, 512, satelliteImg)

	terrainImg, _, err := maps.LoadImage(c.Terrain)
	if err != nil {
		log.Printf("Error loading %s", c.Terrain)
		return nil, err
	}
	m.terrain = maps.NewTile(c.X, c.Y, c.Level, 512, terrainImg)

	log.Printf("%+v", m)

	return &m, nil
}

func (m *MapLayer) Run() {
	for {
		select {
		case c, ok := <-m.In:
			if !ok {
				return
			}
			err := m.Render(c.FileName, c.Nodes, c.Links)
			if err != nil {
				log.Printf("Map render error: %s", err)
			}
		}
	}
}

func (m *MapLayer) Render(fileName string, nodes []types.Node, links []types.Link) error {
	tile := m.satellite

	for _, n := range nodes {
		tile.DrawPoint(onsToMapLoc(&n.Location), 16, color.RGBA{255, 0, 0, 255})
	}

	for _, l := range links {
		n1, n2 := nodes[l.A], nodes[l.B]
		tile.DrawLine(onsToMapLoc(&n1.Location), onsToMapLoc(&n2.Location), color.RGBA{255, 0, 0, 255})
	}

	return maps.SaveImageJPG(tile, fileName)
}

func onsToMapLoc(l *types.Location) base.Location {
	return base.Location{Latitude: l.Lat, Longitude: l.Lng}
}

// CalculateFading calculates the free space fading for a link
func (m *MapLayer) CalculateFading(band config.Band, p1, p2 types.Location) float64 {

	p1m, p2m := onsToMapLoc(&p1), onsToMapLoc(&p2)
	terrain := m.terrain.InterpolateAltitudes(p1m, p2m)
	distance := rf.CalculateDistanceLOS(p1.Lat, p1.Lng, p1.Alt, p2.Lat, p2.Lng, p2.Alt)

	highestImpingement := rf.Distance(terrain[0])
	distanceToImpingement := rf.Distance(0.0)

	for i, v := range terrain {
		if v > float64(highestImpingement) {
			highestImpingement = rf.Distance(v)
			distanceToImpingement = distance * rf.Distance(float64(i)/float64(len(terrain)))
		}
	}

	fmt.Printf("link\n")
	fmt.Printf("  - p1 alt: %02.2f p2 alt: %2.2f gradient: %2.4f\n", p1.Alt, p2.Alt, (p2.Alt-p1.Alt)/float64(len(terrain)))
	fmt.Printf("  - terrain p1: %2.2f terrain p2: %2.2f terrain h: %2.2f\n", terrain[0], terrain[len(terrain)-1], highestImpingement)

	v, err := rf.CalculateFresnelKirckoffDiffractionParam(rf.Frequency(band.Frequency), distanceToImpingement, distance-distanceToImpingement, highestImpingement)
	if err != nil {
		fmt.Printf("MAP layer error: %s", err)
		return 0.0
	}

	f, err := rf.CalculateFresnelKirchoffLossApprox(v)
	if err != nil {
		fmt.Printf("MAP layer error: %s", err)
		return 0.0
	}

	return float64(f)
}

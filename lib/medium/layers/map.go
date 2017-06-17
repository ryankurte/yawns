package layers

import (
	"image/color"
	"log"

	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

type Map struct {
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

func NewMap(c *config.Maps) (*Map, error) {
	m := Map{
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

func (m *Map) Run() {
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

func (m *Map) Render(fileName string, nodes []types.Node, links []types.Link) error {
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

package layers

import (
	"image/color"
	"log"

	"github.com/ryankurte/go-mapbox/lib"
	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"
	"github.com/ryankurte/ons/lib/types"
)

type Map struct {
	In       chan RenderCommand
	mapbox   *mapbox.Mapbox
	baseTile maps.Tile
}

type RenderCommand struct {
	FileName string
	Level    int
	Nodes    []types.Node
	Links    []types.Link
}

func NewMap(token string, a, b types.Location, level int) (*Map, error) {
	m := Map{
		mapbox: mapbox.NewMapbox(token),
		In:     make(chan RenderCommand, 128),
	}
	cache, _ := maps.NewFileCache("/tmp/owns")
	m.mapbox.Maps.SetCache(cache)

	tiles, err := m.mapbox.Maps.GetEnclosingTiles(maps.MapIDSatellite,
		onsToMapLoc(&a), onsToMapLoc(&b),
		uint64(level), maps.MapFormatJpg90, true)
	if err != nil {
		return nil, err
	}

	m.baseTile = maps.StitchTiles(tiles)

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
	tile := m.baseTile

	for _, l := range links {
		n1, n2 := nodes[l.A], nodes[l.B]
		tile.DrawLine(onsToMapLoc(&n1.Location), onsToMapLoc(&n2.Location), color.RGBA{255, 0, 0, 255})
	}

	return maps.SaveImageJPG(tile, fileName)
}

func onsToMapLoc(l *types.Location) base.Location {
	return base.Location{Latitude: l.Lat, Longitude: l.Lng}
}

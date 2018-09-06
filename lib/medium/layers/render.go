package layers

import (
	"image/color"
	"log"

	"github.com/pbnjay/pixfont"

	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

// RenderLayer provides mechanisms for map rendering and visualisation
// Using the provided map tile
type RenderLayer struct {
	In        chan RenderCommand
	satellite maps.Tile
}

type RenderCommand struct {
	FileName string
	Level    int
	Nodes    types.Nodes
	Links    types.Links
}

// NewRenderLayer creates a new instance of the rendering layer based on the provided config
func NewRenderLayer(c *config.Maps) (*RenderLayer, error) {
	r := RenderLayer{
		In: make(chan RenderCommand, 128),
	}

	satelliteImg, _, err := maps.LoadImage(c.Satellite)
	if err != nil {
		log.Printf("Error loading %s", c.Satellite)
		return nil, err
	}
	r.satellite = maps.NewTile(c.X, c.Y, c.Level, 512, satelliteImg)

	return &r, nil
}

// Run launches a render layer thread to process RenderEvents
func (m *RenderLayer) Run() {
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

// Render renders a simulation map with the provided nodes and links
func (m *RenderLayer) Render(fileName string, nodes types.Nodes, links types.Links) error {
	tile := m.satellite

	for _, l := range links {
		n1, n2 := nodes[l.A], nodes[l.B]
		tile.DrawLine(onsToMapLoc(&n1.Location), onsToMapLoc(&n2.Location), color.RGBA{255, 0, 0, 255})

		//		x1, y1, _ := tile.LocationToPixel(onsToMapLoc(&n1.Location))
		//		x2, y2, _ := tile.LocationToPixel(onsToMapLoc(&n2.Location))
		//		xAvg, yAvg := (x1+x2)/2, (y1+y2)/2
		//		note := fmt.Sprintf("%.2f dB", l.Fading)
		//		pixfont.DrawString(tile, int(xAvg)-len(note)*8/2, int(yAvg)-4, note, color.Black)
	}

	for _, n := range nodes {
		tile.DrawPoint(onsToMapLoc(&n.Location), 16, color.RGBA{255, 0, 0, 255})
		x, y, _ := tile.LocationToPixel(onsToMapLoc(&n.Location))
		pixfont.DrawString(tile, int(x)-len(n.Address)*8/2, int(y)-4, n.Address, color.Black)
	}

	return maps.SaveImageJPG(tile, fileName)
}

type Render struct {
	tile maps.Tile
}

func (m *RenderLayer) NewRender() *Render {
	return &Render{tile: m.satellite.Clone()}
}

func (r *Render) Nodes(nodes types.Nodes, c color.RGBA, size uint64) *Render {
	for _, n := range nodes {
		r.tile.DrawPoint(onsToMapLoc(&n.Location), 16, c)
		x, y, _ := r.tile.LocationToPixel(onsToMapLoc(&n.Location))
		pixfont.DrawString(r.tile, int(x)-len(n.Address)*8/2, int(y)-4, n.Address, color.Black)
	}
	return r
}

func (r *Render) Links(nodes types.Nodes, links types.Links, color color.Color) *Render {
	for _, l := range links {
		n1, n2 := nodes[l.A], nodes[l.B]
		r.tile.DrawLine(onsToMapLoc(&n1.Location), onsToMapLoc(&n2.Location), color)
	}
	return r
}

func (r *Render) Finish(fileName string) error {
	return maps.SaveImageJPG(r.tile, fileName)
}

func onsToMapLoc(l *types.Location) base.Location {
	return base.Location{Latitude: l.Lat, Longitude: l.Lng}
}

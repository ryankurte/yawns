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

type RenderLayer struct {
	In        chan RenderCommand
	satellite maps.Tile
}

type RenderCommand struct {
	FileName string
	Level    int
	Nodes    []types.Node
	Links    []types.Link
}

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

func (m *RenderLayer) Render(fileName string, nodes []types.Node, links []types.Link) error {
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

func onsToMapLoc(l *types.Location) base.Location {
	return base.Location{Latitude: l.Lat, Longitude: l.Lng}
}

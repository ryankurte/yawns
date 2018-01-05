package layers

import (
	"fmt"
	"image/color"
	"log"

	"github.com/pbnjay/pixfont"

	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"
	"github.com/ryankurte/go-rf"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

type MapLayer struct {
	In            chan RenderCommand
	satellite     maps.Tile
	terrain       maps.Tile
	defaultOffset float64
}

type RenderCommand struct {
	FileName string
	Level    int
	Nodes    []types.Node
	Links    []types.Link
}

func NewMapLayer(c *config.Maps) (*MapLayer, error) {
	m := MapLayer{
		In:            make(chan RenderCommand, 128),
		defaultOffset: 1.0,
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

// CalculateFading calculates the free space fading for a link
func (m *MapLayer) CalculateFading(band config.Band, p1, p2 types.Location) (float64, error) {
	p1m, p2m := onsToMapLoc(&p1), onsToMapLoc(&p2)

	// Fetch terrain between points
	terrain := m.terrain.InterpolateAltitudes(p1m, p2m)
	if len(terrain) == 0 {
		return 0.0, fmt.Errorf("no terrain between %+v and %+v", p1, p2)
	}
	terrain = rf.SmoothN(1, terrain)

	// Apply default offset where altitudes are not explicitly set
	p1Alt := p1.Alt
	if p1Alt == 0 {
		p1Alt = terrain[0] + m.defaultOffset
	}
	p2Alt := p2.Alt
	if p2Alt == 0 {
		p2Alt = terrain[len(terrain)-1] + m.defaultOffset
	}

	// Calculate LoS distance between two points
	distance := rf.CalculateDistanceLOS(p1.Lat, p1.Lng, p1Alt, p2.Lat, p2.Lng, p2Alt)

	// Calculate equivalent knife edge using bullington figure 12 method
	d1, d2, h := rf.BullingtonFigure12Method(p1Alt, p2Alt, distance, terrain)

	fmt.Printf("\nlink\n")
	fmt.Printf("  - p1 alt: %02.2f p2 alt: %2.2f distance: %2.4f\n", p1Alt, p2Alt, distance)
	fmt.Printf("  - terrain p1: %2.2f terrain p2: %2.2f terrain impingement: %2.2f at: %2.2f\n", terrain[0], terrain[len(terrain)-1], h, d1)

	// Calculate fresnel kirchoff approximation using bullington equivalent knife edge
	v, err := rf.CalculateFresnelKirckoffDiffractionParam(rf.Frequency(band.Frequency), rf.Distance(d1), rf.Distance(d2), rf.Distance(h))
	if err != nil {
		return 0.0, err
	}
	f, err := rf.CalculateFresnelKirchoffLossApprox(v)
	if err != nil {
		return 0.0, err
	}

	fmt.Printf("  - attenuation: %.2f\n", f)

	return float64(f), nil
}

func (m *MapLayer) GraphTerrain(file string, p1, p2 types.Location) error {
	p1m, p2m := onsToMapLoc(&p1), onsToMapLoc(&p2)

	// Fetch terrain between points
	terrain := m.terrain.InterpolateAltitudes(p1m, p2m)
	if len(terrain) == 0 {
		return fmt.Errorf("no terrain between %+v and %+v", p1, p2)
	}
	terrain = rf.SmoothN(1, terrain)

	// Apply default offset where altitudes are not explicitly set
	p1Alt := p1.Alt
	if p1Alt == 0 {
		p1Alt = terrain[0] + m.defaultOffset
	}
	p2Alt := p2.Alt
	if p2Alt == 0 {
		p2Alt = terrain[len(terrain)-1] + m.defaultOffset
	}

	// Calculate LoS distance between two points
	distance := rf.CalculateDistanceLOS(p1.Lat, p1.Lng, p1Alt, p2.Lat, p2.Lng, p2Alt)

	return rf.GraphBullingtonFigure12(file, false, p1Alt, p2Alt, distance, terrain)
}

package layers

import (
	"fmt"
	"log"

	"github.com/ryankurte/go-mapbox/lib/maps"
	"github.com/ryankurte/go-rf"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

type TerrainLayer struct {
	terrain       maps.Tile
	defaultOffset float64
}

func NewTerrainLayer(c *config.Maps) (*TerrainLayer, error) {
	t := TerrainLayer{
		defaultOffset: 1.0,
	}

	terrainImg, _, err := maps.LoadImage(c.Terrain)
	if err != nil {
		log.Printf("Error loading %s", c.Terrain)
		return nil, err
	}
	t.terrain = maps.NewTile(c.X, c.Y, c.Level, 512, terrainImg)

	return &t, nil
}

// CalculateFading calculates the free space fading for a link
func (m *TerrainLayer) CalculateFading(band config.Band, p1, p2 types.Location) (float64, error) {
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

	// Calculate fresnel kirchoff approximation using bullington equivalent knife edge
	v, err := rf.CalculateFresnelKirckoffDiffractionParam(rf.Frequency(band.Frequency), rf.Distance(d1), rf.Distance(d2), rf.Distance(h))
	if err != nil {
		return 0.0, err
	}
	f, err := rf.CalculateFresnelKirchoffLossApprox(v)
	if err != nil {
		return 0.0, err
	}

	return float64(f), nil
}

func (m *TerrainLayer) GraphTerrain(file string, p1, p2 types.Location) error {
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

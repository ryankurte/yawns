package layers

import (
	"image/color"
	"log"

	"github.com/ryankurte/go-mapbox/lib/maps"
	"github.com/ryankurte/go-rf"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

// FoliageLayer implements Weissburg fading using a map tile with foliage areas blacked out.
type FoliageLayer struct {
	foliage maps.Tile
}

// NewFoliageLayer creates a new foliage layer from the provided map configuration
func NewFoliageLayer(c *config.Maps) (*FoliageLayer, error) {
	t := FoliageLayer{}

	foliageImg, _, err := maps.LoadImage(c.Foliage)
	if err != nil {
		log.Printf("Error loading %s", c.Foliage)
		return nil, err
	}
	t.foliage = maps.NewTile(c.X, c.Y, c.Level, 512, foliageImg)

	return &t, nil
}

// CalculateFading calculates the free space fading for a link
func (t *FoliageLayer) CalculateFading(band config.Band, p1, p2 types.Location) (float64, error) {
	p1m, p2m := onsToMapLoc(&p1), onsToMapLoc(&p2)

	// Calculate LoS distance between two points
	distance := rf.CalculateDistanceLOS(p1.Lat, p1.Lng, p1.Alt, p2.Lat, p2.Lng, p2.Alt)

	// Fetch foliage map between points
	foliage, notFoliage := 0.0, 0.0
	t.foliage.InterpolateLocations(p1m, p2m, func(pixel color.Color) color.Color {
		r, g, b, a := pixel.RGBA()
		if r == 0 && g == 0 && b == 0 && a != 0 {
			foliage++
		} else {
			notFoliage++
		}
		return pixel
	})

	impingement := float64(distance) / (foliage + notFoliage) * foliage

	f, err := rf.CalculateFoliageLoss(rf.Frequency(band.Frequency), rf.Distance(impingement))

	return float64(f), err
}

package layers

import (
	"testing"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
	"github.com/stretchr/testify/assert"
)

func TestMapLayer(t *testing.T) {

	// Load example config
	c, err := config.LoadConfigFile("../../../example.yml")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	cfg := config.Maps{
		X:         64584,
		Y:         39988,
		Level:     16,
		Satellite: "../mapbox-satellite-16-64584-39988-9x5-512.jpg",
		Terrain:   "../mapbox-terrain-rgb-16-64584-39988-9x5-512.png",
	}

	band := config.Band{
		Frequency: 433e6,
	}

	l0 := types.Location{Lat: -36.8474505, Lng: 174.773418, Alt: 17.60}
	l1 := types.Location{Lat: -36.845286, Lng: 174.816868, Alt: 2.10}
	l2 := types.Location{Lat: -36.830456, Lng: 174.809947, Alt: 1.0}

	t.Run("Can render out map", func(t *testing.T) {
		renderLayer, err := NewRenderLayer(&cfg)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		links := make(types.Links, 0)

		for i := 0; i < len(c.Nodes); i++ {
			for j := i; j < len(c.Nodes); j++ {
				links = append(links, types.Link{
					A: i,
					B: j,
				})
			}
		}

		renderLayer.Render("./map-render-test-01.png", c.Nodes, links)
	})

	t.Run("Calculates terrain fading", func(t *testing.T) {
		t.SkipNow()

		terrainLayer, err := NewTerrainLayer(&cfg)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		fading, err := terrainLayer.CalculateFading(band, l0, l1)
		assert.Nil(t, err)
		assert.InDelta(t, 6.0, fading, 0.1)

		fading, err = terrainLayer.CalculateFading(band, l1, l2)
		assert.Nil(t, err)
		assert.InDelta(t, 0.0, fading, 0.1)
	})

}

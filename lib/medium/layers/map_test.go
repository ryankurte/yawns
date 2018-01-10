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

	// Rewrite path to images
	c.Medium.Maps.Satellite = c.Medium.Maps.Satellite
	c.Medium.Maps.Terrain = c.Medium.Maps.Terrain

	t.Run("Can render out map", func(t *testing.T) {
		renderLayer, err := NewRenderLayer(&c.Medium.Maps)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		links := make([]types.Link, 0)

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

		terrainLayer, err := NewTerrainLayer(&c.Medium.Maps)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		fading, err := terrainLayer.CalculateFading(c.Medium.Bands["433MHz"], c.Nodes[0].Location, c.Nodes[4].Location)
		assert.Nil(t, err)
		assert.InDelta(t, 6.0, fading, 0.1)

		fading, err = terrainLayer.CalculateFading(c.Medium.Bands["433MHz"], c.Nodes[4].Location, c.Nodes[5].Location)
		assert.Nil(t, err)
		assert.InDelta(t, 0.0, fading, 0.1)
	})

}

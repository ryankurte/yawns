package layers

import (
	"testing"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
	//"github.com/stretchr/testify/assert"
)

func TestMapLayer(t *testing.T) {

	// Load example config
	c, err := config.LoadConfigFile("../../../example.yml")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Rewrite path to images
	c.Medium.Maps.Satellite = "../../../" + c.Medium.Maps.Satellite
	c.Medium.Maps.Terrain = "../../../" + c.Medium.Maps.Terrain

	// Create map layer
	mapLayer, err := NewMap(&c.Medium.Maps)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Run("Can render out map", func(t *testing.T) {
		links := make([]types.Link, 0)

		for i := 0; i < len(c.Nodes); i++ {
			for j := i; j < len(c.Nodes); j++ {
				links = append(links, types.Link{
					A: i,
					B: j,
				})
			}
		}

		mapLayer.Render("./map-render-test-01.png", c.Nodes, links)
	})

}

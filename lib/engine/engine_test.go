package engine

import (
	"math"
	"strconv"
	"testing"
)

import (
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/connector"
)

func FloatEq(a, b float64) bool {
	diff := math.Abs(a - b)
	avg := math.Abs((a + b) / 2)
	if diff > 0.01*avg {
		return false
	}
	return true
}

func TestEngine(t *testing.T) {

	var e *Engine
	connector := connector.NewZMQConnector()

	t.Run("Create from config", func(t *testing.T) {
		c := config.Config{}

		node := config.Node{Address: "TestAddress", Location: config.Location{Lat: 0.0, Lng: 0.0}}
		c.Nodes = append(c.Nodes, node)

		e = NewEngine(connector)

		e.LoadConfig(&c)
	})

	t.Run("Handles location updates", func(t *testing.T) {

		lat := -87.3245
		lng := 35.47

		// Location update data
		updateData := make(map[string]string)
		updateData["lat"] = strconv.FormatFloat(lat, 'f', 6, 64)
		updateData["lon"] = strconv.FormatFloat(lng, 'f', 6, 64)

		err := e.handleUpdate("TestAddress", config.UpdateSetLocation, updateData)
		if err != nil {
			t.Error(err)
		}

		node, err := e.getNode("TestAddress")
		if err != nil {
			t.Error(err)
		}

		if !FloatEq(float64(node.Location.Lat), lat) {
			t.Errorf("Failed to set latitude (actual %f, expected %f)", node.Location.Lat, lat)
		}
		if !FloatEq(float64(node.Location.Lng), lng) {
			t.Errorf("Failed to set longitude (actual %f, expected %f)", node.Location.Lng, lng)
		}

	})

}

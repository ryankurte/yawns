package engine

import (
	"math"
	"strconv"
	"testing"
)

import (
	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/connector"
	"github.com/ryankurte/owns/lib/types"
	"time"
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
	connector := connector.NewZMQConnector(connector.DefaultIPCAddress)

	t.Run("Create from config", func(t *testing.T) {
		cfg := config.Config{}

		node := types.Node{Address: "TestAddress", Location: types.Location{Lat: 0.0, Lng: 0.0}}
		cfg.Nodes = append(cfg.Nodes, node)

		e = NewEngine(&cfg)
		e.BindConnectorChannels(connector.OutputChan, connector.InputChan)

		e.Setup(false)

		go e.Run()
	})

	t.Run("Handles location Events", func(t *testing.T) {

		lat := -87.3245
		lng := 35.47

		// Location Event data
		EventData := make(map[string]string)
		EventData["lat"] = strconv.FormatFloat(lat, 'f', 6, 64)
		EventData["lon"] = strconv.FormatFloat(lng, 'f', 6, 64)

		err := e.handleNodeUpdate(time.Second, "TestAddress", config.UpdateSetLocation, EventData)
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

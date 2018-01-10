package config

import (
	"fmt"
	"testing"

	"github.com/ryankurte/owns/lib/types"
)

func TestConfigLoading(t *testing.T) {

	t.Run("Parses config from example file", func(t *testing.T) {
		t.SkipNow()

		c, err := LoadConfigFile("../../example.yml")
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if c == nil {
			t.Errorf("Config object is nil")
		}

		c.Info()

		fmt.Printf("Config: %+v\n", c)
	})

	t.Run("Marshals config to file", func(t *testing.T) {
		c := Config{}

		node := types.Node{Address: "TestAddress", Location: types.Location{Lat: 0.0, Lng: 0.0, Alt: 100.0}}
		c.Nodes = append(c.Nodes, node)

		UpdateData := make(map[string]string)
		UpdateData["lat"] = "1.0"
		UpdateData["lon"] = "2.0"

		Update := Update{1000, []string{"TestAddress"}, UpdateSetLocation, UpdateData, "Test Comment"}
		c.Updates = append(c.Updates, Update)

		err := WriteConfigFile("/tmp/ons-test.yml", &c)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	})

}

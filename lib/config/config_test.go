package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/ryankurte/yawns/lib/types"
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

		Update := Update{Action: UpdateSetLocation, TimeStamp: 1 * time.Second, Nodes: []string{"TestAddress"}, Data: UpdateData, Comment: "Test Comment"}
		c.Updates = append(c.Updates, Update)

		err := WriteConfigFile("/tmp/ons-test.yml", &c)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	})

}

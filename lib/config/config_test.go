package config

import (
	"fmt"
	"testing"
)

func TestConfigLoading(t *testing.T) {

	t.Run("Marshal Config", func(t *testing.T) {
		c := Config{}

		node := Node{Address: "TestAddress", Location: Location{Lat: 0.0, Lng: 0.0, Alt: 100.0}}
		c.Nodes = append(c.Nodes, node)

		updateData := make(map[string]string)
		updateData["lat"] = "1.0"
		updateData["lon"] = "2.0"

		update := Update{1000, []string{"TestAddress"}, UpdateSetLocation, updateData}
		c.Updates = append(c.Updates, update)

		err := WriteConfigFile("/tmp/ons-test.yml", &c)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	})

	t.Run("Parse config file", func(t *testing.T) {

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

}

/**
 * OpenNetworkSim Helper Package
 * Contains helper functions for use in other module
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package helpers

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/go-yaml/yaml"
)

// ParseFieldToFloat64 grabs a field from a map and converts it to a float64
func ParseFieldToFloat64(name string, data map[string]string) (float64, error) {
	field, ok := data[name]
	if !ok {
		return 0.0, fmt.Errorf("ParseFieldToFloat64 error field %s not found (data: %+v)", name, data)
	}

	fieldFloat, err := strconv.ParseFloat(field, 64)
	if err != nil {
		return 0.0, fmt.Errorf("ParseFieldToFloat64 error field %s is not a float (data: %+v)", name, data)
	}
	return fieldFloat, nil
}

// ParseFieldToInt grabs a field from a map and converts it to an int
func ParseFieldToInt(name string, data map[string]string) (int, error) {
	field, ok := data[name]
	if !ok {
		return 0.0, fmt.Errorf("ParseFieldToInt error field %s not found (data: %+v)", name, data)
	}

	fieldInt, err := strconv.Atoi(field)
	if err != nil {
		return 0.0, fmt.Errorf("ParseFieldToInt error field %s is not a float (data: %+v)", name, data)
	}
	return fieldInt, nil
}

func ReadYAMLFile(name string, data interface{}) error {
	d, err := ioutil.ReadFile(name)
	if err != nil {
		return fmt.Errorf("ReadYAMLFile error loading file (%s)", err)
	}

	err = yaml.Unmarshal(d, data)
	if err != nil {
		return fmt.Errorf("ReadYAMLFile error parsing file (%s)", err)
	}

	return nil
}

func WriteYAMLFile(name string, data interface{}) error {
	d, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("WriteYAMLFile error parsing config (%s)", err)
	}

	err = ioutil.WriteFile(name, d, 0644)
	if err != nil {
		return fmt.Errorf("WriteYAMLFile error writing file (%s)", err)
	}

	return nil
}

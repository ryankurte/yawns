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
	"strconv"
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

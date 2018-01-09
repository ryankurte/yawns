/**
 * OpenNetworkSim Medium Layers package
 * Medium simulation / algorithms are implemented as layers that are included in the simulation
 * based on the provided simulation configuration
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package layers

import (
	"fmt"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/types"
)

// FadingInterface interface for layers implementing fading calculations
type FadingInterface interface {
	CalculateFading(band config.Band, p1, p2 types.Location) (float64, error)
	//CalculateFadingBounds(band config.Band, p1, p2 types.Location) (float64, float64)
}

// InfoInterface allows layers to return arbitrary info structures
type InfoInterface interface {
	GetInfo() interface{}
}

// RenderInterface interface for layers implementing rendering functions
type RenderInterface interface {
	Render(fileName string, nodes []types.Node, links []types.Link) error
}

// LayerManager manages a set of medium layers
type LayerManager struct {
	FadingInterfaces map[string]FadingInterface
	InfoInterfaces   map[string]InfoInterface
	RenderInterface  RenderInterface
}

// NewLayerManager creates a new medium layer manager
func NewLayerManager() *LayerManager {
	return &LayerManager{
		FadingInterfaces: make(map[string]FadingInterface),
		InfoInterfaces:   make(map[string]InfoInterface),
	}
}

// BindLayer binds a layer into the layer manager
// This checks the layer against available interfaces and binds where matches are found
func (lm *LayerManager) BindLayer(name string, layer interface{}) error {
	match := false
	if fading, ok := layer.(FadingInterface); ok {
		lm.FadingInterfaces[name] = fading
		match = true
	}
	if info, ok := layer.(InfoInterface); ok {
		lm.InfoInterfaces[name] = info
		match = true
	}
	if render, ok := layer.(RenderInterface); ok {
		lm.RenderInterface = render
		match = true
	}
	if !match {
		return fmt.Errorf("No matching layer interfaces found for %t", layer)
	}

	return nil
}

func (lm *LayerManager) GetLayer(name string) (interface{}, error) {
	f, ok := lm.FadingInterfaces[name]
	if ok {
		return f, nil
	}
	i, ok := lm.InfoInterfaces[name]
	if ok {
		return i, nil
	}
	return nil, nil
}

// CalculateFading calculates the overall fading using the provided layers
func (lm *LayerManager) CalculateFading(band config.Band, p1, p2 types.Location) (float64, error) {
	fading := 0.0
	layers := make(map[string]float64)

	for name, layer := range lm.FadingInterfaces {
		layerFading, _ := layer.CalculateFading(band, p1, p2)
		fading += layerFading
		layers[name] = layerFading
	}

	fmt.Printf("Fading %+v total: %.2f dB\n", layers, fading)

	return fading, nil
}

func (lm *LayerManager) Render(filename string, nodes []types.Node, links []types.Link) error {
	if lm.RenderInterface != nil {
		return lm.RenderInterface.Render(filename, nodes, links)
	}
	return fmt.Errorf("No render layer bound")
}

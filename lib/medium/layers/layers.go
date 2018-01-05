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

// FadingLayer interface for layers implementing fading calculations
type FadingLayer interface {
	CalculateFading(band config.Band, p1, p2 types.Location) (float64, error)
	//CalculateFadingBounds(band config.Band, p1, p2 types.Location) (float64, float64)
}

// InfoLayer allows layers to return arbitrary info structures
type InfoLayer interface {
	GetInfo() interface{}
}

// RenderLayer interface for layers implementing rendering functions
type RenderLayer interface {
	Render(fileName string, nodes []types.Node, links []types.Link) error
}

// LayerManager manages a set of medium layers
type LayerManager struct {
	fadingLayers map[string]FadingLayer
	infoLayers   map[string]InfoLayer
	renderLayer  RenderLayer
}

// NewLayerManager creates a new medium layer manager
func NewLayerManager() *LayerManager {
	return &LayerManager{
		fadingLayers: make(map[string]FadingLayer),
		infoLayers:   make(map[string]InfoLayer),
	}
}

// BindLayer binds a layer into the layer manager
// This checks the layer against available interfaces and binds where matches are found
func (lm *LayerManager) BindLayer(name string, layer interface{}) error {
	match := false
	if fading, ok := layer.(FadingLayer); ok {
		lm.fadingLayers[name] = fading
		match = true
	}
	if info, ok := layer.(InfoLayer); ok {
		lm.infoLayers[name] = info
		match = true
	}
	if render, ok := layer.(RenderLayer); ok {
		lm.renderLayer = render
		match = true
	}
	if !match {
		return fmt.Errorf("No matching layer interfaces found for %t", layer)
	}

	return nil
}

func (lm *LayerManager) GetLayer(name string) (interface{}, error) {
	f, ok := lm.fadingLayers[name]
	if ok {
		return f, nil
	}
	i, ok := lm.infoLayers[name]
	if ok {
		return i, nil
	}
	return nil, nil
}

// CalculateFading calculates the overall fading using the provided layers
func (lm *LayerManager) CalculateFading(band config.Band, p1, p2 types.Location) (float64, error) {
	fading := 0.0
	fmt.Printf("Fading ")
	for name, layer := range lm.fadingLayers {
		layerFading, _ := layer.CalculateFading(band, p1, p2)
		fading += layerFading
		fmt.Printf("%s: %.2f ", name, layerFading)
	}
	fmt.Printf("total: %.2f\n", fading)

	return fading, nil
}

func (lm *LayerManager) Render(filename string, nodes []types.Node, links []types.Link) error {
	if lm.renderLayer != nil {
		return lm.renderLayer.Render(filename, nodes, links)
	}
	return fmt.Errorf("No render layer bound")
}

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
	CalculateFading(band config.Band, p1, p2 types.Location) float64
	//CalculateFadingBounds(band config.Band, p1, p2 types.Location) (float64, float64)
}

// InfoLayer allows layers to return arbitrary info structures
type InfoLayer interface {
	GetInfo() interface{}
}

// RenderLayer interface for layers implementing rendering functions
type RenderLayer interface {
	Render()
}

// LayerManager manages a set of medium layers
type LayerManager struct {
	fadingLayers []FadingLayer
	infoLayers   []InfoLayer
	renderLayers []RenderLayer
}

// NewLayerManager creates a new medium layer manager
func NewLayerManager() *LayerManager {
	return &LayerManager{
		fadingLayers: make([]FadingLayer, 0),
		infoLayers:   make([]InfoLayer, 0),
		renderLayers: make([]RenderLayer, 0),
	}
}

// BindLayer binds a layer into the layer manager
// This checks the layer against available interfaces and binds where matches are found
func (lm *LayerManager) BindLayer(layer interface{}) error {
	match := false
	if fading, ok := layer.(FadingLayer); ok {
		lm.fadingLayers = append(lm.fadingLayers, fading)
		match = true
	}
	if info, ok := layer.(InfoLayer); ok {
		lm.infoLayers = append(lm.infoLayers, info)
		match = true
	}
	if render, ok := layer.(RenderLayer); ok {
		lm.renderLayers = append(lm.renderLayers, render)
		match = true
	}
	if !match {
		return fmt.Errorf("No matching layer interfaces found for %t", layer)
	}

	return nil
}

// CalculateFading calculates the overall fading using the provided layers
func (lm *LayerManager) CalculateFading(band config.Band, p1, p2 types.Location) float64 {
	fading := 0.0
	for _, layer := range lm.fadingLayers {
		fading += layer.CalculateFading(band, p1, p2)
	}
	return fading
}

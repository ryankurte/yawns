package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/ryankurte/go-mapbox/lib"
	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"

	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/types"
)

type Options struct {
	Config flags.Filename `short:"c" long:"config" description:"ONS Configuration file (used to automatically parse some options)" `
	APIKey string         `short:"a" long:"api-key" description:"Mapbox API key" env:"MAPBOX_TOKEN"`
	Type   string         `short:"t" long:"map-type" description:"Map download type" default:"satellite"`
	Level  int            `short:"l" long:"level" description:"Map level" default:"16" required:"yes"`

	Output    string `short:"o" long:"output-dir" description:"Output directory" default:"ons-map" default:""`
	Cache     string `short:"d" long:"cache-dir" description:"Cache directory" default:"/tmp/ons/"`
	Flatten   bool   `long:"flatten-terrain" description:"Flattens terrain images to greyscale for human consumption"`
	NoHighDPI bool   `long:"no-high-dpi" description:"Uses standard (not high DPI) tiles"`
}

func main() {
	fmt.Printf("OWNS Map Fetching Utility\n")

	options := Options{}

	parser := flags.NewParser(&options, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	if options.APIKey == "" {
		fmt.Printf("Mapbox API key must be specified\n")
		os.Exit(-1)
	}

	if options.Config == "" {
		fmt.Printf("Configuration file must be specified\n")
		os.Exit(-1)
	}

	// Create mapbox connector
	mbox := mapbox.NewMapbox(options.APIKey)
	cache, _ := maps.NewFileCache(options.Cache)
	mbox.Maps.SetCache(cache)

	// Load configuration file
	config, err := config.LoadConfigFile(string(options.Config))
	if err != nil {
		fmt.Printf("Error loading config file %s\n", err)
		os.Exit(-1)
	}

	p1, p2 := types.GetNodeBounds(config.Nodes)

	p1a := base.Location{Latitude: p1.Lat, Longitude: p1.Lng}
	p2a := base.Location{Latitude: p2.Lat, Longitude: p2.Lng}

	fmt.Printf("Fetching tiles...\n")

	mapID := maps.MapIDSatellite
	mapFormat := maps.MapFormatJpg90
	extension := "jpg"

	switch options.Type {
	case "terrain":
		mapID = maps.MapIDTerrainRGB
		mapFormat = maps.MapFormatPngRaw
		extension = "png"
	case "outdoors":
		mapID = maps.MapIDOutdoors
		mapFormat = maps.MapFormatPng
		extension = "png"
	}

	tiles, err := mbox.Maps.GetEnclosingTiles(mapID,
		p1a, p2a, uint64(options.Level), mapFormat, !options.NoHighDPI)
	if err != nil {
		fmt.Printf("Error fetching map tiles: %s\n", err)
		os.Exit(-1)
	}

	tile := maps.StitchTiles(tiles)
	xCount, yCount := tile.Bounds().Dx()/int(tile.Size), tile.Bounds().Dy()/int(tile.Size)

	fmt.Printf("\nTile Information:\n")
	fmt.Printf("  - Mapbox ID: %s\n", string(mapID))
	fmt.Printf("  - X: %d Y: %d Level: %d\n", tile.X, tile.Y, tile.Level)
	fmt.Printf("  - Base tile size: %d\n", tile.Size)
	fmt.Printf("  - Stitched tiles X: %d Y: %d\n", xCount, yCount)

	if options.Flatten {
		maxAlt := tile.GetHighestAltitude()
		tile = tile.FlattenAltitudes(maxAlt + 1)
		fmt.Printf("  - Flattened to maximum altitude of: %.2fm\n", maxAlt)
	}

	fmt.Printf("\n")

	fileName := fmt.Sprintf("%s%s-%d-%d-%d-%dx%d-%d.%s", options.Output, strings.Replace(string(mapID), ".", "-", -1), tile.Level, tile.X, tile.Y, xCount, yCount, tile.Size, extension)
	switch mapFormat {
	case maps.MapFormatJpg90:
		err = maps.SaveImageJPG(tile, fileName)
	case maps.MapFormatPng:
		err = maps.SaveImagePNG(tile, fileName)
	case maps.MapFormatPngRaw:
		err = maps.SaveImagePNG(tile, fileName)
	}

	if err != nil {
		fmt.Printf("Error saving map output: %s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Output map written to: %s\n", fileName)
}

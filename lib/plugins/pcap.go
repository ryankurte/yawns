/**
 * PCap file output plugin
 * File format information from https://wiki.wireshark.org/Development/LibpcapFileFormat
 *
 * Copyright 2017 Ryan Kurte
 */

package plugins

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryankurte/go-pcapng"
	"github.com/ryankurte/go-pcapng/types"

	"github.com/ryankurte/owns/lib/config"
)

const (
	// FileName configuration option
	FileName string = "file"
	// LinkType configuration option
	LinkType string = "linktype"

	// Default (no config) pcap file name
	DefaultFileName = "./owns.pcap"

	// Default (no config) link type
	DefaultLinkType = types.LinkTypeIEEE802_15_4
)

// PCAPPlugin is a plugin to write PCAP files from the simulation
type PCAPPlugin struct {
	fileWriter   *pcapng.FileWriter
	interfaceIDs map[string]int
	bands        map[string]config.Band
	startTime    time.Time
}

// NewPCAPPlugin creates a new PCAP file writer plugin
func NewPCAPPlugin(bands map[string]config.Band, startTime time.Time, options map[string]interface{}) (*PCAPPlugin, error) {
	var err error

	// Fetch file name
	fileName, err := GetOptionString(FileName, DefaultFileName, options)
	if err != nil {
		return nil, err
	}

	// Fetch link type override
	linkType, err := GetOptionUint(LinkType, uint32(DefaultLinkType), options)
	if err != nil {
		return nil, err
	}

	// Open capture file
	writer, err := pcapng.NewFileWriter(fileName)
	if err != nil {
		return nil, err
	}

	// Write section header
	sectionOpts := types.SectionHeaderOptions{
		Application: "Open Wireless Network Sim",
		Comment:     time.Now().String(),
	}
	if err := writer.WriteSectionHeader(sectionOpts); err != nil {
		return nil, err
	}

	interfaceIDs := make(map[string]int)
	index := 0

	// Write a header for each band
	for k, b := range bands {
		desc, err := json.Marshal(b)
		if err != nil {
			return nil, err
		}

		interfaceOpts := types.InterfaceOptions{
			Name:        k,
			Description: string(desc),
			Speed:       uint64(b.Baud),
		}

		if err := writer.WriteInterfaceDescription(uint16(linkType), interfaceOpts); err != nil {
			return nil, err
		}

		interfaceIDs[k] = index
		index++
	}

	return &PCAPPlugin{fileWriter: writer, interfaceIDs: interfaceIDs, bands: bands, startTime: startTime}, nil
}

// Received logs a received packet
func (p *PCAPPlugin) Received(d time.Duration, band string, address string, message []byte) error {
	interfaceID, ok := p.interfaceIDs[band]
	if !ok {
		return fmt.Errorf("Unrecognised band name (%s)", band)
	}

	_, ok = p.bands[band]
	if !ok {
		return fmt.Errorf("Unable to locate band (%s)", band)
	}

	packetOpts := types.EnhancedPacketOptions{
		//OriginalLength: uint32(len(message) + int(bandInfo.PacketOverhead)),
		Comment: fmt.Sprintf("Simulator address: %s", address),
	}

	return p.fileWriter.WriteEnhancedPacketBlock(uint32(interfaceID), p.startTime.Add(d), message, packetOpts)
}

// Close closes the pcap file
func (p *PCAPPlugin) Close() {
	p.fileWriter.Close()
}

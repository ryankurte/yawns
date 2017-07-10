/**
 * PCap file output plugin
 * File format information from https://wiki.wireshark.org/Development/LibpcapFileFormat
 *
 * Copyright 2017 Ryan Kurte
 */

package plugins

import (
	"bufio"
	"encoding/binary"
	"os"
	"time"
)

const (
	// LinkTypeIEEE802_15_4 IEEE802.15.4 link type
	LinkTypeIEEE802_15_4 uint32 = 195
	// LinkTypePrivate Private link type
	LinkTypePrivate uint32 = 147
	// FileName configuration option
	FileName string = "file"
	// LinkType configuration option
	LinkType string = "linktype"
)

var DefaultFileName = "./owns.pcap"
var DefaultLinkType = LinkTypePrivate

// GlobalHeader PCap file global header
type GlobalHeader struct {
	Magic          uint32
	VersionMajor   uint16
	VersionMinor   uint16
	Timezone       uint32
	TimeAccuracy   uint32
	SnapshotLength uint32
	Network        uint32
}

// BuildGlobalHeader builds a global header instance
func BuildGlobalHeader(linkType uint32) GlobalHeader {
	return GlobalHeader{
		Magic:        0xa1b2c3d4,
		VersionMajor: 2,
		VersionMinor: 4,
		Timezone:     0,
		TimeAccuracy: 0,
		Network:      linkType,
	}
}

// PacketHeader PCap file packet header
type PacketHeader struct {
	Seconds        uint32
	Micros         uint32
	IncludedLength uint32
	OriginalLength uint32
}

// PCAPPlugin is a plugin to write PCAP files from the simulation
type PCAPPlugin struct {
	f *os.File
	b *bufio.Writer
}

// NewPCAPPlugin creates a new PCAP file writer plugin
func NewPCAPPlugin(options map[string]interface{}) (*PCAPPlugin, error) {
	p := PCAPPlugin{}
	var err error

	// Fetch file name
	fileName, err := GetOptionString(FileName, DefaultFileName, options)
	if err != nil {
		return nil, err
	}

	// Fetch link type override
	linkType, err := GetOptionUint(LinkType, DefaultLinkType, options)
	if err != nil {
		return nil, err
	}

	// Open capture file
	p.f, err = os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	p.b = bufio.NewWriter(p.f)

	// Write file header
	if err = p.writeGlobalHeader(linkType); err != nil {
		return nil, err
	}

	return &p, nil
}

// Received logs a received packet
func (p *PCAPPlugin) Received(d time.Duration, address string, message []byte) error {
	return p.writePacket(d, message)
}

// Close closes the pcap file
func (p *PCAPPlugin) Close() {
	p.b.Flush()
	p.f.Close()
}

// Helper to write the global header that starts a pcap file
func (p *PCAPPlugin) writeGlobalHeader(linkType uint32) error {
	// Constant header components
	header := BuildGlobalHeader(linkType)
	return binary.Write(p.b, binary.LittleEndian, &header)
}

// Helper to write a packet to the pcap file
func (p *PCAPPlugin) writePacket(d time.Duration, data []byte) error {

	header := PacketHeader{
		Seconds:        uint32(d.Seconds()),
		Micros:         uint32(d.Nanoseconds() % 1e9),
		IncludedLength: uint32(len(data)),
		OriginalLength: uint32(len(data)),
	}

	// Write the header
	if err := binary.Write(p.b, binary.LittleEndian, &header); err != nil {
		return err
	}
	// Write the data
	if _, err := p.b.Write(data); err != nil {
		return err
	}

	return nil
}

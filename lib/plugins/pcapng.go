package plugins

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

type BlockHeader struct {
	Type   uint32
	Length uint32
}

type BlockTrailer struct {
	Length uint32
}

type Block struct {
	BlockHeader
	Data []byte
	BlockTrailer
}

const (
	Magic                  uint32 = 0x1A2B3C4D
	BlockTypeSectionHeader uint32 = 0x0A0D0D0A
	MajorVersion           uint16 = 1
	MinorVersion           uint16 = 0
	SectionLengthDefault   uint64 = 0xFFFFFFFFFFFFFFFF
)

type SectionHeaderBlock struct {
	Magic         uint32
	VersionMajor  uint16
	VersionMinor  uint16
	SectionLength uint64
	//Options       []Block
}

// PCAPPlugin is a plugin to write PCAP files from the simulation
type PCAPNGPlugin struct {
	f *os.File
	b *bufio.Writer
}

func writeSectionHeader(w io.Writer, options []Block) error {
	sectionBuff := bytes.NewBuffer(nil)
	sectionHeader := SectionHeaderBlock{
		Magic:         Magic,
		VersionMajor:  MajorVersion,
		VersionMinor:  MinorVersion,
		SectionLength: SectionLengthDefault,
		//Options:       options,
	}
	err := binary.Write(sectionBuff, binary.LittleEndian, sectionHeader)
	if err != nil {
		return err
	}

	return writeBlock(w, BlockTypeSectionHeader, sectionBuff.Bytes())
}

func writeBlock(w io.Writer, blockType uint32, data []byte) error {
	length := uint32(len(data)) + 12
	if err := binary.Write(w, binary.LittleEndian, &BlockHeader{Type: blockType, Length: length}); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &BlockTrailer{Length: length}); err != nil {
		return err
	}
	return nil
}

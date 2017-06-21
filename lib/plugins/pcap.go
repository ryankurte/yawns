package plugins

import (
	"bufio"
	"encoding/binary"
	"os"
	"time"
)

const (
	// LinkTypeIEEE802_14_4 IEEE802.14.5 link type
	LinkTypeIEEE802_14_4 uint32 = 195
	// LinkTypePrivate Private link type
	LinkTypePrivate uint32 = 147
)

type PCAPPlugin struct {
	f *os.File
	b *bufio.Writer
}

func NewPACPPlugin(pcapFile string) (*PCAPPlugin, error) {
	p := PCAPPlugin{}
	var err error

	// Open capture file
	p.f, err = os.OpenFile(pcapFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	p.b = bufio.NewWriter(p.f)

	// Write header
	p.writeGlobalHeader(LinkTypePrivate)

	return &p, nil
}

func (p *PCAPPlugin) Received(t time.Time, address string, message []byte) {
	p.writePacket(t, message)
}

func (p *PCAPPlugin) Close() {
	p.b.Flush()
	p.f.Close()
}

func (p *PCAPPlugin) writeGlobalHeader(linkType uint32) {
	// Magic no.
	p.b.Write([]byte{0xa1, 0xb2, 0xc3, 0xd4})
	// Major version
	p.b.Write([]byte{0x02, 0x00})
	// Minor version
	p.b.Write([]byte{0x04, 0x00})
	// Timezone (all caps in UTC, always 0)
	p.b.Write([]byte{0x00, 0x00, 0x00, 0x00})
	// Time accuracy, also always 0
	p.b.Write([]byte{0x00, 0x00, 0x00, 0x00})
	// Snapshot length
	p.b.Write([]byte{0xff, 0xff, 0x00, 0x00})
	// Data link type
	binary.Write(p.b, binary.LittleEndian, linkType)
}

func (p *PCAPPlugin) writePacket(t time.Time, data []byte) {
	sec := t.UnixNano() / 1e9
	micro := (t.UnixNano() % 1e9) / 1e3

	// Capture time seconds component
	binary.Write(p.b, binary.LittleEndian, uint32(sec))
	// Capture time microsecond component
	binary.Write(p.b, binary.LittleEndian, uint32(micro))

	// Included and original data lengths
	binary.Write(p.b, binary.LittleEndian, uint32(len(data)))
	binary.Write(p.b, binary.LittleEndian, uint32(len(data)))

	// Actual data
	p.b.Write(data)
}

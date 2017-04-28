package types

import (
	"time"
)

type Packet struct {
	Address  string
	BandName string
	Data     []byte
	Sent     time.Time
}

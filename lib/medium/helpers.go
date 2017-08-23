package medium

import (
	"time"

	"github.com/ryankurte/owns/lib/types"
)

func (m *Medium) SetTransceiverState(now time.Time, nodeIndex int, band string, state types.TransceiverState) {
	transceiver := m.transceivers[nodeIndex][band]
	transceiver.SetState(now, state)
	m.transceivers[nodeIndex][band] = transceiver
}

package medium

import (
	"time"
)

func (m *Medium) SetTransceiverState(now time.Time, nodeIndex int, band string, state TransceiverState) {
	transceiver := m.transceivers[nodeIndex][band]
	transceiver.SetState(now, state)
	m.transceivers[nodeIndex][band] = transceiver
}

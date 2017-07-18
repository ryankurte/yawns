package medium

import (
//"log"
)

func (m *Medium) SetTransceiverState(nodeIndex int, band string, state TransceiverState) {
	//log.Printf("Updating node '%s' transceiver: '%s' from state: '%s' to: '%s'",
	//	(*m.nodes)[nodeIndex].Address, band, m.transceivers[nodeIndex][band], state)
	m.transceivers[nodeIndex][band] = state
}

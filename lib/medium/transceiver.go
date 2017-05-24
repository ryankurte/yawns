package medium

// TransceiverState is the state of a virtual transceiver
type TransceiverState string

// Allowed transceiver states
const (
	TransceiverStateIdle         TransceiverState = "idle"
	TransceiverStateReceive      TransceiverState = "receive"
	TransceiverStateReceiving    TransceiverState = "receiving"
	TransceiverStateTransmitting TransceiverState = "transmitting"
)

func (ts *TransceiverState) Set(state TransceiverState) { *ts = state }

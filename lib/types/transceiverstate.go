package types

// TransceiverState is the state of a virtual transceiver
type TransceiverState string

// Allowed transceiver states
const (
	TransceiverStateOff          TransceiverState = "off"
	TransceiverStateIdle         TransceiverState = "idle"
	TransceiverStateSleep        TransceiverState = "sleep"
	TransceiverStateReceive      TransceiverState = "receive"
	TransceiverStateReceiving    TransceiverState = "receiving"
	TransceiverStateTransmitting TransceiverState = "transmitting"
)

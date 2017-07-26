package medium

import (
	"time"
)

// TransceiverState is the state of a virtual transceiver
type TransceiverState string

// Allowed transceiver states
const (
	TransceiverStateIdle         TransceiverState = "idle"
	TransceiverStateSleep        TransceiverState = "sleep"
	TransceiverStateReceive      TransceiverState = "receive"
	TransceiverStateReceiving    TransceiverState = "receiving"
	TransceiverStateTransmitting TransceiverState = "transmitting"
)

type Transceiver struct {
	// Current transceiver state
	State TransceiverState

	lastTime time.Time

	// Time spent in idle mode
	IdleTime time.Duration
	// Time spent in sleep mode
	SleepTime time.Duration
	// Time spent in receive (listening) mode
	ReceiveTime time.Duration
	// Time spent receiving packets
	ReceivingTime time.Duration
	// Time spent transmitting packets
	TransmittingTime time.Duration
}

func NewTransceiver(startTime time.Time) *Transceiver {
	return &Transceiver{
		State:    TransceiverStateIdle,
		lastTime: startTime,
	}
}

func (t *Transceiver) SetState(now time.Time, state TransceiverState) {
	lastState := t.State
	stateTime := now.Sub(t.lastTime)

	switch lastState {
	case TransceiverStateIdle:
		t.IdleTime += stateTime
	case TransceiverStateSleep:
		t.SleepTime += stateTime
	case TransceiverStateReceive:
		t.ReceiveTime += stateTime
	case TransceiverStateReceiving:
		t.ReceivingTime += stateTime
	case TransceiverStateTransmitting:
		t.TransmittingTime += stateTime
	}

	t.State = state
	t.lastTime = now
}

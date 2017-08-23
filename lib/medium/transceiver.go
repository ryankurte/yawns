package medium

import (
	"time"

	"github.com/ryankurte/owns/lib/types"
)

type Transceiver struct {
	// Current transceiver state
	State types.TransceiverState

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
		State:    types.TransceiverStateIdle,
		lastTime: startTime,
	}
}

func (t *Transceiver) SetState(now time.Time, state types.TransceiverState) {
	lastState := t.State
	stateTime := now.Sub(t.lastTime)

	switch lastState {
	case types.TransceiverStateIdle:
		t.IdleTime += stateTime
	case types.TransceiverStateSleep:
		t.SleepTime += stateTime
	case types.TransceiverStateReceive:
		t.ReceiveTime += stateTime
	case types.TransceiverStateReceiving:
		t.ReceivingTime += stateTime
	case types.TransceiverStateTransmitting:
		t.TransmittingTime += stateTime
	}

	t.State = state
	t.lastTime = now
}

package medium

import (
	"time"

	"github.com/ryankurte/yawns/lib/types"
)

type Transceiver struct {
	// Current transceiver state
	State types.TransceiverState

	lastTime time.Time

	Stats TransceiverStats
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
	case types.TransceiverStateOff:
		t.Stats.OffTime += stateTime
	case types.TransceiverStateIdle:
		t.Stats.IdleTime += stateTime
	case types.TransceiverStateSleep:
		t.Stats.SleepTime += stateTime
	case types.TransceiverStateReceive:
		t.Stats.ReceiveTime += stateTime
	case types.TransceiverStateReceiving:
		t.Stats.ReceivingTime += stateTime
	case types.TransceiverStateTransmitting:
		t.Stats.TransmittingTime += stateTime
	}

	t.State = state
	t.lastTime = now
}

package plugins

import (
	"fmt"
	"sync"
	"time"

	"github.com/ryankurte/owns/lib/config"
)

type Field string
type Fields map[string]Field
type FieldMap map[string]Fields

// StateManager plugin tracks endpoint state for testing
type StateManager struct {
	fields     FieldMap
	fieldMutex sync.Mutex
	events     []StateEvent
	eventMutex sync.Mutex
}

// StateEvent is a simulation state event record
type StateEvent struct {
	Type     config.UpdateAction
	Time     time.Duration
	Address  string
	Key      string
	Expected string
	Actual   string
	Result   bool
	Error    error
}

// NewStateManager creates a new StateManger instance
func NewStateManager(addresses []string) StateManager {
	fields := make(FieldMap)
	for _, v := range addresses {
		fields[v] = make(Fields)
	}
	events := make([]StateEvent, 0)
	return StateManager{fields: fields, fieldMutex: sync.Mutex{}, events: events, eventMutex: sync.Mutex{}}
}

// OnEvent is called to handle simulation events
func (sm *StateManager) OnEvent(d time.Duration, eventType config.UpdateAction, address string, data map[string]string) {
	se := StateEvent{Type: eventType, Time: d, Address: address}
	var ok bool

	switch eventType {
	case config.UpdateCheckState:
		se.Key, ok = data["key"]
		if !ok {
			return
		}
		se.Expected, ok = data["value"]
		if !ok {
			return
		}
		se.Actual, se.Error = sm.getField(se.Address, se.Key)
		if se.Error == nil && se.Actual == se.Expected {
			se.Result = true
		} else {
			se.Result = false
		}

		sm.eventMutex.Lock()
		sm.events = append(sm.events, se)
		sm.eventMutex.Unlock()
	}
}

func (sm *StateManager) setField(address, key, value string) {
	sm.fieldMutex.Lock()
	defer sm.fieldMutex.Unlock()

	fields, ok := sm.fields[address]
	if !ok {
		return
	}

	fields[key] = Field(value)
	sm.fields[address] = fields
}

func (sm *StateManager) getField(address, key string) (string, error) {
	sm.fieldMutex.Lock()
	defer sm.fieldMutex.Unlock()

	fields, ok := sm.fields[address]
	if !ok {
		return "", fmt.Errorf("State not found for address '%s'", address)
	}

	value, ok := fields[key]
	if !ok {
		return "", fmt.Errorf("Field not found for address '%s' key '%s'", address, key)
	}

	return string(value), nil
}

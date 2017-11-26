package plugins

import (
	"fmt"
	"sync"
	"time"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/messages"
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

// OnMessage is is called to handle simulation messages
// This allows the plugin to receive, handle (and respond to) simulation messages
func (sm *StateManager) OnMessage(d time.Duration, message interface{}) error {

	switch m := message.(type) {
	case messages.FieldSet:
		sm.setField(m.Address, m.Name, string(m.Data))
	}

	return nil
}

// OnUpdate is called to handle simulation updates
// This allows fields to be set and checked at simulation time
func (sm *StateManager) OnUpdate(d time.Duration, eventType config.UpdateAction, address string, data map[string]string) error {
	se := StateEvent{Type: eventType, Time: d, Address: address, Result: false}
	var ok bool

	switch eventType {
	case config.UpdateCheckState:
		if se.Key, ok = data["key"]; !ok {
			se.Error = fmt.Errorf("No key in state update for address '%s'", address)
			break
		}
		if se.Expected, ok = data["value"]; !ok {
			se.Error = fmt.Errorf("No value in state updatefor address '%s'", address)
			break
		}
		se.Actual, se.Error = sm.getField(se.Address, se.Key)
		if se.Error == nil && se.Actual == se.Expected {
			se.Result = true
		}
	}

	sm.eventMutex.Lock()
	sm.events = append(sm.events, se)
	sm.eventMutex.Unlock()

	return nil
}

// setField sets the value of a field in the map
// A mutex is used here to ensure partial updates cannot occur
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

// getField fetches the value of a field from the map
// This is protected by the fieldMutex to ensure reads cannot occur during write operations
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

package plugins

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/messages"
)

func TestState(t *testing.T) {
	addresses := []string{"a", "b"}
	key, val, addr := "testName1", "testData1", addresses[0]
	sm := NewStateManager(addresses)

	t.Run("Handles set state messages", func(t *testing.T) {
		m := messages.FieldSet{
			BaseMessage: messages.NewBaseMessage(addresses[0]),
			Name:        key,
			Data:        []byte(val),
		}

		err := sm.OnMessage(time.Second, m)
		assert.Nil(t, err)

		val2, err := sm.getField(addresses[0], key)
		assert.Nil(t, err)
		assert.EqualValues(t, val, val2)
	})

	t.Run("Successful state comparison", func(t *testing.T) {
		data := map[string]string{
			"key":   key,
			"value": val,
		}

		err := sm.OnUpdate(time.Second, config.UpdateCheckState, addr, data)
		assert.Nil(t, err)

		assert.EqualValues(t, 1, len(sm.events))
		assert.EqualValues(t, true, sm.events[0].Result)
		assert.EqualValues(t, val, sm.events[0].Actual)
		assert.EqualValues(t, val, sm.events[0].Expected)
	})

	t.Run("Unsuccessful state comparison (invalid value)", func(t *testing.T) {
		data := map[string]string{
			"key":   key,
			"value": "notvalue",
		}

		err := sm.OnUpdate(time.Second, config.UpdateCheckState, addr, data)
		assert.Nil(t, err)

		assert.EqualValues(t, 2, len(sm.events))
		assert.EqualValues(t, false, sm.events[1].Result)
		assert.EqualValues(t, val, sm.events[1].Actual)
		assert.EqualValues(t, data["key"], sm.events[1].Key)
		assert.EqualValues(t, data["value"], sm.events[1].Expected)
	})

	t.Run("Unsuccessful state comparison (no data)", func(t *testing.T) {
		data := map[string]string{
			"key":   "notkey",
			"value": val,
		}

		err := sm.OnUpdate(time.Second, config.UpdateCheckState, addr, data)
		assert.Nil(t, err)

		assert.EqualValues(t, 3, len(sm.events))
		assert.EqualValues(t, false, sm.events[2].Result)
		assert.EqualValues(t, "", sm.events[2].Actual)
		assert.EqualValues(t, data["key"], sm.events[2].Key)
		assert.EqualValues(t, data["value"], sm.events[2].Expected)
	})
}

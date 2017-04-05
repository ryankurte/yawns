package medium

import (
	"testing"
)

import (
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/messages"
	"time"
)

func TestMedium(t *testing.T) {

	// TODO: might be better to statically create config than load example?
	c, err := config.LoadConfigFile("../../example.yml")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	m := NewMedium(c)

	go m.Run()

	t.Run("Loads nodes and creates paths", func(t *testing.T) {
		if len(m.nodes) != len(c.Nodes) {
			t.Errorf("Expected 4%d nodes from config file", len(c.Nodes))
		}

		if len(m.Links) != 10 {
			t.Errorf("Invalid link count (actual: %d expected: %d)", len(m.Links), 10)
		}
	})

	t.Run("Writes medium information to file", func(t *testing.T) {
		err := m.WriteYML("/tmp/ons-test-medium.yml")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Sends messages based on link budgets", func(t *testing.T) {

		// Node 1 can communicate with nodes 3 and 5
		message := messages.NewMessage(messages.Packet, "0x0001", []byte("test"))
		m.inCh <- message

		time.Sleep(100 * time.Millisecond)

		// First message will be send complete
		resp := <-m.outCh
		if addr := "0x0001"; resp.GetAddress() != addr {
			t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
		}
		if messageType := messages.PacketSent; resp.GetType() != messageType {
			t.Errorf("Incorrect message type (actual: %s expected: %s)", resp.GetType(), messageType)
		}

		// Responses should be in order
		resp = <-m.outCh
		if addr := "0x0002"; resp.GetAddress() != addr {
			t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
		}

		resp = <-m.outCh
		if addr := "0x0003"; resp.GetAddress() != addr {
			t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
		}

		resp = <-m.outCh
		if addr := "0x0004"; resp.GetAddress() != addr {
			t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
		}

	})

}

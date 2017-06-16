/**
 * OpenNetworkSim CZMQ Radio Driver Tests
 * Uses the libons.go cgo wrapper around libons to test the native ons connector
 * against the ONS controller
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package libons

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ryankurte/owns/lib/connector"
	"github.com/ryankurte/owns/lib/messages"
	//	"github.com/ryankurte/owns/lib/protocol"
)

func TestLibONS(t *testing.T) {

	clientAddress := "fakeClient"

	var server *connector.ZMQConnector
	var client *ONSConnector

	timeout := 1 * time.Second
	port := fmt.Sprintf("inproc:///ons-%s", uuid.NewV4())

	t.Run("Bind ZMQ Connector", func(t *testing.T) {
		server = connector.NewZMQConnector(port)
		log.Printf("Bound to port: %s", port)
	})

	t.Run("Init client", func(t *testing.T) {
		client = NewONSConnector()
		err := client.Init(port, clientAddress)
		if err != nil {
			t.Error(err)
			t.FailNow()
			t.Skip()
		}
	})

	t.Run("Client sends registration packet", func(t *testing.T) {
		select {
		case msg := <-server.OutputChan:
			reg, ok := msg.(*messages.Register)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, reg.Address)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}
	})

	t.Run("Client can message server", func(t *testing.T) {

		data := "Test Client Data String"
		client.Send([]byte(data))

		time.Sleep(100 * time.Millisecond)
		select {
		case msg := <-server.OutputChan:
			packet, ok := msg.(*messages.Packet)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, packet.Address)
			assert.EqualValues(t, data, packet.Data)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}
	})

	t.Run("Client starts with no messages", func(t *testing.T) {
		if client.CheckReceive() {
			t.Errorf("Client appears to have received message")
			t.FailNow()
		}
	})

	t.Run("Server can message client", func(t *testing.T) {

		data := "Test Server Data String"

		packet := messages.Packet{Message: messages.Message{Address: clientAddress}, Data: []byte(data)}
		server.InputChan <- packet

		time.Sleep(100 * time.Millisecond)

		if !client.CheckReceive() {
			t.Errorf("Receive callback not called")
			t.FailNow()
		}

		message, err := client.GetReceived()
		assert.Nil(t, err)

		if string(message) != data {
			t.Errorf("Data mismatch (expected '%s' received '%s')", data, message)
		}

	})

	t.Run("Client can request rssi", func(t *testing.T) {

		respond := func(t *testing.T, value float32) {
			select {
			case msg, ok := <-server.OutputChan:
				assert.True(t, ok)
				req, ok := msg.(*messages.RSSIRequest)
				assert.True(t, ok)

				resp := messages.RSSIResponse{Message: messages.Message{Address: req.Address}, RSSI: value}
				server.InputChan <- resp

			case <-time.After(timeout):
				t.Errorf("Timeout")
				t.FailNow()
			}
		}

		timer := time.AfterFunc(time.Second, func() {
			t.Errorf("Timeout")
			t.FailNow()
		})

		log.Printf("CCA Check 1")
		go respond(t, 10.0)
		time.Sleep(100)

		rssi, err := client.GetRSSI()
		assert.Nil(t, err)
		assert.InDelta(t, 10, rssi, 0.01)

		log.Printf("CCA Check 2")
		go respond(t, 76.5)
		time.Sleep(100)

		rssi, err = client.GetRSSI()
		assert.Nil(t, err)
		assert.InDelta(t, 76.5, rssi, 0.01)

		timer.Stop()
	})

	t.Run("Exit client", func(t *testing.T) {
		client.Close()
	})

	t.Run("Exit connector", func(t *testing.T) {
		server.Exit()
	})
}

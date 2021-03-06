/**
 * OpenNetworkSim CZMQ Radio Driver Tests
 * Uses the libyawns.go cgo wrapper around libyawns to test the native ons connector
 * against the ONS controller
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package libyawns

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ryankurte/yawns/lib/connector"
	"github.com/ryankurte/yawns/lib/messages"
	"github.com/ryankurte/yawns/lib/types"
)

func TestLibONS(t *testing.T) {

	clientAddress := "fakeClient"

	var server *connector.ZMQConnector
	var client *ONSConnector
	var radio *ONSRadio

	timeout := 1 * time.Second
	port := fmt.Sprintf("inproc:///ons-%s", uuid.NewV4())
	band := "test-band"

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
			reg, ok := msg.(messages.Register)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, reg.Address)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}
	})

	t.Run("Init radio", func(t *testing.T) {
		r, err := client.RadioInit(band)
		if err != nil {
			t.Error(err)
			t.FailNow()
			t.Skip()
		}
		radio = r
	})

	t.Run("Client can message server", func(t *testing.T) {

		data := "Test Client Data String"
		radio.Send(0, []byte(data))

		time.Sleep(100 * time.Millisecond)
		select {
		case msg := <-server.OutputChan:
			packet, ok := msg.(messages.Packet)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, packet.Address)
			assert.EqualValues(t, data, packet.Data)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}
	})

	t.Run("Client starts with no messages", func(t *testing.T) {
		if radio.CheckReceive() {
			t.Errorf("Client appears to have received message")
			t.FailNow()
		}
	})

	t.Run("Server can message client", func(t *testing.T) {

		data := "Test Server Data String"

		packet := messages.Packet{
			BaseMessage: messages.BaseMessage{Address: clientAddress},
			RFInfo:      messages.NewRFInfo(band, 0),
			Data:        []byte(data),
		}
		server.InputChan <- packet

		time.Sleep(100 * time.Millisecond)

		if !radio.CheckReceive() {
			t.Errorf("Receive callback not called")
			t.FailNow()
		}

		message, err := radio.GetReceived()
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
				req, ok := msg.(messages.RSSIRequest)
				assert.True(t, ok)

				resp := messages.RSSIResponse{
					BaseMessage: messages.BaseMessage{Address: req.Address},
					RFInfo:      messages.NewRFInfo(band, 0),
					RSSI:        value}
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

		rssi, err := radio.GetRSSI(0)
		assert.Nil(t, err)
		assert.InDelta(t, 10, rssi, 0.01)

		log.Printf("CCA Check 2")
		go respond(t, 76.5)
		time.Sleep(100)

		rssi, err = radio.GetRSSI(0)
		assert.Nil(t, err)
		assert.InDelta(t, 76.5, rssi, 0.01)

		timer.Stop()
	})

	t.Run("Client can request radio states", func(t *testing.T) {

		respond := func(t *testing.T, state types.TransceiverState) {
			select {
			case msg, ok := <-server.OutputChan:
				assert.True(t, ok)
				req, ok := msg.(messages.StateRequest)
				assert.True(t, ok)

				resp := messages.StateResponse{
					BaseMessage: messages.BaseMessage{Address: req.Address},
					RFInfo:      messages.NewRFInfo(band, 0),
					State:       state,
				}
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

		log.Printf("State Check 1")
		go respond(t, types.TransceiverStateIdle)
		time.Sleep(100)

		state, err := radio.GetState()
		assert.Nil(t, err)
		assert.EqualValues(t, 2, state)

		log.Printf("State Check 2")
		go respond(t, types.TransceiverStateReceiving)
		time.Sleep(100)

		state, err = radio.GetState()
		assert.Nil(t, err)
		assert.EqualValues(t, 4, state)

		timer.Stop()
	})

	t.Run("Client can set radio state", func(t *testing.T) {
		radio.StartReceive(7)

		time.Sleep(100 * time.Millisecond)
		select {
		case msg := <-server.OutputChan:
			packet, ok := msg.(messages.StateSet)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, packet.Address)
			assert.EqualValues(t, types.TransceiverStateReceive, packet.State)
			assert.EqualValues(t, band, packet.Band)
			assert.EqualValues(t, 7, packet.Channel)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}

		radio.StopReceive()

		time.Sleep(100 * time.Millisecond)
		select {
		case msg := <-server.OutputChan:
			packet, ok := msg.(messages.StateSet)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, packet.Address)
			assert.EqualValues(t, types.TransceiverStateIdle, packet.State)
			assert.EqualValues(t, band, packet.Band)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}

	})

	t.Run("Client can set fields", func(t *testing.T) {

		name := "test-name"
		data := "test-data"

		client.SetField(name, data)

		time.Sleep(100 * time.Millisecond)
		select {
		case msg := <-server.OutputChan:
			packet, ok := msg.(messages.FieldSet)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, packet.Address)
			assert.EqualValues(t, name, packet.Name)
			assert.EqualValues(t, data, packet.Data)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}
	})

	t.Run("Client can set fields with formatting", func(t *testing.T) {

		name := "test-name"
		format := "0x%.4x"
		data := uint32(1024)

		client.SetFieldf(name, format, data)

		time.Sleep(100 * time.Millisecond)
		select {
		case msg := <-server.OutputChan:
			packet, ok := msg.(messages.FieldSet)
			assert.True(t, ok)
			assert.EqualValues(t, clientAddress, packet.Address)
			assert.EqualValues(t, name, packet.Name)
			assert.EqualValues(t, fmt.Sprintf(format, data), packet.Data)

		case <-time.After(timeout):
			t.Errorf("Timeout")
			t.FailNow()
		}
	})

	t.Run("Exit radio", func(t *testing.T) {
		client.CloseRadio(radio)
	})

	t.Run("Exit client", func(t *testing.T) {
		client.Close()
	})

	t.Run("Exit connector", func(t *testing.T) {
		server.Exit()
	})
}

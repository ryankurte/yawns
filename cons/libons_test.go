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
)

import (
	"github.com/ryankurte/ons/lib/connector"
	"github.com/ryankurte/ons/lib/messages"
	"github.com/ryankurte/ons/lib/protocol"
	"github.com/satori/go.uuid"
)

func TestLibONS(t *testing.T) {

	clientAddress := "fakeClient"

	var server *connector.ZMQConnector
	var client *ONSConnector

	port := fmt.Sprintf("inproc:///ons-%s", uuid.NewV4())

	t.Run("Bind ZMQ Connector", func(t *testing.T) {
		server = connector.NewZMQConnector(port)
		log.Printf("Connector port: %s", port)
	})

	t.Run("Test init client", func(t *testing.T) {
		client = NewONSConnector()
		err := client.Init(port, clientAddress)
		if err != nil {
			t.Error(err)
			t.FailNow()
			t.Skip()
		}
	})

	t.Run("Client sends registration packet", func(t *testing.T) {

		timer := time.AfterFunc(time.Second, func() {
			t.Errorf("Timeout")
			t.FailNow()
		})

		msg := <-server.OutputChan

		log.Printf("Message %+v", msg)

		if msg.GetAddress() != clientAddress {
			t.Errorf("Received address mismatch (expected '%s' received '%s')", clientAddress, msg.GetAddress())
		}

		if msg.GetType() != messages.Connected {
			t.Errorf("OnConnected message type mismatch (expected '%s' received '%s')", messages.Connected, msg.GetType())
		}

		timer.Stop()
	})

	t.Run("Client can message server", func(t *testing.T) {

		data := "Test Client Data String"
		client.Send([]byte(data))

		time.Sleep(100 * time.Millisecond)

		timer := time.AfterFunc(time.Second, func() {
			t.Errorf("Timeout")
			t.FailNow()
		})

		msg := <-server.OutputChan

		log.Printf("Message %+v", msg)

		if msg.GetAddress() != clientAddress {
			t.Errorf("Received address mismatch (expected '%s' received '%s')", clientAddress, msg.GetAddress())
		}

		if string(msg.GetData()) != data {
			t.Errorf("Data mismatch (expected '%s' received '%s')", data, string(msg.GetData()))
		}

		if msg.GetType() != messages.Packet {
			t.Errorf("OnConnected message type mismatch (expected '%s' received '%s')", messages.Packet, msg.GetType())
		}

		timer.Stop()
	})

	t.Run("Client starts with no messages", func(t *testing.T) {
		if client.CheckReceive() {
			t.Errorf("Client appears to have received message")
			t.FailNow()
		}
	})

	t.Run("Server can message client", func(t *testing.T) {

		data := "Test Server Data String"
		server.Send(clientAddress, []byte(data))

		time.Sleep(100 * time.Millisecond)

		if !client.CheckReceive() {
			t.Errorf("Receive callback not called")
			t.FailNow()
		}

		message, err := client.GetReceived()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if string(message) != data {
			t.Errorf("Data mismatch (expected '%s' received '%s')", data, message)
		}

	})

	t.Run("Client can request cca", func(t *testing.T) {

		respond := func(name string, value bool) {
			select {
			case msg, ok := <-server.OutputChan:
				if !ok {
					return
				}
				if msg.GetType() == messages.CCAReq {
					resp := messages.NewMessage(messages.CCAResp, msg.GetAddress(), []byte{})
					resp.SetCCA(value)

					log.Printf("Response (instance %s) writing %+v", name, resp)

					server.InputChan <- resp
				}
			}
		}

		timer := time.AfterFunc(time.Second, func() {
			t.Errorf("Timeout")
			t.FailNow()
		})

		log.Printf("CCA Check 1")
		go respond("check-1", false)
		time.Sleep(100)

		cca, err := client.GetCCA()
		if err != nil {
			t.Error(err)
		}
		if cca != false {
			t.Errorf("CCA error")
		}

		log.Printf("CCA Check 2")
		go respond("check-2", true)
		time.Sleep(100)

		cca, err = client.GetCCA()
		if err != nil {
			t.Error(err)
		}
		if cca != true {
			t.Errorf("CCA error")
		}

		timer.Stop()
	})

	t.Run("Exit client", func(t *testing.T) {
		client.Close()
	})

	t.Run("Exit connector", func(t *testing.T) {
		server.Exit()
	})
}

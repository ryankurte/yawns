package libons

import (
	"testing"
	"time"
)

import (
	"github.com/ryankurte/ons/connectors/zmq"
)

type TestServerReceiver struct {
	Received  bool
	Address   string
	Connected string
	Messsage  []byte
}

func (tc *TestServerReceiver) Receive(address string, message []byte) {
	tc.Address = address
	tc.Messsage = message
	tc.Received = true
}

func (tc *TestServerReceiver) OnConnect(address string) {
	tc.Connected = address
}

func TestZMQ(t *testing.T) {

	clientAddress := "fakeClient"

	sr := TestServerReceiver{}

	var server *zmq.ZMQConnector
	var client *ONSConnector

	t.Run("Bind ZMQ Connector", func(t *testing.T) {
		c := zmq.NewZMQConnector()
		c.Init("inproc://test", &sr)
		server = c
	})

	t.Run("Test init client", func(t *testing.T) {
		c := NewONSConnector()
		c.Init("inproc://test", clientAddress)
		client = c
	})

	t.Run("Client can message server", func(t *testing.T) {
		sr.Received = false

		data := "Test Client Data String"
		client.Send([]byte(data))

		time.Sleep(100 * time.Millisecond)

		if !sr.Received {
			t.Errorf("Receive callback not called")
			t.FailNow()
		}

		if sr.Address != clientAddress {
			t.Errorf("Received address mismatch (expected '%s' received '%s')", clientAddress, sr.Address)
		}

		if string(sr.Messsage) != data {
			t.Errorf("Data mismatch (expected '%s' received '%s')", data, sr.Messsage)
		}

		if sr.Connected != clientAddress {
			t.Errorf("OnConnected address mismatch (expected '%s' received '%s')", clientAddress, sr.Address)
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

	t.Run("Exit client", func(t *testing.T) {
		client.Close()
	})

	t.Run("Exit connector", func(t *testing.T) {
		server.Exit()
	})
}

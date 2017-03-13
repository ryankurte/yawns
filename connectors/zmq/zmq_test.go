package zmq

import (
	"testing"
	"time"
)

type TestClientReceiver struct {
	Received bool
	Messsage []byte
}

func (tc *TestClientReceiver) Receive(message []byte) {
	tc.Messsage = message
	tc.Received = true
}

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

func (tc *TestServerReceiver) OnConnected(address string) {
	tc.Connected = address
}

func TestZMQ(t *testing.T) {

	clientAddress := "fakeClient"
	serverAddress := "fakeServer"

	sr := TestServerReceiver{}
	cr := TestClientReceiver{}

	var server *ZMQConnector
	var client *ZMQClient

	t.Run("Bind ZMQ Connector", func(t *testing.T) {
		c := NewZMQConnector(serverAddress, "inproc://test", &sr)
		go c.Run()
		server = c
	})

	t.Run("Test connect client", func(t *testing.T) {
		c := NewZMQClient(clientAddress, "inproc://test", &cr)
		go c.Run()
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
			t.Errorf("Received address mismatch (expected %s received %s)", clientAddress, sr.Address)
		}

		if string(sr.Messsage) != data {
			t.Errorf("Data mismatch (expected %s received %s)", data, sr.Messsage)
		}

		if sr.Connected != clientAddress {
			t.Errorf("OnConnected address mismatch (expected %s received %s)", clientAddress, sr.Address)
		}

	})

	t.Run("Server can message client", func(t *testing.T) {
		cr.Received = false

		data := "Test Server Data String"
		server.Send(clientAddress, []byte(data))

		time.Sleep(100 * time.Millisecond)

		if !cr.Received {
			t.Errorf("Receive callback not called")
			t.FailNow()
		}

		if string(cr.Messsage) != data {
			t.Errorf("Data mismatch (expected %s received %s)", data, cr.Messsage)
		}

	})

	t.Run("Exit connector", func(t *testing.T) {
		server.Exit()
	})
}

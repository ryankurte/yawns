/**
 * OpenNetworkSim Connector Protocol Definitions
 * Defines protocol message constants, construction, and parsing. This must match the definitions in libons
 * to allow the c language connector to interact with the ons server.
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package connector

import (
	"fmt"
)

import (
	"github.com/ryankurte/ons/lib/messages"

	"github.com/golang/protobuf/proto"
	"github.com/ryankurte/ons/lib/protocol"
)

// handleIncoming handles incoming messages from external sources (ie. from nodes to ONS)
func (c *ZMQConnector) handleIncoming(data [][]byte) error {

	if len(data) != 3 {
		return fmt.Errorf("Error parsing message, required 3 parts")
	}

	// Fetch ZMQ client ID
	clientID := data[0]

	// Decode message
	message := protocol.Base{}
	err := proto.Unmarshal(data[1], &message)
	if err != nil {
		return fmt.Errorf("Error parsing protobuf message (%s)", err)
	}

	// Register message is a special case as no address is available for lookup
	if m, ok := message.GetMessage().(*protocol.Base_Register); ok {
		// Bind address to ID lookup for sending
		address := m.Register.Address

		if _, ok := c.clients[address]; !ok {
			// Save to list
			c.clients[address] = clientID
			// Send connected event
			c.OutputChan <- messages.NewMessage(messages.Connected, address, []byte{})
		}

		return nil
	}

	// Perform ZMQ Client ID to address lookup
	address := c.findClientAddressByID(clientID)
	if address == "" {
		return fmt.Errorf("Received message for unknown clientID (%+v)", clientID)
	}

	// Handle messages
	switch m := message.GetMessage().(type) {
	case *protocol.Base_Deregister:

	case *protocol.Base_Packet:
		c.OutputChan <- messages.NewMessage(messages.Packet, address, m.Packet.Data)

	case *protocol.Base_SendComplete:
		//c.OutputChan <- messages.NewMessage(messages.SendComplete, address, m.Packet.Data)

	case *protocol.Base_StartReceive:

	case *protocol.Base_StopReceive:

	case *protocol.Base_Event:
		c.OutputChan <- messages.NewMessage(messages.Event, address, []byte(m.Event.Data))

	case *protocol.Base_RssiReq:
		c.OutputChan <- messages.NewMessage(messages.CCAReq, address, []byte{})

	default:
		return fmt.Errorf("Received unhandled message type (%t)", m)
	}

	return nil
}

// handleOutgoing handles outgoing messages (ie. from ONS to nodes)
func (c *ZMQConnector) handleOutgoing(message *messages.Message) error {

	switch message.GetType() {
	case messages.CCAResp:
		// Build and write CCA response packet
		dataOut := make([]byte, 1)
		if message.GetCCA() {
			dataOut[0] = 1
		} else {
			dataOut[0] = 0
		}
		c.SendMsg(message.GetAddress(), onsMessageIDCCAResp, dataOut)

	default:

	}

	return nil
}

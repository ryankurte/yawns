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
	"strconv"
)

import (
	"github.com/ryankurte/ons/lib/messages"
)

const (
	// onsMessageIDRegister Registration message type, this message binds a socket ID to a client address
	onsMessageIDRegister int = 1
	// ONSMessagePacket Packet message type, used for transferring virtual network packets
	onsMessageIDPacket int = 2
	// onsMessageIDCCAReq CCA request type, requests CCA information from the ONS server
	onsMessageIDCCAReq int = 3
	// onsMessageIDCCAResp CCA response type, response with CCA info from the ONS server
	onsMessageIDCCAResp int = 4
	//onsMessageIDSetMode sets the radio mode for the medium
	onsMessageIDSetMode int = 5
)

// receive Parses an ONS message
func (c *ZMQConnector) handleClientReceive(data [][]byte) error {

	if len(data) != 3 {
		return fmt.Errorf("Error parsing message, required 3 parts")
	}

	// Fetch ZMQ client ID
	clientID := data[0]

	// Fetch message type
	messageType, err := strconv.Atoi(string(data[1]))
	if err != nil {
		return fmt.Errorf("Error parsing message type (%+v)", data[1])
	}

	// All ONS messages have data[2]
	body := data[2]

	// Handle message type
	switch messageType {
	case onsMessageIDRegister:
		// Bind address to ID lookup for sending
		address := string(body)
		_, ok := c.clients[address]
		if !ok {

			// Save to list
			c.clients[address] = clientID

			// Send connected event
			c.OutputChan <- messages.NewMessage(messages.Connected, address, []byte{})

		}

	case onsMessageIDPacket:
		address := c.findClientAddressByID(clientID)
		if address == "" {
			return fmt.Errorf("Received message for unknown clientID (%+v)", clientID)
		}

		// Write data packet message
		c.OutputChan <- messages.NewMessage(messages.Packet, address, body)

	case onsMessageIDCCAReq:
		address := c.findClientAddressByID(clientID)
		if address == "" {
			return fmt.Errorf("Received message for unknown clientID (%+v)", clientID)
		}

		// Write CCA request message
		c.OutputChan <- messages.NewMessage(messages.CCAReq, address, []byte{})

	default:
		return fmt.Errorf("Received unknown packet type (%d)", messageType)
	}

	return nil
}

func (c *ZMQConnector) handleMessageReceive(message *messages.Message) error {

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

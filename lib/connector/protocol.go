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

const (
	// ONSMessageRegister Registration message type, this message binds a socket ID to a client address
	ONSMessageRegister int = 1
	// ONSMessagePacket Packet message type, used for transferring virtual network packets
	ONSMessagePacket int = 2
	// ONSMessageCCAReq CCA request type, requests CCA information from the ONS server
	ONSMessageCCAReq int = 3
	// ONSMessageCCAResp CCA response type, response with CCA info from the ONS server
	ONSMessageCCAResp int = 4
	//ONSMessageSetMode sets the radio mode for the medium
	ONSMessageSetMode int = 5
)

// receive Parses an ONS message
func (c *ZMQConnector) receive(data [][]byte) error {

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
	case ONSMessageRegister:
		// Bind address to ID lookup for sending
		address := string(body)
		_, ok := c.clients[address]
		if !ok {
			// Call OnConnect handler if required
			c.handler.OnConnect(address)
			c.clients[address] = clientID
		}
		return nil

	case ONSMessagePacket:
		address := c.findClientAddressByID(clientID)
		if address == "" {
			return fmt.Errorf("Received message for unknown clientID (%+v)", clientID)
		}
		c.handler.Receive(address, body)
		return nil

	case ONSMessageCCAReq:
		address := c.findClientAddressByID(clientID)
		if address == "" {
			return fmt.Errorf("Received message for unknown clientID (%+v)", clientID)
		}
		cca := c.handler.GetCCA(address)
		dataOut := make([]byte, 1)
		if cca {
			dataOut[0] = 1
		} else {
			dataOut[0] = 0
		}
		c.SendMsg(address, ONSMessageCCAResp, dataOut)
		return nil

	default:
		return fmt.Errorf("Recieved unknown packet type (%d)", messageType)
	}

}

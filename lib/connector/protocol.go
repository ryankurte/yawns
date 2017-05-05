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
// This maps from Protobuf to ONS messages
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
			c.OutputChan <- &messages.Register{
				Message: messages.Message{Address: address},
			}
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

	// Receive a packet from a device
	case *protocol.Base_Packet:
		c.OutputChan <- &messages.Packet{
			Message: messages.Message{Address: address},
			Data:    m.Packet.Data,
		}

	// Signal that a device has entered receive mode
	case *protocol.Base_StartReceive:
		c.OutputChan <- &messages.StartReceive{
			Message: messages.Message{Address: address},
			RFInfo: messages.RFInfo{
				Band:    m.StartReceive.Info.Band,
				Channel: m.StartReceive.Info.Channel,
			},
		}

	case *protocol.Base_StopReceive:
		c.OutputChan <- &messages.StopReceive{
			Message: messages.Message{Address: address},
		}

	case *protocol.Base_Event:
		c.OutputChan <- &messages.Event{
			Message: messages.Message{Address: address},
			Data:    m.Event.Data,
		}

	case *protocol.Base_RssiReq:
		c.OutputChan <- &messages.RSSIRequest{
			Message: messages.Message{Address: address},
			RFInfo: messages.RFInfo{
				Band:    m.RssiReq.Info.Band,
				Channel: m.RssiReq.Info.Channel,
			},
		}

	default:
		return fmt.Errorf("Received unhandled message type (%t)", m)
	}

	return nil
}

// handleOutgoing handles outgoing messages (ie. from ONS to nodes)
// This maps from ONS messages to protobufs for external use
func (c *ZMQConnector) handleOutgoing(message interface{}) error {

	base := protocol.Base{}
	m, ok := message.(*messages.Message)
	if !ok {
		return fmt.Errorf("Connector error: message must be of base type messages.Message")
	}
	address := m.Address

	switch m := message.(type) {
	case messages.Packet:
		base.Message = &protocol.Base_Packet{
			Packet: &protocol.Packet{
				Data: m.Data,
			},
		}

	case messages.RSSIResponse:
		base.Message = &protocol.Base_RssiResp{
			RssiResp: &protocol.RSSIResp{
				Rssi: m.RSSI,
			},
		}
	case messages.SendComplete:
		base.Message = &protocol.Base_SendComplete{
			SendComplete: &protocol.SendComplete{},
		}

	default:
		return fmt.Errorf("Connector error: unsupported derived message type: %t", message)
	}

	data, err := proto.Marshal(&base)
	if err != nil {
		return err
	}

	c.sendMsg(address, data)

	return nil
}

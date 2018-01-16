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

	"github.com/golang/protobuf/proto"

	"github.com/ryankurte/owns/lib/messages"
	"github.com/ryankurte/owns/lib/protocol"
	"github.com/ryankurte/owns/lib/types"
)

// handleIncoming handles incoming messages from external sources (ie. from nodes to ONS)
// This maps from Protobuf to ONS messages
func (c *ZMQConnector) handleIncoming(data [][]byte) error {

	if len(data) != 2 {
		return fmt.Errorf("Error parsing message, required 2 parts")
	}

	// Fetch ZMQ client ID
	clientID := data[0]

	// Decode message
	message := protocol.Base{}
	if err := proto.Unmarshal(data[1], &message); err != nil {
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
			c.OutputChan <- messages.Register{
				BaseMessage: messages.BaseMessage{Address: address},
			}
		}

		return nil
	}

	// Perform ZMQ Client ID to address lookup
	address := c.findClientAddressByID(clientID)
	if address == "" {
		return fmt.Errorf("Received message for unknown clientID (%s)", clientID)
	}

	//log.Printf("Incoming From: %s Message: %+v", address, &message)

	// Handle messages
	switch m := message.GetMessage().(type) {
	case *protocol.Base_Deregister:

	// Receive a packet from a device
	case *protocol.Base_Packet:
		c.OutputChan <- messages.Packet{
			BaseMessage: messages.BaseMessage{Address: address},
			RFInfo: messages.RFInfo{
				Band:    m.Packet.Info.Band,
				Channel: m.Packet.Info.Channel},
			Data: m.Packet.Data,
		}

	// Signal that a device has entered receive mode
	case *protocol.Base_StateSet:
		state := types.TransceiverStateIdle
		switch m.StateSet.State {
		case protocol.RFState_IDLE:
			state = types.TransceiverStateIdle
		case protocol.RFState_RECEIVE:
			state = types.TransceiverStateReceive
		case protocol.RFState_RECEIVING:
			state = types.TransceiverStateReceiving
		case protocol.RFState_TRANSMITTING:
			state = types.TransceiverStateTransmitting
		case protocol.RFState_SLEEP:
			state = types.TransceiverStateSleep
		}

		c.OutputChan <- messages.StateSet{
			BaseMessage: messages.BaseMessage{Address: address},
			RFInfo: messages.RFInfo{
				Band:    m.StateSet.Info.Band,
				Channel: m.StateSet.Info.Channel,
			},
			State: state,
		}

	case *protocol.Base_Event:
		c.OutputChan <- messages.Event{
			BaseMessage: messages.BaseMessage{Address: address},
			Data:        m.Event.Data,
		}

	case *protocol.Base_RssiReq:
		c.OutputChan <- messages.RSSIRequest{
			BaseMessage: messages.BaseMessage{Address: address},
			RFInfo: messages.RFInfo{
				Band:    m.RssiReq.Info.Band,
				Channel: m.RssiReq.Info.Channel,
			},
		}

	case *protocol.Base_StateReq:
		c.OutputChan <- messages.StateRequest{
			BaseMessage: messages.BaseMessage{Address: address},
			RFInfo: messages.RFInfo{
				Band: m.StateReq.Info.Band,
			},
		}

	case *protocol.Base_FieldSet:
		c.OutputChan <- messages.FieldSet{
			BaseMessage: messages.BaseMessage{Address: address},
			Name:        m.FieldSet.Name,
			Data:        m.FieldSet.Data,
		}

	case *protocol.Base_FieldReq:
		c.OutputChan <- messages.FieldGet{
			BaseMessage: messages.BaseMessage{Address: address},
			Name:        m.FieldReq.Name,
		}

	default:
		return fmt.Errorf("[WARNING] Connector.handleIncoming: unhandled message type (%t)", m)
	}

	return nil
}

// handleOutgoing handles outgoing messages (ie. from ONS to nodes)
// This maps from ONS messages to protobufs for external use
func (c *ZMQConnector) handleOutgoing(message interface{}) error {

	base := protocol.Base{}
	address := ""

	switch m := message.(type) {
	case messages.Packet:
		address = m.Address
		base.Message = &protocol.Base_Packet{
			Packet: &protocol.Packet{
				Info: &protocol.RFInfo{Band: m.Band},
				Data: m.Data,
			},
		}

	case messages.StateResponse:
		address = m.Address
		state := protocol.RFState_IDLE
		switch m.State {
		case types.TransceiverStateIdle:
			state = protocol.RFState_IDLE
		case types.TransceiverStateReceive:
			state = protocol.RFState_RECEIVE
		case types.TransceiverStateReceiving:
			state = protocol.RFState_RECEIVING
		case types.TransceiverStateTransmitting:
			state = protocol.RFState_TRANSMITTING
		case types.TransceiverStateSleep:
			state = protocol.RFState_SLEEP
		}

		base.Message = &protocol.Base_StateResp{
			StateResp: &protocol.StateResp{
				Info:  &protocol.RFInfo{Band: m.Band},
				State: state,
			},
		}

	case messages.RSSIResponse:
		address = m.Address
		base.Message = &protocol.Base_RssiResp{
			RssiResp: &protocol.RSSIResp{
				Info: &protocol.RFInfo{Band: m.Band, Channel: m.Channel},
				Rssi: m.RSSI,
			},
		}
	case messages.SendComplete:
		address = m.Address
		base.Message = &protocol.Base_SendComplete{
			SendComplete: &protocol.SendComplete{
				Info: &protocol.RFInfo{Band: m.Band},
			},
		}

	case messages.FieldResp:
		address = m.Address
		base.Message = &protocol.Base_FieldResp{
			FieldResp: &protocol.FieldResp{
				Name: m.Name,
				Data: m.Data,
			},
		}

	default:
		return fmt.Errorf("[WARNING] Connector.handleOutgoing: unsupported message type (%T)", message)
	}

	//log.Printf("Outgoing Message: %+v", message)

	data, err := proto.Marshal(&base)
	if err != nil {
		return err
	}

	c.sendMsg(address, data)

	return nil
}

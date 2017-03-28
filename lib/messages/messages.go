/**
 * OpenNetworkSim Message Definitions
 * These messages are used for communication between components of the simulator
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package messages

// Type defines the type of message being passed
type Type string

const (
	// Connected message indicates a node has connected
	Connected Type = "connected"
	// Packet Packet message type
	Packet Type = "packet"
	// CCAReq CCA request type
	CCAReq Type = "cca-request"
	// CCAResp CCA response type
	CCAResp Type = "cca-response"
	//SetMode sets the radio mode for the medium
	SetMode Type = "set-mode"
)

// Message type used for communication with connector module
type Message struct {
	messageType Type
	address     string
	data        []byte
	cca         bool
}

// NewMessage creates a message
func NewMessage(messageType Type, address string, data []byte) *Message {
	return &Message{
		messageType: messageType,
		address:     address,
		data:        data,
	}
}

// GetType fetches the type of the message
func (message *Message) GetType() Type { return message.messageType }


// GetAddress fetches the address of the origin/destination of the message
func (message *Message) GetAddress() string { return message.address }

// GetData fetches message data
func (message *Message) GetData() []byte { return message.data }

// MessageCCA is a CCA message
type MessageCCA struct {
	*Message
	cca bool
}

// GetCCA Fetch Clear Channel Acknowledgement from a CCA message
func (message *Message) GetCCA() bool { return message.cca }

// SetCCA Set Clear Channel Acknowledgement for a CCA message
func (message *Message) SetCCA(cca bool) { message.cca = cca }

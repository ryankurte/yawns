/**
 * OpenNetworkSim Message Definitions
 * These messages are used for communication between components of the simulator
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package messages

import ()

// Type defines the type of message being passed
type Type string

const (
	// ConnectedID message indicates a node has connected
	ConnectedID Type = "connected"
	// Packet Packet message type
	PacketID Type = "packet"
	// PacketSentID Packet sent message type
	PacketSentID Type = "packet-sent"
	// CCAReqID CCA request type
	CCAReqID Type = "cca-request"
	// CCARespID CCA response type
	CCARespID Type = "cca-response"
	// SetModeID sets the radio mode for the medium
	SetModeID Type = "set-mode"
	// EventID is a node event message (used for central logging)
	EventID Type = "event"
)

// Message base type used for internal communication
type Message struct {
	Address string
}

type RFInfo struct {
	Band    string
	Channel int32
	RSSI    float64
}

func NewRFInfo(band string, channel int32) RFInfo {
	return RFInfo{band, channel, 0.0}
}

type Register struct {
	Message
}

type Deregister struct {
	Message
}

type Packet struct {
	Message
	RFInfo
	Data []byte
}

func NewPacket(address string, data []byte, rfInfo RFInfo) *Packet {
	return &Packet{
		Message: Message{
			Address: address,
		},
		RFInfo: rfInfo,
		Data:   data,
	}
}

type RSSIRequest struct {
	Message
	RFInfo
}

type RSSIResponse struct {
	Message
	RFInfo
	RSSI float32
}

type SendComplete struct {
	Message
	RFInfo
}

func NewSendComplete(address, bandName string, channel int32) *SendComplete {
	return &SendComplete{
		Message: Message{
			Address: address,
		},
		RFInfo: NewRFInfo(bandName, channel),
	}
}

type StartReceive struct {
	Message
	RFInfo
}

type StopReceive struct {
	Message
	RFInfo
}

type Event struct {
	Message
	Data string
}

// GetAddress fetches the address of the origin/destination of the message
func (message *Message) GetAddress() string { return message.Address }

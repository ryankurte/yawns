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

// Message Generic message interface type to provide slightly more clarity than interface{}
type Message interface {
	GetAddress() string
}

// BaseMessage base type used for internal communication
type BaseMessage struct {
	Address string
}

// GetAddress fetches the address of the origin/destination of the message
func (message *BaseMessage) GetAddress() string { return message.Address }

type RFInfo struct {
	Band    string
	Channel int32
	RSSI    float64
}

func NewRFInfo(band string, channel int32) RFInfo {
	return RFInfo{band, channel, 0.0}
}

type Register struct {
	BaseMessage
}

type Deregister struct {
	BaseMessage
}

type Packet struct {
	BaseMessage
	RFInfo
	Data []byte
}

func NewPacket(address string, data []byte, rfInfo RFInfo) *Packet {
	return &Packet{
		BaseMessage: BaseMessage{
			Address: address,
		},
		RFInfo: rfInfo,
		Data:   data,
	}
}

type RSSIRequest struct {
	BaseMessage
	RFInfo
}

type RSSIResponse struct {
	BaseMessage
	RFInfo
	RSSI float32
}

type SendComplete struct {
	BaseMessage
	RFInfo
}

func NewSendComplete(address, bandName string, channel int32) *SendComplete {
	return &SendComplete{
		BaseMessage: BaseMessage{
			Address: address,
		},
		RFInfo: NewRFInfo(bandName, channel),
	}
}

type StartReceive struct {
	BaseMessage
	RFInfo
}

type StopReceive struct {
	BaseMessage
	RFInfo
}

type Event struct {
	BaseMessage
	Data string
}

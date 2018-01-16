/**
 * OpenNetworkSim Message Definitions
 * These messages are used for communication between components of the simulator
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package messages

import (
	"github.com/ryankurte/owns/lib/types"
)

// Type defines the type of message being passed
type Type string

const (
	// ConnectedID message indicates a node has connected
	ConnectedID Type = "connected"
	// PacketID Packet message type
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

func NewBaseMessage(address string) BaseMessage {
	return BaseMessage{Address: address}
}

// GetAddress fetches the address of the origin/destination of the message
func (message *BaseMessage) GetAddress() string { return message.Address }

// RFInfo structure encodes RF packet information
type RFInfo struct {
	Band    string
	Channel int32
	RSSI    float64
}

func NewRFInfo(band string, channel int32) RFInfo {
	return RFInfo{band, channel, 0.0}
}

// Register message sent when a device registers with the simulator
type Register struct {
	BaseMessage
}

// Deregister message sent when a device is deregistered from the simulator
type Deregister struct {
	BaseMessage
}

// Packet encodes an RF packet to be sent or received
type Packet struct {
	BaseMessage
	RFInfo
	Data []byte
}

func NewPacket(address string, data []byte, rfInfo RFInfo) Packet {
	return Packet{
		BaseMessage: BaseMessage{
			Address: address,
		},
		RFInfo: rfInfo,
		Data:   data,
	}
}

// RSSIRequest is a message from a node requesting RSSI data for a given band and channel
type RSSIRequest struct {
	BaseMessage
	RFInfo
}

// RSSIResponse is a message from the simulator to a node containing RSSI data for a given band and channel
type RSSIResponse struct {
	BaseMessage
	RFInfo
	RSSI float32
}

type StateSet struct {
	BaseMessage
	RFInfo
	State types.TransceiverState
}

type StateRequest struct {
	BaseMessage
	RFInfo
}

type StateResponse struct {
	BaseMessage
	RFInfo
	State types.TransceiverState
}

type SendComplete struct {
	BaseMessage
	RFInfo
}

func NewSendComplete(address, bandName string, channel int32) SendComplete {
	return SendComplete{
		BaseMessage: BaseMessage{
			Address: address,
		},
		RFInfo: NewRFInfo(bandName, channel),
	}
}

type FieldSet struct {
	BaseMessage
	Name string
	Data string
}

type FieldGet struct {
	BaseMessage
	Name string
}

type FieldResp struct {
	BaseMessage
	Name string
	Data string
}

type Event struct {
	BaseMessage
	Address string
	Type    string
	Data    string
}

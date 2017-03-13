/**
 * OpenNetworkSim CZMQ Radio Driver Test Wrapper
 * This wraps the C library to allow go unit testing against the go connector
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package lib

/*
#include <stdint.h>
#include "ons/ons.h"
#cgo LDFLAGS: ./build/libons.a -lzmq -lczmq -lpthread
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// ONSConnector is an instance of the ONS connector
type ONSConnector struct {
	ons C.struct_ons_s
}

// Init the ONS connector
func (c *ONSConnector) Init(serverAddress string, localAddress string) {
	C.ONS_init(&c.ons, C.CString(serverAddress), C.CString(localAddress))
}

// Send a data packet using the connector
func (c *ONSConnector) Send(data []byte) {
	p := *(*C.uint8_t)(unsafe.Pointer(&data[0]))

	C.ONS_send(&c.ons, &p, C.uint16_t(len(data)))
}

// CheckReceive Check for data packet send completion
func (c *ONSConnector) CheckSend() bool {
	return false
}

// CheckReceive Check whether a packet has been received
func (c *ONSConnector) CheckReceive() bool {
	return false
}

// GetReceived Fetch a received packet
func (c *ONSConnector) GetReceived() ([]byte, error) {
	// Create C objects for calling
	data := make([]C.uint8_t, 256)
	maxLen := C.uint16_t(len(data))
	dataPtr := *(*C.uint8_t)(unsafe.Pointer(&data[0]))
	len := C.uint16_t(0)

	// Call C method
	res := C.ONS_get_received(&c.ons, maxLen, &dataPtr, &len)

	// Check response
	if res <= 0 {
		return []byte{}, fmt.Errorf("ONS_get_received error %d", res)
	}

	// Convert to go data
	safeData := make([]byte, len)
	for i := range data {
		safeData[i] = byte(data[i])
	}

	return safeData, nil
}

// Close the ONS connector
func (c *ONSConnector) Close() {
	C.ONS_close(&c.ons)
}

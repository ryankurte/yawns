/**
 * OpenNetworkSim CZMQ Radio Driver Test Wrapper
 * This wraps the C library to allow go unit testing against the go connector
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package libons

/*
#include <stdint.h>
#include "ons/ons.h"
#cgo LDFLAGS: -L./build/ -lons -lzmq -lczmq -lpthread
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

// NewONSConnector creates an ONS connector
func NewONSConnector() *ONSConnector {
	return &ONSConnector{C.struct_ons_s{}}
}

// Init the ONS connector
func (c *ONSConnector) Init(serverAddress string, localAddress string) {
	sa := C.CString(serverAddress)
	la := C.CString(localAddress)

	C.ONS_init(&c.ons, sa, la)

	C.free(unsafe.Pointer(sa))
	C.free(unsafe.Pointer(la))
}

// Send a data packet using the connector
func (c *ONSConnector) Send(data []byte) {

	typedData := make([]C.uint8_t, len(data))
	ptr := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	length := C.uint16_t(len(data))

	for i := range data {
		typedData[i] = C.uint8_t(data[i])
	}

	C.ONS_send(&c.ons, ptr, length)
}

// CheckSend Check for data packet send completion
func (c *ONSConnector) CheckSend() bool {
	res := C.ONS_check_send(&c.ons)
	if res > 0 {
		return true
	}
	return false
}

// CheckReceive Check whether a packet has been received
func (c *ONSConnector) CheckReceive() bool {
	res := C.ONS_check_receive(&c.ons)
	if res > 0 {
		return true
	}
	return false
}

// GetReceived Fetch a received packet
func (c *ONSConnector) GetReceived() ([]byte, error) {
	// Create C objects for calling
	data := make([]C.uint8_t, C.ONS_BUFFER_LENGTH)
	dataPtr := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	maxLen := C.uint16_t(len(data))
	len := C.uint16_t(0)

	// Call C method
	res := C.ONS_get_received(&c.ons, maxLen, dataPtr, &len)

	// Check response
	if res <= 0 {
		return []byte{}, fmt.Errorf("ONS_get_received error %d", res)
	}

	// Convert to go data
	safeData := make([]byte, len)
	for i := range safeData {
		safeData[i] = byte(data[i])
	}

	return safeData, nil
}

// Close the ONS connector
func (c *ONSConnector) Close() {
	C.ONS_close(&c.ons)
}

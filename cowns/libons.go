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
#include "owns/owns.h"
#cgo LDFLAGS: -L./build -L./cowns/build -lowns -lzmq -lczmq -lpthread -lprotobuf-c
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

type ONSRadio struct {
	radio C.struct_ons_radio_s
}

// NewONSConnector creates an ONS connector
func NewONSConnector() *ONSConnector {
	return &ONSConnector{C.struct_ons_s{}}
}

// Init the ONS connector
func (c *ONSConnector) Init(serverAddress string, localAddress string) error {
	sa := C.CString(serverAddress)
	la := C.CString(localAddress)
	conf := C.struct_ons_config_s{
		intercept_signals: true,
	}

	res := C.ONS_init(&c.ons, sa, la, &conf)

	C.free(unsafe.Pointer(sa))
	C.free(unsafe.Pointer(la))

	if res != 0 {
		return fmt.Errorf("Error initialising ONSC")
	}
	return nil
}

// RadioInit Initialise a virtual radio using the connector
func (c *ONSConnector) RadioInit(band string) (*ONSRadio, error) {
	r := ONSRadio{
		radio: C.struct_ons_radio_s{},
	}
	bandString := C.CString(band)

	res := C.ONS_radio_init(&c.ons, &r.radio, bandString)
	C.free(unsafe.Pointer(bandString))
	if res != 0 {
		return nil, fmt.Errorf("Error creating virtual radio for band: %s", band)
	}
	return &r, nil
}

// CloseRadio Closes an ONS virtual radio
func (c *ONSConnector) CloseRadio(r *ONSRadio) {
	C.ONS_radio_close(&c.ons, &r.radio)
}

// Close the ONS connector
func (c *ONSConnector) Close() {
	C.ONS_close(&c.ons)
}

// Send a data packet using the connector
func (r *ONSRadio) Send(channel int, data []byte) {
	typedData := make([]C.uint8_t, len(data))
	ptr := (*C.uint8_t)(unsafe.Pointer(&typedData[0]))
	length := C.uint16_t(len(data))
	c := C.int32_t(channel)

	for i := range data {
		typedData[i] = C.uint8_t(data[i])
	}

	C.ONS_radio_send(&r.radio, c, ptr, length)
}

// CheckSend Check for data packet send completion
func (r *ONSRadio) CheckSend() bool {
	res := C.ONS_radio_check_send(&r.radio)
	if res > 0 {
		return true
	}
	return false
}

// StartReceive Puts the virtual radio into receive mode
func (r *ONSRadio) StartReceive(channel int) bool {
	c := C.int32_t(channel)
	res := C.ONS_radio_start_receive(&r.radio, c)
	if res > 0 {
		return true
	}
	return false
}

// StopReceive putsthe virtual radio in idle mode
func (r *ONSRadio) StopReceive() bool {
	res := C.ONS_radio_stop_receive(&r.radio)
	if res > 0 {
		return true
	}
	return false
}

// CheckReceive Check whether a packet has been received
func (r *ONSRadio) CheckReceive() bool {
	res := C.ONS_radio_check_receive(&r.radio)
	if res > 0 {
		return true
	}
	return false
}

// GetReceived Fetch a received packet
func (r *ONSRadio) GetReceived() ([]byte, error) {
	// Create C objects for calling
	data := make([]C.uint8_t, C.ONS_BUFFER_LENGTH)
	dataPtr := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	maxLen := C.uint16_t(len(data))
	len := C.uint16_t(0)

	// Call C method
	res := C.ONS_radio_get_received(&r.radio, maxLen, dataPtr, &len)

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

// GetRSSI Check fetches RSSI for the device
func (r *ONSRadio) GetRSSI(channel int) (float32, error) {
	rssi := C.float(0.0)
	rssiPtr := (*C.float)(unsafe.Pointer(&rssi))
	c := C.int32_t(channel)

	res := C.ONS_radio_get_rssi(&r.radio, c, rssiPtr)
	if res < 0 {
		return float32(rssi), fmt.Errorf("GetRSSI error %d", res)
	}
	if res == 0 {
		return float32(rssi), nil
	}
	return float32(rssi), nil
}

// GetState Check fetches state for the device
func (r *ONSRadio) GetState() (uint32, error) {
	state := C.uint32_t(0)
	statePtr := (*C.uint32_t)(unsafe.Pointer(&state))

	res := C.ONS_radio_get_state(&r.radio, statePtr)
	if res < 0 {
		return uint32(state), fmt.Errorf("GetState error %d", res)
	}
	if res == 0 {
		return uint32(state), nil
	}
	return uint32(state), nil
}

// SetField Sets a field in the simulation
func (conn *ONSConnector) SetField(name string, data []byte) error {
	n := C.CString(name)
	typedData := make([]C.uint8_t, len(data))
	ptr := (*C.uint8_t)(unsafe.Pointer(&typedData[0]))
	length := C.size_t(len(data))

	for i := range data {
		typedData[i] = C.uint8_t(data[i])
	}

	res := C.ONS_set_field(&conn.ons, n, ptr, length)
	if res < 0 {
		return fmt.Errorf("SetField error %d", res)
	}
	return nil
}

/**
 * OpenNetworkSim Plugin Manager
 * Plugins should implement one or more of the interfaces defined in this module
 * The PluginManager detects these on binding and calls the associated methods as appropriate during operation
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package plugins

import (
	"fmt"
)

// ConnectHandler interface should be implemented by plugins that need to detect
// when a node is connected
type ConnectHandler interface {
	Connected(address string) error
}

// ReceiveHandler interface should be implemented by plugins to receive packets sent by any node
type ReceiveHandler interface {
	Received(address string, message []byte) error
}

// SendHandler interface should be implemented by plugins to receive packets distributed by the simulator
// Note that packets will be repeated based on connectivity
type SendHandler interface {
	Send(address string, message []byte) error
}

// EventHandler interface should be implemented by plugins to handle simulation Events
type EventHandler interface {
	Event(eventType, address string, data map[string]string)
}

// PluginManager manages plugins and calls underlying handlers when required
type PluginManager struct {
	connectHandlers []ConnectHandler
	receiveHandlers []ReceiveHandler
	sendHandlers    []SendHandler
	eventHandlers   []EventHandler
}

// NewPluginManager creates an empty plugin manager instance
func NewPluginManager() *PluginManager {
	return &PluginManager{}
}

// BindPlugin Bind a plugin to the plugin manager
func (pm *PluginManager) BindPlugin(plugin interface{}) error {
	bound := 0

	connected, ok := plugin.(ConnectHandler)
	if ok {
		pm.connectHandlers = append(pm.connectHandlers, connected)
		bound++
	}

	receive, ok := plugin.(ReceiveHandler)
	if ok {
		pm.receiveHandlers = append(pm.receiveHandlers, receive)
		bound++
	}

	send, ok := plugin.(SendHandler)
	if ok {
		pm.sendHandlers = append(pm.sendHandlers, send)
		bound++
	}

	event, ok := plugin.(EventHandler)
	if ok {
		pm.eventHandlers = append(pm.eventHandlers, event)
		bound++
	}

	if bound == 0 {
		return fmt.Errorf("PluginManager.BindPlugin error: interface (%+v) does not implement any plugin methods", plugin)
	}
	return nil
}

// OnConnected calls bound plugin ConnectHandlers
func (pm *PluginManager) OnConnected(address string) {
	for _, h := range pm.connectHandlers {
		h.Connected(address)
	}
}

// OnReceived calls bound plugin ReceiveHandlers
func (pm *PluginManager) OnReceived(address string, data []byte) {
	for _, h := range pm.receiveHandlers {
		h.Received(address, data)
	}
}

// OnSend calls bound plugin SendHandlers
func (pm *PluginManager) OnSend(address string, data []byte) {
	for _, h := range pm.sendHandlers {
		h.Send(address, data)
	}
}

// OnEvent calls bound plugin EventHandlers
func (pm *PluginManager) OnEvent(eventType, address string, data map[string]string) {
	for _, h := range pm.eventHandlers {
		h.Event(eventType, address, data)
	}
}

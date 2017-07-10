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
	"log"
	"time"
)

// ConnectHandler interface should be implemented by plugins that need to detect
// when a node is connected
type ConnectHandler interface {
	Connected(d time.Duration, address string) error
}

// ReceiveHandler interface should be implemented by plugins to receive packets sent by any node
type ReceiveHandler interface {
	Received(d time.Duration, address string, message []byte) error
}

// SendHandler interface should be implemented by plugins to receive packets distributed by the simulator
// Note that packets will be repeated based on connectivity
type SendHandler interface {
	Send(d time.Duration, address string, message []byte) error
}

// EventHandler interface should be implemented by plugins to handle simulation Events
type EventHandler interface {
	Event(d time.Duration, eventType, address string, data map[string]string)
}

// CloseHandler interface should be implemented by plugins to handle plugin closing at simulation exit
type CloseHandler interface {
	Close()
}

// PluginManager manages plugins and calls underlying handlers when required
type PluginManager struct {
	connectHandlers []ConnectHandler
	receiveHandlers []ReceiveHandler
	sendHandlers    []SendHandler
	eventHandlers   []EventHandler
	closeHandlers   []CloseHandler
}

// NewPluginManager creates an empty plugin manager instance
func NewPluginManager() *PluginManager {
	return &PluginManager{}
}

// BindPlugin Bind a plugin to the plugin manager
func (pm *PluginManager) BindPlugin(plugin interface{}) error {
	bound := 0

	if connected, ok := plugin.(ConnectHandler); ok {
		log.Printf("PluginManager.BindPlugin: Bound connectHandler")
		pm.connectHandlers = append(pm.connectHandlers, connected)
		bound++
	}

	if receive, ok := plugin.(ReceiveHandler); ok {
		log.Printf("PluginManager.BindPlugin: Bound receiveHandler")
		pm.receiveHandlers = append(pm.receiveHandlers, receive)
		bound++
	}

	if send, ok := plugin.(SendHandler); ok {
		log.Printf("PluginManager.BindPlugin: Bound sendHandler")
		pm.sendHandlers = append(pm.sendHandlers, send)
		bound++
	}

	if event, ok := plugin.(EventHandler); ok {
		log.Printf("PluginManager.BindPlugin: Bound eventHandler")
		pm.eventHandlers = append(pm.eventHandlers, event)
		bound++
	}

	if close, ok := plugin.(CloseHandler); ok {
		log.Printf("PluginManager.BindPlugin: Bound closeHandler")
		pm.closeHandlers = append(pm.closeHandlers, close)
	}

	if bound == 0 {
		return fmt.Errorf("PluginManager.BindPlugin error: interface (%+v) does not implement any plugin methods", plugin)
	}
	return nil
}

// OnConnected calls bound plugin ConnectHandlers
func (pm *PluginManager) OnConnected(d time.Duration, address string) {
	log.Printf("PluginManager.OnConnected called")
	for _, h := range pm.connectHandlers {
		h.Connected(d, address)
	}
}

// OnReceived calls bound plugin ReceiveHandlers
func (pm *PluginManager) OnReceived(d time.Duration, address string, data []byte) {
	log.Printf("PluginManager.OnReceived called")
	for _, h := range pm.receiveHandlers {
		h.Received(d, address, data)
	}
}

// OnSend calls bound plugin SendHandlers
func (pm *PluginManager) OnSend(d time.Duration, address string, data []byte) {
	log.Printf("PluginManager.OnSend called")
	for _, h := range pm.sendHandlers {
		h.Send(d, address, data)
	}
}

// OnEvent calls bound plugin EventHandlers
func (pm *PluginManager) OnEvent(d time.Duration, eventType, address string, data map[string]string) {
	log.Printf("PluginManager.OnEvent called")
	for _, h := range pm.eventHandlers {
		h.Event(d, eventType, address, data)
	}
}

// OnClose calls bound plugin CloseHandlers
func (pm *PluginManager) OnClose() {
	log.Printf("PluginManager.OnClose called")
	for _, h := range pm.closeHandlers {
		h.Close()
	}
}

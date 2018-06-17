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
	"time"

	"github.com/ryankurte/yawns/lib/config"
)

// ConnectHandler interface should be implemented by plugins that need to detect
// when a node is connected
type ConnectHandler interface {
	Connected(d time.Duration, address string) error
}

// DisconnectHandler interface should be implemented by plugins that need to detect
// when a node is connected
type DisconnectHandler interface {
	Disconnected(d time.Duration, address string) error
}

// ReceiveHandler interface should be implemented by plugins to receive packets sent by any node
type ReceiveHandler interface {
	Received(d time.Duration, band, address string, message []byte) error
}

// SendHandler interface should be implemented by plugins to receive packets distributed by the simulator
// Note that packets will be repeated based on connectivity
type SendHandler interface {
	Send(d time.Duration, band, address string, message []byte) error
}

// EventHandler interface should be implemented by plugins to receive events from nodes
type EventHandler interface {
	OnEvent(d time.Duration, address string, data string) error
}

// MessageHandler interface should be implemented by plugins to receive raw simulation messages
type MessageHandler interface {
	OnMessage(d time.Duration, message interface{}) error
}

// UpdateHandler interface should be implemented by plugins to handle simulation Events
type UpdateHandler interface {
	OnUpdate(d time.Duration, eventType config.UpdateAction, address string, data map[string]string) error
}

// CloseHandler interface should be implemented by plugins to handle plugin closing at simulation exit
type CloseHandler interface {
	Close()
}

// PluginManager manages plugins and calls underlying handlers when required
type PluginManager struct {
	connectHandlers    []ConnectHandler
	disconnectHandlers []DisconnectHandler
	receiveHandlers    []ReceiveHandler
	sendHandlers       []SendHandler
	eventHandlers      []EventHandler
	messageHandlers    []MessageHandler
	updateHandlers     []UpdateHandler
	closeHandlers      []CloseHandler
}

// NewPluginManager creates an empty plugin manager instance
func NewPluginManager() *PluginManager {
	return &PluginManager{}
}

// BindPlugin Bind a plugin to the plugin manager
func (pm *PluginManager) BindPlugin(plugin interface{}) error {
	bound := 0

	if connected, ok := plugin.(ConnectHandler); ok {
		pm.connectHandlers = append(pm.connectHandlers, connected)
		bound++
	}

	if disconnected, ok := plugin.(DisconnectHandler); ok {
		pm.disconnectHandlers = append(pm.disconnectHandlers, disconnected)
		bound++
	}

	if receive, ok := plugin.(ReceiveHandler); ok {
		pm.receiveHandlers = append(pm.receiveHandlers, receive)
		bound++
	}

	if send, ok := plugin.(SendHandler); ok {
		pm.sendHandlers = append(pm.sendHandlers, send)
		bound++
	}

	if event, ok := plugin.(EventHandler); ok {
		pm.eventHandlers = append(pm.eventHandlers, event)
		bound++
	}

	if message, ok := plugin.(MessageHandler); ok {
		pm.messageHandlers = append(pm.messageHandlers, message)
		bound++
	}

	if update, ok := plugin.(UpdateHandler); ok {
		pm.updateHandlers = append(pm.updateHandlers, update)
		bound++
	}

	if close, ok := plugin.(CloseHandler); ok {
		pm.closeHandlers = append(pm.closeHandlers, close)
	}

	if bound == 0 {
		return fmt.Errorf("PluginManager.BindPlugin error: interface (%+v) does not implement any plugin methods", plugin)
	}
	return nil
}

// OnConnected calls bound plugin ConnectHandlers
func (pm *PluginManager) OnConnected(d time.Duration, address string) {
	for _, h := range pm.connectHandlers {
		h.Connected(d, address)
	}
}

// OnReceived calls bound plugin ReceiveHandlers
func (pm *PluginManager) OnReceived(d time.Duration, band, address string, data []byte) {
	for _, h := range pm.receiveHandlers {
		h.Received(d, band, address, data)
	}
}

// OnSend calls bound plugin SendHandlers
func (pm *PluginManager) OnSend(d time.Duration, band, address string, data []byte) {
	for _, h := range pm.sendHandlers {
		h.Send(d, band, address, data)
	}
}

// OnEvent calls bound plugin EventHandlers
func (pm *PluginManager) OnEvent(d time.Duration, address string, data string) {
	for _, h := range pm.eventHandlers {
		h.OnEvent(d, address, data)
	}
}

// OnMessage calls bound plugin MessageHandlers
func (pm *PluginManager) OnMessage(d time.Duration, message interface{}) {
	for _, h := range pm.messageHandlers {
		h.OnMessage(d, message)
	}
}

// onUpdate calls bound plugin UpdateHandlers
func (pm *PluginManager) OnUpdate(d time.Duration, eventType config.UpdateAction, address string, data map[string]string) {
	for _, h := range pm.updateHandlers {
		h.OnUpdate(d, eventType, address, data)
	}
}

// OnClose calls bound plugin CloseHandlers
func (pm *PluginManager) OnClose() {
	for _, h := range pm.closeHandlers {
		h.Close()
	}
}

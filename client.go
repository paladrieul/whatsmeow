// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/// Package whatsmeow implements aatsApp web client.
package whatsmeow

import (
	.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// EventHandler is a function that can handle events from WhatsApp.
type EventHandler func(evt interface{})

// Client is the main WhatsApp web client struct.
type Client struct {
	Store   *store.Device
	Log     waLog.Logger
	RecvLog waLog.Logger
	SendLog waLog.Logger

	// AutoTrustIdentity specifies whether to automatically trust new identities.
	AutoTrustIdentity bool

	// EmitAppStateEventsOnFullSync specifies whether to emit app state events on full sync.
	EmitAppStateEventsOnFullSync bool

	// GetMessageForRetry is used to fetch the original message when a retry receipt is received.
	GetMessageForRetry func(requester, to types.JID, id types.MessageID) *waProto.Message

	// PrePairCallback is called before pairing is completed.
	PrePairCallback func(jid types.JID, platform, businessName string) bool

	eventHandlersLock sync.RWMutex
	eventHandlers     []wrappedEventHandler
	nextHandlerID     uint32

	uniqueID  string
	identityID []byte

	connectionState atomic.Int32
	lastConnect     time.Time

	ctx       context.Context
	cancel    context.CancelFunc
	ctxCancel context.CancelCauseFunc
}

type wrappedEventHandler struct {
	fn EventHandler
	id uint32
}

// NewClient creates a new WhatsApp web client with the given device store and logger.
func NewClient(deviceStore *store.Device, log waLog.Logger) *Client {
	if log == nil {
		log = waLog.Noop
	}
	return &Client{
		Store:             deviceStore,
		Log:               log,
		RecvLog:           log.Sub("Recv"),
		SendLog:           log.Sub("Send"),
		AutoTrustIdentity: false,
	}
}

// AddEventHandler adds a new event handler and returns its ID.
// The ID can be used to remove the handler later with RemoveEventHandler.
func (cli *Client) AddEventHandler(handler EventHandler) uint32 {
	id := atomic.AddUint32(&cli.nextHandlerID, 1)
	cli.eventHandlersLock.Lock()
	cli.eventHandlers = append(cli.eventHandlers, wrappedEventHandler{fn: handler, id: id})
	cli.eventHandlersLock.Unlock()
	return id
}

// RemoveEventHandler removes the event handler with the given ID.
// Returns true if the handler was found and removed.
func (cli *Client) RemoveEventHandler(id uint32) bool {
	cli.eventHandlersLock.Lock()
	defer cli.eventHandlersLock.Unlock()
	for i, handler := range cli.eventHandlers {
		if handler.id == id {
			cli.eventHandlers = append(cli.eventHandlers[:i], cli.eventHandlers[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveAllEventHandlers removes all registered event handlers.
func (cli *Client) RemoveAllEventHandlers() {
	cli.eventHandlersLock.Lock()
	cli.eventHandlers = nil
	cli.eventHandlersLock.Unlock()
}

// dispatchEvent sends an event to all registered event handlers.
func (cli *Client) dispatchEvent(evt interface{}) {
	cli.eventHandlersLock.RLock()
	handlers := cli.eventHandlers
	cli.eventHandlersLock.RUnlock()
	for _, handler := range handlers {
		handle(handler.fn, evt)
	}
}

func handle(fn EventHandler, evt interface{}) {
	defer func() {
		if err := recover(); err != nil {
			_ = err // TODO: log panic
		}
	}()
	fn(evt)
}

// IsLoggedIn returns true if the client is currently logged in.
func (cli *Client) IsLoggedIn() bool {
	return cli.Store != nil && cli.Store.ID != nil
}

// IsConnected returns true if the client is currently connected to WhatsApp servers.
func (cli *Client) IsConnected() bool {
	return cli.connectionState.Load() == int32(events.ConnectStateConnected)
}

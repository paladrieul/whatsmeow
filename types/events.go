// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package types contains various types used by whatsmeow.
package types

import (
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

// JID represents a WhatsApp JID (ber ID).
type JID struct {
	User   string
	Agent  uint8
	Device uint8
	Server string
	AD     bool
}

// String returns the string representation of the JID.
func (j JID) String() string {
	if j.AD {
		return j.User + "." + string(rune('0'+j.Agent))":" + string(rune('0'+j.Device)) + "@" + j.Server
	}
	if len(j.User) > 0 {
		return j.User + "@" + j.Server
	}
	return j.Server
}

// MessageInfo contains metadata about a received{
	// The JID of the chat where the message was sent.
	Chat JID
	// The JID of the sender.
	Sender JID
	// Whether the message was sent by the current user.
	IsFromMe bool
	// Whether the message was sent in a group.
	IsGroup bool
	// The unique message ID.
	ID string
	// The server timestamp of the message.
	Timestamp time.Time
	// Push name (display name) of the sender.
	PushName string
}

// Message is the event dispatched when a new message is received.
type Message struct {
	Info    MessageInfo
	Message *waProto.Message
	// RawMessage contains the raw decrypted message bytes before proto unmarshalling.
	RawMessage []byte
}

// Receipt is the event dispatched when a message receipt is received.
type Receipt struct {
	// The chat JID where the receipt was sent.
	Chat JID
	// The sender JID of the receipt.
	Sender JID
	// The message IDs that were receipted.
	MessageIDs []string
	// The type of receipt (read, delivered, etc.).
	Type ReceiptType
	// The timestamp of the receipt.
	Timestamp time.Time
}

// ReceiptType represents the type of a message receipt.
type ReceiptType string

const (
	// ReceiptTypeDelivered means the message was delivered to the device.
	ReceiptTypeDelivered ReceiptType = "delivered"
	// ReceiptTypeRead means the message was read by the user.
	ReceiptTypeRead ReceiptType = "read"
	// ReceiptTypePlayed means the media message was played.
	ReceiptTypePlayed ReceiptType = "played"
)

// Connected is the event dispatched when the client successfully connects to WhatsApp.
type Connected struct{}

// Disconnected is the event dispatched when the client disconnects from WhatsApp.
type Disconnected struct {
	// Whether the disconnection was due to a logout.
	LoggedOut bool
}

// QR is the event dispatched when a QR code is available for scanning.
type QR struct {
	// Codes contains the QR code strings to display. Multiple codes may be
	// provided as the QR code refreshes periodically.
	Codes []string
}

// PairSuccess is the event dispatched when QR pairing succeeds.
type PairSuccess struct {
	// ID is the JID of the paired device.
	ID JID
	// BusinessName is the business name if the account is a business account.
	BusinessName string
	// Platform is the platform string returned by the server.
	Platform string
}

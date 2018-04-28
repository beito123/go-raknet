package server

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	raknet "github.com/beito123/go-raknet"
	"github.com/satori/go.uuid"
)

// Handler handles packets, connections and more from Raknet server
type Handler interface {

	// OpenConn received a login packet from client
	OpenConn()

	// HandleRawPacket handles a raw packet no processed in Raknet server
	HandleRawPacket(uid uuid.UUID, pk raknet.Packet) error

	// HandlePacket handles a packet processed in Raknet server
	HandlePacket(uid uuid.UUID, pk raknet.Packet) error
}

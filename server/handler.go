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
	"net"

	raknet "github.com/beito123/go-raknet"
)

// Handler handles packets, connections and more from Raknet server
type Handler interface {

	// StartServer is called when the server is started
	StartServer()

	// CloseServer is called when the server is closed
	CloseServer()

	// HandlePing is called when a ping packet is received
	HandlePing(addr net.Addr)

	// OpenPreConn is called when a new client is created before
	OpenPreConn(addr net.Addr)

	// OpenConn is called when a new client is created
	OpenConn(uid int64, addr net.Addr)

	// ClosePreConn is called when a client is closed before
	ClosePreConn(uid int64)

	// CloseConn is called when a client is closed
	CloseConn(uid int64)

	// Timeout is called when a client is timed out
	Timeout(uid int64)

	// BlockedAddress is called when a client is added blocked address
	AddBlockedAddress(ip net.IP, reason string)

	// BlockedAddress is called when a client is removed blocked address
	RemoveBlockedAddress(ip net.IP, reason string)

	// HandleSendPacket handles a packet sent from the server to a client
	HandleSendPacket(addr net.Addr, pk raknet.Packet)

	// HandleRawPacket handles a raw packet no processed in Raknet server
	HandleRawPacket(addr net.Addr, pk raknet.Packet)

	// HandlePacket handles a message packet
	HandlePacket(uid int64, pk raknet.Packet)

	// HandleUnknownPacket handles a unknown packet
	HandleUnknownPacket(uid int64, pk raknet.Packet)
}

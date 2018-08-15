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

// Handler handles processing from server
type Handler interface {

	// Start is called when the server is started
	StartServer()

	// Close is called when the server is closed
	CloseServer()

	// HandlePing is called when a ping packet is received
	HandlePing(addr net.Addr)

	// OpenedPreConn is called when a new client is created before
	OpenedPreConn(addr net.Addr)

	// OpenedConn is called when a new client is created
	OpenedConn(uid int64, addr net.Addr)

	// ClosedPreConn is called when a client is closed before
	ClosedPreConn(uid int64)

	// ClosedConn is called when a client is closed
	ClosedConn(uid int64)

	// Timeout is called when a client is timed out
	Timedout(uid int64)

	// AddedBlockedAddress is called when a client is added blocked address
	AddedBlockedAddress(ip net.IP, reason string)

	// RemovedBlockedAddress is called when a client is removed blocked address
	RemovedBlockedAddress(ip net.IP)

	// HandleSendPacket handles a packet sent from the server to a client
	HandleSendPacket(addr net.Addr, pk raknet.Packet)

	// HandleRawPacket handles a raw packet no processed in Raknet server
	HandleRawPacket(addr net.Addr, pk raknet.Packet)

	// HandlePacket handles a message packet
	HandlePacket(uid int64, pk raknet.Packet)

	// HandleUnknownPacket handles a unknown packet
	HandleUnknownPacket(uid int64, pk raknet.Packet)
}

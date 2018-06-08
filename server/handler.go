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

	// StartServer is called on the server is started
	StartServer()

	// CloseServer is called on the server is closed
	CloseServer()

	// HandlePing is called on a ping packet is received
	HandlePing(addr net.Addr)

	// OpenPreConn is called before a new session is created
	OpenPreConn(addr net.Addr)

	// OpenConn is called on a new session is created
	OpenConn(uid int64, addr net.Addr)

	// ClosePreConn is called before a session is closed
	ClosePreConn(uid int64)

	// CloseConn is called on a session is closed
	CloseConn(uid int64)

	// HandleRawPacket handles a raw packet no processed in Raknet server
	HandleRawPacket(addr net.Addr, pk raknet.Packet)

	// HandlePacket handles a message packet
	HandlePacket(uid int64, pk raknet.Packet)

	// HandleUnknownPacket handles a unknown packet
	HandleUnknownPacket(uid int64, pk raknet.Packet)
}

/*
// StartServer is called on the server is started
func (Handler) StartServer() {
}

// CloseServer is called on the server is closed
func (Handler) CloseServer() {
}

// HandlePing is called on a ping packet is received
func (Handler) HandlePing(addr *net.Addr) {
}

// OpenPreConn is called before a new session is created
func (Handler) OpenPreConn(addr *net.Addr) {
}

// OpenConn is called on a new session is created
func (Handler) OpenConn(uid int64, addr *net.Addr) {
}

// ClosePreConn is called before a session is closed
func (Handler) ClosePreConn(uid int64) {
}

// CloseConn is called on a session is closed
func (Handler) CloseConn(uid int64) {
}

// HandleRawPacket handles a raw packet no processed in Raknet server
func (Handler) HandleRawPacket(uid int64, pk raknet.Packet) {
}

// HandlePacket handles a message packet
func (Handler) HandlePacket(uid int64, pk raknet.Packet) {
}

// HandleUnknownPacket handles a unknown packet
func (Handler) HandleUnknownPacket(uid int64, pk raknet.Packet) {
}
*/

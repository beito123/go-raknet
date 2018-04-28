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
	"time"

	"github.com/beito123/go-raknet/binary"

	"github.com/beito123/go-raknet/protocol"

	raknet "github.com/beito123/go-raknet"
	"github.com/satori/go.uuid"
)

//

// Session
type Session struct {
	Addr   *net.UDPAddr
	Conn   *net.UDPConn
	UUID   uuid.UUID
	Logger raknet.Logger
	Server *Server

	messageIndex binary.Triad
	splitId      binary.Triad
}

func (session *Session) handlePacket(pk raknet.Packet) {
	//
}

func (session *Session) handleCustomPacket(pk *protocol.CustomPacket) {
	//
}

func (session *Session) SendPacket(pk raknet.Packet, rea raknet.Reliability, channel int) error {
	return nil
}

func (session *Session) SendRawPacket(pk raknet.Packet) {
	session.Server.SendPacket(session.Addr, pk.Bytes())
}

func (session *Session) update() error {
	return nil
}

func (session *Session) Close() error {
	return nil
}

func (session *Session) SetDeadline(t time.Time) error {
	return nil
}

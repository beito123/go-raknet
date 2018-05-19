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
	"context"
	"errors"
	"net"

	"github.com/beito123/go-raknet/binary"
	"github.com/beito123/go-raknet/protocol"

	raknet "github.com/beito123/go-raknet"
)

//

var (
	errSessionClosed = errors.New("session closed")
)

type SessionState int

const (
	StateDisconected SessionState = iota
	StateHandshaking
	StateConnected
)

// Session
type Session struct {
	Addr   *net.UDPAddr
	Conn   *net.UDPConn
	Logger raknet.Logger
	Server *Server
	GUID   int64
	MTU    int

	messageIndex binary.Triad
	splitId      binary.Triad

	ctx context.Context

	state SessionState
}

func (session *Session) init() {
	session.state = StateConnected
}

func (session *Session) State() SessionState {
	return session.state
}

func (session *Session) handlePacket(pk raknet.Packet) {
	if session.State() == StateDisconected {
		return
	}

}

func (session *Session) handleCustomPacket(pk *protocol.CustomPacket) {
	if session.State() == StateDisconected {
		return
	}

}

func (session *Session) handleACKPacket(pk *protocol.ACK) {
	if session.State() == StateDisconected {
		return
	}

}

func (session *Session) SendPacket(pk raknet.Packet, rea raknet.Reliability, channel int) error {
	return nil
}

func (session *Session) SendRawPacket(pk raknet.Packet) {
	session.Server.SendPacket(session.Addr, pk)
}

func (session *Session) update() {
	select {
	case <-session.ctx.Done():
		return
	default:
	}

	//
}

// Close closes the session
// Notice!: Don't use this close function for close session
// Use CloseSession in Server instead of it
func (session *Session) Close() error {
	if session.State() == StateDisconected {
		return errSessionClosed
	}

	//session.Server.CloseSession(session.UUID, "Disconnected from server")
	session.state = StateDisconected

	return nil
}

func (session *Session) close() {
	if session.State() == StateDisconected {
		return
	}

	session.state = StateDisconected

	//
}

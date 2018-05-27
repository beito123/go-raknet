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
	Addr              *net.UDPAddr
	Conn              *net.UDPConn
	Logger            raknet.Logger
	Server            *Server
	GUID              int64
	MTU               int
	LatencyEnabled    bool
	LatencyIdentifier int64

	messageIndex binary.Triad
	splitId      binary.Triad

	reliablePackets []int

	sendQueue         []*protocol.EncapsulatedPacket
	recoveryQueue     [][]*protocol.EncapsulatedPacket
	ackReceiptPackets map[int]*protocol.EncapsulatedPacket

	sendSequenceNumber    int
	receiveSequenceNumber int

	ctx context.Context

	State SessionState
}

func (session *Session) SystemAddress() *raknet.SystemAddress {
	return raknet.NewSystemAddressBytes([]byte(session.Addr.IP), uint16(session.Addr.Port))
}

func (session *Session) init() {
	//
}

func (session *Session) handlePacket(pk raknet.Packet) {
	if session.State == StateDisconected {
		return
	}

	switch npk := pk.(type) {
	case *protocol.ConnectedPing:
		err := npk.Decode()
		if err != nil {
			session.Logger.Warn(err)
			return
		}

		pong := &protocol.ConnectedPong{
			Time: npk.Time,
		}

		err = pong.Encode()
		if err != nil {
			session.Logger.Warn(err)
			return
		}

		err = session.SendPacket(pong, raknet.Unreliable, raknet.DefaultChannel)
		if err != nil {
			session.Logger.Warn(err)
		}
	case *protocol.ConnectedPong:
		err := npk.Decode()
		if err != nil {
			session.Logger.Warn(err)
			return
		}

		if session.LatencyEnabled {
			// TODO: writes
		}
	case *protocol.ConnectionRequestAccepted:
		if session.State != StateHandshaking {
			return
		}

		err := npk.Decode()
		if err != nil {
			session.Logger.Warn(err) // remove

			session.Server.CloseSession(session.Addr, "Failed to login")
			return
		}

		hpk := &protocol.NewIncomingConnection{
			ServerAddress:   *session.SystemAddress(),
			ClientTimestamp: npk.ServerTimestamp,
			ServerTimestamp: npk.ClientTimestamp,
		}

		err = session.SendPacket(hpk, raknet.ReliableOrderedWithACKReceipt, raknet.DefaultChannel)
		if err != nil {
			session.Server.CloseSession(session.Addr, "Failed to login")
		}
	case *protocol.DisconnectionNotification:
		err := npk.Decode()
		if err != nil {
			session.Logger.Warn(err)
			return
		}

		session.Server.CloseSession(session.Addr, "Server disconnected")
	default:
		if npk.ID() >= protocol.IDUserPacketEnum { // user packet
			if session.Server.Handler != nil {
				session.Server.Handler.HandlePacket(session.GUID, npk)
			}
		} else { // unknown packet
			if session.Server.Handler != nil {
				session.Server.Handler.HandleUnknownPacket(session.GUID, npk)
			}
		}
	}
}

func (session *Session) handleCustomPacket(pk *protocol.CustomPacket) {
	if session.State == StateDisconected {
		return
	}

}

func (session *Session) handleACKPacket(pk *protocol.ACK) {
	if session.State == StateDisconected {
		return
	}

}

func (session *Session) SendPacket(pk raknet.Packet, rea raknet.Reliability, channel int) error {
	return nil
}

func (session *Session) SendRawPacket(pk raknet.Packet) {
	session.Server.SendPacket(session.Addr, pk)
}

func (session *Session) update() bool {
	select {
	case <-session.ctx.Done():
		return false
	default:
	}

	if session.State == StateDisconected {
		return false
	}

	//

	return true
}

// Close closes the session
func (session *Session) Close() error {
	if session.State == StateDisconected {
		return errSessionClosed
	}

	//session.Server.CloseSession(session.UUID, "Disconnected from server")
	session.State = StateDisconected

	return nil
}

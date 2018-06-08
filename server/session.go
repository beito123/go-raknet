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
	"github.com/beito123/go-raknet/util"
	"github.com/orcaman/concurrent-map"

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
	splitID      int

	reliablePackets []int

	sendQueue         *util.Queue
	recoveryQueue     cmap.ConcurrentMap // [][]*protocol.EncapsulatedPacket
	ackReceiptPackets map[int]*protocol.EncapsulatedPacket

	sendSequenceNumber    int
	receiveSequenceNumber int

	orderSendIndex       map[int]binary.Triad
	orderReceiveIndex    map[int]binary.Triad
	sequenceSendIndex    map[int]binary.Triad
	sequenceReceiveIndex map[int]binary.Triad

	ctx context.Context

	State SessionState
}

func (session *Session) SystemAddress() *raknet.SystemAddress {
	return raknet.NewSystemAddressBytes([]byte(session.Addr.IP), uint16(session.Addr.Port))
}

func (session *Session) init() {
	session.sendQueue = util.NewQueue()
	session.recoveryQueue = cmap.New()

	session.orderSendIndex = make(map[int]binary.Triad, raknet.MaxChannels)
	session.orderReceiveIndex = make(map[int]binary.Triad, raknet.MaxChannels)
	session.sequenceSendIndex = make(map[int]binary.Triad, raknet.MaxChannels)
	session.sequenceReceiveIndex = make(map[int]binary.Triad, raknet.MaxChannels)
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

		_, err = session.SendPacket(pong, raknet.Unreliable, raknet.DefaultChannel)
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
			ServerAddress:   session.SystemAddress(),
			ClientTimestamp: npk.ServerTimestamp,
			ServerTimestamp: npk.ClientTimestamp,
		}

		_, err = session.SendPacket(hpk, raknet.ReliableOrderedWithACKReceipt, raknet.DefaultChannel)
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

	if session.Server.Handler != nil {
		session.Server.Handler.HandlePacket(session.GUID, pk)
	}

}

func (session *Session) handleACKPacket(pk *protocol.ACK) {
	if session.State == StateDisconected {
		return
	}

}

func (session *Session) addSendQueue(epk *protocol.EncapsulatedPacket) {
	session.sendQueue.Add(epk)
}

func (session *Session) pollSendQueue() (*protocol.EncapsulatedPacket, error) {
	item, _ := session.sendQueue.Poll()

	epk, ok := item.(*protocol.EncapsulatedPacket)
	if !ok {
		return nil, errors.New("invalid a value")
	}

	return epk, nil
}

func (session *Session) bumpMessageIndex() binary.Triad {
	session.messageIndex = session.messageIndex.Bump()
	return session.messageIndex
}

func (session *Session) bumpOrderSendIndex(channel int) binary.Triad {
	session.orderSendIndex[channel] = session.orderSendIndex[channel].Bump()
	return session.orderSendIndex[channel]
}

func (session *Session) bumpSequenceSendIndex(channel int) binary.Triad {
	session.sequenceSendIndex[channel] = (session.sequenceSendIndex[channel]).Bump()
	return session.sequenceSendIndex[channel]
}

func (session *Session) bumpSplitID() uint16 {
	session.splitID = (session.splitID + 1) % 65536
	return uint16(session.splitID)
}

func (session *Session) SendPacket(pk raknet.Packet, reliability raknet.Reliability, channel int) (protocol.EncapsulatedPacket, error) {
	if channel >= raknet.MaxChannels {
		return protocol.EncapsulatedPacket{}, errors.New("invalid channel")
	}

	epk := &protocol.EncapsulatedPacket{
		Reliability:  reliability,
		OrderChannel: byte(channel),
		Payload:      pk.Bytes(),
	}

	if reliability.IsReliable() {
		epk.MessageIndex = session.bumpMessageIndex()
	}

	if reliability.IsOrdered() || reliability.IsSequenced() {
		if reliability.IsOrdered() {
			epk.OrderIndex = session.bumpSequenceSendIndex(channel)
		} else {
			epk.OrderIndex = session.bumpOrderSendIndex(channel)
		}

		//session.Logger.Debug("Bumped" + )
	}

	if needSplit(epk.Reliability, pk, session.MTU) {
		epk.SplitID = session.bumpSplitID()

		for _, spk := range session.splitPacket(epk) {
			session.addSendQueue(spk)
		}
	} else {
		session.addSendQueue(epk)
	}

	return *epk, nil // returns clone of epk
}

func (session *Session) splitPacket(epk *protocol.EncapsulatedPacket) []*protocol.EncapsulatedPacket {
	exp := util.SplitBytesSlice(epk.Payload,
		session.MTU-(protocol.CalcCPacketBaseSize()+protocol.CalcEPacketSize(epk.Reliability, true, []byte{})))

	spk := make([]*protocol.EncapsulatedPacket, len(exp))

	for i := 0; i < len(exp); i++ {
		npk := &protocol.EncapsulatedPacket{
			Reliability: epk.Reliability,
			Payload:     exp[i],
		}

		if epk.Reliability.IsReliable() {
			npk.MessageIndex = session.bumpMessageIndex()
		} else {
			npk.MessageIndex = epk.MessageIndex
		}

		if epk.Reliability.IsOrdered() || epk.Reliability.IsSequenced() {
			npk.OrderChannel = epk.OrderChannel
			npk.OrderIndex = epk.OrderIndex
		}

		npk.Split = true
		npk.SplitCount = int32(len(exp))
		npk.SplitID = epk.SplitID
		npk.SplitIndex = int32(i)

		spk[i] = npk
	}

	return spk
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

	// send a disconnection notification packet

	return nil
}

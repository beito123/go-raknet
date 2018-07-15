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
	// Addr is the client's address to connect
	Addr *net.UDPAddr

	// Conn is a connection for the client
	Conn *net.UDPConn

	// Logger is a logger
	Logger raknet.Logger

	// Server is the server instance.
	Server *Server

	// GUID is a GUID
	GUID int64

	// MTU is the max packet size to receive and send
	MTU int

	// LatencyEnabled enables measuring a latency time
	LatencyEnabled bool

	// LatencyID is a latency identifier
	LatencyID int64

	State SessionState

	// messageIndex is a message index of EncapsulatedPacket
	messageIndex binary.Triad

	// splitID is a handling split id
	splitID int

	// reliablePackets contains handled reliable packets
	reliablePackets map[binary.Triad]bool

	// splitQueue contains split packets with split id
	// It's used to handle split packets
	splitQueue map[uint16]*SplitPacket

	// sendQueue is a queue contained packets to send
	sendQueue util.IntMap // map[int]*protocol.EncapsulatedPacket

	// sendQueueOffset is a offset of sendQueue
	sendQueueOffset int

	// recoveryQueue is a recovery queue by NACK to send the client
	// When the session received NACK packet, add lost packets.
	recoveryQueue util.IntMap // map[int][]*protocol.EncapsulatedPacket

	// ackReceiptPackets contains ack index of sent packets to the client
	ackReceiptPackets map[int]*protocol.EncapsulatedPacket

	// sendSequenceNumber is sent the newest sequence number to the client
	// It's used in CustomPacket
	sendSequenceNumber binary.Triad

	// receiveSequenceNumber is received the newest sequence number from the client
	// It's used in CustomPacket
	receiveSequenceNumber binary.Triad

	// orderSendIndex contains the newest sent order indexes to client with order channel.
	// It's used in EncapsulatedPacket
	orderSendIndex map[int]binary.Triad

	// orderReceiveIndex contains the newest received order indexes to client with order channel.
	// It's used in EncapsulatedPacket
	orderReceiveIndex map[int]binary.Triad

	// sequenceSendIndex contains the newest sent sequence indexes to client with order channel.
	// It's used in EncapsulatedPacket
	sequenceSendIndex map[int]binary.Triad

	// sequenceReceiveIndex contains the newest received sequence indexes to client with order channel.
	// It's used in EncapsulatedPacket
	sequenceReceiveIndex map[int]binary.Triad

	// handleQueue contains ordered packets with order channel
	// It's used to handle ordered packet in the order
	handleQueue map[byte]map[binary.Triad]*protocol.EncapsulatedPacket

	// PacketReceivedCount is sent a packet counter
	// It's used to check packets count on every second
	PacketSentCount int

	// PacketReceivedCount is received a packet counter
	// It's used to check packets count on every second
	PacketReceivedCount int

	// LastPacketSendTime is the last time sent a packet
	LastPacketSendTime time.Time

	// LastPacketReceiveTime is the last time received a packet
	LastPacketReceiveTime time.Time

	// LastRecoverySendTime is the last time sent a lost packet
	LastRecoverySendTime time.Time

	// LastKeepAliveSendTime is the last time sent DetectLostConnection packet
	LastKeepAliveSendTime time.Time

	// LastPacketCounterResetTime is the last time reset packet counters
	LastPacketCounterResetTime time.Time

	// handshakeRecord is a records of a handshake packet
	// It's used to detect whether the client is connected
	handshakeRecord *raknet.Record

	ctx context.Context
}

func (session *Session) SystemAddress() *raknet.SystemAddress {
	return raknet.NewSystemAddressBytes([]byte(session.Addr.IP), uint16(session.Addr.Port))
}

func (session *Session) Init() {
	session.reliablePackets = make(map[binary.Triad]bool)
	session.splitQueue = make(map[uint16]*SplitPacket)

	session.sendQueue = util.NewIntMap()
	session.recoveryQueue = util.NewIntMap()

	session.ackReceiptPackets = make(map[int]*protocol.EncapsulatedPacket)

	session.orderSendIndex = make(map[int]binary.Triad, raknet.MaxChannels)
	session.orderReceiveIndex = make(map[int]binary.Triad, raknet.MaxChannels)
	session.sequenceSendIndex = make(map[int]binary.Triad, raknet.MaxChannels)
	session.sequenceReceiveIndex = make(map[int]binary.Triad, raknet.MaxChannels)

	session.handleQueue = make(map[byte]map[binary.Triad]*protocol.EncapsulatedPacket)

	session.LastPacketSendTime = time.Now()
	session.LastPacketReceiveTime = time.Now()
	session.LastRecoverySendTime = time.Now()
	session.LastKeepAliveSendTime = time.Now()
	session.LastPacketCounterResetTime = time.Now()
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
			for _, hand := range session.Server.Handlers {
				hand.HandlePacket(session.GUID, npk)
			}
		} else { // unknown packet
			for _, hand := range session.Server.Handlers {
				hand.HandleUnknownPacket(session.GUID, npk)
			}
		}
	}
}

func (session *Session) handleCustomPacket(pk *protocol.CustomPacket) {
	if session.State == StateDisconected {
		return
	}

	for _, handler := range session.Server.Handlers { // Debug: I'll remove
		handler.HandlePacket(session.GUID, pk)
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

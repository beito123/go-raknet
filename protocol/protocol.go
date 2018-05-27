package protocol

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import "github.com/beito123/go-raknet"

type Protocol struct {
	packets []raknet.Packet
}

func (protocol *Protocol) RegisterPackets() {
	protocol.packets = make([]raknet.Packet, 0xff)

	protocol.packets[IDConnectedPing] = &ConnectedPing{}
	protocol.packets[IDUnconnectedPing] = &UnconnectedPing{}
	protocol.packets[IDUnconnectedPingOpenConnections] = &UnconnectedPingOpenConnections{}
	protocol.packets[IDOpenConnectionRequest1] = &OpenConnectionRequestOne{}
	protocol.packets[IDOpenConnectionReply1] = &OpenConnectionResponseOne{}
	protocol.packets[IDOpenConnectionRequest2] = &OpenConnectionRequestTwo{}
	protocol.packets[IDOpenConnectionReply2] = &OpenConnectionResponseTwo{}
	protocol.packets[IDConnectionRequest] = &ConnectionRequest{}
	protocol.packets[IDConnectionRequestAccepted] = &ConnectionRequestAccepted{}
	protocol.packets[IDAlreadyConnected] = &AlreadyConnected{}
	protocol.packets[IDNewIncomingConnection] = &NewIncomingConnection{}
	protocol.packets[IDNoFreeIncomingConnections] = &NoFreeIncomingConnections{}
	protocol.packets[IDDisconnectionNotification] = &DisconnectionNotification{}
	protocol.packets[IDConnectionBanned] = &ConnectionBanned{}
	protocol.packets[IDIncompatibleProtocolVersion] = &IncompatibleProtocol{}
	protocol.packets[IDUnconnectedPong] = &UnconnectedPong{}
	protocol.packets[IDCustom0] = NewCustomPacket(IDCustom0)
	protocol.packets[IDCustom1] = NewCustomPacket(IDCustom1)
	protocol.packets[IDCustom2] = NewCustomPacket(IDCustom2)
	protocol.packets[IDCustom3] = NewCustomPacket(IDCustom3)
	protocol.packets[IDCustom4] = NewCustomPacket(IDCustom4)
	protocol.packets[IDCustom5] = NewCustomPacket(IDCustom5)
	protocol.packets[IDCustom6] = NewCustomPacket(IDCustom6)
	protocol.packets[IDCustom7] = NewCustomPacket(IDCustom7)
	protocol.packets[IDCustom8] = NewCustomPacket(IDCustom8)
	protocol.packets[IDCustom9] = NewCustomPacket(IDCustom9)
	protocol.packets[IDCustomA] = NewCustomPacket(IDCustomA)
	protocol.packets[IDCustomB] = NewCustomPacket(IDCustomB)
	protocol.packets[IDCustomC] = NewCustomPacket(IDCustomC)
	protocol.packets[IDCustomD] = NewCustomPacket(IDCustomD)
	protocol.packets[IDCustomE] = NewCustomPacket(IDCustomE)
	protocol.packets[IDCustomF] = NewCustomPacket(IDCustomF)

}

func (protocol *Protocol) Packet(id byte) (pk raknet.Packet, ok bool) {
	pk = protocol.packets[id]
	if pk == nil {
		return nil, false
	}

	return pk, true
}

func (protocol *Protocol) Packets() []raknet.Packet {
	return protocol.packets
}

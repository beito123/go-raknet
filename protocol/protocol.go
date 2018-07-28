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
	protocol.packets[IDConnectedPong] = &ConnectedPong{}
	protocol.packets[IDDetectLostConnections] = &DetectLostConnections{}
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
	protocol.packets[IDACK] = &Acknowledge{
		Type: TypeACK,
	}
	protocol.packets[IDNACK] = &Acknowledge{
		Type: TypeNACK,
	}

	for i := IDCustom0; i <= IDCustomF; i++ {
		protocol.packets[i] = NewCustomPacket(byte(i))
	}

}

func (protocol *Protocol) Packet(id byte) (pk raknet.Packet, ok bool) {
	pk = protocol.packets[id]

	return pk, pk != nil
}

func (protocol *Protocol) Packets() []raknet.Packet {
	return protocol.packets
}

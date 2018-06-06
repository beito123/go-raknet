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
	raknet "github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/protocol"
)

func needSplit(reliability raknet.Reliability, pk raknet.Packet, mtu int) bool {
	return (protocol.CalcCPacketBaseSize() + 
	protocol.CalcEPacketSize(reliability, false, pk.Bytes())) > mtu
}

//func splitPacket(epk *protocol.EncapsulatedPacket, mtu int)

// SplitPacket is used to easily assemble split packets
type SplitPacket struct {
	SplitID     int
	SplitCount  int
	Reliability raknet.Reliability

	Packets map[int]raknet.Packet
}

package pk

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
}

func (protocol *Protocol) Packet(id byte) raknet.Packet {
	return protocol.packets[id]
}

func (protocol *Protocol) Packets() []raknet.Packet {
	return protocol.packets
}

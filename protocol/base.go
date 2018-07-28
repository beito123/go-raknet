package protocol

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/binary"
)

type BasePacket struct {
	binary.RaknetStream
}

func (base *BasePacket) Encode(pk raknet.Packet) error {
	err := base.PutByte(pk.ID())
	if err != nil {
		return err
	}

	return nil
}

func (base *BasePacket) Decode(pk raknet.Packet) error {
	base.Skip(1) // for id

	return nil
}

func (base *BasePacket) Bytes() []byte {
	return base.RaknetStream.AllBytes()
}

<<<<<<< HEAD
func NewRawPacket(b []byte) *RawPacket {
=======
func NewRaknetPacket(id byte) *RaknetPacket {
	return &RaknetPacket{
		id: id,
	}
}

func NewRaknetPacketBytes(b []byte) *RaknetPacket {
>>>>>>> origin/master
	id := byte(0xff)
	if len(b) > 0 {
		id = b[0]
	}

<<<<<<< HEAD
	pk := &RawPacket{
		id: id,
	}
=======
	pk := NewRaknetPacket(id)
>>>>>>> origin/master

	pk.SetBytes(b)

	return pk
}

<<<<<<< HEAD
type RawPacket struct {
=======
type RaknetPacket struct {
>>>>>>> origin/master
	BasePacket

	id byte
}

<<<<<<< HEAD
func (pk *RawPacket) ID() byte {
	return pk.id
}

func (pk *RawPacket) Encode() error {
	return nil
}

func (pk *RawPacket) Decode() error {
	return nil
}

func (pk *RawPacket) New() raknet.Packet {
	return &RawPacket{
=======
func (pk *RaknetPacket) ID() byte {
	return pk.id
}

func (pk *RaknetPacket) Encode() error {
	return nil
}

func (pk *RaknetPacket) Decode() error {
	return nil
}

func (pk *RaknetPacket) New() raknet.Packet {
	return &RaknetPacket{
>>>>>>> origin/master
		id: pk.id,
	}
}

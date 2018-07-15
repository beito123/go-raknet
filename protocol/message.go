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
	"errors"

	"github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/binary"
)

const (
	EPacketMinBufferLen                         = 3
	EPacketBitFlagLen                           = 1
	EPacketPayloadLengthLen                     = 2
	EPacketMessageIndexLen                      = 3
	EPacketOrderIndexAndOrderChannelLen         = 4
	EPacketSplitCountAndSplitIDAndSplitIndexLen = 10
)

const (
	ReliabilityPosition = 5 //
)

const (
	FlagReliability = 224 // 0b11100000
	FlagSplit       = 16  // 0b00010000
)

type EncapsulatedPacket struct {
	Buf *binary.RaknetStream

	Reliability  raknet.Reliability
	Split        bool
	MessageIndex binary.Triad
	OrderIndex   binary.Triad
	OrderChannel byte
	SplitCount   int32
	SplitID      uint16
	SplitIndex   int32

	Payload []byte

	Record *raknet.Record
}

func (epk *EncapsulatedPacket) Encode() error {
	if epk.Buf == nil {
		epk.Buf = binary.NewStream()
	}

	flags := epk.Reliability.ToBinary() << ReliabilityPosition
	if epk.Split {
		flags |= FlagSplit
	}

	err := epk.Buf.PutByte(flags)
	if err != nil {
		return err
	}

	err = epk.Buf.PutShort(uint16(len(epk.Payload) << 3))
	if err != nil {
		return err
	}

	if epk.Reliability.IsReliable() {
		err = epk.Buf.PutLTriad(epk.MessageIndex)
		if err != nil {
			return err
		}
	}

	if epk.Reliability.IsOrdered() || epk.Reliability.IsSequenced() {
		err = epk.Buf.PutLTriad(epk.OrderIndex)
		if err != nil {
			return err
		}

		err = epk.Buf.PutByte(epk.OrderChannel)
		if err != nil {
			return err
		}
	}

	if epk.Split {
		err = epk.Buf.PutInt(epk.SplitCount)
		if err != nil {
			return err
		}

		err = epk.Buf.PutShort(epk.SplitID)
		if err != nil {
			return err
		}

		err = epk.Buf.PutInt(epk.SplitIndex)
		if err != nil {
			return err
		}
	}

	err = epk.Buf.Put(epk.Payload)
	if err != nil {
		return err
	}

	return nil
}

func (epk *EncapsulatedPacket) Decode() error {
	if epk.Buf == nil {
		return errors.New("no sets buffer")
	}

	flags, err := epk.Buf.Byte()
	if err != nil {
		return err
	}

	epk.Reliability = raknet.ReliabilityBinary(flags >> ReliabilityPosition)
	epk.Split = (flags & FlagSplit) > 0

	payloadLen, err := epk.Buf.Short()
	if err != nil {
		return err
	}

	length := int(payloadLen / 8)

	if epk.Reliability.IsReliable() {
		epk.MessageIndex, err = epk.Buf.LTriad()
		if err != nil {
			return err
		}
	}

	if epk.Reliability.IsOrdered() || epk.Reliability.IsSequenced() {
		epk.OrderIndex, err = epk.Buf.LTriad()
		if err != nil {
			return err
		}

		epk.OrderChannel, err = epk.Buf.Byte()
		if err != nil {
			return err
		}
	}

	if epk.Split {
		epk.SplitCount, err = epk.Buf.Int()
		if err != nil {
			return err
		}

		epk.SplitID, err = epk.Buf.Short()
		if err != nil {
			return err
		}

		epk.SplitIndex, err = epk.Buf.Int()
		if err != nil {
			return err
		}
	}

	epk.Payload = epk.Buf.Get(length)

	return nil
}

func (epk *EncapsulatedPacket) CalcSize() int {
	return CalcEPacketSize(epk.Reliability, epk.Split, epk.Payload)
}

func CalcEPacketSize(reliability raknet.Reliability, split bool, payload []byte) int {
	var size int
	size += EPacketBitFlagLen
	size += EPacketPayloadLengthLen

	if reliability.IsReliable() {
		size += EPacketMessageIndexLen
	}

	if reliability.IsOrdered() || reliability.IsSequenced() {
		size += EPacketOrderIndexAndOrderChannelLen
	}

	if split {
		size += EPacketSplitCountAndSplitIDAndSplitIndexLen
	}

	size += len(payload)

	return size
}

func NewCustomPacket(id byte) *CustomPacket {
	return &CustomPacket{
		id: id,
	}
}

type CustomPacket struct {
	BasePacket

	id byte

	Index    binary.Triad
	Messages []*EncapsulatedPacket

	Records []*raknet.Record
}

func (pk *CustomPacket) ID() byte {
	return pk.id
}

func (pk *CustomPacket) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLTriad(pk.Index)
	if err != nil {
		return err
	}

	for _, epk := range pk.Messages {
		epk.Buf = &pk.RaknetStream

		if epk.Reliability.IsNeededACK() {
			epk.Record = &raknet.Record{
				Index: int(pk.Index),
			}

			pk.Records = append(pk.Records, epk.Record)
		}

		err = epk.Encode()
		if err != nil {
			return err
		}
	}

	return nil
}

func (pk *CustomPacket) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	pk.Index, err = pk.LTriad()
	if err != nil {
		return err
	}

	for pk.Len() >= EPacketMinBufferLen {
		epk := &EncapsulatedPacket{
			Buf: &pk.RaknetStream,
		}

		err = epk.Decode()
		if err != nil {
			return err
		}

		if epk.Reliability.IsNeededACK() {
			epk.Record = &raknet.Record{
				Index: int(pk.Index),
			}

			pk.Records = append(pk.Records, epk.Record)
		}

		pk.Messages = append(pk.Messages, epk)
	}

	return nil
}

func (pk *CustomPacket) RemoveUnreliables() {
	var epks []*EncapsulatedPacket

	for _, epk := range pk.Messages {
		if epk.Reliability.IsReliable() {
			epks = append(epks, epk)
		}
	}

	pk.Messages = epks
}

func (pk *CustomPacket) CalcSize() int {
	size := CalcCPacketBaseSize()
	for _, epk := range pk.Messages {
		size += epk.CalcSize()
	}

	return size
}

func (pk *CustomPacket) New() raknet.Packet {
	return NewCustomPacket(pk.id)
}

func CalcCPacketBaseSize() int {
	return 1 + 3 // pk id + index
}

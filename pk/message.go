package pk

import (
	"github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/binary"
)

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

const (
	MinBufferLen                         = 3
	BitFlagLen                           = 1
	PayloadLengthLen                     = 2
	MessageIndexLen                      = 3
	OrderIndexAndOrderChannelLen         = 4
	SplitCountAndSplitIDAndSplitIndexLen = 10
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
}

func (epk *EncapsulatedPacket) Encode() error {
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
	var flags byte

	err := epk.Buf.Byte(&flags)
	if err != nil {
		return err
	}

	epk.Reliability = raknet.ReliabilityBinary(flags >> ReliabilityPosition)
	epk.Split = (flags & FlagSplit) > 0

	var payloadLen uint16
	err = epk.Buf.Short(&payloadLen)
	if err != nil {
		return err
	}

	length := int(payloadLen / 8)

	if epk.Reliability.IsReliable() {
		err = epk.Buf.LTriad(&epk.MessageIndex)
		if err != nil {
			return err
		}
	}

	if epk.Reliability.IsOrdered() || epk.Reliability.IsSequenced() {
		err = epk.Buf.LTriad(&epk.OrderIndex)
		if err != nil {
			return err
		}

		err = epk.Buf.Byte(&epk.OrderChannel)
		if err != nil {
			return err
		}
	}

	if epk.Split {
		err = epk.Buf.Int(&epk.SplitCount)
		if err != nil {
			return err
		}

		err = epk.Buf.Short(&epk.SplitID)
		if err != nil {
			return err
		}

		err = epk.Buf.Int(&epk.SplitIndex)
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
	size += BitFlagLen
	size += PayloadLengthLen

	if reliability.IsReliable() {
		size += MessageIndexLen
	}

	if reliability.IsOrdered() || reliability.IsSequenced() {
		size += OrderIndexAndOrderChannelLen
	}

	if split {
		size += SplitCountAndSplitIDAndSplitIndexLen
	}

	size += len(payload)

	return size
}

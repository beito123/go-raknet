package pk

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"github.com/beito123/go-raknet/binary"
	"github.com/beito123/go-raknet"
	"bytes"
	"errors"
)

type BasePacket struct {
	binary.RaknetStream
}

func (base *BasePacket) Encode(pk raknet.Packet) error {
	if base.Buffer == nil {
		base.Buffer = &bytes.Buffer{}
	}

	err := base.PutByte(pk.ID())
	if err != nil {
		return err
	}

	return nil
}

func (base *BasePacket) Decode(pk raknet.Packet) error {
	if base.Buffer == nil {
		return errors.New("the buffer isn't set")
	}

	base.Skip(1) // for id

	return nil
}
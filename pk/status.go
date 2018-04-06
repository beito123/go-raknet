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

type ConnectedPing struct {
	BasePacket
	time int64
}

func (pk ConnectedPing) ID() byte {
	return IDConnectedPing
}

func (pk *ConnectedPing) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.time)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectedPing) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.time)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectedPing) New() raknet.Packet {
	return new(ConnectedPing)
}

type ConnectedPong struct {
	BasePacket
	time int64
}

func (pk ConnectedPong) ID() byte {
	return IDConnectedPong
}

func (pk *ConnectedPong) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.time)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectedPong) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.time)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectedPong) New() raknet.Packet {
	return new(ConnectedPong)
}

type UnconnectedPing struct {
	BasePacket
	Timestamp int64
	Magic []byte
	PingID int64
	ConnectionMagic []byte
	Connection raknet.ConnectionType
}

func (pk UnconnectedPing) ID() byte {
	return IDUnconnectedPing
}

func (pk *UnconnectedPing) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.Timestamp)
	if err != nil {
		return err
	}

	err = pk.Put(raknet.Magic)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.PingID)
	if err != nil {
		return err
	}

	err = pk.Put(raknet.ConnctionTypeMagic)
	if err != nil {
		return err
	}

	err = pk.PutConnectionType(pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *UnconnectedPing) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.Timestamp)
	if err != nil {
		return err
	}

	pk.Magic = pk.Get(len(raknet.Magic))

	err = pk.Long(&pk.PingID)
	if err != nil {
		return err
	}

	if pk.Len() < len(raknet.ConnctionTypeMagic) {
		pk.Connection = raknet.ConnectionVanilla

		return nil
	}

	pk.ConnectionMagic = pk.Get(len(raknet.ConnctionTypeMagic))

	err = pk.ConnectionType(&pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *UnconnectedPing) New() raknet.Packet {
	return new(UnconnectedPing)
}

type UnconnectedPingOpenConnections struct {
	UnconnectedPing
}

func (pk UnconnectedPingOpenConnections) ID() byte {
	return IDUnconnectedPingOpenConnections
}

func (pk *UnconnectedPingOpenConnections) New() raknet.Packet {
	return new(UnconnectedPingOpenConnections)
}

type UnconnectedPong struct {
	BasePacket
	Timestamp int64
	PongID int64
	Magic []byte
	Identifier raknet.Identifier
	ConnectionMagic []byte
	Connection raknet.ConnectionType
}

func (pk UnconnectedPong) ID() byte {
	return IDUnconnectedPong
}

func (pk *UnconnectedPong) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.Timestamp)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.PongID)
	if err != nil {
		return err
	}

	err = pk.Put(raknet.Magic)
	if err != nil {
		return err
	}

	err = pk.PutString(pk.Identifier.Build())
	if err != nil {
		return err
	}

	err = pk.Put(raknet.ConnctionTypeMagic)
	if err != nil {
		return err
	}

	err = pk.PutConnectionType(pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *UnconnectedPong) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.Timestamp)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.PongID)
	if err != nil {
		return err
	}

	pk.Magic = pk.Get(len(raknet.Magic))

	var identifier string

	err = pk.String(&identifier)
	if err != nil {
		return err
	}

	if pk.Len() >= len(raknet.ConnctionTypeMagic) {
		pk.ConnectionMagic = pk.Get(len(raknet.ConnctionTypeMagic))

		err = pk.ConnectionType(&pk.Connection)
		if err != nil {
			return err
		}
	} else {
		pk.Connection = raknet.ConnectionVanilla
	}

	pk.Identifier = raknet.Identifier{
		Identifier: identifier,
		ConnectionType: pk.Connection,
	}

	return nil
}

func (pk *UnconnectedPong) New() raknet.Packet {
	return new(UnconnectedPong)
}
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

type ConnectionBanned struct {
	BasePacket

	ServerGUID int64
}

func (ConnectionBanned) ID() byte {
	return IDConnectionBanned
}

func (pk *ConnectionBanned) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ServerGUID)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectionBanned) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.ServerGUID)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectionBanned) New() raknet.Packet {
	return new(ConnectionBanned)
}

type ConnectionRequest struct {
	BasePacket

	ClientGuid int64
	Timestamp  int64

	// UseSecurity is Raknet's built in security function
	// But go-raknet doesn't support.
	UseSecurity bool
}

func (pk ConnectionRequest) ID() byte {
	return IDConnectionRequest
}

func (pk *ConnectionRequest) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ClientGuid)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.Timestamp)
	if err != nil {
		return err
	}

	err = pk.PutBool(pk.UseSecurity)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectionRequest) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.ClientGuid)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.Timestamp)
	if err != nil {
		return err
	}

	err = pk.Bool(&pk.UseSecurity)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectionRequest) New() raknet.Packet {
	return new(ConnectionRequest)
}

type ConnectionRequestAccepted struct {
	BasePacket

	ClientAddress     raknet.SystemAddress
	SystemIndex       uint16                 // unknown
	InternalAddresses []raknet.SystemAddress // unknown
	ClientTimestamp   int64
	ServerTimestamp   int64
}

func (pk ConnectionRequestAccepted) ID() byte {
	return IDConnectionRequestAccepted
}

func (pk *ConnectionRequestAccepted) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutAddressSystemAddress(pk.ClientAddress)
	if err != nil {
		return err
	}

	err = pk.PutShort(pk.SystemIndex)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		var address raknet.SystemAddress
		if i < len(pk.InternalAddresses) {
			address = pk.InternalAddresses[i]
		} else {
			address = *raknet.NewSystemAddress("0.0.0.0", 0)
		}

		err = pk.PutAddressSystemAddress(address)
		if err != nil {
			return err
		}
	}

	err = pk.PutLong(pk.ClientTimestamp)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ServerTimestamp)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectionRequestAccepted) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.AddressSystemAddress(&pk.ClientAddress)
	if err != nil {
		return err
	}

	err = pk.Short(&pk.SystemIndex)
	if err != nil {
		return err
	}

	pk.InternalAddresses = make([]raknet.SystemAddress, 10)
	for i := 0; i < 10; i++ {
		err = pk.AddressSystemAddress(&pk.InternalAddresses[i])
		if err != nil {
			return err
		}
	}

	err = pk.Long(&pk.ClientTimestamp)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.ServerTimestamp)
	if err != nil {
		return err
	}

	return nil
}

func (pk *ConnectionRequestAccepted) New() raknet.Packet {
	return new(ConnectionRequestAccepted)
}

type IncompatibleProtocol struct {
	BasePacket

	NetworkProtocol byte
	Magic           bool
	ServerGuid      int64
}

func (pk IncompatibleProtocol) ID() byte {
	return IDIncompatibleProtocolVersion
}

func (pk *IncompatibleProtocol) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutByte(pk.NetworkProtocol)
	if err != nil {
		return err
	}

	err = pk.PutMagic()
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ServerGuid)
	if err != nil {
		return err
	}

	return nil
}

func (pk *IncompatibleProtocol) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	err = pk.Byte(&pk.NetworkProtocol)
	if err != nil {
		return err
	}

	pk.Magic = pk.CheckMagic()

	err = pk.Long(&pk.ServerGuid)
	if err != nil {
		return err
	}

	return nil
}

func (pk *IncompatibleProtocol) New() raknet.Packet {
	return new(IncompatibleProtocol)
}

const (
	MTUPadding = 18 // id(1byte) + magic(16bytes) + protocol(1byte)
)

type OpenConnectionRequestOne struct {
	BasePacket

	Magic           bool
	ProtocolVersion byte
	MTU             int
}

func (pk OpenConnectionRequestOne) ID() byte {
	return IDOpenConnectionRequest1
}

func (pk *OpenConnectionRequestOne) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutMagic()
	if err != nil {
		return err
	}

	err = pk.PutByte(pk.ProtocolVersion)
	if err != nil {
		return err
	}

	padding := pk.MTU - MTUPadding

	err = pk.Pad(padding)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionRequestOne) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	pk.Magic = pk.CheckMagic()

	err = pk.PutByte(pk.ProtocolVersion)
	if err != nil {
		return err
	}

	pk.MTU = pk.Len() + MTUPadding

	pk.Get(pk.Len()) // reads left bytes

	return nil
}

func (pk *OpenConnectionRequestOne) New() raknet.Packet {
	return new(OpenConnectionRequestOne)
}

type OpenConnectionRequestTwo struct {
	BasePacket

	Magic      bool
	Address    raknet.SystemAddress
	MTU        uint16
	ClientGuid int64
	Connection raknet.ConnectionType
}

func (pk OpenConnectionRequestTwo) ID() byte {
	return IDOpenConnectionRequest2
}

func (pk *OpenConnectionRequestTwo) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutMagic()
	if err != nil {
		return err
	}

	err = pk.PutAddressSystemAddress(pk.Address)
	if err != nil {
		return err
	}

	err = pk.PutShort(pk.MTU)
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ClientGuid)
	if err != nil {
		return err
	}

	err = pk.PutConnectionType(pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionRequestTwo) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	pk.Magic = pk.CheckMagic()

	err = pk.AddressSystemAddress(&pk.Address)
	if err != nil {
		return err
	}

	err = pk.Short(&pk.MTU)
	if err != nil {
		return err
	}

	err = pk.Long(&pk.ClientGuid)
	if err != nil {
		return err
	}

	err = pk.ConnectionType(&pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionRequestTwo) New() raknet.Packet {
	return new(OpenConnectionRequestTwo)
}

type OpenConnectionResponseOne struct {
	BasePacket

	Magic       bool
	ServerGuid  int64
	UseSecurity bool
	MTU         uint16
}

func (pk OpenConnectionResponseOne) ID() byte {
	return IDOpenConnectionReply1
}

func (pk *OpenConnectionResponseOne) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutMagic()
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ServerGuid)
	if err != nil {
		return err
	}

	err = pk.PutBool(pk.UseSecurity)
	if err != nil {
		return err
	}

	err = pk.PutShort(pk.MTU)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionResponseOne) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	pk.Magic = pk.CheckMagic()

	err = pk.Long(&pk.ServerGuid)
	if err != nil {
		return err
	}

	err = pk.Bool(&pk.UseSecurity)
	if err != nil {
		return err
	}

	err = pk.Short(&pk.MTU)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionResponseOne) New() raknet.Packet {
	return new(OpenConnectionResponseOne)
}

type OpenConnectionResponseTwo struct {
	BasePacket

	Magic             bool
	ServerGuid        int64
	ClientAddress     raknet.SystemAddress
	MTU               uint16
	EncrtptionEnabled bool
	Connection        raknet.ConnectionType
}

func (pk OpenConnectionResponseTwo) ID() byte {
	return IDOpenConnectionReply2
}

func (pk *OpenConnectionResponseTwo) Encode() error {
	err := pk.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	err = pk.PutMagic()
	if err != nil {
		return err
	}

	err = pk.PutLong(pk.ServerGuid)
	if err != nil {
		return err
	}

	err = pk.PutAddressSystemAddress(pk.ClientAddress)
	if err != nil {
		return err
	}

	err = pk.PutShort(pk.MTU)
	if err != nil {
		return err
	}

	err = pk.PutBool(pk.EncrtptionEnabled)
	if err != nil {
		return err
	}

	err = pk.PutConnectionType(pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionResponseTwo) Decode() error {
	err := pk.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	pk.Magic = pk.CheckMagic()

	err = pk.Long(&pk.ServerGuid)
	if err != nil {
		return err
	}

	err = pk.AddressSystemAddress(&pk.ClientAddress)
	if err != nil {
		return err
	}

	err = pk.Short(&pk.MTU)
	if err != nil {
		return err
	}

	err = pk.Bool(&pk.EncrtptionEnabled)
	if err != nil {
		return err
	}

	err = pk.ConnectionType(&pk.Connection)
	if err != nil {
		return err
	}

	return nil
}

func (pk *OpenConnectionResponseTwo) New() raknet.Packet {
	return new(OpenConnectionResponseTwo)
}

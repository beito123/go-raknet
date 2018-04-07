package binary

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"github.com/beito123/binary"
	"strconv"
	"strings"
	"github.com/beito123/go-raknet"
	"github.com/satori/go.uuid"
	"errors"
)

// RaknetStream is binary stream for Raknet
type RaknetStream struct {
	binary.Stream
}

// Triad sets triad got from buffer to value
func (bs *RaknetStream) Triad(value *Triad) error {
	return Read(bs.Buffer, BigEndian, value)
}

// PutTriad puts triad from value to buffer
func (bs *RaknetStream) PutTriad(value Triad) error {
	return Write(bs.Buffer, BigEndian, value)
}

// LTriad sets triad got from buffer as LittleEndian to value
func (bs *RaknetStream) LTriad(value *Triad) error {
	return Read(bs.Buffer, LittleEndian, value)
}

// PutLTriad puts triad from value to buffer as LittleEndian
func (bs *RaknetStream) PutLTriad(value Triad) error {
	return Write(bs.Buffer, LittleEndian, value)
}

// String sets string(len short, str string) got from buffer to value
func (bs *RaknetStream) String(value *string) error {
	var n uint16
	err := bs.Short(&n)
	if err != nil {
		return err
	}

	*value = string(bs.Get(int(n)))
	return nil
}

// PutString puts string(len short, str string) to Buffer
func (bs *RaknetStream) PutString(value string) error {
	n := uint16(len(value))
	err := bs.PutShort(n)
	if err != nil {
		return err
	}
	return bs.Put([]byte(value))
}

// Address sets address got from Buffer to addr and port
// address(version byte, address byte x4, port ushort)
func (bs *RaknetStream) Address(addr *string, port *uint16) error {
	var version byte
	err := bs.Byte(&version)
	if err != nil {
		return err
	}

	var address string

	if version == 4 {
		var bytes byte
		for i := 0; i < 4; i++ {
			err = bs.Byte(&bytes)
			if err != nil {
				return err
			}

			address = address + strconv.Itoa(int(^bytes&0xff))
			if i < 3 {
				address = address + "."
			}
		}
		addr = &address

		err = bs.Short(port)
		if err != nil {
			return err
		}
	} else {
		// IPv6
	}

	return nil
}

// PutAddress puts address to Buffer
// address(version byte, address byte x4, port ushort)
func (bs *RaknetStream) PutAddress(addr string, port uint16, version byte) error {
	err := bs.PutByte(version)
	if err != nil {
		return err
	}

	if version == 4 {
		for _, str := range strings.Split(addr, ".") {
			i, _ := strconv.Atoi(str)
			err = bs.PutByte(^byte(i) & 0xff)
			if err != nil {
				return err
			}
		}
		err = bs.PutShort(port)
		if err != nil {
			return err
		}
	} else {
		// ipv6
	}

	return nil
}

// AddressSystemAddress sets address got from Buffer to SystemAddress
func (bs *RaknetStream) AddressSystemAddress(addr *raknet.SystemAddress) error {
	var add string
	var port uint16

	err := bs.Address(&add, &port)
	if err != nil {
		return err
	}

	naddr := raknet.NewSystemAddress(add, port)

	*addr = *naddr

	return nil
}

// PutAddressSystemAddress puts address from UDPAddr to Buffer
func (bs *RaknetStream) PutAddressSystemAddress(addr raknet.SystemAddress) error {
	return bs.PutAddress(addr.IP.String(), addr.Port, byte(addr.Version()))
}

// UUID reads UUID
func (bs *RaknetStream) UUID(uid *uuid.UUID) error {
	u, err := uuid.FromBytes(bs.Get(16))
	if err != nil {
		return err
	}

	*uid = u

	return nil
}

// PutUUID writes UUID
func (bs *RaknetStream) PutUUID(uid uuid.UUID) error {
	return bs.Put(uid.Bytes())
}

// ConnectionType reads ConnectionType
func (bs *RaknetStream) ConnectionType(typ *raknet.ConnectionType) error {
	var ntyp raknet.ConnectionType

	err := bs.UUID(&ntyp.UUID)
	if err != nil {
		return err
	}

	err = bs.String(&ntyp.Name)
	if err != nil {
		return err
	}

	err = bs.String(&ntyp.Lang)
	if err != nil {
		return err
	}

	err = bs.String(&ntyp.Version)
	if err != nil {
		return err
	}

	var metaLen byte
	err = bs.Byte(&metaLen)
	if err != nil {
		return err
	}

	ntyp.Metadata = raknet.Metadata{}

	var key, value string
	for i := byte(0); i < metaLen; i++ {
		err = bs.String(&key)
		if err != nil {
			return err
		}

		err = bs.String(&value)
		if err != nil {
			return err
		}

		_, ok := ntyp.Metadata[key]
		if ok { // if exists already
			return errors.New("duplicate key")
		}

		ntyp.Metadata[key] = value
	}

	return nil
}

// PutConnectionType writes ConnectionType
func (bs *RaknetStream) PutConnectionType(typ raknet.ConnectionType) error {
	err := bs.PutUUID(typ.UUID)
	if err != nil {
		return err
	}

	err = bs.PutString(typ.Name)
	if err != nil {
		return err
	}

	err = bs.PutString(typ.Lang)
	if err != nil {
		return err
	}

	err = bs.PutString(typ.Version)
	if err != nil {
		return err
	}

	if len(typ.Metadata) > raknet.MaxMetadataValues {
		return errors.New("too many metadata values")
	}

	err = bs.PutByte(byte(len(typ.Metadata)))
	if err != nil {
		return err
	}

	for k, v := range typ.Metadata {
		err = bs.PutString(k)
		if err != nil {
			return err
		}

		err = bs.PutString(v)
		if err != nil {
			return err
		}
	}

	return nil
}
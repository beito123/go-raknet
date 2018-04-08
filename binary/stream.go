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
	"errors"
	"github.com/beito123/binary"
	"github.com/beito123/go-raknet"
	"github.com/satori/go.uuid"
	"strconv"
	"strings"
)

// RaknetStream is binary stream for Raknet
type RaknetStream struct {
	binary.Stream
}

// Triad sets triad got from buffer to value
func (rs *RaknetStream) Triad(value *Triad) error {
	return Read(rs, BigEndian, value)
}

// PutTriad puts triad from value to buffer
func (rs *RaknetStream) PutTriad(value Triad) error {
	return Write(rs, BigEndian, value)
}

// LTriad sets triad got from buffer as LittleEndian to value
func (rs *RaknetStream) LTriad(value *Triad) error {
	return Read(rs, LittleEndian, value)
}

// PutLTriad puts triad from value to buffer as LittleEndian
func (rs *RaknetStream) PutLTriad(value Triad) error {
	return Write(rs, LittleEndian, value)
}

// IsMagic returns whether 16bytes is Raknet magic
func (rs *RaknetStream) IsMagic() bool {
	return string(rs.Get(len(raknet.Magic))) == string(raknet.Magic) // bad hack? :P// but fast..
}

// PutMagic write Raknet Magic
func (rs *RaknetStream) PutMagic() error {
	return rs.Put(raknet.Magic)
}

// String sets string(len short, str string) got from buffer to value
func (rs *RaknetStream) String(value *string) error {
	var n uint16
	err := rs.Short(&n)
	if err != nil {
		return err
	}

	*value = string(rs.Get(int(n)))
	return nil
}

// PutString puts string(len short, str string) to Buffer
func (rs *RaknetStream) PutString(value string) error {
	n := uint16(len(value))
	err := rs.PutShort(n)
	if err != nil {
		return err
	}
	return rs.Put([]byte(value))
}

// Address sets address got from Buffer to addr and port
// address(version byte, address byte x4, port ushort)
func (rs *RaknetStream) Address(addr *string, port *uint16) error {
	var version byte
	err := rs.Byte(&version)
	if err != nil {
		return err
	}

	var address string

	if version == 4 {
		var bytes byte
		for i := 0; i < 4; i++ {
			err = rs.Byte(&bytes)
			if err != nil {
				return err
			}

			address = address + strconv.Itoa(int(^bytes&0xff))
			if i < 3 {
				address = address + "."
			}
		}
		addr = &address

		err = rs.Short(port)
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
func (rs *RaknetStream) PutAddress(addr string, port uint16, version byte) error {
	err := rs.PutByte(version)
	if err != nil {
		return err
	}

	if version == 4 {
		for _, str := range strings.Split(addr, ".") {
			i, _ := strconv.Atoi(str)
			err = rs.PutByte(^byte(i) & 0xff)
			if err != nil {
				return err
			}
		}
		err = rs.PutShort(port)
		if err != nil {
			return err
		}
	} else {
		// ipv6
	}

	return nil
}

// AddressSystemAddress sets address got from Buffer to SystemAddress
func (rs *RaknetStream) AddressSystemAddress(addr *raknet.SystemAddress) error {
	var add string
	var port uint16

	err := rs.Address(&add, &port)
	if err != nil {
		return err
	}

	naddr := raknet.NewSystemAddress(add, port)

	*addr = *naddr

	return nil
}

// PutAddressSystemAddress puts address from UDPAddr to Buffer
func (rs *RaknetStream) PutAddressSystemAddress(addr raknet.SystemAddress) error {
	return rs.PutAddress(addr.IP.String(), addr.Port, byte(addr.Version()))
}

// UUID reads UUID
func (rs *RaknetStream) UUID(uid *uuid.UUID) error {
	u, err := uuid.FromBytes(rs.Get(16))
	if err != nil {
		return err
	}

	*uid = u

	return nil
}

// PutUUID writes UUID
func (rs *RaknetStream) PutUUID(uid uuid.UUID) error {
	return rs.Put(uid.Bytes())
}

// ConnectionType reads ConnectionType
func (rs *RaknetStream) ConnectionType(typ *raknet.ConnectionType) error {
	var ntyp raknet.ConnectionType

	err := rs.UUID(&ntyp.UUID)
	if err != nil {
		return err
	}

	err = rs.String(&ntyp.Name)
	if err != nil {
		return err
	}

	err = rs.String(&ntyp.Lang)
	if err != nil {
		return err
	}

	err = rs.String(&ntyp.Version)
	if err != nil {
		return err
	}

	var metaLen byte
	err = rs.Byte(&metaLen)
	if err != nil {
		return err
	}

	ntyp.Metadata = raknet.Metadata{}

	var key, value string
	for i := byte(0); i < metaLen; i++ {
		err = rs.String(&key)
		if err != nil {
			return err
		}

		err = rs.String(&value)
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
func (rs *RaknetStream) PutConnectionType(typ raknet.ConnectionType) error {
	err := rs.PutUUID(typ.UUID)
	if err != nil {
		return err
	}

	err = rs.PutString(typ.Name)
	if err != nil {
		return err
	}

	err = rs.PutString(typ.Lang)
	if err != nil {
		return err
	}

	err = rs.PutString(typ.Version)
	if err != nil {
		return err
	}

	if len(typ.Metadata) > raknet.MaxMetadataValues {
		return errors.New("too many metadata values")
	}

	err = rs.PutByte(byte(len(typ.Metadata)))
	if err != nil {
		return err
	}

	for k, v := range typ.Metadata {
		err = rs.PutString(k)
		if err != nil {
			return err
		}

		err = rs.PutString(v)
		if err != nil {
			return err
		}
	}

	return nil
}

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
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/beito123/binary"
	"github.com/beito123/go-raknet"
	"github.com/satori/go.uuid"
)

func NewStream() *RaknetStream {
	return NewStreamBytes([]byte{})
}

func NewStreamBytes(b []byte) *RaknetStream {
	return &RaknetStream{
		Stream: *binary.NewStreamBytes(b),
	}
}

// RaknetStream is binary stream for Raknet
type RaknetStream struct {
	binary.Stream
}

// Triad sets triad got from buffer to value
func (rs *RaknetStream) Triad() (Triad, error) {
	return ReadETriad(rs.Get(TriadSize))
}

// PutTriad puts triad from value to buffer
func (rs *RaknetStream) PutTriad(value Triad) error {
	return Write(rs, BigEndian, value)
}

// LTriad sets triad got from buffer as LittleEndian to value
func (rs *RaknetStream) LTriad() (Triad, error) {
	return ReadELTriad(rs.Get(TriadSize))
}

// PutLTriad puts triad from value to buffer as LittleEndian
func (rs *RaknetStream) PutLTriad(value Triad) error {
	return Write(rs, LittleEndian, value)
}

// CheckMagic returns whether 16bytes is Raknet magic
func (rs *RaknetStream) CheckMagic() bool {
	return bytes.Equal(rs.Get(len(raknet.Magic)), raknet.Magic)
}

// PutMagic write Raknet Magic
func (rs *RaknetStream) PutMagic() error {
	return rs.Put(raknet.Magic)
}

// String gets string(len short, str string) from the buffer
func (rs *RaknetStream) String() (string, error) {
	n, err := rs.Short()
	if err != nil {
		return "", err
	}

	return string(rs.Get(int(n))), nil
}

// PutString puts string(len short, str string) to the buffer
func (rs *RaknetStream) PutString(value string) error {
	b := []byte(value)

	err := rs.PutShort(uint16(len(b)))
	if err != nil {
		return err
	}

	return rs.Put(b)
}

// Address sets address got from Buffer to addr and port
// address(version byte, address byte x4, port ushort)
func (rs *RaknetStream) Address() (addr string, port uint16, err error) {
	ver, err := rs.Byte()
	if err != nil {
		return "", 0, err
	}

	if ver == 4 {
		for i := 0; i < 4; i++ {
			bytes, err := rs.Byte()
			if err != nil {
				return "", 0, err
			}

			addr := addr + strconv.Itoa(int(^bytes&0xff))
			if i < 3 {
				addr += "."
			}
		}

		port, err = rs.Short()
		if err != nil {
			return "", 0, err
		}

		return addr, port, nil
	} else {
		// IPv6
	}

	return "", 0, nil
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
func (rs *RaknetStream) AddressSystemAddress() (*raknet.SystemAddress, error) {
	addr, port, err := rs.Address()
	if err != nil {
		return nil, err
	}

	naddr := raknet.NewSystemAddress(addr, port)

	return naddr, nil
}

// PutAddressSystemAddress puts address from UDPAddr to Buffer
func (rs *RaknetStream) PutAddressSystemAddress(addr *raknet.SystemAddress) error {
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
func (rs *RaknetStream) ConnectionType() (*raknet.ConnectionType, error) {
	var ntyp raknet.ConnectionType

	err := rs.UUID(&ntyp.UUID)
	if err != nil {
		return nil, err
	}

	ntyp.Name, err = rs.String()
	if err != nil {
		return nil, err
	}

	ntyp.Lang, err = rs.String()
	if err != nil {
		return nil, err
	}

	ntyp.Version, err = rs.String()
	if err != nil {
		return nil, err
	}

	metaLen, err := rs.Byte()
	if err != nil {
		return nil, err
	}

	ntyp.Metadata = raknet.Metadata{}

	for i := byte(0); i < metaLen; i++ {
		key, err := rs.String()
		if err != nil {
			return nil, err
		}

		value, err := rs.String()
		if err != nil {
			return nil, err
		}

		_, ok := ntyp.Metadata[key]
		if ok { // if exists already
			return nil, errors.New("duplicate key")
		}

		ntyp.Metadata[key] = value
	}

	return nil, nil
}

// PutConnectionType writes ConnectionType
func (rs *RaknetStream) PutConnectionType(typ *raknet.ConnectionType) error {
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

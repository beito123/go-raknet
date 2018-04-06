package raknet

import (
	"net"
	"strconv"
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

	// Version is version of go-raknet library
	Version = "1.0.0"

	// ProtocolVersion is version of Raknet protocol
	ProtocolVersion = 8
)

// SystemAddress is internal address for Raknet
type SystemAddress struct {
	IP   net.IP
	Port uint16
}

// SetLoopback sets loopback address
func (addr *SystemAddress) SetLoopback() {
	if len(addr.IP) == net.IPv4len {
		addr.IP = net.ParseIP("127.0.0.1")
	} else {
		addr.IP = net.IPv6loopback // "::1"
	}
}

// IsLoopback returns whether this is loopback address
func (addr *SystemAddress) IsLoopback() bool {
	return addr.IP.IsLoopback()
}

// Version returns the ip address version (4 or 6)
func (addr *SystemAddress) Version() int {
	if len(addr.IP) == net.IPv6len {
		return 6
	}

	return 4
}

// Equal returns whether sub is the same address
func (addr *SystemAddress) Equal(sub *SystemAddress) bool {
	return addr.IP.Equal(sub.IP) && addr.Port == sub.Port
}

// String returns as string
// Format: 192.168.11.1:8080, [fc00::]:8080
func (addr *SystemAddress) String() string {
	if len(addr.IP) == net.IPv6len {
		return "[" + addr.IP.String() + "]:" + strconv.Itoa(int(addr.Port))
	}

	return addr.IP.String() + ":" + strconv.Itoa(int(addr.Port))
}

// NewSystemAddress returns a new SystemAddress from string
func NewSystemAddress(addr string, port uint16) *SystemAddress {
	return &SystemAddress{
		IP:   net.ParseIP(addr),
		Port: port,
	}
}

// NewSystemAddress returns a new SystemAddress from bytes
func NewSystemAddressBytes(addr []byte, port uint16) *SystemAddress {
	return &SystemAddress{
		IP:   net.IP(addr).To16(),
		Port: port,
	}
}
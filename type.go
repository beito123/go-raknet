package raknet

import (
	"github.com/satori/go.uuid"
	"net"
	"strconv"
)

/*
	RaknetProtocol Interface
*/

// Packet is basic packet interface
type Packet interface {
	ID() byte
	Encode() error
	Decode() error
	New() Packet
}

// Protocol is packet protocol interface
type Protocol interface {
	RegisterPackets()
	Packet(id byte) Packet
	Packets() []Packet
}

/*
	ConnectionType
*/

const MaxMetadataValues = 0xff

// Metadata is metadata for ConnectionType
type Metadata map[string]string

// ConnectionTypeMagic is the magic for ConnectionType.
var ConnctionTypeMagic = []byte{0x03, 0x08, 0x05, 0x0b, 0x43, 0x54, 0x49}

// ConnectionVanilla is a connection from a vanilla client or an unknown implementation
var ConnectionVanilla = ConnectionType{
	Name:      "Vanilla",
	IsVanilla: true,
}

// ConnectionGoRaknet is a go-raknet connection.
var ConnectionGoRaknet = ConnectionType{
	UUID:    uuid.FromStringOrNil("495248b9-d485-4389-acd0-175fdb2233cf"),
	Name:    "GoRaknet",
	Lang:    "Go",
	Version: "2.9.2",
}

// ConnectionType is connection info struct
type ConnectionType struct {
	UUID      uuid.UUID
	Name      string
	Lang      string
	Version   string
	Metadata  Metadata
	IsVanilla bool
}

// Identifier represents an identifier sent from a server on the local network
type Identifier struct {
	Identifier     string
	ConnectionType ConnectionType
}

func (id *Identifier) Build() string {
	return id.Identifier
}

/*
	SystemAddress
*/

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

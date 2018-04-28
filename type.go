package raknet

import (
	"net"
	"strconv"

	"github.com/satori/go.uuid"
)

/*
	Logger
*/

// Logger is a supported logger in Raknet server
type Logger interface {
	Info(msg ...interface{})
	Warn(msg ...interface{})
	Fatal(msg ...interface{})
	Debug(msg ...interface{})
}

/*
	RaknetProtocol Interface
*/

// Packet is basic packet interface
type Packet interface {
	ID() byte
	Encode() error
	Decode() error
	Bytes() []byte
	SetBytes([]byte)
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

/*
	Reliability
*/

// Reliability decides reliable and ordered of packet when sending
// Thanks: http://www.jenkinssoftware.com/raknet/manual/Doxygen/PacketPriority_8h.html#e41fa01235e99dced384d137fa874a7e
type Reliability int

const (
	// Unreliable is normal UDP packet.
	Unreliable Reliability = iota

	// UnreliableSequenced is the same as Unreliable. but it has Sequenced.
	UnreliableSequenced

	Reliable
	ReliableOrdered
	ReliableSequenced
	UnreliableWithACKReceipt
	ReliableWithACKReceipt
	ReliableOrderedWithACKReceipt
)

// IsReliable returns whether reliability has reliable
func (r Reliability) IsReliable() bool {
	return (r & 0x04) > 0
}

// IsOrdered returns whether reliability has ordered
func (r Reliability) IsOrdered() bool {
	return (r & 0x02) > 0
}

// IsSequenced returns whether reliability has sequenced
func (r Reliability) IsSequenced() bool {
	return (r & 0x01) > 0
}

// IsNeededACK returns whether reliability need ack
func (r Reliability) IsNeededACK() bool {
	return r == UnreliableWithACKReceipt ||
		r == ReliableWithACKReceipt ||
		r == ReliableOrderedWithACKReceipt
}

// ToBinary encode reliability to bytes
func (r Reliability) ToBinary() byte {
	var b byte

	if r.IsReliable() {
		b |= 1 << 2
	}

	if r.IsOrdered() {
		b |= 1 << 1
	}

	if r.IsSequenced() {
		b |= 1
	}

	return b
}

// ReliabilityBinary returns Reliability from binary
func ReliabilityBinary(b byte) Reliability {
	if (b & 0x04) > 0 { // Reliable
		if (b & 0x02) > 0 { // Ordered
			return ReliableOrdered
		} else if (b & 0x01) > 0 { // Sequenced
			return ReliableSequenced
		} else {
			return Reliable
		}
	} else { // Unreliable
		if (b & 0x01) > 0 { // Sequenced
			return UnreliableSequenced
		}
	}

	return Unreliable
}

/*
	Records
*/

// Record is ack numbers container for acknowledge packet
type Record struct {
	Index    int
	EndIndex int
}

// IsRanged returns whether the record is range
func (rec *Record) IsRanged() bool {
	return rec.Index < rec.EndIndex
}

// Count returns the number of Record
func (rec *Record) Count() int {
	if !rec.IsRanged() {
		return 1
	}

	return (rec.EndIndex - rec.Index) + 1
}

// Numbers returns numbers recorded in Record as array
func (rec *Record) Numbers() []int {
	count := rec.Count()
	if count == 1 {
		return []int{rec.Index}
	}

	numbers := make([]int, count)
	for i := 0; i < count; i++ {
		numbers[i] = rec.Index + i
	}

	return numbers
}

type SessionState int

const (
	StateDisconected SessionState = iota
	StateHandshaking
	StateConnected
)
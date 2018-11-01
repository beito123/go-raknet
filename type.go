package raknet

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"net"
	"strconv"
	"time"

	"github.com/satori/go.uuid"
)

/*
	ConnectionType
*/

const MaxMetadataValues = 0xff

// Metadata is metadata for ConnectionType
type Metadata map[string]string

// ConnectionTypeMagic is the magic for ConnectionType.
var ConnctionTypeMagic = []byte{0x03, 0x08, 0x05, 0x0b, 0x43, 0x54, 0x49}

// ConnectionVanilla is a connection from a vanilla client or an unknown implementation
var ConnectionVanilla = &ConnectionType{
	Name:      "Vanilla",
	IsVanilla: true,
}

// ConnectionGoRaknet is a connection from go-raknet.
var ConnectionGoRaknet = &ConnectionType{
	UUID:    uuid.FromStringOrNil("495248b9-d485-4389-acd0-175fdb2233cf"),
	Name:    "GoRaknet",
	Lang:    "Go",
	Version: "1.0.0",
}

// ConnectionType used to signify which implementation of the Raknet protocol
type ConnectionType struct {
	UUID      uuid.UUID
	Name      string
	Lang      string
	Version   string
	Metadata  Metadata
	IsVanilla bool
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

// Equals returns whether the record equals record sub
func (record *Record) Equals(sub *Record) bool {
	return record.Index == sub.Index && record.EndIndex == sub.EndIndex
}

// IsRanged returns whether the record is range
func (record *Record) IsRanged() bool {
	return record.Index < record.EndIndex
}

// Count returns the number of Record
func (record *Record) Count() int {
	if !record.IsRanged() {
		return 1
	}

	return (record.EndIndex - record.Index) + 1
}

// Numbers returns numbers recorded in Record as array
func (record *Record) Numbers() []int {
	count := record.Count()
	if count == 1 {
		return []int{record.Index}
	}

	numbers := make([]int, count)
	for i := 0; i < count; i++ {
		numbers[i] = record.Index + i
	}

	return numbers
}

/*
	Latency
*/

type Latency struct {
	// totalLatency is the total latency time
	TotalLatency time.Duration

	// latency is the average latency time
	Latency time.Duration

	// lastLatency is the last latency time
	LastLatency time.Duration

	// lowestLatency is the lowest latency time
	LowestLatency time.Duration

	// highestLatency is the highest latency time
	HighestLatency time.Duration

	counter int
}

func (lat *Latency) AddRaw(raw time.Duration) {
	lat.LastLatency = raw

	if lat.counter == 0 { //first
		lat.LowestLatency = raw
		lat.HighestLatency = raw
	} else {
		if raw < lat.LowestLatency {
			lat.LowestLatency = raw
		} else if raw < lat.HighestLatency {
			lat.HighestLatency = raw
		}
	}

	lat.counter++

	lat.TotalLatency += raw
	lat.Latency = time.Duration(int64(lat.TotalLatency) / int64(lat.counter))
}

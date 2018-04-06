package pk

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */


// Ref: http://www.jenkinssoftware.com/raknet/manual/Doxygen/PacketPriority_8h.html#e41fa01235e99dced384d137fa874a7e

// Reliability decides reliable and ordered of packet when sending
type Reliability int

const (
	// Unreliable is normal UDP packet.
	Unreliable                    Reliability = iota

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
	return r == Reliable || r == ReliableOrdered ||
		r == ReliableSequenced || r == ReliableWithACKReceipt ||
		r == ReliableOrderedWithACKReceipt
}

// IsOrdered returns whether reliability has ordered
func (r Reliability) IsOrdered() bool {
	return r == UnreliableSequenced || r == ReliableOrdered ||
		r == ReliableSequenced || r == ReliableOrderedWithACKReceipt
}

// IsSequenced returns whether reliability has sequenced
func (r Reliability) IsSequenced() bool {
	return r == UnreliableSequenced || r == ReliableSequenced
}

// IsNeededACK returns whether reliability need ack
func (r Reliability) IsNeededACK() bool {
	return r == UnreliableWithACKReceipt || r == ReliableWithACKReceipt ||
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


package raknet

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

/*
	Logger
*/

// Logger is a logger interface supported
type Logger interface {
	Info(msg ...interface{})
	Warn(msg ...interface{})
	Fatal(msg ...interface{})
	Debug(msg ...interface{})
}

/*
	Raknet's protocol interfaces
*/

// Packet is a basic packet interface
type Packet interface {

	// ID returns a packet id
	ID() byte

	// Encode encodes the packet
	Encode() error

	// Decode decodes the packet
	Decode() error

	// Bytes returns encoded bytes
	Bytes() []byte

	// SetBytes sets bytes to decode
	SetBytes([]byte)

	// New returns new packet
	New() Packet
}

// Protocol is Raknet protocol's packets manager
type Protocol interface {

	// RegisterPackets registers packets
	// It's called in starting the server
	RegisterPackets()

	// Packet returns a packet with id
	Packet(id byte) Packet

	//Packets returns all registered packets
	Packets() []Packet
}

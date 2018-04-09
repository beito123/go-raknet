package raknet

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

// Magic is Raknet offline message data id
// using offline connection in Raknet
var Magic = []byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

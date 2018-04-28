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

	// NetworkProtocol is a version of Raknet protocol
	NetworkProtocol = 8

	// MaxMTU is the maximum size of MTU
	MaxMTU = 1492

	// MinMTU is the minimum size of MTU
	MinMTU = 400

	// MaxChannel is the maximum size of Channel
	MaxChannel = 32

	// DefaultChannel is default channel
	DefaultChannel = 0

	// MaxSplitCount is the maximum size that can split
	MaxSplitCount = 128

	// MaxSplitsPerQueue is the maximum size of Queue
	MaxSplitsPerQueue = 4
)

// Magic is Raknet offline message data id
// using offline connection in Raknet
var Magic = []byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

// MaxPacketsPerSecond is the maximum size that can send per second
var MaxPacketsPerSecond = 500

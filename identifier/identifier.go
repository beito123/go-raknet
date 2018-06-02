package identifier

import raknet "github.com/beito123/go-raknet"

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

// Identifier ...
type Identifier interface {
	ConnectionType() *raknet.ConnectionType
	Build() string
}

// Base represents an identifier sent from a server on the local network
type Base struct {
	Identifier string
	Connection *raknet.ConnectionType
}

func (id Base) ConnectionType() *raknet.ConnectionType {
	return id.Connection
}

func (id Base) Build() string {
	return id.Identifier
}

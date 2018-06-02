package identifier

import (
	"strconv"

	"github.com/beito123/go-raknet"
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
	MinecraftHeader      = "MCPE"
	MinecraftSeparator   = ";"
	MinecraftCountLegacy = 6
	MinecraftCount       = 9
)

//var MinecraftVersionTagAlphabet = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "."}

type Minecraft struct {
	Connection *raknet.ConnectionType

	ServerName        string
	ServerProtocol    int
	VersionTag        string
	OnlinePlayerCount int
	MaxPlayerCount    int
	GUID              int64
	WorldName         string
	Gamemode          string
	Legacy            bool
}

func (id Minecraft) ConnectionType() *raknet.ConnectionType {
	return id.Connection
}

func (id Minecraft) Build() string {
	if id.Legacy {
		return MinecraftHeader + MinecraftSeparator +
			id.ServerName + MinecraftSeparator +
			strconv.Itoa(id.ServerProtocol) + MinecraftSeparator +
			id.VersionTag + MinecraftSeparator +
			strconv.Itoa(id.OnlinePlayerCount) + MinecraftSeparator +
			strconv.Itoa(id.MaxPlayerCount)
	}

	return MinecraftHeader + MinecraftSeparator +
		id.ServerName + MinecraftSeparator +
		strconv.Itoa(id.ServerProtocol) + MinecraftSeparator +
		id.VersionTag + MinecraftSeparator +
		strconv.Itoa(id.OnlinePlayerCount) + MinecraftSeparator +
		strconv.Itoa(id.MaxPlayerCount) + MinecraftSeparator +
		strconv.FormatInt(id.GUID, 10) + MinecraftSeparator +
		id.WorldName + MinecraftSeparator +
		id.Gamemode
}

package main

import (
	"bufio"
	"context"
	"net"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/beito123/binary"
	"github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/identifier"
	"github.com/satori/go.uuid"

	"github.com/beito123/go-raknet/server"
)

func main() {
	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.TextFormatter{},
		Level:     logrus.DebugLevel,
	}

	uid, _ := uuid.NewV4()

	id := identifier.Minecraft{
		Connection:        &raknet.ConnectionGoRaknet,
		ServerName:        "Go-raknet server",
		ServerProtocol:    raknet.NetworkProtocol,
		VersionTag:        "1.0.0",
		OnlinePlayerCount: 0,
		MaxPlayerCount:    10,
		GUID:              binary.ReadLong(uid.Bytes()[0:8]),
		WorldName:         "world",
		Gamemode:          "0",
		Legacy:            false,
	}

	ser := &server.Server{
		Logger:              logger,
		MaxConnections:      10,
		MTU:                 1472,
		Identifier:          id,
		UUID:                uid,
		BroadcastingEnabled: true,
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("Starting the server...")

	addr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 19133,
	}

	go ser.ListenAndServe(ctx, addr)

	logger.Info("Enter to stop the server")

	bufio.NewScanner(os.Stdin).Scan() // wait input anything

	cancel()

	logger.Info("Stopping the server...")
}

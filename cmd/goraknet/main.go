package main

import (
	"bufio"
	"context"
	"flag"
	"net"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/beito123/binary"
	"github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/identifier"
	"github.com/satori/go.uuid"

	"github.com/beito123/go-raknet/server"
	"github.com/mattn/go-colorable"
)

func main() {
	var (
		port          int
		maxConnection int
		monitor       string
		help          bool
	)

	// goraknet -port <server port> -maxconnections <max connections> -monitor <ip addr>
	flag.IntVar(&port, "port", 19132, "a server port")
	flag.IntVar(&maxConnection, "maxconnections", 10, "max connections")
	flag.StringVar(&monitor, "monitor", "", "monitor ip")
	flag.BoolVar(&help, "help", false, "help")
	flag.Parse()

	logger := &logrus.Logger{
		Out: colorable.NewColorableStdout(),
		Formatter: &logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     true,
		},
		Level: logrus.DebugLevel,
	}

	if help {
		logger.Info("Usage: goraknet -port <a server port> -maxconnections <max connections> -monitor <ip addr>")
		logger.Info("-port: a server port (default: 19132)")
		logger.Info("-maxconnections: max connections of server (default: 15)")
		logger.Info("-monitor: monitor a IP address (example: 192.168.xx.xx)")
		return
	}

	if port < 0 || port > 65535 {
		logger.Errorln("invaild a port, wants 0-65535")
		return
	}

	uid, _ := uuid.NewV4()

	// For MCBE server
	id := identifier.Minecraft{
		Connection:        &raknet.ConnectionGoRaknet,
		ServerName:        "Go-Raknet server",
		ServerProtocol:    raknet.NetworkProtocol,
		VersionTag:        "1.0.0",
		OnlinePlayerCount: 0,
		MaxPlayerCount:    maxConnection,
		GUID:              binary.ReadLong(uid.Bytes()[0:8]),
		WorldName:         "world",
		Gamemode:          "0",
		Legacy:            false,
	}

	ser := &server.Server{
		Logger:              logger,
		MaxConnections:      maxConnection,
		MTU:                 1472,
		Identifier:          id,
		UUID:                uid,
		BroadcastingEnabled: true,
	}

	if len(monitor) > 0 {
		ip := net.ParseIP(monitor)
		if ip == nil {
			logger.Fatal("Failed to parse the monitor IP address")
			return
		}

		ser.Handler = &MonitorHandler{
			MonitorIP: ip,
			Path:      "./",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("Starting the server...")
	logger.Debug("ip: 0.0.0.0, port:" + strconv.Itoa(port))

	addr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}

	go ser.ListenAndServe(ctx, addr)

	logger.Info("Enter to stop the server")

	bufio.NewScanner(os.Stdin).Scan() // wait input anything

	cancel()

	logger.Info("Stopping the server...")
}

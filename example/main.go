package main

import (
	"bufio"
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

	//Command: goraknet -p <server port> -m <max connections> -M <ip addr>
	flag.IntVar(&port, "p", 19132, "a server port")
	flag.IntVar(&maxConnection, "m", 10, "max connections")
	flag.StringVar(&monitor, "M", "", "monitor ip")
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
		logger.Info("Usage: goraknet -p <a server port> -m <max connections> -M <ip addr>")
		logger.Info("-p: set server's port. (default: 19132)")
		logger.Info("-m: set server's max connections. (default: 10)")
		logger.Info("-M: set server's IP address.")
		return
	}

	if port < 0 || port > 65535 {
		logger.Errorln("invaild a port, it must be between 0 and 65535")
		return
	}

	uid, _ := uuid.NewV4()

	// For MCBE server
	id := identifier.Minecraft{
		Connection:        &raknet.ConnectionGoRaknet,
		ServerName:        "Go-Raknet server",
		ServerProtocol:    280,
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
			logger.Fatal("Failed to parse the monitor's IP address")
			return
		}

		ser.Handlers = append(ser.Handlers, &MonitorHandler{
			MonitorIP: ip,
			Path:      "./monitor.txt",
		})
	}

	logger.Info("Starting the server...")
	logger.Debug("address: 0.0.0.0:" + strconv.Itoa(port))

	go ser.Start("0.0.0.0", port)

	logger.Info("Enter to stop the server")

	bufio.NewScanner(os.Stdin).Scan() // wait until any command is executed.

	logger.Info("Stopping the server...")

	ser.Shutdown()
}

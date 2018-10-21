package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"

	raknet "github.com/beito123/go-raknet"
)

// MonitorHandler is a simple monitor handler
type MonitorHandler struct {
	MonitorIP net.IP
	Path      string
	out       chan string
	stopped   chan bool
	number    int
	targets   map[int64]net.Addr
}

// StartServer is called on the server is started
func (hand *MonitorHandler) StartServer() {
	file, err := os.OpenFile(hand.Path, os.O_CREATE|os.O_WRONLY /*|os.O_APPEND*/, 0666)
	if err != nil {
		panic("Couldn't create a file")
	}

	file.WriteString("## Monitor Target IP: " + hand.MonitorIP.String() + " Time: " + time.Now().Format(time.UnixDate) + "\n\n")

	hand.stopped = make(chan bool)

	hand.out = make(chan string, 10)

	go func() {
		defer file.Close()

		for {
			select {
			case text := <-hand.out:
				_, err := file.WriteString(text)
				if err != nil {
					panic(err)
				}
			case <-hand.stopped:
				return
			}
		}
	}()
}

// CloseServer is called on the server is closed
func (hand *MonitorHandler) CloseServer() {
	<-hand.stopped
}

// HandlePing is called on a ping packet is received
func (hand *MonitorHandler) HandlePing(addr net.Addr) {
	if hand.IsTargetAddr(addr) {
		hand.out <- "# Received a ping packet from the monitor target (" + addr.String() + ")\n\n"
	}
}

// OpenedPreConn is called before a new session is created
func (MonitorHandler) OpenedPreConn(addr net.Addr) {
}

// OpenConn is called on a new session is created
func (hand *MonitorHandler) OpenedConn(uid int64, addr net.Addr) {
	if hand.IsTargetAddr(addr) {
		hand.out <- "# Connected the monitor target from " + addr.String() + "\n\n"

		hand.targets[uid] = addr
	}
}

// ClosePreConn is called before a session is closed
func (MonitorHandler) ClosedPreConn(uid int64) {
}

// CloseConn is called on a session is closed
func (hand *MonitorHandler) ClosedConn(uid int64) {
	if hand.IsTarget(uid) {
		hand.out <- "# Disconnected the monitor target connected from " + hand.targets[uid].String() + "\n\n"

		delete(hand.targets, uid)
	}
}

// Timeout is called when a client is timed out
func (hand *MonitorHandler) Timedout(uid int64) {
	if hand.IsTarget(uid) {
		hand.out <- "# The session timed out"
	}
}

// BlockedAddress is called when a client is added blocked address
func (hand *MonitorHandler) AddedBlockedAddress(ip net.IP, reason string) {
	if hand.IsTargetIP(ip) {
		hand.out <- "# Added the target ip to blocked address"
	}
}

// BlockedAddress is called when a client is removed blocked address
func (hand *MonitorHandler) RemovedBlockedAddress(ip net.IP) {
	if hand.IsTargetIP(ip) {
		hand.out <- "# Removed the target ip from blocked address"
	}
}

func (hand *MonitorHandler) HandleSendPacket(addr net.Addr, pk raknet.Packet) {
	if hand.IsTargetAddr(addr) {
		hand.out <- "HandleSendPacket: \n"
		hand.out <- hand.dump("Server", "Client", getPacketName(pk), pk.Bytes())
	}
}

// HandleRawPacket handles a raw packet no processed in Raknet server
func (hand *MonitorHandler) HandleRawPacket(addr net.Addr, pk raknet.Packet) {
	if hand.IsTargetAddr(addr) {
		hand.out <- "HandleRawPacket: \n"
		hand.out <- hand.dump("Client", "Server", getPacketName(pk), pk.Bytes())
	}
}

// HandlePacket handles a message packet
func (hand *MonitorHandler) HandlePacket(uid int64, pk raknet.Packet) {
	if hand.IsTarget(uid) {
		hand.out <- "HandlePacket: \n"
		hand.out <- hand.dump("Client", "Server", getPacketName(pk), pk.Bytes())
	}
}

// HandleUnknownPacket handles a unknown packet
func (hand *MonitorHandler) HandleUnknownPacket(uid int64, pk raknet.Packet) {
	if hand.IsTarget(uid) {
		hand.out <- "HandleUnknownPacket: \n"
		hand.out <- hand.dump("Client", "Server", getPacketName(pk), pk.Bytes())
	}
}

func (hand *MonitorHandler) IsTarget(uid int64) bool {
	_, ok := hand.targets[uid]
	return ok
}

func (hand *MonitorHandler) IsTargetAddr(addr net.Addr) bool {
	naddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return false
	}

	return hand.IsTargetIP(naddr.IP)
}

func (hand *MonitorHandler) IsTargetIP(ip net.IP) bool {
	return hand.MonitorIP.Equal(ip)
}

func getPacketName(pk raknet.Packet) string {
	return fmt.Sprintf("%s (0x%x)", getTypeName(pk), pk.ID())
}

// Thank you : https://stackoverflow.com/questions/35790935/using-reflection-in-go-to-get-the-name-of-a-struct
// From Stackoverflow author: icza(https://stackoverflow.com/users/1705598/icza), questioner: Daniele D
func getTypeName(v interface{}) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func (hand *MonitorHandler) dump(from string, to string, name string, bytes []byte) string {
	// #1 [Client -> Server] UnknownPacket (0xff)
	hand.number++
	return fmt.Sprintf("#%d [%s -> %s] %s \n%s\n\n", hand.number, from, to, name, hex.Dump(bytes))
}

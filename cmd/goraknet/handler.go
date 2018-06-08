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
	path := hand.Path + "/monitor.txt"

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY /*|os.O_APPEND*/, 0666)
	if err != nil {
		panic("Couldn't create a file")
	}

	file.WriteString("# Monitor TargetIP:" + hand.MonitorIP.String() + " #" + time.Now().Format(time.UnixDate) + "\n\n")

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

// OpenPreConn is called before a new session is created
func (MonitorHandler) OpenPreConn(addr net.Addr) {
}

// OpenConn is called on a new session is created
func (hand *MonitorHandler) OpenConn(uid int64, addr net.Addr) {
	if hand.IsTargetAddr(addr) {
		hand.out <- "# Connected the monitor target from " + addr.String() + "\n\n"

		hand.targets[uid] = addr
	}
}

// ClosePreConn is called before a session is closed
func (MonitorHandler) ClosePreConn(uid int64) {
}

// CloseConn is called on a session is closed
func (hand *MonitorHandler) CloseConn(uid int64) {
	if hand.IsTarget(uid) {
		hand.out <- "# Disconnected the monitor target connected from " + hand.targets[uid].String() + "\n\n"

		delete(hand.targets, uid)
	}
}

func (hand *MonitorHandler) HandleSendPacket(addr net.Addr, pk raknet.Packet) {
	if hand.IsTargetAddr(addr) {
		hand.out <- hand.dump("Server", "Client", getPacketName(pk), pk.Bytes())
	}
}

// HandleRawPacket handles a raw packet no processed in Raknet server
func (hand *MonitorHandler) HandleRawPacket(addr net.Addr, pk raknet.Packet) {
	if hand.IsTargetAddr(addr) {
		hand.out <- hand.dump("Client", "Server", getPacketName(pk), pk.Bytes())
	}
}

// HandlePacket handles a message packet
func (hand *MonitorHandler) HandlePacket(uid int64, pk raknet.Packet) {
	hand.out <- hand.dump("Client", "Server", getPacketName(pk), pk.Bytes())
}

// HandleUnknownPacket handles a unknown packet
func (hand *MonitorHandler) HandleUnknownPacket(uid int64, pk raknet.Packet) {
	hand.out <- hand.dump("Client", "Server", getPacketName(pk), pk.Bytes())
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

	return hand.MonitorIP.Equal(naddr.IP)
}

func getPacketName(pk raknet.Packet) string {
	return fmt.Sprintf("%s (0x%x)", getTypeName(pk), pk.ID())
}

// Thank you : https://stackoverflow.com/questions/35790935/using-reflection-in-go-to-get-the-name-of-a-struct
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

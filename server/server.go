package server

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/beito123/go-raknet/identifier"

	"github.com/beito123/binary"

	raknet "github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/protocol"
	"github.com/orcaman/concurrent-map"
	"github.com/satori/go.uuid"
)

// ServerState is a server state
type ServerState int

const (
	StateNew ServerState = iota
	StateRunning
	StateClosed
)

var (
	errAlreadyRunning           = errors.New("server has been already running")
	errServerClosed             = errors.New("server closed")
	errInvalidMaxAsyncTaskCount = errors.New("invalid max async task count")
)

type Handlers []Handler

type Server struct {
	Logger         utils.Logger
	Handlers        Handlers
	MaxConnections int
	MTU            int
	Identifier     identifier.Minecraft
	protocol       *protocol.Protocol

	cancel         context.CancelFunc

	UUID uuid.UUID

	// BroadcastingEnabled broadcast the server for the outside
	// if it enabled, the server send UnconnectedPong when received UnconnectPing.
	BroadcastingEnabled bool

	conn   *net.UDPConn
	port   uint16
	state  ServerState
	uid    int64
	pongid int64

	sessions         cmap.ConcurrentMap
	blockedAddresses cmap.ConcurrentMap
}

func (s *Server) Cancel() context.CancelFunc {
	return s.cancel
}

func (s *Server) SetCancel(cancel context.CancelFunc) {
	s.cancel = cancel
}

func (s *Server) Shutdown() {
	if s.IsRunning() {
		f := s.Cancel()
		f()
	}
}

func (ser *Server) State() ServerState {
	return ser.state
}

func (ser *Server) IsRunning() bool {
	return ser.state == StateRunning
}

func (ser *Server) IsClosed() bool {
	return ser.state == StateClosed
}

func (ser *Server) init() {
	// init maps
	ser.sessions = cmap.New()
	ser.blockedAddresses = cmap.New()

	// readly protocols
	ser.protocol = new(protocol.Protocol)
	ser.protocol.RegisterPackets()

	ser.uid = binary.ReadLong(ser.UUID.Bytes()[:8])
	ser.pongid = binary.ReadLong(ser.UUID.Bytes()[8:16])

	//if no set, set default
	if ser.MTU < raknet.MinMTU || ser.MTU > raknet.MaxMTU {
		ser.MTU = raknet.MaxMTU
	}
}

func (ser *Server) Start(ip string, port int) {
	ctx,cancel := context.WithCancel(context.Background())
	ser.SetCancel(cancel)
	go ser.ListenAndServe(ctx, &net.UDPAddr{IP:net.ParseIP(ip), Port:port,})
}

func (ser *Server) ListenAndServe(ctx context.Context, addr *net.UDPAddr) error {
	switch ser.State() {
	case StateRunning:
		return errAlreadyRunning
	case StateClosed:
		return errServerClosed
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	return ser.Serve(ctx, conn)
}

// Serve serves a Raknet server
func (ser *Server) Serve(ctx context.Context, l *net.UDPConn) error {
	switch ser.State() {
	case StateRunning:
		return errAlreadyRunning
	case StateClosed:
		return errServerClosed
	}

	ser.conn = l

	ser.init()

	// Waits close command from context.Context
	go func() {
		<-ctx.Done()

		ser.state = StateClosed

		err := ser.conn.Close()
		if err != nil {
			ser.Logger.Warn(err)
		}

		for _,handler := range ser.Handlers {
			handler.CloseServer()
		}
	}()

	// Updates the sessions connected already
	// in another thread
	go func() {
		for {
			time.Sleep(1 * time.Nanosecond) // lower cpu usage

			select {
			case <-ctx.Done():
				break
			default:
			}

			err := ser.RangeSessions(func(key string, session *Session) bool {
				if !session.update() { // if it is closed, remove session from server
					ser.removeSession(session.Addr)
				}

				return true
			})

			if err != nil {
				ser.Logger.Warn(err)
				break
			}
		}
	}()

	for _,handler := range ser.Handlers {
		handler.StartServer()
	}

	// Reads packets from udp socket, and handles them
	// in main thread
	var buf = make([]byte, 2048)
	for {
		_, addr, err := l.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				//Shutting down listener
				ser.Logger.Info("Shutting down listenner")
				return nil
			default:
				return err
			}
		}

		ser.Logger.Debug("Connection:" + addr.String())

		if len(buf) <= 0 {
			continue
		}

		ser.handlePacket(ctx, addr, buf)
	}
}

func (ser *Server) handlePacket(ctx context.Context, addr *net.UDPAddr, b []byte) {
	if len(b) <= 0 {
		return
	}

	// check blocked address
	if ser.HasBlockedAddress(addr.IP) {
		return
	}

	ser.Logger.Debug("handle a packet from " + addr.String())

	// new packet

	pk, ok := ser.protocol.Packet(b[0])
	if !ok {
		ser.Logger.Warn("unknown packet id: 0x", pk.ID())
		return
	}

	pk.SetBytes(b)

	fmt.Printf("%#v", b)

	switch npk := pk.(type) {
	case *protocol.UnconnectedPing, *protocol.UnconnectedPingOpenConnections:
		if !ser.BroadcastingEnabled {
			return
		}

		ping := npk.(*protocol.UnconnectedPing)

		err := ping.Decode()
		if err != nil {
			ser.Logger.Warn(err)
			return
		}

		if pk.ID() != protocol.IDUnconnectedPing &&
			(ser.Count() >= ser.MaxConnections || ser.MaxConnections >= 0) {
			return
		}

		if !ping.Magic {
			return
		}

		pong := &protocol.UnconnectedPong{
			Timestamp:  ping.Timestamp,
			PongID:     ser.pongid,
			Identifier: ser.Identifier,
			Connection: ser.Identifier.ConnectionType(),
		}

		err = pong.Encode()
		if err != nil {
			ser.Logger.Warn(err)
			return
		}

		fmt.Printf("%#v", pong.Bytes())

		ser.SendPacket(addr, pong)

		return
	case *protocol.OpenConnectionRequestOne:
		err := npk.Decode()
		if err != nil {
			ser.Logger.Warn(err)
		}

		session, ok := ser.GetSession(addr)
		if ok {
			if session.State == StateConnected {
				ser.CloseSession(session.Addr, "Client re-instantiated connection")
			}
		}

		if !npk.Magic {
			return
		}

		epk := ser.validateNewConnection(addr)
		if epk != nil {
			err = epk.Encode()
			if err != nil {
				ser.Logger.Warn(err)
				return
			}

			ser.SendPacket(addr, epk)
			return
		}

		if npk.ProtocolVersion != raknet.NetworkProtocol {
			rpk := &protocol.IncompatibleProtocol{}

			rpk.NetworkProtocol = raknet.NetworkProtocol
			rpk.ServerGuid = ser.uid

			err = rpk.Encode()
			if err != nil {
				ser.Logger.Warn(err)
				return
			}

			ser.SendPacket(addr, rpk)
			return
		}

		if npk.MTU > raknet.MaxMTU {
			return
		}

		rpk := &protocol.OpenConnectionResponseOne{}
		rpk.ServerGuid = ser.uid
		rpk.MTU = uint16(npk.MTU)

		err = rpk.Encode()
		if err != nil {
			ser.Logger.Warn(err)
			return
		}

		ser.SendPacket(addr, rpk)

		return
	case *protocol.OpenConnectionRequestTwo:
		err := npk.Decode()
		if err != nil {
			return
		}

		epk := ser.validateNewConnection(addr)
		if epk != nil {
			err = epk.Encode()
			if err != nil {
				ser.Logger.Warn(err)
				return
			}

			ser.SendPacket(addr, epk)
			return
		}

		if ser.HasSessionGUID(npk.ClientGuid) {
			rpk := &protocol.AlreadyConnected{}

			err = rpk.Encode()
			if err != nil {
				ser.Logger.Warn(err)
				return
			}

			ser.SendPacket(addr, rpk)
			return
		}

		if npk.MTU > raknet.MaxMTU {
			return
		}

		rpk := &protocol.OpenConnectionResponseTwo{}
		rpk.ServerGuid = ser.uid
		rpk.ClientAddress = ser.newSystemAddress(addr)
		rpk.MTU = npk.MTU
		rpk.EncrtptionEnabled = false

		err = rpk.Encode()
		if err != nil {
			return
		}

		// handler: pre connect

		session := &Session{
			Addr:   addr,
			Conn:   ser.conn,
			GUID:   npk.ClientGuid,
			Logger: ser.Logger,
			MTU:    ser.MTU,
			State:  StateHandshaking,
			ctx:    ctx,
		}

		ser.storeSession(addr, session)

		ser.SendPacket(addr, rpk)

		return
	}

	session, ok := ser.GetSession(addr)
	if !ok {
		ser.Logger.Debug("Invalid connctions from " + addr.String())
		return
	}

	switch npk := pk.(type) {
	case *protocol.ACK:
		session.handleACKPacket(npk)
	case *protocol.CustomPacket:
		session.handleCustomPacket(npk)
	default:
		session.handlePacket(npk)
	}

}

func (ser *Server) newSystemAddress(addr *net.UDPAddr) *raknet.SystemAddress {
	return &raknet.SystemAddress{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}
}

// validateNewConnection returns error packets if the sender has problems
func (ser *Server) validateNewConnection(addr *net.UDPAddr) raknet.Packet {
	if ser.HasSession(addr) {
		return &protocol.AlreadyConnected{}
	} else if ser.Count() >= ser.MaxConnections && ser.MaxConnections >= 0 {
		return &protocol.NoFreeIncomingConnections{}
	} else if ser.HasBlockedAddress(addr.IP) {
		return &protocol.ConnectionBanned{}
	}

	return nil
}

func (ser *Server) Count() int {
	return ser.sessions.Count()
}

/*
func (ser *Server) Sessions() map[string]*Sessions {
	return ser.sessions.Items()
}*/

func (ser *Server) storeSession(addr net.Addr, session *Session) {
	ser.sessions.Set(addr.String(), session)
}

func (ser *Server) restoreSession(addr net.Addr) (*Session, bool) {
	value, ok := ser.sessions.Get(addr.String())
	if !ok {
		return nil, false
	}

	session, ok := value.(*Session)
	if !ok {
		ser.Logger.Warn("Invaild value, wants *Session")
		return nil, false
	}

	return session, true
}

func (ser *Server) existSession(addr net.Addr) bool {
	_, b := ser.restoreSession(addr)
	return b
}

func (ser *Server) removeSession(addr net.Addr) {
	ser.sessions.Remove(addr.String())
}

// RangeSessions processes for the sessions instead of "for range".
// if f returns false, stops the loop.
//
// key is string(returned net.Addr.String()), value is *Session
// returns errors if key and value is invalid.
func (ser *Server) RangeSessions(f func(key string, value *Session) bool) error {
	for item := range ser.sessions.IterBuffered() {
		session, ok := item.Val.(*Session)
		if !ok {
			return errors.New("invalid value, wants *Session")
		}

		if !f(item.Key, session) {
			break
		}
	}

	return nil
}

func (ser *Server) GetSessionGUID(guid int64) (*Session, bool) {
	var session *Session

	err := ser.RangeSessions(func(key string, sess *Session) bool {
		if sess.GUID == guid {
			session = sess
			return false
		}

		return true
	})

	if err != nil {
		ser.Logger.Warn(err)
		return nil, false
	}

	return session, session != nil
}

func (ser *Server) HasSession(addr net.Addr) bool {
	return ser.existSession(addr)
}

func (ser *Server) HasSessionGUID(guid int64) bool {
	_, b := ser.GetSessionGUID(guid)

	return b
}

func (ser *Server) GetSession(addr *net.UDPAddr) (*Session, bool) {
	session, ok := ser.restoreSession(addr)
	if !ok {
		return nil, false
	}

	if session.State == StateDisconected {
		ser.CloseSession(addr, "Already closed")

		session = nil
	}

	return session, true
}

func (ser *Server) closeSession(session *Session) {
	ser.removeSession(session.Addr)

	session.SendPacket(&protocol.DisconnectionNotification{}, raknet.Unreliable, 0)
}

func (ser *Server) CloseSession(addr *net.UDPAddr, reason string) error {
	session, ok := ser.restoreSession(addr)
	if !ok {
		return errors.New("couldn't find the session")
	}

	ser.closeSession(session)

	return nil
}

func (ser *Server) CloseSessionGUID(guid int64, reason string) error {
	session, ok := ser.GetSessionGUID(guid)
	if !ok {
		return errors.New("couldn't find the session")
	}

	ser.closeSession(session)

	return nil
}

func (ser *Server) SendPacket(addr *net.UDPAddr, pk raknet.Packet) {
	ser.SendRawPacket(addr, pk.Bytes())
}

func (ser *Server) SendRawPacket(addr *net.UDPAddr, b []byte) {
	go func() {  // TODO: rewrite
		ser.conn.WriteToUDP(b, addr)
	}()
}

func (ser *Server) HasBlockedAddress(ip net.IP) bool {
	return ser.blockedAddresses.Has(ip.String())
}

func (ser *Server) SetBlockedAddress(ip net.IP, exp Expire) {
	ser.blockedAddresses.Set(ip.String(), exp)
}

func (ser *Server) RemoveBlockedAddress(ip net.IP) {
	ser.blockedAddresses.Remove(ip.String())
}

func (ser *Server) packet(b []byte) (raknet.Packet, error) {
	if len(b) <= 0 {
		return nil, errors.New("no enough bytes")
	}

	pk, ok := ser.protocol.Packet(b[0])
	if !ok {
		return nil, errors.New("unknown packet")
	}

	return pk, nil
}

func equalUDPAddr(a *net.UDPAddr, b *net.UDPAddr) bool {
	if a.IP.Equal(b.IP) && a.Port == b.Port {
		return true
	}

	return false
}


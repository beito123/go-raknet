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
	"net"
	"time"

	"github.com/beito123/binary"

	raknet "github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/protocol"
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

type Server struct {
	Logger         raknet.Logger
	Handlers       []Handler
	MaxConnections int
	MTU            int
	Identifier     raknet.Identifier
	protocol       protocol.Protocol

	UUID uuid.UUID

	// BroadcastingEnabled broadcast the server for the outside
	// if it enabled, the server send UnconnectedPong when received UnconnectPing.
	BroadcastingEnabled bool

	conn   *net.UDPConn
	port   uint16
	state  ServerState
	pongid int64

	Sessions map[uuid.UUID]*Session
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

func (ser *Server) ListenAndServe(ctx context.Context, addr *net.UDPAddr) error {
	if ser.IsRunning() {
		return errAlreadyRunning
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	return ser.Serve(ctx, conn)
}

// Serve serves a Raknet server
func (ser *Server) Serve(ctx context.Context, l *net.UDPConn) error {
	if ser.IsRunning() {
		return errAlreadyRunning
	} else if ser.IsClosed() {
		return errServerClosed
	}

	ser.conn = l

	//init
	protocol := new(protocol.Protocol)
	protocol.RegisterPackets()

	ser.pongid = binary.ReadLong(ser.UUID.Bytes()[:8])

	// Waits done from context.Context
	go func() {
		<-ctx.Done()

		err := ser.Close()
		if err != nil {
			ser.Logger.Warn(err)
		}
	}()

	// Updates the sessions connected already
	// in another thread
	go func() {
		for ser.IsRunning() {
			for _, session := range ser.Sessions {
				err := session.update()
				if err != nil {
					ser.Logger.Warn(err)

					continue
				}
			}

			time.Sleep(1 * time.Nanosecond) // lower cpu usage
		}
	}()

	// Reads packets from udp socket, and handles them
	// in main thread
	var buf []byte
	for {
		_, addr, err := l.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				//Shutting down listener
				return nil
			default:
				return err
			}
		}

		ser.handlePacket(addr, buf)
	}

	return nil
}

func (ser *Server) handlePacket(addr *net.UDPAddr, b []byte) {
	// check blocked address

	// new packet
	pk, err := ser.Packet(b)
	if err != nil {
		ser.Logger.Warn(err)
		return
	}

	// is offline message packet
	if pk.ID() == protocol.IDUnconnectedPing || pk.ID() == protocol.IDUnconnectedPingOpenConnections {
		if ser.BroadcastingEnabled {
			return
		}

		pk, ok := pk.(*protocol.UnconnectedPing)
		if !ok {
			return
		}

		err := pk.Decode()
		if err != nil {
			return // bad packet, ignore
		}

		if pk.ID() == protocol.IDUnconnectedPing ||
			(len(ser.Sessions) < ser.MaxConnections || ser.MaxConnections < 0) {
			if !pk.Magic {
				return
			}

			pong := &protocol.UnconnectedPong{
				Timestamp:  pk.Timestamp,
				PongID:     ser.pongid,
				Identifier: ser.Identifier,
			}

			err := pong.Encode()
			if err != nil {
				ser.Logger.Warn(err)
				return
			}

			ser.SendPacket(addr, pong.Bytes())
		}
	}

	// create new session
	session := ser.GetSession(addr)
	if session == nil {
		session = ser.newSession(addr)
	}

	for _, hand := range ser.Handlers {
		hand.HandleRawPacket(session.UUID, pk)
	}

	session.handlePacket(pk)
}

func (ser *Server) GetSession(addr *net.UDPAddr) *Session {
	var session *Session
	for _, sess := range ser.Sessions {
		if equalUDPAddr(addr, sess.Addr) {
			session = sess
			break
		}
	}

	return session
}

func (ser *Server) newSession(addr *net.UDPAddr) *Session {
	uid, _ := uuid.NewV4()

	return &Session{
		Addr:   addr,
		Conn:   ser.conn,
		UUID:   uid,
		Logger: ser.Logger,
	}
}

func (ser *Server) CloseSessionAddr(addr *net.UDPAddr, reason string) error {
	session := ser.GetSession(addr)
	if session == nil {
		return errors.New("couldn't find a session")
	}

	return ser.CloseSession(session.UUID, reason)
}

func (ser *Server) CloseSession(uid uuid.UUID, reason string) error {
	session, ok := ser.Sessions[uid]
	if !ok {
		return errors.New("couldn't find a session")
	}

	delete(ser.Sessions, uid)

	session.SendPacket(&protocol.DisconnectionNotification{}, raknet.Unreliable, 0)

	return nil
}

func (ser *Server) SendPacket(addr *net.UDPAddr, b []byte) {
	go func() {
		ser.conn.WriteToUDP(b, addr)
	}()
}

func (ser *Server) Close() error {
	if !ser.IsRunning() {
		if ser.IsClosed() {
			return errServerClosed
		} else {
			return errors.New("no running the server")
		}
	}

	ser.state = StateClosed

	for _, session := range ser.Sessions {
		err := session.Close()
		if err != nil {
			return err
		}
	}

	err := ser.conn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (ser *Server) Packet(b []byte) (raknet.Packet, error) {
	if len(b) <= 0 {
		return nil, errors.New("no enough bytes")
	}

	pk := ser.protocol.Packet(b[0])
	if pk == nil {
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

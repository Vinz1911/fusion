//  Fusion
//
//  Created by Vinzenz Weist on 17.06.21.
//  Copyright © 2021 Vinzenz Weist. All rights reserved.
//

// Package network encapsulates the logic required to set up network connections
// and communication, including TCP and TLS encrypted connections, along with handling
// different types of messages such as text and binary.
// It uses the fusion network protocol to ensuring reliability and structure.
package network

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strconv"
)

// Predefined constants for identifying connection and message types.
const (
	TCPConnection uint8 = 0x0
	TLSConnection uint8 = 0x1

	TextMessage   uint8 = 0x1
	BinaryMessage uint8 = 0x2
)

// Private predefined constant for the maximum buffer size and pingMessage.
const (
	maximum 	uint32 = 0x8000
	pingMessage uint8 = 0x3
)

// Listener struct represents a TCP based connection listener that handles incoming
// pure TCP connections or TLS encrypted connections.
type Listener struct {
	frame    frame
	listener net.Listener

	Cert string
	Key  string

	Ready     func(conn net.Conn)
	Message   func(conn net.Conn, data []byte, opcode uint8)
	Failed    func(conn net.Conn, err error)
	Cancelled func(conn net.Conn)
}

// Start initiates the listener to start accepting incoming connections on the specified port.
// The parameter decides whether it's a TCP or TLS connection based on predefined constants.
func (listener *Listener) Start(parameter uint8, port uint16) (err error) {
	switch parameter {
	case TCPConnection:
		listener.listener, err = net.Listen("tcp6", ":" + strconv.Itoa(int(port)))
		if err != nil { return err }
	case TLSConnection:
		var cer tls.Certificate; cer, err = tls.LoadX509KeyPair(listener.Cert, listener.Key)
		if err != nil { return err }
		var config = &tls.Config{Certificates: []tls.Certificate{cer}}
		listener.listener, err = tls.Listen("tcp6", ":" + strconv.Itoa(int(port)), config)
		if err != nil { return err }
	}
	defer func() { if closed := listener.listener.Close(); closed != nil && err == nil { err = closed } }()
	for {
		var conn net.Conn; conn, err = listener.listener.Accept()
		if err != nil { return err }
		go listener.receiveMessage(conn)
	}
}

// Cancel stops the listener from accepting new connections and closes any existing ones.
func (listener *Listener) Cancel() {
	if listener.listener == nil { return }
	var err = listener.listener.Close()
	if err != nil { listener.Failed(nil, err) }
	listener.listener = nil
}

// SendMessage sends a message through the specified connection.
func (listener *Listener) SendMessage(conn net.Conn, messageType uint8, data []byte) {
	listener.processingSend(conn, data, messageType)
}

// processingSend is a helper function to create and send a message frame over a connection.
func (listener *Listener) processingSend(conn net.Conn, data []byte, opcode uint8) {
	if listener.listener == nil { return }
	var message, err = listener.frame.create(data, opcode)
	if err != nil { listener.Failed(conn, err); listener.remove(conn) }
	_, err = conn.Write(message)
	if err != nil { listener.Failed(conn, err) }
}

// processingParse is a helper function to parse a message frame from the connection data.
func (listener *Listener) processingParse(conn net.Conn, frame *frame, data []byte) error {
	if listener.listener == nil { return errors.New(parsingFailed) }
	var err = frame.parse(data, func(data []byte, opcode uint8) {
		listener.Message(conn, data, opcode)
		if opcode == pingMessage { listener.processingSend(conn, data, pingMessage) }
	})
	return err
}

// remove terminates a specific connection and triggers the Cancelled callback.
func (listener *Listener) remove(conn net.Conn) {
	var err = conn.Close()
	if err != nil { listener.Failed(conn, err) }
	listener.Cancelled(conn)
}

// receiveMessage handles all incoming data for a connection and tracks broken connections.
func (listener *Listener) receiveMessage(conn net.Conn) {
	var frame = frame{}
	listener.Ready(conn)
	var buffer = make([]byte, maximum)
	for {
		var size, err = conn.Read(buffer)
		if err != nil { if err == io.EOF { listener.Cancelled(conn) } else { listener.Failed(conn, err) }; break }
		err = listener.processingParse(conn, &frame, buffer[:size])
		if err != nil { listener.Failed(conn, err); listener.remove(conn); break }
	}
}
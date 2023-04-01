//  Fusion
//
//  Created by Vinzenz Weist on 17.06.21.
//  Copyright © 2021 Vinzenz Weist. All rights reserved.
//

package network

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strconv"
)

const (
	TCPConnection uint8 = 0x0
	TLSConnection uint8 = 0x1

	TextMessage   uint8 = 0x1
	BinaryMessage uint8 = 0x2
	PingMessage   uint8 = 0x3
)

// Listener is a tcp based connection listener
// this is for handling incoming pure tcp connections
type Listener struct {
	frame    frame
	listener net.Listener

	Cert string
	Key  string

	Ready     func(conn net.Conn)
	Message   func(conn net.Conn, text *string, binary []byte)
	Failed    func(conn net.Conn, err error)
	Cancelled func(conn net.Conn)
}

// Start the NetworkGO connection listener
// waits for incoming connections
func (listener *Listener) Start(parameter uint8, port uint16) error {
	switch parameter {
	case TCPConnection:
		var err error
		listener.listener, err = net.Listen("tcp", ":" + strconv.Itoa(int(port)))
		if err != nil { return err }
	case TLSConnection:
		var cer, err = tls.LoadX509KeyPair(listener.Cert, listener.Key)
		if err != nil { return err }
		var config = &tls.Config{Certificates: []tls.Certificate{cer}}
		listener.listener, err = tls.Listen("tcp", ":" + strconv.Itoa(int(port)), config)
		if err != nil { return err }
	}
	defer listener.listener.Close()
	for {
		var conn, err = listener.listener.Accept()
		if err != nil { return err }
		go listener.receiveMessage(conn)
	}
}

// Cancel closes all connections and stops
// the listener from accepting new connections
func (listener *Listener) Cancel() {
	if listener.listener == nil { return }
	var err = listener.listener.Close()
	if err != nil { listener.Failed(nil, err) }
	listener.listener = nil
}

// SendTextMessage is for sending a text based message
func (listener *Listener) SendTextMessage(conn net.Conn, str string) {
	listener.processingSend(conn, []byte(str), TextMessage)
}

// SendBinaryMessage is for sending a text based message
func (listener *Listener) SendBinaryMessage(conn net.Conn, data []byte) {
	listener.processingSend(conn, data, BinaryMessage)
}

/// MARK: - Private API

// create and send message frame
func (listener *Listener) processingSend(conn net.Conn, data []byte, opcode uint8) {
	if listener.listener == nil { return }
	var message, err = listener.frame.create(data, opcode)
	if err != nil { listener.Failed(conn, err); listener.remove(conn) }
	_, err = conn.Write(message)
	if err != nil { listener.Failed(conn, err) }
}

// parse a message frame
func (listener *Listener) processingParse(conn net.Conn, frame *frame, data []byte) error {
	if listener.listener == nil { return errors.New(parsingFailed) }
	var err = frame.parse(data, func(text *string, data []byte, ping []byte) {
		listener.Message(conn, text, data)
		if ping != nil { listener.sendPong(conn, ping) }
	})
	return err
}

// remove is for terminating a specific connection
func (listener *Listener) remove(conn net.Conn) {
	var err = conn.Close()
	if err != nil { listener.Failed(conn, err) }
	listener.Cancelled(conn)
}

// sendPong is for sending a pong based message
func (listener *Listener) sendPong(conn net.Conn, data []byte) {
	listener.processingSend(conn, data, PingMessage)
}

// receiveMessage is handling all incoming input
// keeps track broken connections
func (listener *Listener) receiveMessage(conn net.Conn) {
	var frame = frame{}
	listener.Ready(conn)
	var buffer = make([]byte, 0x2000)
	for {
		var size, err = conn.Read(buffer)
		if err != nil { if err == io.EOF { listener.Cancelled(conn) } else { listener.Failed(conn, err) }; break }
		err = listener.processingParse(conn, &frame, buffer[:size])
		if err != nil { listener.Failed(conn, err); listener.remove(conn); break }
	}
}
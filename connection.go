// Package network

// Copyright 2021 Vinzenz Weist. All rights reserved.
// Use of this source code is risked by yourself.
// license that can be found in the LICENSE file.

package network

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strconv"
)

const (
	TCPConnection 	uint8 = 0x0
	TLSConnection 	uint8 = 0x1

	TextMessage 	uint8 = 0x1
	BinaryMessage 	uint8 = 0x2
	PingMessage		uint8 = 0x3
)

// Connection is a a tcp based connection
// this is for handling pure tcp connections
type Connection struct {
	frame    	frame
	listener 	net.Listener

	Cert 		string
	Key			string

	Ready 		func(conn net.Conn)
	Message 	func(conn net.Conn, text *string, binary []byte)
	Failed    	func(conn net.Conn, err error)
	Cancelled 	func(conn net.Conn)
}

// Start the NetworkGO connection listener
// waits for incoming connections
func (connection *Connection) Start(parameter uint8, port uint16) error {
	switch parameter {
	case TCPConnection:
		var err error
		connection.listener, err = net.Listen("tcp", ":" + strconv.Itoa(int(port)))
		if err != nil { return  err }
	case TLSConnection:
		cer, err := tls.LoadX509KeyPair(connection.Cert, connection.Key)
		if err != nil { return err }
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		connection.listener, err = tls.Listen("tcp", ":" + strconv.Itoa(int(port)), config)
		if err != nil { return err }
	}
	defer connection.listener.Close()
	for {
		conn, err := connection.listener.Accept()
		if err != nil { return err }
		go connection.receiveMessage(conn)
	}
}

// Cancel closes all connections and stops
// the listener from accepting new connections
func (connection *Connection) Cancel() {
	if connection.listener == nil { return }
	err := connection.listener.Close()
	if err != nil { connection.Failed(nil, err) }
	connection.listener = nil
}

// SendTextMessage is for sending a text based message
func (connection *Connection) SendTextMessage(conn net.Conn, str string) {
	connection.processingSend(conn, []byte(str), TextMessage)
}

// SendBinaryMessage is for sending a text based message
func (connection *Connection) SendBinaryMessage(conn net.Conn, data []byte) {
	connection.processingSend(conn, data, BinaryMessage)
}

/// MARK: - Private API

// create and send message frame
func (connection *Connection) processingSend(conn net.Conn, data []byte, opcode uint8) {
	if connection.listener == nil { return }
	message, err := connection.frame.create(data, opcode)
	if err != nil { connection.Failed(conn, err); connection.remove(conn) }
	_, err = conn.Write(message)
	if err != nil { connection.Failed(conn, err) }
}

// parse a message frame
func (connection *Connection) processingParse(conn net.Conn, frame *frame, data []byte) error {
	if connection.listener == nil { return errors.New(parsingFailed) }
	err := frame.parse(data, func(text *string, data []byte, ping []byte) {
		connection.Message(conn, text, data)
		if ping != nil { connection.sendPong(conn, ping) }
	})
	return err
}

// remove is for terminating a specific connection
func (connection *Connection) remove(conn net.Conn) {
	err := conn.Close()
	if err != nil { connection.Failed(conn, err) }
	connection.Cancelled(conn)
}

// sendPong is for sending a pong based message
func (connection *Connection) sendPong(conn net.Conn, data []byte) {
	connection.processingSend(conn, data, PingMessage)
}

// receiveMessage is handling all incoming input
// keeps track broken connections
func (connection *Connection) receiveMessage(conn net.Conn) {
	if connection.listener == nil { return }
	frame := frame{}
	connection.Ready(conn)
	buffer := make([]byte, 0x2000)
	for {
		size, err := conn.Read(buffer)
		if err != nil { if err == io.EOF { connection.Cancelled(conn) } else { connection.Failed(conn, err) }; break }
		err = connection.processingParse(conn, &frame, buffer[:size])
		if err != nil { connection.Failed(conn, err); connection.remove(conn); break }
	}
}
// Package network

// Copyright 2021 Vinzenz Weist. All rights reserved.
// Use of this source code is risked by yourself.
// license that can be found in the LICENSE file.

package network

import (
	"encoding/binary"
	"errors"
	"math"
)

// frame is message creator & parser
// fast and reliable
type frame struct {
	buffer []byte
}

const(
	overheadByteCount int = 0x5
	frameByteCount    int = math.MaxUint32

	parsingFailed string = "message parsing failed"
	readBufferOverflow string = "read buffer overflow"
	writeBufferOverflow string = "write buffer overflow"
)

// create is for creating compliant message frames
// returns message as data frame
func (*frame) create(data []byte, opcode uint8) (message []byte, err error) {
	if len(data) > frameByteCount - overheadByteCount { return nil, errors.New(writeBufferOverflow) }
	var frame []byte; var length = make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data) + overheadByteCount))
	frame = append(frame, opcode)
	frame = append(frame, length...)
	frame = append(frame, data...)
	return frame, nil
}

// parse is for parsing data back into compliant messages
// returns parsed data as 'Data' or 'String'
func (frame *frame) parse(data []byte, completion func(text *string, data []byte, ping []byte)) error {
	frame.buffer = append(frame.buffer, data...)
	length := frame.extractSize(); if length == nil { return nil }
	if len(frame.buffer) > frameByteCount { return errors.New(readBufferOverflow) }
	if len(frame.buffer) < overheadByteCount { return nil }
	if len(frame.buffer) < *length { return nil }
	for len(frame.buffer) >= *length && *length != 0 {
		bytes, err := frame.extractMessage(frame.buffer)
		if err != nil { return err }
		switch frame.buffer[0] {
		case TextMessage: result := string(bytes); completion(&result, nil, nil)
		case BinaryMessage: completion(nil, bytes, nil)
		case PingMessage: completion(nil, nil, bytes)
		default: return errors.New(parsingFailed)
		}
		if len(frame.buffer) <= *length { frame.buffer = []byte{} } else { frame.buffer = frame.buffer[*length:] }
	}
	return nil
}

// extract the message frame size from the data
// if not possible it returns zero
func (frame *frame) extractSize() *int {
	if len(frame.buffer) < overheadByteCount { return nil }
	size := frame.buffer[1:overheadByteCount]
	length := int(binary.BigEndian.Uint32(size))
	return &length
}

// extract the message and remove the overhead
// if not possible it returns nil
func (frame *frame) extractMessage(data []byte) (message []byte, err error) {
	if len(data) < overheadByteCount { return nil, errors.New(parsingFailed) }
	length := frame.extractSize(); if length == nil { return nil, errors.New(parsingFailed) }
	if *length < overheadByteCount { return nil, errors.New(parsingFailed) }
	return data[overheadByteCount:*length], nil
}
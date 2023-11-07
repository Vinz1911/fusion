//  Fusion
//
//  Created by Vinzenz Weist on 17.06.21.
//  Copyright © 2021 Vinzenz Weist. All rights reserved.
//

package network

import (
	"encoding/binary"
	"errors"
)

// frame is message creator & parser
// fast and reliable
type frame struct {
	buffer []byte
}

const (
	opcodeByteCount   uint32 = 0x1
	controlByteCount  uint32 = 0x5
	frameByteCount    uint32 = 0xFFFFFFFF
)

const (
	parsingFailed       	string = "message parsing failed"
	readBufferOverflow  	string = "read buffer overflow"
	writeBufferOverflow 	string = "write buffer overflow"
	sizeExtractionFailed	string = "size extraction failed"
)

// create is for creating compliant message frames
// returns message as data frame
func (*frame) create(data []byte, opcode uint8) (message []byte, err error) {
	if uint32(len(data)) > frameByteCount - controlByteCount { return nil, errors.New(writeBufferOverflow) }
	var frame []byte
	var length = make([]byte, 0x4); binary.BigEndian.PutUint32(length, uint32(len(data)) + controlByteCount)
	frame = append(frame, opcode)
	frame = append(frame, length...)
	frame = append(frame, data...)
	return frame, nil
}

// parse is for parsing data back into compliant messages
// returns parsed data as '[]Byte' or 'String'
func (frame *frame) parse(data []byte, completion func(data []byte, opcode uint8)) error {
	frame.buffer = append(frame.buffer, data...)
	var length, err = frame.extractSize(); if err != nil { return nil }
	if len(frame.buffer) > int(frameByteCount) { return errors.New(readBufferOverflow) }
	if len(frame.buffer) < int(controlByteCount) { return nil }; if len(frame.buffer) < length { return nil }
	for len(frame.buffer) >= length && length != 0 {
		var bytes, err = frame.extractMessage(length); if err != nil { return err }
		switch frame.buffer[0] {
		case TextMessage: completion(bytes, TextMessage)
		case BinaryMessage: completion(bytes, BinaryMessage)
		case pingMessage: completion(bytes, pingMessage)
		default: return errors.New(parsingFailed) }
		if len(frame.buffer) <= length { frame.buffer = []byte{} } else { frame.buffer = frame.buffer[length:] }
	}; return nil
}

// extract the message frame size from the data
// if not possible it returns zero
func (frame *frame) extractSize() (length int, err error) {
	if len(frame.buffer) < int(controlByteCount) { return 0x0, errors.New(sizeExtractionFailed) }
	var size = frame.buffer[opcodeByteCount:controlByteCount]
	return int(binary.BigEndian.Uint32(size)), nil
}

// extract the message and remove the overhead
// if not possible it returns nil
func (frame *frame) extractMessage(length int) (message []byte, err error) {
	if len(frame.buffer) < int(controlByteCount) { return nil, errors.New(parsingFailed) }
	if length < int(controlByteCount) { return nil, errors.New(parsingFailed) }
	return frame.buffer[controlByteCount:length], nil
}
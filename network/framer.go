//  Fusion
//
//  Created by Vinzenz Weist on 17.06.21.
//  Copyright © 2021 Vinzenz Weist. All rights reserved.
//

// Package network encapsulates functionalities related to network communication.
// It provides a structure and associated methods to create and parse data frames
// for messaging across the fusion network protocol ensuring reliability and structure.
package network

import (
	"encoding/binary"
	"errors"
)

// frame is a type that encapsulates operations related to the creation and parsing
// of message frames over the network. It aims to ensure fast and reliable communication
// by adhering to a specified frame structure.
type framer struct {
	buffer []byte
}

// Predefined constants to maintain consistency in frame size and errors.
const (
	opcodeConstant   uint32 = 0x1
	controlConstant  uint32 = 0x5
	frameConstant    uint32 = 0xFFFFFFFF
)

// Predefined error messages for common frame parsing issues.
const (
	parsingFailed       	string = "message parsing failed"
	readBufferOverflow  	string = "read buffer overflow"
	writeBufferOverflow 	string = "write buffer overflow"
	sizeExtractionFailed	string = "size extraction failed"
)

// create is a method to construct a compliant frame for sending a message.
// It takes in the message data and an opcode indicating the type of message.
// It returns a slice of bytes representing the framed message or an error if the frame could not be created.
func (*framer) create(data []byte, opcode uint8) (message []byte, err error) {
	if uint32(len(data)) > frameConstant - controlConstant { return nil, errors.New(writeBufferOverflow) }
	var frame []byte
	var length = make([]byte, 0x4); binary.BigEndian.PutUint32(length, uint32(len(data)) + controlConstant)
	frame = append(frame, opcode)
	frame = append(frame, length...)
	frame = append(frame, data...)
	return frame, nil
}

// parse is a method for converting a sequence of bytes received over the network back into a structured message.
// It appends the incoming data to the frame's buffer and attempts to extract complete messages based on the frame structure.
// The completion callback is called with the parsed data and opcode when a message is successfully parsed.
func (frame *framer) parse(data []byte, completion func(data []byte, opcode uint8)) error {
	frame.buffer = append(frame.buffer, data...)
	var length, err = frame.extractSize(); if err != nil { return nil }
	if len(frame.buffer) > int(frameConstant) { return errors.New(readBufferOverflow) }
	if len(frame.buffer) < int(controlConstant) { return nil }; if len(frame.buffer) < length { return nil }
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

// extractSize attempts to determine the size of the next message in the buffer.
// It returns the size as an integer and an error if the size cannot be determined.
func (frame *framer) extractSize() (length int, err error) {
	if len(frame.buffer) < int(controlConstant) { return 0x0, errors.New(sizeExtractionFailed) }
	var size = frame.buffer[opcodeConstant:controlConstant]
	return int(binary.BigEndian.Uint32(size)), nil
}

// extractMessage attempts to extract the message from the frame buffer based on the input length.
// It returns the extracted message as byte array and an error if the message cannot be extracted.
func (frame *framer) extractMessage(length int) (message []byte, err error) {
	if len(frame.buffer) < int(controlConstant) { return nil, errors.New(parsingFailed) }
	if length < int(controlConstant) { return nil, errors.New(parsingFailed) }
	return frame.buffer[controlConstant:length], nil
}
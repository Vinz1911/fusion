//  Fusion
//
//  Created by Vinzenz Weist on 17.06.21.
//  Copyright © 2021 Vinzenz Weist. All rights reserved.
//

package network

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
	overheadByteCount uint32 = 0x25
	frameByteCount    uint32 = 0xFFFFFFFF
)

const (
	hashMismatch        string = "message hash does not match"
	parsingFailed       string = "message parsing failed"
	readBufferOverflow  string = "read buffer overflow"
	writeBufferOverflow string = "write buffer overflow"
)

// create is for creating compliant message frames
// returns message as data frame
func (*frame) create(data []byte, opcode uint8) (message []byte, err error) {
	if uint32(len(data)) > frameByteCount - overheadByteCount { return nil, errors.New(writeBufferOverflow) }
	var frame []byte
	var length = make([]byte, 0x4); binary.BigEndian.PutUint32(length, uint32(len(data)) + overheadByteCount)
	frame = append(frame, opcode)
	frame = append(frame, length...); var hash = sha256.Sum256(frame[:controlByteCount])
	frame = append(frame, hash[:]...)
	frame = append(frame, data...)
	return frame, nil
}

// parse is for parsing data back into compliant messages
// returns parsed data as 'Data' or 'String'
func (frame *frame) parse(data []byte, completion func(text *string, data []byte, ping []byte)) error {
	frame.buffer = append(frame.buffer, data...)
	var length = frame.extractSize()
	if length == nil { return nil }
	if uint32(len(frame.buffer)) > frameByteCount { return errors.New(readBufferOverflow) }
	if uint32(len(frame.buffer)) < overheadByteCount { return nil }
	if len(frame.buffer) < *length { return nil }
	for len(frame.buffer) >= *length && *length != 0 {
		var digest = sha256.Sum256(frame.buffer[:controlByteCount])
		if hex.EncodeToString(digest[:]) != *frame.extractHash() { return errors.New(hashMismatch) }
		var bytes, err = frame.extractMessage(frame.buffer)
		if err != nil { return err }
		switch frame.buffer[0] {
		case TextMessage: var result = string(bytes); completion(&result, nil, nil)
		case BinaryMessage: completion(nil, bytes, nil)
		case PingMessage: completion(nil, nil, bytes)
		default: return errors.New(parsingFailed) }
		if len(frame.buffer) <= *length { frame.buffer = []byte{} } else { frame.buffer = frame.buffer[*length:] }
	}
	return nil
}

// MARK: - Private API -

// extract the message hash from the data
// if not possible it returns nil
func (frame *frame) extractHash() *string {
	if uint32(len(frame.buffer)) < overheadByteCount { return nil }
	var hash = frame.buffer[controlByteCount:overheadByteCount]
	var digest = hex.EncodeToString(hash[:])
	return &digest
}

// extract the message frame size from the data
// if not possible it returns zero
func (frame *frame) extractSize() *int {
	if uint32(len(frame.buffer)) < overheadByteCount { return nil }
	var size = frame.buffer[opcodeByteCount:controlByteCount]
	var length = int(binary.BigEndian.Uint32(size))
	return &length
}

// extract the message and remove the overhead
// if not possible it returns nil
func (frame *frame) extractMessage(data []byte) (message []byte, err error) {
	if uint32(len(data)) < overheadByteCount { return nil, errors.New(parsingFailed) }
	var length = frame.extractSize()
	if length == nil { return nil, errors.New(parsingFailed) }
	if uint32(*length) < overheadByteCount { return nil, errors.New(parsingFailed) }
	return data[overheadByteCount:*length], nil
}

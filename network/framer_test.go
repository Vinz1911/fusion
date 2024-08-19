package network

import (
	"bytes"
	"testing"
)

// TestFramerText is for testing the frame creator and parser.
// Includes testing for multiple message extractions from single byte buffer.
func TestFramerText(test *testing.T) {
	framer := framer{}
	message := "Hello World!"
	var bytesBlock []byte
	data, err := framer.create([]byte(message), TextMessage)
	if err != nil { test.Errorf("parsing failed") }
	for i := 0; i < 10000; i++ { bytesBlock = append(bytesBlock, data...) }
	err = framer.parse(bytesBlock, func(data []byte, opcode uint8) {
		if opcode == TextMessage { if string(data) != message { test.Errorf("parsing failed") } }
	})
	if err != nil { test.Errorf("parsing failed") }
}

// TestFramerBinary is for testing the frame creator and parser.
// Includes testing for multiple message extractions from single byte buffer.
func TestFramerBinary(test *testing.T) {
	framer := framer{}
	message := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	var bytesBlock []byte
	data, err := framer.create(message, BinaryMessage)
	if err != nil { test.Errorf("parsing failed") }
	for i := 0; i < 10000; i++ { bytesBlock = append(bytesBlock, data...) }
	err = framer.parse(bytesBlock, func(data []byte, opcode uint8) {
		if opcode != BinaryMessage { if !bytes.Equal(data, message) { test.Errorf("parsing failed") } }
	})
	if err != nil { test.Errorf("parsing failed") }
}
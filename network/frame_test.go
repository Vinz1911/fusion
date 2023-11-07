package network

import (
	"bytes"
	"testing"
)

// TestFramerText is for testing the frame creator and parser
// includes testing for multiple message extraction from single byte buffer
func TestFramerText(test *testing.T) {
	frame := frame{}
	message := "Hello World!"
	var bytesBlock []byte
	data, err := frame.create([]byte(message), TextMessage)
	if err != nil { test.Errorf("parsing failed") }
	for i := 0; i < 1000; i++ { bytesBlock = append(bytesBlock, data...) }
	err = frame.parse(bytesBlock, func(text *string, data []byte, ping []byte) {
		if text != nil { if *text != message { test.Errorf("parsing failed") } }
	})
	if err != nil { test.Errorf("parsing failed") }
}

// TestFramerBinary is for testing the frame creator and parser
// includes testing for multiple message extraction from single byte buffer
func TestFramerBinary(test *testing.T) {
	frame := frame{}
	message := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	var bytesBlock []byte
	data, err := frame.create(message, BinaryMessage)
	if err != nil { test.Errorf("parsing failed") }
	for i := 0; i < 1000; i++ { bytesBlock = append(bytesBlock, data...) }
	err = frame.parse(bytesBlock, func(text *string, binary []byte, ping []byte) {
		if binary != nil { if !bytes.Equal(binary, message) { test.Errorf("parsing failed") } }
	})
	if err != nil { test.Errorf("parsing failed") }
}
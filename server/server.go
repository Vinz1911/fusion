package main

import (
	"github.com/fatih/color"
	"math"
	"net"
	"strconv"
)

func main() {
	printError := color.New(color.FgHiRed).PrintlnFunc()
	printLog := color.New(color.FgHiGreen).PrintlnFunc()
	printWarn := color.New(color.FgHiYellow).PrintlnFunc()
	printInfo := color.New(color.FgHiCyan).PrintlnFunc()

	listener := Listener{}
	port := uint16(7878)

	listener.Ready = func(socket net.Conn) {
		printInfo("[INFO]: new connection from address:", socket.RemoteAddr())
	}

	listener.Message = func(conn net.Conn, text *string, binary []byte) {
		if text != nil {
			size := math.Min(float64(atoi(*text)), 0xFFFFFF)
			size = math.Max(size, 0)
			message := make([]byte, int(size))
			listener.SendBinaryMessage(conn, message)
		}
		if binary != nil {
			size := len(binary)
			message := itoa(size)
			listener.SendTextMessage(conn, message)
		}
	}

	listener.Failed = func(conn net.Conn, err error) {
		printError("[ERROR]: ", err)
	}

	listener.Cancelled = func(conn net.Conn) {
		printWarn("[WARN]: connection closed from address:", conn.RemoteAddr())
	}

	printLog("[SERVER]: Network-GO Listener")
	printLog("[PORT]:", port)
	printLog("[VERSION]: v1.0.0")
	err := listener.Start(TCPConnection, port)
	if err != nil {
		printError("[ERROR]:", err)
	}
}

// convert a string to an integer
func atoi(str string) int {
	value, err := strconv.Atoi(str)
	if err != nil {
		value = 0
	}
	return value
}

// convert an integer to a string
func itoa(integer int) string {
	value := strconv.Itoa(integer)
	return value
}

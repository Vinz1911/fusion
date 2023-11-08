# Fusion

`Fusion` is a library which implements the `Fusion Framing Protocol (FFP)`. 
The `Fusion Framing Protocol (FFP)` is proprietary networking protocol which uses a small and lightweight header with a performance as fast as raw tcp performance. 
Built directly on top of Go's `net.Listener` with support for plain tcp and tls encrypted connections. 
The implementation for the Client is [FusionKit](https://github.com/Vinz1911/fusionkit) written in swift on top of `Network.framework` to ensure maximum performance.

## License:
[![License](https://img.shields.io/badge/license-GPLv3-blue.svg?longCache=true&style=flat)](https://github.com/Vinz1911/network-go/blob/main/LICENSE)

## Golang Version:
[![Golang 1.16](https://img.shields.io/badge/Golang-1.16-00ADD8.svg?logo=go&style=flat)](https://golang.org) [![Golang 1.16](https://img.shields.io/badge/Modules-Support-00ADD8.svg?logo=go&style=flat)](https://blog.golang.org/using-go-modules)

## Installation:
### Go Modules
[Go Modules](https://blog.golang.org/using-go-modules). Just add this repo to your go.sum file.

## Import:
```go
// import the framework
import "github.com/Vinz1911/fusion/network"

// create a new connection listener
listener := network.Listener{}

// ...
```

## Callback:
```go
// create a new connection listener
listener := network.Listener{}

// listener is ready
listener.Ready = func(socket net.Conn) { }

// listener received message from a connection
listener.Message = func(conn net.Conn, binary []byte, opcode uint8) {
	// message is text based
    if opcode == network.TextMessage { println(string(binary)) }
    // message is binary based
    if binary == network.BinaryMessage { println(len(binary)) }
}

// listener connection failed
listener.Failed = func(conn net.Conn, err error) { }

// listener connection cancelled
listener.Cancelled = func(conn net.Conn) { }

// start listener
err := listener.Start(network.TCPConnection, uint16(7878))
```

## Author:
👨🏼‍💻 [Vinzenz Weist](https://github.com/Vinz1911)
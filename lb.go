package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
)

const (
	// Default ConnPort front end
	defaultConnPort = "8080"
	ConnType        = "tcp"
)

// Ref: https://lawlessguy.wordpress.com/2013/07/23/filling-a-slice-using-command-line-flags-in-go-golang/
type intslice []int

// Now, for our new type, implement the two methods of the flag. Value interface...
// The first method is String() string
func (i *intslice) String() string {
	return fmt.Sprintf("%d", *i)
}

// The second method is Set(value string) error
func (i *intslice) Set(value string) error {
	//fmt.Printf("%s\n", value)
	tmp, err := strconv.Atoi(value)
	if err == nil {
		*i = append(*i, tmp)
	}
	return err
}

var backendPorts intslice
var currBackendPortIndex int

func getBackendPort() string {
	i := currBackendPortIndex
	currBackendPortIndex += 1
	if currBackendPortIndex >= len(backendPorts) {
		currBackendPortIndex = 0
	}
	return fmt.Sprintf("%d", backendPorts[i])
}

func main() {
	flag.Var(&backendPorts, "p", "Ports to be used in backend")
	frontendPortPtr := flag.String("f", defaultConnPort, "Port to be used as frontend")
	flag.Parse()
	if len(backendPorts) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Listen for incoming connections
	l, err := net.Listen(ConnType, "0.0.0.0:"+*frontendPortPtr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(2)
	}
	// Close the listener when the application closes
	defer l.Close()
	fmt.Println("Listening on port", *frontendPortPtr)
	for {
		// Listen for an incoming connection
		frontConn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			continue
		}
		// Handle connection in a new goroutine
		go handleRequest(frontConn)
	}
}

// Handle incoming request
func handleRequest(frontConn net.Conn) {
	var wg sync.WaitGroup

	defer frontConn.Close()

	backConn, err := net.Dial(ConnType, "localhost:"+getBackendPort())
	if err != nil {
		fmt.Println("Error connecting to backend:", err.Error())
		return
	}
	defer backConn.Close()

	fmt.Println("Handling", frontConn.RemoteAddr(), "to", backConn.RemoteAddr())

	// Start an copy buffer goroutine from input to output connection.
	// Copy buffer until either connection is closed, then calls wg.Done.
	connectionTransfer := func(connFrom net.Conn, connTo net.Conn) {
		var err error = nil
		var readDone = false
		var readLen int
		var writeOffset int
		var writeLen int

		// Make a buffer to hold incoming data
		buf := make([]byte, 1024)

		for !readDone && err == nil {
			// Read from connFrom, up to buf size
			readLen, err = connFrom.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading from", connFrom.RemoteAddr(), err.Error())
					break
				}
				// Graceful closing of socket. Still may need to write
				readDone = true
			}

			// Write to connTo
			writeOffset = 0
			for writeOffset < readLen {
				writeLen, err = connTo.Write(buf[writeOffset:readLen])
				if err != nil {
					fmt.Println("Error writing to", connTo.RemoteAddr(), err.Error())
					break
				}
				writeOffset += writeLen
			}
			//fmt.Println("Wrote", writeOffset, "to", connTo.RemoteAddr())
		}
		wg.Done()
	}

	wg.Add(2)
	go connectionTransfer(frontConn, backConn)
	go connectionTransfer(backConn, frontConn)

	// Defer will close the connections
	wg.Wait()

	fmt.Println("Finished handling", frontConn.RemoteAddr(), "to", backConn.RemoteAddr())
}

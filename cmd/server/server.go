package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type TransparentServerReader struct {
	conn net.Conn
}

func (tsr *TransparentServerReader) Read(p []byte) (n int, err error) {
	fmt.Printf("Being asked to read %d bytes.\n", len(p))
	time.Sleep(5 * time.Second)
	return tsr.conn.Read(p)
}

func main() {
	listener, err := net.Listen("tcp4", "0.0.0.0:5001")

	if err != nil {
		fmt.Printf("Error: Could not create the server socket: %v\n", err)
		os.Exit(-1)
	}

	for {
		accepted, err := listener.Accept()
		defer accepted.Close()
		if err != nil {
			fmt.Printf("Error accepting client connection: %v\nGoing around again ...\n", err)
			continue
		}
		size, err := io.ReadAll(&TransparentServerReader{accepted})
		if err != nil {
			fmt.Printf("There was an error reading: %v\n", err)
			continue
		}
		fmt.Printf("Finished reading %d bytes from the sender.\n", size)
	}
}

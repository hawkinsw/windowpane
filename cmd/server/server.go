package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func GetTCPInfo(basicConn net.Conn) (*unix.TCPInfo, error) {
	tcpConn, ok := basicConn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("Could convert to tcp connection!")
	}
	var info *unix.TCPInfo = nil
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return nil, err
	}
	rawConn.Control(func(fd uintptr) {
		info, err = unix.GetsockoptTCPInfo(int(fd), unix.SOL_TCP, unix.TCP_INFO)
	})
	return info, err
}

type TransparentServerReader struct {
	conn net.Conn
}

func (tsr *TransparentServerReader) Read(p []byte) (n int, err error) {
	fmt.Printf("Being asked to read %d bytes.\n", len(p))

	if info, err := GetTCPInfo(tsr.conn); err == nil {
		fmt.Printf("Rcv_space: %d\n", info.Snd_cwnd)
	} else {
		fmt.Printf("Error: %v\n", err)
	}

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

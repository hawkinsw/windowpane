package main

import (
	"fmt"
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

type TransparentClientSender struct {
	conn         net.Conn
	totalWritten uint64
}

func (tsc *TransparentClientSender) Write(p []byte) (n int, err error) {
	fmt.Printf("Being asked to write %d bytes.\n", len(p))
	doneChan := make(chan interface{})
	var written int = 0
	go func() {
		written, err = tsc.conn.Write(p[:10])
		doneChan <- struct{}{}
	}()
allDone:
	for {
		select {
		case <-doneChan:
			tsc.totalWritten += uint64(written)
			fmt.Printf("The send is done (%d).\n", tsc.totalWritten)
			break allDone
		case <-time.Tick(1 * time.Second):
			fmt.Printf("The send is (still) blocked (%d).\n", tsc.totalWritten)
		}
	}
	return written, err
}

func main() {
	conn, err := net.Dial("tcp4", "127.0.0.1:5001")
	defer conn.Close()

	if err != nil {
		fmt.Printf("Error: Could not create the server socket: %v\n", err)
		os.Exit(-1)
	}

	sendSize := 5000000
	data := make([]byte, sendSize)

	success := true
	sender := &TransparentClientSender{conn, 0}
	written := 0
	for written < sendSize {
		justWritten, err := sender.Write(data[written:])
		if err != nil {
			success = false
			break
		}
		written += justWritten
	}
	if !success {
		fmt.Printf("There was an error sending!\n")
	} else {
		fmt.Printf("Successful sending!\n")
	}
}

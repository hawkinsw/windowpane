package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hawkinsw/windowpane/window"
)

type TransparentClientSender struct {
	conn         net.Conn
	totalWritten uint64
}

func (tsc *TransparentClientSender) Write(p []byte) (written int, err error) {
	if window.GetSenderWindowAvailable() {
		if sndWnd, err := window.GetSenderWindow(tsc.conn); err == nil {
			fmt.Printf("Receiver window: %d\n", sndWnd)
		} else {
			fmt.Printf("Receiver window: -1\n")
		}
	}

	doneChan := make(chan interface{})
	go func() {
		err = nil
		written = 0
		// We will attempt to write until we either
		// a) send data or
		// b) get an error.
		// Either one is satisfactory, in some broad sense of the term!
		for written == 0 && err == nil {
			written, err = tsc.conn.Write(p[:10])
		}
		doneChan <- struct{}{}
	}()
allDone:
	for {
		select {
		case <-doneChan:
			tsc.totalWritten += uint64(written)
			break allDone
		case <-time.Tick(1 * time.Second):
			if window.GetSenderWindowAvailable() {
				if sndWnd, err := window.GetSenderWindow(tsc.conn); err == nil {
					fmt.Printf("Receiver window: %d\n", sndWnd)
				} else {
					fmt.Printf("Receiver window: -1\n")
				}
			}
		}
	}
	return
}

var (
	serverIp   = flag.String("ip", "127.0.0.1", "Server IP address")
	serverPort = flag.Uint("port", 5001, "Server port")
)

func main() {
	flag.Parse()
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", *serverIp, *serverPort))

	if err != nil {
		fmt.Printf("Error: Could not create the server socket (to %s:%d): %v\n", *serverIp, *serverPort, err)
		os.Exit(-1)
	}
	defer conn.Close()

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
	}
}

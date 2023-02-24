//go:build darwin

package window

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

func GetSenderWindowAvailable() bool {
	return true
}

func GetSenderWindow(basicConn net.Conn) (uint32, error) {
	if info, err := getTCPInfo(basicConn); err == nil {
		return info.Snd_wnd, nil
	} else {
		return 0, err
	}
}

/*
 * Blatant self plagiarism from goresponsiveness.
 */
func getTCPInfo(basicConn net.Conn) (*unix.TCPConnectionInfo, error) {
	tcpConn, ok := basicConn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf(
			"OOPS: Could not get the TCP info for the connection (not a TCP connection)",
		)
	}
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return nil, err
	}
	var info *unix.TCPConnectionInfo = nil
	rawConn.Control(func(fd uintptr) {
		info, err = unix.GetsockoptTCPConnectionInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_CONNECTION_INFO)
	})
	return info, err
}
